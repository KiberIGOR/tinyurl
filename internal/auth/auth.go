// Package auth реализует JWT-авторизацию и middleware для HTTP API.
package auth

import (
	"context"
	"net/http"
)

type ctxKey string

// KeyUserID — ключ user id в context.Request.Context().
const KeyUserID ctxKey = "user_id"

// AuthService выпускает JWT и выделяет идентификаторы пользователей.
type AuthService interface {
	BuildJWTString(ctx context.Context, id int) (string, error)
	GetUserID(ctx context.Context, tokenString string) int
	GetLastID(ctx context.Context) int
}

var auth AuthService

// Initialize задаёт реализацию AuthService для middleware.
func Initialize(a AuthService) {
	auth = a
}

// AuthMiddleware проверяет cookie Token или создаёт нового пользователя (GET /api/user/urls).
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
