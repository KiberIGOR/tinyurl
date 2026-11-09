package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"log"

	"github.com/KiberIGOR/tinyurl/internal/audit"
	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/compress"
	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
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
	// go build -ldflags "-X main.buildVersion=1.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o shortener ./cmd/shortener
	fmt.Println("Build version:", buildLabel(buildVersion))
	fmt.Println("Build date:", buildLabel(buildDate))
	fmt.Println("Build commit:", buildLabel(buildCommit))

	cfg := config.Parse()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err)
	}
	var store service.URLRepository
	var userRepo service.UserRepository
	var pinger handler.Pinger
	var err error
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
		var fileStore *repository.FileMemory
		fileStore, err = repository.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		defer fileStore.Close()
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
	if cfg.AuditFile != "" {
		var fileObserver *audit.FileObserver
		fileObserver, err = audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			log.Fatal("audit file: ", err)
		}
		auditBroker.Subscribe(fileObserver)
	}
	if cfg.AuditURL != "" {
		auditBroker.Subscribe(audit.NewHTTPObserver(cfg.AuditURL, nil))
	}

	h := handler.New(svc, pinger, auditBroker)

	auth.Initialize(service.NewAuthService(userRepo, cfg.SecretKey))

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

	// go func() {
	// 	log.Println("pprof listening on :6060")
	// 	if err := http.ListenAndServe(":6060", nil); err != nil {
	// 		log.Print("pprof: ", err)
	// 	}
	// }()

	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
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
