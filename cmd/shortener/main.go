package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	urls := make(map[string]string)

	parseFlags()

	r := chi.NewRouter()
	r.Post("/", handler.PostUrlHandler(&urls,flagRedirectAddr))
	r.Get("/{id}", handler.GetUrlHandler(&urls))
	err := http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		panic(err)
	}
}
