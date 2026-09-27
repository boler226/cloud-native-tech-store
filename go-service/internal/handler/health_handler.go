package handler

import (
	"net/http"
	"runtime"
	"time"

	"tech-store-go/internal/config"
	"tech-store-go/internal/httpx"
)

type healthResponse struct {
	Status        string `json:"status"`
	Service       string `json:"service"`
	Runtime       string `json:"runtime"`
	UptimeSeconds int    `json:"uptimeSeconds"`
	Timestamp     string `json:"timestamp"`
}

type HealthHandler struct {
	startedAt time.Time
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{startedAt: time.Now()}
}

// Health: GET /health -> 200 (перевірка, що сервер живий).
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, healthResponse{
		Status:        "ok",
		Service:       config.ServiceName,
		Runtime:       runtime.Version(),
		UptimeSeconds: int(time.Since(h.startedAt).Seconds()),
		Timestamp:     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	})
}
