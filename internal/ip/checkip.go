// Package ip проверяет IP клиента для внутренних эндпоинтов.
package ip

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var trusted *net.IPNet

// Initialize задаёт доверенную подсеть в формате CIDR. Пустая строка запрещает доступ всем.
func Initialize(subnet string) error {
	if subnet == "" {
		trusted = nil
		return nil
	}
	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("trusted subnet: %w", err)
	}
	trusted = network
	return nil
}

// IPMiddleware проверяет заголовок X-Real-IP относительно доверенной подсети.
func IPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if trusted == nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		raw := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if raw == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		clientIP := net.ParseIP(raw)
		if clientIP == nil || !trusted.Contains(clientIP) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
