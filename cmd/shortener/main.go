package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/handler"
)

func main() {
	urls := make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.PostUrlHandler(&urls))
	mux.HandleFunc("/{id}", handler.GetUrlHandler(&urls))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
