package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

var ( 
	mu sync.Mutex
)

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}

func GetUrlHandler(urls map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")
		mu.Lock()
		url, ok := urls[id]
		mu.Unlock()
		if !ok {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostUrlHandler(urls map[string]string,redirect string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
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
		//решение коллизии
		var id string
		for {
			id, err = generateID()
			if err != nil {
				http.Error(res, "can't generate id", http.StatusInternalServerError)
				return
			}
			mu.Lock()
    	if _, exists := urls[id]; !exists {
        urls[id] = string(body)
        mu.Unlock()
        break
    	}
    	mu.Unlock()
		}
		shortURL, err := url.JoinPath(redirect,id)
		if err != nil {
			http.Error(res, "can't join path", http.StatusInternalServerError)
			return
		}
		res.Header().Set("content-type", "text/plain")
		res.Header().Set("content-length", fmt.Sprintf("%d", len(shortURL)))
		res.WriteHeader(http.StatusCreated)
		_,err = res.Write([]byte(shortURL))
		if err != nil {
			http.Error(res, "can't write body", http.StatusInternalServerError)
			return
		}
	}
}
