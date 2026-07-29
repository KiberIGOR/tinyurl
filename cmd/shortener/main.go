package main

import (
	"net/http"

	"log"

	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/KiberIGOR/tinyurl/internal/compress"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	if err:=logger.Initialize("info");err!=nil {
		log.Fatal(err)
	}

	store,err := repository.NewMemory(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	svc := service.NewShortener(store, cfg.BaseURL)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Post("/", logger.RequestLogger(compress.GzipMiddleware(h.PostUrlHandler)))
	r.Post("/api/shorten", logger.RequestLogger(compress.GzipMiddleware(h.PostJsonUrlHandler)))
	r.Get("/{id}",logger.RequestLogger(compress.GzipMiddleware(h.GetURL)))
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
