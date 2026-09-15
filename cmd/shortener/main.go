package main

import (
	"net/http"

	"log"

	"database/sql"

	"github.com/KiberIGOR/tinyurl/internal/compress"
	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Parse()

	if err:=logger.Initialize("info");err!=nil {
		log.Fatal(err)
	}

  db, err := sql.Open("pgx", cfg.DataBaseDSN)
	if err != nil {
        panic(err)
    }
  defer db.Close()

	store,err := repository.NewMemory(cfg.FileStoragePath)
	dbstore := repository.NewDB(db)
	if err != nil {
		log.Fatal(err)
	}
	svc := service.NewShortener(store, cfg.BaseURL)
	h := handler.New(svc, dbstore)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)
	
	r.Post("/", h.PostUrlHandler)
	r.Post("/api/shorten", h.PostJsonUrlHandler)
	r.Get("/{id}", h.GetURL)
	r.Get("/ping", h.GetPingHandler)
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
