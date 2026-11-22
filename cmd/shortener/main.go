package main

import (
	"context"
	"errors"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/KiberIGOR/tinyurl/internal/audit"
	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/compress"
	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/facade"
	"github.com/KiberIGOR/tinyurl/internal/grpc"
	grpcpb "github.com/KiberIGOR/tinyurl/internal/grpc/pb"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/ip"
	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/KiberIGOR/tinyurl/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	// _ "net/http/pprof" // подключаем пакет pprof
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func buildLabel(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// stub
type nilPinger struct{}

func (nilPinger) Ping(ctx context.Context) error {
	return errors.New("database is not configured")
}

func main() {
	fmt.Println("Build version:", buildLabel(buildVersion))
	fmt.Println("Build date:", buildLabel(buildDate))
	fmt.Println("Build commit:", buildLabel(buildCommit))

	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err)
	}
	var store service.URLRepository
	var userRepo service.UserRepository
	var pinger handler.Pinger
	var fileStore *repository.FileMemory
	switch {
	case cfg.DataBaseDSN != "":
		initCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err = runMigrations(cfg.DataBaseDSN); err != nil {
			log.Fatal("migrate: ", err)
		}
		var pool *pgxpool.Pool
		pool, err = newPool(initCtx, cfg.DataBaseDSN)
		if err != nil {
			log.Fatal("pool: ", err)
		}
		defer pool.Close()
		pg := repository.NewDB(pool)
		store = pg
		userRepo = pg
		pinger = pg
	case cfg.FileStoragePath != "":
		fileStore, err = repository.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		store = fileStore
		userRepo = fileStore
		pinger = nilPinger{}
	default:
		memory := repository.NewMemory()
		store = memory
		userRepo = memory
		pinger = nilPinger{}
	}
	svc := service.NewShortener(store, cfg.BaseURL)

	auditBroker := audit.NewBroker()
	var auditFile *audit.FileObserver
	if cfg.AuditFile != "" {
		auditFile, err = audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			log.Fatal("audit file: ", err)
		}
		auditBroker.Subscribe(auditFile)
	}
	if cfg.AuditURL != "" {
		auditBroker.Subscribe(audit.NewHTTPObserver(cfg.AuditURL, nil))
	}

	api := facade.NewShortener(svc, auditBroker)
	h := handler.New(api, svc, pinger, auditBroker)

	auth.Initialize(service.NewAuthService(userRepo, cfg.SecretKey))
	if err = ip.Initialize(cfg.TrustedSubnet); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(auth.SetCookieMiddleware)
		r.Post("/", h.PostURLHandler)
		r.Post("/api/shorten", h.PostJSONURLHandler)
		r.Post("/api/shorten/batch", h.PostBatchHandler)
		r.Get("/ping", h.GetPingHandler)
		r.Get("/{id}", h.GetURL)
	})
	r.With(auth.AuthMiddleware).Get("/api/user/urls", h.GetUsersURLs)
	r.With(auth.AuthMiddleware).Delete("/api/user/urls", h.DeleteUsersURLs)
	r.With(ip.IPMiddleware).Get("/api/internal/stats", h.GetStatus)

	// go func() {
	// 	log.Println("pprof listening on :6060")
	// 	if err := http.ListenAndServe(":6060", nil); err != nil {
	// 		log.Print("pprof: ", err)
	// 	}
	// }()

	srv, tlsCfg, err := newHTTPServer(cfg, r)
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv, grpcLis, err := newGRPCServer(cfg, tlsCfg, api)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if serveErr := grpcSrv.Serve(grpcLis); serveErr != nil {
			logger.Log.Error("grpc serve", zap.Error(serveErr))
		}
	}()

	go func() {
		var listenErr error
		if cfg.EnableHTTPS {
			listenErr = srv.ListenAndServeTLS("", "")
		} else {
			listenErr = srv.ListenAndServe()
		}
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			log.Fatal(listenErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("server shutdown", zap.Error(err))
	}
	grpcSrv.GracefulStop()

	h.Wait()

	if auditFile != nil {
		if err = auditFile.Close(); err != nil {
			logger.Log.Error("audit file close", zap.Error(err))
		}
	}
	if fileStore != nil {
		if err = fileStore.Close(); err != nil {
			logger.Log.Error("file storage close", zap.Error(err))
		}
	}
}

func newHTTPServer(cfg *config.Config, handler http.Handler) (*http.Server, *tls.Config, error) {
	if !cfg.EnableHTTPS {
		return &http.Server{Addr: cfg.Address, Handler: handler}, nil, nil
	}

	host, err := tlsHostFromBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("https: %w", err)
	}

	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
	}
	tlsCfg := manager.TLSConfig()
	return &http.Server{
		Addr:      cfg.Address,
		Handler:   handler,
		TLSConfig: tlsCfg,
	}, tlsCfg, nil
}

func newGRPCServer(cfg *config.Config, tlsCfg *tls.Config, api *facade.Shortener) (*googlegrpc.Server, net.Listener, error) {
	opts := []googlegrpc.ServerOption{googlegrpc.UnaryInterceptor(auth.UnaryServerInterceptor)}
	if cfg.EnableHTTPS && tlsCfg != nil {
		opts = append(opts, googlegrpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	srv := googlegrpc.NewServer(opts...)
	grpcpb.RegisterShortenerServiceServer(srv, grpc.NewShortenerServer(api))
	lis, err := net.Listen("tcp", cfg.GRPCListenAddr())
	if err != nil {
		return nil, nil, err
	}
	return srv, lis, nil
}

func tlsHostFromBaseURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base_url: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return "", errors.New("base_url must contain a host for HTTPS")
	}
	return host, nil
}

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 10 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	//Пингуем с таймаутом, чтобы сразу падать, если БД не достпна
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
