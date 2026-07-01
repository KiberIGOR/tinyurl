package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	urls := make(map[string]string)

	r := chi.NewRouter()
	r.Post("/", handler.PostUrlHandler(&urls))
	r.Get("/{id}", handler.GetUrlHandler(&urls))
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
