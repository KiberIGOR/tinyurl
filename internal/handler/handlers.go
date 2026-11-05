// Package handler реализует HTTP API сервиса сокращения URL (Практикум, трек Go).
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/KiberIGOR/tinyurl/internal/audit"
	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"go.uber.org/zap"
)

func respondInternalError(res http.ResponseWriter, err error) {
	logger.Log.Error("internal server error", zap.Error(err))
	http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func respondBadRequest(res http.ResponseWriter, err error) {
	logger.Log.Error("bad request", zap.Error(err))
	http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// Shortener — доменная логика сокращения и получения URL (для подмены в тестах).
type Shortener interface {
	Shorten(ctx context.Context, originalURL string, userID int) (shortURL string, err error)
	Resolve(ctx context.Context, id string) (originalURL string, err error)
	BatchShorten(ctx context.Context, batch []model.BatchRequest, userID int) ([]model.BatchResponse, error)
	ResolveByUserID(ctx context.Context, userID int) ([]model.UserResponse, error)
	DeleteByUserID(ctx context.Context, userID int, shortURLCh chan string)
}

// Pinger проверяет доступность бэкенда (например, PostgreSQL).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler обрабатывает HTTP-запросы к API сокращателя.
type Handler struct {
	shortener Shortener
	pinger    Pinger
	audit     *audit.Broker
}

// New создаёт Handler. auditBroker может быть nil — аудит отключён.
func New(shortener Shortener, pinger Pinger, auditBroker *audit.Broker) *Handler {
	return &Handler{shortener: shortener, pinger: pinger, audit: auditBroker}
}

func userIDFromContext(ctx context.Context) int {
	userID, _ := ctx.Value(auth.KeyUserID).(int)
	return userID
}

func (h *Handler) publishAudit(ctx context.Context, event audit.Event) {
	if h.audit == nil {
		return
	}
	h.audit.Publish(ctx, event)
}

// GetURL обрабатывает GET /{id}: редирект на original URL (307) или 410, если ссылка удалена.
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	id := req.PathValue("id")
	originalURL, err := h.shortener.Resolve(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrGone) {
			http.Error(res, http.StatusText(http.StatusGone), http.StatusGone)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		respondInternalError(res, err)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
	h.publishAudit(ctx, audit.NewFollowEvent(time.Now().Unix(), userIDFromContext(req.Context()), originalURL))
}

// PostURLHandler обрабатывает POST /: тело text/plain с URL, ответ — полный short URL (201/409).
func (h *Handler) PostURLHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostURLHandler error: no user id"))
		return
	}
	content := req.Header.Get("content-type")
	if content != "text/plain" && content != "text/plain;charset=UTF-8" {
		http.Error(res, "Only content-type: text/plain are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
		return
	}
	if len(body) == 0 {
		http.Error(res, "URL is required", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.Shorten(ctx, string(body), userID)
	status := http.StatusCreated
	if err != nil {
		if !errors.Is(err, service.ErrConflict) {
			respondInternalError(res, err)
			return
		}
		status = http.StatusConflict
	}
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("content-length", strconv.Itoa(len(shortURL)))
	res.WriteHeader(status)
	if _, err = res.Write([]byte(shortURL)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.publishAudit(ctx, audit.NewShortenEvent(time.Now().Unix(), userID, string(body)))
}

// PostJSONURLHandler обрабатывает POST /api/shorten: JSON {"url":"..."} → {"result":"..."}.
func (h *Handler) PostJSONURLHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}
	content := req.Header.Get("content-type")
	if content != "application/json" {
		http.Error(res, "Only content-type: application/json are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
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
	shortURL, err := h.shortener.Shorten(ctx, bodyReq.URL, userID)
	status := http.StatusCreated
	if err != nil {
		if !errors.Is(err, service.ErrConflict) {
			respondInternalError(res, err)
			return
		}
		status = http.StatusConflict
	}
	resp, err := json.Marshal(model.Response{URL: shortURL})
	if err != nil {
		respondInternalError(res, err)
		return
	}
	res.Header().Set("content-type", "application/json")
	res.Header().Set("content-length", strconv.Itoa(len(resp)))
	res.WriteHeader(status)
	if _, err = res.Write(resp); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.publishAudit(ctx, audit.NewShortenEvent(time.Now().Unix(), userID, bodyReq.URL))
}

// GetPingHandler обрабатывает GET /ping: 200, если Pinger.Ping успешен.
func (h *Handler) GetPingHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		respondInternalError(res, err)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// PostBatchHandler обрабатывает POST /api/shorten/batch: массив пар correlation_id / original_url.
func (h *Handler) PostBatchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostBatchHandler error: no user id"))
		return
	}
	content := req.Header.Get("content-type")
	if content != "application/json" {
		http.Error(res, "Only content-type: application/json are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
		return
	}
	var bodyReq []model.BatchRequest
	err = json.Unmarshal(body, &bodyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(bodyReq) == 0 {
		http.Error(res, "URL is required", http.StatusBadRequest)
		return
	}
	shortURL, err := h.shortener.BatchShorten(ctx, bodyReq, userID)
	if err != nil {
		respondInternalError(res, err)
		return
	}
	resp, err := json.Marshal(shortURL)
	if err != nil {
		respondInternalError(res, err)
		return
	}
	res.Header().Set("content-type", "application/json")
	res.Header().Set("content-length", strconv.Itoa(len(resp)))
	res.WriteHeader(http.StatusCreated)
	if _, err = res.Write(resp); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetUsersURLs обрабатывает GET /api/user/urls: JSON списка ссылок пользователя (200/204/401).
func (h *Handler) GetUsersURLs(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok || userID <= 0 {
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	urls, err := h.shortener.ResolveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoContent) {
			res.WriteHeader(http.StatusNoContent)
			return
		}
		respondInternalError(res, err)
		return
	}
	out, err := json.Marshal(urls)
	if err != nil {
		respondInternalError(res, err)
		return
	}
	res.Header().Set("content-type", "application/json")
	res.Header().Set("content-length", strconv.Itoa(len(out)))
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(out); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// DeleteUsersURLs обрабатывает DELETE /api/user/urls: JSON-массив short id, ответ 202 Accepted.
func (h *Handler) DeleteUsersURLs(res http.ResponseWriter, req *http.Request) {
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok || userID <= 0 {
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	content := req.Header.Get("content-type")
	if content != "application/json" {
		http.Error(res, "Only content-type: application/json are allowed!", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
		return
	}
	var bodyReq []model.DeleteRequest
	if err := json.Unmarshal(body, &bodyReq); err != nil {
		respondBadRequest(res, err)
		return
	}
	shortURLCh := make(chan string, len(bodyReq))
	for _, item := range bodyReq {
		shortURLCh <- string(item)
	}
	close(shortURLCh)
	go h.shortener.DeleteByUserID(context.Background(), userID, shortURLCh)

	res.WriteHeader(http.StatusAccepted)
}
