package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
)

func main() {
	cfg := config.Parse()

	store := repository.NewMemory()
	svc := service.NewShortener(store, cfg.BaseURL)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Post("/", h.PostUrlHandler)
	r.Get("/{id}", h.GetURL)
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
