// Package handler (concurrency) — HTTP-ендпойнти /io, /cpu/sequential,
// /cpu/parallel для дослідження моделей конкурентності.
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tech-store-go/internal/httpx"
	"tech-store-go/internal/service"
)

// Дефолти підібрано так, щоб один запит тривав ~0.3-0.5с у типовому
// середовищі; під своє залізо підберіть власні значення через query.
const (
	defaultIODelayMs = 200
	maxIODelayMs     = 30_000

	defaultCPUIterationsPerPass = 300_000_000
	maxCPUIterationsPerPass     = 20_000_000_000
	defaultCPUPasses            = 4
	maxCPUPasses                = 64
)

type ConcurrencyHandler struct {
	svc *service.ConcurrencyService
}

func NewConcurrencyHandler(svc *service.ConcurrencyService) *ConcurrencyHandler {
	return &ConcurrencyHandler{svc: svc}
}

// IO: GET /io?delayMs=200 -> 200.
// time.Sleep не блокує інші горутини (див. коментар у ConcurrencyService.Wait).
func (h *ConcurrencyHandler) IO(w http.ResponseWriter, r *http.Request) {
	delayMs, ok := parseIntQuery(w, r, "delayMs", defaultIODelayMs, 0, maxIODelayMs)
	if !ok {
		return
	}
	start := time.Now()
	h.svc.Wait(time.Duration(delayMs) * time.Millisecond)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"endpoint":         "io",
		"runtime":          "go",
		"delayMs":          delayMs,
		"actualDurationMs": time.Since(start).Milliseconds(),
	})
}

// CPUSequential: GET /cpu/sequential -> 200.
// passes важких проходів один за одним у ЄДИНІЙ горутині обробника запиту.
func (h *ConcurrencyHandler) CPUSequential(w http.ResponseWriter, r *http.Request) {
	iterations, passes, ok := parseCPUParams(w, r)
	if !ok {
		return
	}
	start := time.Now()
	result := h.svc.CPUSequential(iterations, passes)
	httpx.WriteJSON(w, http.StatusOK, cpuResponse("cpu-sequential", iterations, passes, result, start))
}

// CPUParallel: GET /cpu/parallel -> 200.
// Той самий сумарний обсяг роботи, що й /cpu/sequential, але passes окремих
// горутин виконуються одночасно.
func (h *ConcurrencyHandler) CPUParallel(w http.ResponseWriter, r *http.Request) {
	iterations, passes, ok := parseCPUParams(w, r)
	if !ok {
		return
	}
	start := time.Now()
	result := h.svc.CPUParallel(iterations, passes)
	httpx.WriteJSON(w, http.StatusOK, cpuResponse("cpu-parallel", iterations, passes, result, start))
}

func cpuResponse(endpoint string, iterationsPerPass int64, passes int, result int64, start time.Time) map[string]any {
	return map[string]any{
		"endpoint":          endpoint,
		"runtime":           "go",
		"iterationsPerPass": iterationsPerPass,
		"passes":            passes,
		"totalIterations":   iterationsPerPass * int64(passes),
		"result":            result,
		"durationMs":        time.Since(start).Milliseconds(),
	}
}

func parseIntQuery(w http.ResponseWriter, r *http.Request, name string, def, min, max int64) (int64, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def, true
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Validation failed", fmt.Sprintf("%s must be an integer", name))
		return 0, false
	}
	if v < min || v > max {
		httpx.WriteError(w, http.StatusBadRequest, "Validation failed",
			fmt.Sprintf("%s must be between %d and %d", name, min, max))
		return 0, false
	}
	return v, true
}

func parseCPUParams(w http.ResponseWriter, r *http.Request) (int64, int, bool) {
	iterations, ok := parseIntQuery(w, r, "iterations", defaultCPUIterationsPerPass, 1, maxCPUIterationsPerPass)
	if !ok {
		return 0, 0, false
	}
	passes, ok := parseIntQuery(w, r, "passes", defaultCPUPasses, 1, maxCPUPasses)
	if !ok {
		return 0, 0, false
	}
	return iterations, int(passes), true
}
