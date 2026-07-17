package main

import (
	"net/http"

	"log"

	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Parse()

	if err:=logger.Initialize("info");err!=nil {
		log.Fatal(err)
	}

	store := repository.NewMemory()
	svc := service.NewShortener(store, cfg.BaseURL)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Post("/", logger.RequestLogger(h.PostUrlHandler))
	r.Get("/{id}",logger.RequestLogger(h.GetURL))
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
