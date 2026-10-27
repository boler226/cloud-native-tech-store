// Package middleware — обгортки навколо http.Handler (логування, відновлення після panic).
package middleware

import (
	"log"
	"net/http"
	"time"

	"tech-store-go/internal/httpx"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging — Middleware №1: виводить метод, шлях, статус і тривалість кожного запиту.
// Формат: "GET /api/users/10 200 4.2ms".
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		ms := float64(time.Since(start).Microseconds()) / 1000
		log.Printf("%s %s %d %.1fms", r.Method, r.URL.RequestURI(), rec.status, ms)
	})
}

// Recover перехоплює panic і повертає 500 замість обриву з'єднання.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
