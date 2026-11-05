package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/mocks"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"go.uber.org/mock/gomock"
)

func BenchmarkPostURLHandler(b *testing.B) {
	store := repository.NewMemory()
	h := New(service.NewShortener(store, "http://localhost:8080/"), mocks.NewMockPinger(gomock.NewController(b)), nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf("https://practicum.yandex.ru/bench/%d", i)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("content-type", "text/plain")
		ctx := context.WithValue(req.Context(), auth.KeyUserID, 1)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h.PostURLHandler(w, req)
		if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

func BenchmarkGetURLHandler(b *testing.B) {
	store := repository.NewMemory()
	ctx := context.Background()
	const short = "benchGet1"
	if _, err := store.Save(ctx, short, "https://practicum.yandex.ru/", 1); err != nil {
		b.Fatal(err)
	}
	h := New(service.NewShortener(store, "http://localhost:8080/"), mocks.NewMockPinger(gomock.NewController(b)), nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/"+short, nil)
		req.SetPathValue("id", short)
		w := httptest.NewRecorder()
		h.GetURL(w, req)
		if w.Code != http.StatusTemporaryRedirect {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}
