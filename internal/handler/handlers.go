package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/KiberIGOR/tinyurl/internal/service"
)

type Shortener interface {
	Shorten(originalURL string) (shortURL string, err error)
	Resolve(id string) (originalURL string, err error)
}

type Handler struct {
	shortener Shortener
}

func New(shortener Shortener) *Handler {
	return &Handler{shortener: shortener}
}

func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	originalURL, err := h.shortener.Resolve(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) PostUrlHandler(res http.ResponseWriter, req *http.Request) {
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
	shortURL, err := h.shortener.Shorten(string(body))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("content-length", strconv.Itoa(len(shortURL)))
	res.WriteHeader(http.StatusCreated)
	if _, err = res.Write([]byte(shortURL)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
