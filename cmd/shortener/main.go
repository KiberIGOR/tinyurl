package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"log"

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
)

// stub
type nilPinger struct{}

func (nilPinger) Ping(ctx context.Context) error {
	return errors.New("database is not configured")
}

func main() {
	cfg := config.Parse()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err)
	}
	var store service.URLRepository
	var pinger handler.Pinger
	var err error
	switch {
	case cfg.DataBaseDSN != "":
		initCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := runMigrations(cfg.DataBaseDSN); err != nil {
			log.Fatal("migrate: ", err)
		}
		pool, err := newPool(initCtx, cfg.DataBaseDSN)
		if err != nil {
			log.Fatal("pool: ", err)
		}
		defer pool.Close()
		pg := repository.NewDB(pool)
		store = pg
		pinger = pg
	case cfg.FileStoragePath != "":
		store, err = repository.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		pinger = nilPinger{}
	default:
		store = repository.NewMemory()
		pinger = nilPinger{}
	}
	svc := service.NewShortener(store, cfg.BaseURL)
	h := handler.New(svc, pinger)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)

	r.Post("/", h.PostURLHandler)
	r.Post("/api/shorten", h.PostJSONURLHandler)
	r.Post("/api/shorten/batch", h.PostBatchHandler)
	r.Get("/ping", h.GetPingHandler)
	r.Get("/{id}", h.GetURL)
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
