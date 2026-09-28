package auth

import (
	"context"
	"io"
	"net/http"
)

type ctxKey string

const KeyUserID ctxKey = "user_id"

type AuthService interface {
	BuildJWTString(ctx context.Context, id int) (string, error)
	GetUserID(ctx context.Context, tokenString string) int
	GetLastID(ctx context.Context) int
}

var auth AuthService

func Initialize(a AuthService) {
	auth = a
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "no auth Cookie")
			return
		}
		tokenString := cookie.Value
		id := auth.GetUserID(r.Context(), tokenString)
		if id <= 0 {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), KeyUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
