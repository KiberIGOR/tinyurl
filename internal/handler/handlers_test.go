package handler

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/mocks"
	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func randomUserID(t *testing.T) int {
	t.Helper()
	var b [4]byte
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	return int(binary.BigEndian.Uint32(b[:])%999999) + 1
}

func requestWithUserID(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), auth.KeyUserID, userID)
	return r.WithContext(ctx)
}

func TestGetUrlHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
	}
	tests := []struct {
		name   string
		want   want
		method string
		path   string
	}{
		{
			name: "positive test #1",
			want: want{
				code:     307,
				response: ``,
				location: "https://practicum.yandex.ru/",
			},
			method: http.MethodGet,
			path:   "EwHXdJfB",
		},
		{
			name: "negative test #2",
			want: want{
				code:     400,
				response: "URL not found\n",
				location: "",
			},
			method: http.MethodGet,
			path:   "EwHXdJfA",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store, err := repository.NewFile("fileMemoryTest.txt")
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockPinger(ctrl)

			require.NoError(t, err)
			defer store.Close()
			if test.want.code == http.StatusTemporaryRedirect {
				if _, ok := store.Get(context.Background(), test.path); !ok {
					_, err = store.Save(context.Background(), test.path, test.want.location, randomUserID(t))
					require.NoError(t, err)
				}
			}
			h := New(service.NewShortener(store, "http://localhost:8080/"), m)

			request := httptest.NewRequest(test.method, "/"+test.path, nil)
			request.SetPathValue("id", test.path)
			w := httptest.NewRecorder()
			h.GetURL(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.location, res.Header.Get("location"))
		})
	}
}

func TestPostUrlHandler(t *testing.T) {
	const redirect = "http://localhost:8080/"

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		contentType string
		data        string
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				response:    redirect,
				contentType: "text/plain",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #2 other contentType",
			want: want{
				code:        400,
				response:    "Only content-type: text/plain are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:        400,
				response:    "URL is required\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := randomUserID(t)
			request := httptest.NewRequest(test.method, "/", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			request = requestWithUserID(request, userID)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockPinger(ctrl)

			w := httptest.NewRecorder()
			store, err := repository.NewFile("fileMemoryTest.txt")
			require.NoError(t, err)
			defer store.Close()
			h := New(service.NewShortener(store, redirect), m)
			h.PostURLHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			body := string(resBody)
			if test.want.code == http.StatusCreated {
				assert.True(t, strings.HasPrefix(body, test.want.response),
					"response body should start with %q, got %q", test.want.response, body)
				id := strings.TrimPrefix(body, test.want.response)
				assert.NotEmpty(t, id)
				originalURL, ok := store.Get(context.Background(), id)
				assert.True(t, ok)
				assert.Equal(t, test.data, originalURL)
			} else {
				assert.Equal(t, test.want.response, body)
			}

			assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
		})
	}
}

func TestPostJsonUrlHandler(t *testing.T) {
	const redirect = "http://localhost:8080/"

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		contentType string
		data        string
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				response:    redirect,
				contentType: "application/json",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"url": "https://practicum.yandex.ru"}`,
		},
		{
			name: "negative test #2 other contentType",
			want: want{
				code:        400,
				response:    "Only content-type: application/json are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:        400,
				response:    "unexpected end of JSON input\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        ``,
		},
		{
			name: "negative test #4 no URL",
			want: want{
				code:        400,
				response:    "URL is required\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"url": ""}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := randomUserID(t)
			request := httptest.NewRequest(test.method, "/api/shorten", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			request = requestWithUserID(request, userID)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockPinger(ctrl)

			w := httptest.NewRecorder()
			store, err := repository.NewFile("fileMemoryTest.txt")
			require.NoError(t, err)
			defer store.Close()
			h := New(service.NewShortener(store, redirect), m)
			h.PostJSONURLHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if test.want.code == http.StatusCreated {
				var resp model.Response
				var req model.Request
				err = json.Unmarshal(resBody, &resp)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(resp.URL, test.want.response),
					"response body should start with %q, got %q", test.want.response, resp.URL)
				id := strings.TrimPrefix(resp.URL, test.want.response)
				assert.NotEmpty(t, id)
				originalURL, ok := store.Get(context.Background(), id)
				assert.True(t, ok)
				err = json.Unmarshal([]byte(test.data), &req)
				require.NoError(t, err)
				assert.Equal(t, req.URL, originalURL)
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}

			assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
		})
	}
}

func TestPostBatchHandler(t *testing.T) {
	const redirect = "http://localhost:8080/"

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		contentType string
		data        string
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				response:    redirect,
				contentType: "application/json",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data: `[
    {
        "correlation_id": "1",
        "original_url": "https://practicum.yandex.ru"
    },
    {
        "correlation_id": "2",
        "original_url": "https://practicum.yandex.ru"
    }
]`,
		},
		{
			name: "negative test #2 other contentType",
			want: want{
				code:        400,
				response:    "Only content-type: application/json are allowed!\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:        400,
				response:    "unexpected end of JSON input\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        ``,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := randomUserID(t)
			request := httptest.NewRequest(test.method, "/api/shorten/batch", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			request = requestWithUserID(request, userID)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockPinger(ctrl)

			w := httptest.NewRecorder()
			store, err := repository.NewFile("fileMemoryTest.txt")
			require.NoError(t, err)
			defer store.Close()
			h := New(service.NewShortener(store, redirect), m)
			h.PostBatchHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if test.want.code == http.StatusCreated {
				var resp []model.BatchResponse
				var req []model.BatchRequest
				err = json.Unmarshal(resBody, &resp)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(resp[0].ShortURL, test.want.response),
					"response body should start with %q, got %q", test.want.response, resp[0].ShortURL)
				id := strings.TrimPrefix(resp[0].ShortURL, test.want.response)
				assert.NotEmpty(t, id)
				originalURL, ok := store.Get(context.Background(), id)
				assert.True(t, ok)
				err = json.Unmarshal([]byte(test.data), &req)
				require.NoError(t, err)
				assert.Equal(t, req[0].OriginalURL, originalURL)
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}

			assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
		})
	}
}

func TestGetUsersURLs(t *testing.T) {
	const redirect = "http://localhost:8080/"

	type want struct {
		code        int
		responseShort    string
		responseOriginal    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		userID        int
	}{
		{
			name: "positive test #1",
			want: want{
				code:        200,
				responseShort:    "EwHXdJfT",
				responseOriginal:    "https://practicum.yandex.ru",
				contentType: "application/json",
			},
			method:      http.MethodGet,
			userID: 1,
		},
		{
			name: "positive test #2",
			want: want{
				code:        204,
				responseShort:    "EwHXdJfT",
				responseOriginal:    "https://practicum.yandex.ru",
				contentType: "application/json",
			},
			method:      http.MethodGet,
			userID: 246,
		},
		{
			name: "negative test #1",
			want: want{
				code:        401,
				responseShort:    "EwHXdJfT",
				responseOriginal:    "https://practicum.yandex.ru",
				contentType: "application/json",
			},
			method:      http.MethodGet,
			userID: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := test.userID
			request := httptest.NewRequest(test.method, "/api/user/urls", nil)
			request = requestWithUserID(request, userID)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockPinger(ctrl)

			w := httptest.NewRecorder()
			store, err := repository.NewFile("fileMemoryTest.txt")
			require.NoError(t, err)
			defer store.Close()
			if test.want.code == http.StatusOK {
				if _, ok := store.Get(context.Background(), test.want.responseShort); !ok {
					_, err = store.Save(context.Background(), test.want.responseShort, test.want.responseOriginal, userID)
					require.NoError(t, err)
				}
			}
			h := New(service.NewShortener(store, redirect), m)
			h.GetUsersURLs(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			
			if test.want.code == http.StatusOK {
				var resp []model.UserResponse
				err = json.Unmarshal(resBody, &resp)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(resp[0].ShortURL, redirect),
					"response body should start with %q, got %q", redirect, resp[0].ShortURL)
				id := strings.TrimPrefix(resp[0].ShortURL, redirect)
				assert.NotEmpty(t, id)
				assert.Equal(t,id, test.want.responseShort)
				assert.Equal(t,resp[0].OriginalURL, test.want.responseOriginal)
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}