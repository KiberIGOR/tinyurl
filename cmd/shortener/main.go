package main

import (
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/handler"
)
func main() {
	mux:=http.NewServeMux()
	mux.HandleFunc("/", handler.PostUrlHandler)
	mux.HandleFunc("/{id}", handler.GetUrlHandler)
	err := http.ListenAndServe(":8080", mux)
	if err!= nil {
		panic(err)
	}
}
