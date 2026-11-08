package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/handler"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
)

type okPinger struct{}

func (okPinger) Ping(context.Context) error { return nil }

func newExampleHandler() *handler.Handler {
	store := repository.NewMemory()
	svc := service.NewShortener(store, "http://localhost:8080/")
	return handler.New(svc, okPinger{}, nil)
}

func withUserID(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), auth.KeyUserID, userID)
	return req.WithContext(ctx)
}

// POST / — сокращение URL (text/plain).
func ExampleHandler_PostURLHandler() {
	h := newExampleHandler()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	req.Header.Set("content-type", "text/plain")
	req = withUserID(req, 1)

	w := httptest.NewRecorder()
	h.PostURLHandler(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.HasPrefix(w.Body.String(), "http://localhost:8080/"))
	// Output:
	// 201
	// true
}

// POST /api/shorten — сокращение URL (JSON).
func ExampleHandler_PostJSONURLHandler() {
	h := newExampleHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://practicum.yandex.ru/"}`))
	req.Header.Set("content-type", "application/json")
	req = withUserID(req, 1)

	w := httptest.NewRecorder()
	h.PostJSONURLHandler(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), `"result"`))
	// Output:
	// 201
	// true
}

// GET /{id} — переход по сокращённой ссылке (307 Temporary Redirect).
func ExampleHandler_GetURL() {
	store := repository.NewMemory()
	ctx := context.Background()
	const shortID = "abc12345"
	_, _ = store.Save(ctx, shortID, "https://practicum.yandex.ru/", 1)

	h := handler.New(service.NewShortener(store, "http://localhost:8080/"), okPinger{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	req.SetPathValue("id", shortID)
	w := httptest.NewRecorder()
	h.GetURL(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))
	// Output:
	// 307
	// https://practicum.yandex.ru/
}

// GET /ping — проверка доступности хранилища.
func ExampleHandler_GetPingHandler() {
	h := newExampleHandler()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h.GetPingHandler(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// POST /api/shorten/batch — пакетное сокращение.
func ExampleHandler_PostBatchHandler() {
	h := newExampleHandler()

	body := `[{"correlation_id":"1","original_url":"https://a.example/"},{"correlation_id":"2","original_url":"https://b.example/"}]`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req = withUserID(req, 1)

	w := httptest.NewRecorder()
	h.PostBatchHandler(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), "correlation_id"))
	// Output:
	// 201
	// true
}

// GET /api/user/urls — список URL пользователя.
func ExampleHandler_GetUsersURLs() {
	store := repository.NewMemory()
	ctx := context.Background()
	_, _ = store.Save(ctx, "link0001", "https://practicum.yandex.ru/", 42)

	h := handler.New(service.NewShortener(store, "http://localhost:8080/"), okPinger{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = withUserID(req, 42)
	w := httptest.NewRecorder()
	h.GetUsersURLs(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), "original_url"))
	// Output:
	// 200
	// true
}

// DELETE /api/user/urls — асинхронное удаление (202 Accepted).
func ExampleHandler_DeleteUsersURLs() {
	store := repository.NewMemory()
	ctx := context.Background()
	_, _ = store.Save(ctx, "del00001", "https://practicum.yandex.ru/", 5)

	h := handler.New(service.NewShortener(store, "http://localhost:8080/"), okPinger{}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["del00001"]`))
	req.Header.Set("content-type", "application/json")
	req = withUserID(req, 5)
	w := httptest.NewRecorder()
	h.DeleteUsersURLs(w, req)

	fmt.Println(w.Code)
	// Output:
	// 202
}
