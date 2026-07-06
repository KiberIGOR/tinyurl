package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/config"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/go-chi/chi/v5"
	"log"
)

func main() {
	cfg := config.Parse()

	urls := make(map[string]string)

	r := chi.NewRouter()
	r.Post("/", handler.PostUrlHandler(urls, cfg.BaseURL))
	r.Get("/{id}", handler.GetUrlHandler(urls))
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
