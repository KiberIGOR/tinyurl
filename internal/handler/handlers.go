package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/service"
)

type Shortener interface {
	Shorten(ctx context.Context, originalURL string) (shortURL string, err error)
	Resolve(ctx context.Context, id string) (originalURL string, err error)
	MassiveShorten(ctx context.Context, MassiveOriginalURL []model.MassiveRequest) ([]model.MassiveResponse, error)
}
type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	shortener Shortener
	pinger Pinger
}

func New(shortener Shortener, pinger Pinger) *Handler {
	return &Handler{shortener: shortener, pinger: pinger}
}

func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	id := req.PathValue("id")
	originalURL, err := h.shortener.Resolve(ctx, id)
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

func (h *Handler) PostURLHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
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
	shortURL, err := h.shortener.Shorten(ctx,string(body))
	status := http.StatusCreated
	if err != nil {
		if !errors.Is(err, service.ErrConflict) {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		status = http.StatusConflict
	}
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("content-length", strconv.Itoa(len(shortURL)))
	res.WriteHeader(status)
	if _, err = res.Write([]byte(shortURL)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) PostJSONURLHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	content := req.Header.Get("content-type")
	if content != "application/json" {
		http.Error(res, "Only content-type: application/json are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var bodyReq model.Request
	err = json.Unmarshal(body, &bodyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(bodyReq.URL) == 0 {
		http.Error(res, "URL is required", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.Shorten(ctx,bodyReq.URL)
	status := http.StatusCreated
	if err != nil {
		if !errors.Is(err, service.ErrConflict) {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		status = http.StatusConflict
	}
	resp, err := json.Marshal(model.Response{URL: shortURL})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("content-type", "application/json")
	res.Header().Set("content-length", strconv.Itoa(len(resp)))
	res.WriteHeader(status)
	if _, err = res.Write(resp); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) GetPingHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	if err := h.pinger.Ping(ctx); err !=nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) PostJSONMassiveURLHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	content := req.Header.Get("content-type")
	if content != "application/json" {
		http.Error(res, "Only content-type: application/json are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var bodyReq []model.MassiveRequest
	err = json.Unmarshal(body, &bodyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(bodyReq) == 0 {
		http.Error(res, "URL is required", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.MassiveShorten(ctx, bodyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(shortURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("content-type", "application/json")
	res.Header().Set("content-length", strconv.Itoa(len(resp)))
	res.WriteHeader(http.StatusCreated)
	if _, err = res.Write(resp); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}