// Package router збирає всі маршрути в один http.Handler.
package router

import (
	"net/http"

	"tech-store-go/internal/controller"
	"tech-store-go/internal/handler"
	"tech-store-go/internal/httpx"
	"tech-store-go/internal/middleware"
)

// New збирає всі маршрути. Router відповідає лише за endpoint-и, HTTP-методи,
// зв'язок із контролерами та підключення middleware — жодної бізнес-логіки.
func New(
	health *handler.HealthHandler,
	products *handler.ProductHandler,
	concurrency *handler.ConcurrencyHandler,
	users *controller.UserController,
) http.Handler {
	mux := http.NewServeMux()

	// Go 1.22+: метод і параметри шляху вказуються прямо у патерні.
	mux.HandleFunc("GET /health", health.Health)

	mux.HandleFunc("GET /products", products.List)
	mux.HandleFunc("POST /products", products.Create)
	mux.HandleFunc("GET /products/{id}", products.Get)
	mux.HandleFunc("PUT /products/{id}", products.Update)
	mux.HandleFunc("DELETE /products/{id}", products.Delete)

	// Лабораторна 2: дослідження моделей конкурентності.
	mux.HandleFunc("GET /io", concurrency.IO)
	mux.HandleFunc("GET /cpu/sequential", concurrency.CPUSequential)
	mux.HandleFunc("GET /cpu/parallel", concurrency.CPUParallel)

	// Лабораторна 4: користувачі. Кожен маршрут: middleware валідації -> контролер.
	// httpx.Wrap перетворює помилку, повернену контролером, на HTTP-відповідь.
	mux.Handle("GET /api/users", httpx.Wrap(users.List))
	mux.Handle("GET /api/users/{id}", middleware.ValidateID(httpx.Wrap(users.Get)))
	mux.Handle("POST /api/users", middleware.ValidateUserBody(httpx.Wrap(users.Create)))
	mux.Handle("PUT /api/users/{id}", middleware.ValidateID(middleware.ValidateUserBody(httpx.Wrap(users.Update))))
	mux.Handle("DELETE /api/users/{id}", middleware.ValidateID(httpx.Wrap(users.Delete)))

	// Патерни без методу спрацьовують, коли шлях існує, а метод — ні (405).
	mux.HandleFunc("/health", methodNotAllowed("GET"))
	mux.HandleFunc("/products", methodNotAllowed("GET", "POST"))
	mux.HandleFunc("/products/{id}", methodNotAllowed("GET", "PUT", "DELETE"))
	mux.HandleFunc("/io", methodNotAllowed("GET"))
	mux.HandleFunc("/cpu/sequential", methodNotAllowed("GET"))
	mux.HandleFunc("/cpu/parallel", methodNotAllowed("GET"))
	mux.HandleFunc("/api/users", methodNotAllowed("GET", "POST"))
	mux.HandleFunc("/api/users/{id}", methodNotAllowed("GET", "PUT", "DELETE"))

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
