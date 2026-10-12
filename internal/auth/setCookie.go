package auth

import (
	"context"
	"net/http"

	"github.com/KiberIGOR/tinyurl/internal/logger"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"go.uber.org/zap"
)

func SetCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Token")
		var id int
		if err != nil {
			id = setCookie(w, r)
		} else {
			id = auth.GetUserID(r.Context(), cookie.Value)
			if id <= 0 {
				id = setCookie(w, r)
			}
		}
		if id <= 0 {
			return
		}
		ctx := context.WithValue(r.Context(), KeyUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setCookie(w http.ResponseWriter, r *http.Request) int {
	id := auth.GetLastID(r.Context())
	logger.Log.Info("new user created", zap.Int("user_id", id))

	token, err := auth.BuildJWTString(r.Context(), id)
	if err != nil {
		logger.Log.Error("failed to build JWT for new user", zap.Int("user_id", id), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return -1
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(service.TokenExp.Seconds()),
	})
	return id
}
