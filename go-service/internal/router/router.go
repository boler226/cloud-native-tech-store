// Package router збирає всі маршрути в один http.Handler.
package router

import (
	"net/http"

	"tech-store-go/internal/handler"
	"tech-store-go/internal/httpx"
	"tech-store-go/internal/middleware"
)

func New(health *handler.HealthHandler, products *handler.ProductHandler) http.Handler {
	mux := http.NewServeMux()

	// Go 1.22+: метод і параметри шляху вказуються прямо у патерні.
	mux.HandleFunc("GET /health", health.Health)

	mux.HandleFunc("GET /products", products.List)
	mux.HandleFunc("POST /products", products.Create)
	mux.HandleFunc("GET /products/{id}", products.Get)
	mux.HandleFunc("PUT /products/{id}", products.Update)
	mux.HandleFunc("DELETE /products/{id}", products.Delete)

	// Патерни без методу спрацьовують, коли шлях існує, а метод — ні (405).
	mux.HandleFunc("/health", methodNotAllowed("GET"))
	mux.HandleFunc("/products", methodNotAllowed("GET", "POST"))
	mux.HandleFunc("/products/{id}", methodNotAllowed("GET", "PUT", "DELETE"))

	// Усе інше — 404 у форматі JSON.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, http.StatusNotFound, "Route not found")
	})

	return middleware.Logging(middleware.Recover(mux))
}

func methodNotAllowed(allowed ...string) http.HandlerFunc {
	header := ""
	for i, m := range allowed {
		if i > 0 {
			header += ", "
		}
		header += m
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", header)
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
