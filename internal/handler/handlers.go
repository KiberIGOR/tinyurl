package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)
//Сгенерировано ии
func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}

func GetUrlHandler(urls *map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
			return
		}
		id := req.PathValue("id")
		url, ok := (*urls)[id]
		if !ok {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostUrlHandler(urls *map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
			return
		}
		content := req.Header.Get("content-type")
		if content != "text/plain" && content != "text/plain;charset=UTF-8" {
			http.Error(res, "Only content-type: text/plain are allowed!", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if len(body) == 0 {
			http.Error(res, "URL is required", http.StatusBadRequest)
			return
		}
		id, err := generateID()
		if err != nil {
			http.Error(res, "can't generate id", http.StatusInternalServerError)
			return
		}
		(*urls)[id] = string(body)
		shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
		res.Header().Set("content-type", "text/plain")
		res.Header().Set("content-length", fmt.Sprintf("%d", len(shortURL)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}
