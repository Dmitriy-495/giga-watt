// Package httpserver содержит общие технические механизмы HTTP-слоя
// backend Giga Watt (не относится к предметной области).
package httpserver

import "net/http"

// WithCORS оборачивает handler, добавляя заголовки CORS для запросов
// с frontend (Nuxt dev-сервер и т.п.), и обрабатывает preflight-запросы
// OPTIONS.
func WithCORS(allowOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
