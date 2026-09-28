package auth

import (
	"context"
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
		var id int
		if err != nil {
			id = setCookie(w, r)
		} else {
			id = auth.GetUserID(r.Context(), cookie.Value)
			if id <= 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		if id <= 0 {
			return
		}
		ctx := context.WithValue(r.Context(), KeyUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
