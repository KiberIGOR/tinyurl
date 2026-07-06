package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			request := httptest.NewRequest(test.method, "/"+test.path, nil)
			request.SetPathValue("id", test.path)
			w := httptest.NewRecorder()
			urls := map[string]string{
				"EwHXdJfB":"https://practicum.yandex.ru/",
			}
			h:=GetUrlHandler(urls)
			h(w, request)

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
				code:201,
				contentType:"text/plain",
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
			urls := make(map[string]string)
			redirect := "http://localhost:8080"

			h:=PostUrlHandler(urls, redirect)
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody,err :=io.ReadAll(res.Body)
			require.NoError(t,err)
			if res.StatusCode!=http.StatusCreated {
				assert.Equal(t, test.want.response, string(resBody))
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			
		})
	}
}
