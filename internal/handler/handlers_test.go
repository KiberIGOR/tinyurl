package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
func TestGetUrlHandler(t *testing.T) {
	type want struct {
        code        int
        response    string
				location string
  }
	tests := []struct {
		name string
		want want
		method string
		path string
	}{
		{
			name: "positive test #1",
			want: want{
					code:        307,
					response:    ``,
					location: "https://practicum.yandex.ru/",
			},
			method: http.MethodGet,
			path: "EwHXdJfB",
    },
		{
			name: "negative test #2",
			want: want{
					code:        400,
					response:    "URL not found\n",
					location: "",
			},
			method: http.MethodGet,
			path: "EwHXdJfA",
    },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := repository.NewMemory()
			store.Save("EwHXdJfB", "https://practicum.yandex.ru/")
			h := New(service.NewShortener(store, "http://localhost:8080/"))

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
		code int
		response string
		contentType string
	}
	tests := []struct {
		name string
		want want
		method string
		contentType string
		data string
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				response:    redirect,
				contentType: "text/plain",
			},
			method: http.MethodPost,
			contentType: "text/plain",
			data: "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #2 other contentType",
			want: want{
				code:400,
				response:"Only content-type: text/plain are allowed!\n",
				contentType:"text/plain; charset=utf-8",
			},
			method: http.MethodPost,
			contentType: "application/json",
			data: "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:400,
				response:"URL is required\n",
				contentType:"text/plain; charset=utf-8",
			},
			method: http.MethodPost,
			contentType: "text/plain",
			data: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method,"/", strings.NewReader(test.data))
			request.Header.Set("content-type",test.contentType)

			w:=httptest.NewRecorder()
			store := repository.NewMemory()
			h := New(service.NewShortener(store, redirect))
			h.PostUrlHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody,err :=io.ReadAll(res.Body)
			require.NoError(t,err)
			body:=string(resBody)
			if test.want.code == http.StatusCreated {
				assert.True(t, strings.HasPrefix(body, test.want.response),
					"response body should start with %q, got %q", test.want.response, body)
				id := strings.TrimPrefix(body, test.want.response)
				assert.NotEmpty(t, id)
				originalURL, ok := store.Get(id)
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
		code int
		response string
		contentType string
	}
	tests := []struct {
		name string
		want want
		method string
		contentType string
		data string
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				response:    redirect,
				contentType: "application/json",
			},
			method: http.MethodPost,
			contentType: "application/json",
			data: `{"url": "https://practicum.yandex.ru"}`,
		},
		{
			name: "negative test #2 other contentType",
			want: want{
				code:400,
				response:"Only content-type: application/json are allowed!\n",
				contentType:"text/plain; charset=utf-8",
			},
			method: http.MethodPost,
			contentType: "text/plain",
			data: "https://practicum.yandex.ru/",
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:400,
				response:"unexpected end of JSON input\n",
				contentType:"text/plain; charset=utf-8",
			},
			method: http.MethodPost,
			contentType: "application/json",
			data: ``,
		},
		{
			name: "negative test #4 no URL",
			want: want{
				code:400,
				response:"URL is required\n",
				contentType:"text/plain; charset=utf-8",
			},
			method: http.MethodPost,
			contentType: "application/json",
			data: `{"url": ""}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method,"/", strings.NewReader(test.data))
			request.Header.Set("content-type",test.contentType)

			w:=httptest.NewRecorder()
			store := repository.NewMemory()
			h := New(service.NewShortener(store, redirect))
			h.PostJsonUrlHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody,err :=io.ReadAll(res.Body)
			require.NoError(t,err)
			if test.want.code == http.StatusCreated {
				var resp model.Response
				var req model.Request
				err = json.Unmarshal(resBody, &resp)
				require.NoError(t,err)
				assert.True(t, strings.HasPrefix(resp.URL, test.want.response),
					"response body should start with %q, got %q", test.want.response, resp.URL)
				id := strings.TrimPrefix(resp.URL, test.want.response)
				assert.NotEmpty(t, id)
				originalURL, ok := store.Get(id)
				assert.True(t, ok)
				err = json.Unmarshal([]byte(test.data), &req)
				assert.Equal(t, req.URL, originalURL)
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}

			assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
		})
	}
}