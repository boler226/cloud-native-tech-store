package pipeline

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// Для work=0 сума value*value по 1..n має замкнений вигляд n(n+1)(2n+1)/6.
func sumOfSquares(n int) int64 {
	nn := int64(n)
	return nn * (nn + 1) * (2*nn + 1) / 6
}

func TestRunCorrectness(t *testing.T) {
	for _, work := range []int{0, 50} {
		for _, workers := range []int{1, 2, 4, 8, 16} {
			for _, buffer := range []int{0, 10, 100, 1000} {
				name := fmt.Sprintf("work=%d/workers=%d/buffer=%d", work, workers, buffer)
				t.Run(name, func(t *testing.T) {
					cfg := Config{Tasks: 2000, Workers: workers, Buffer: buffer, Work: work, Verify: true}
					stats, err := Run(context.Background(), cfg)
					if err != nil {
						t.Fatal(err)
					}
					if !stats.OK() {
						t.Fatalf("pipeline result is not OK: %+v", stats)
					}
					if stats.Completed != cfg.Tasks {
						t.Fatalf("completed = %d, want %d", stats.Completed, cfg.Tasks)
					}
					if work == 0 && stats.Total != sumOfSquares(cfg.Tasks) {
						t.Fatalf("total = %d, want %d", stats.Total, sumOfSquares(cfg.Tasks))
					}
				})
			}
		}
	}
}

func TestRunCancellation(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// Свідомо величезна задача — за 20мс вона не встигне завершитись.
	cfg := Config{Tasks: 50_000_000, Workers: 4, Buffer: 10, Work: 1000}
	stats, err := Run(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !stats.Cancelled {
		t.Fatalf("expected Cancelled=true, got %+v", stats)
	}
	if stats.Completed >= cfg.Tasks {
		t.Fatalf("expected partial completion, got %d", stats.Completed)
	}
	if stats.OK() {
		t.Fatal("cancelled run must not be reported as OK")
	}

	// Усі горутини pipeline мають завершитись (даємо рантайму трохи часу).
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Fatalf("goroutine leak: before=%d after=%d", before, after)
	}
}

func TestValidate(t *testing.T) {
	bad := []Config{
		{Tasks: 0, Workers: 1},
		{Tasks: 1, Workers: 0},
		{Tasks: 1, Workers: 1, Buffer: -1},
		{Tasks: 1, Workers: 1, Work: -1},
	}
	for _, cfg := range bad {
		if _, err := Run(context.Background(), cfg); err == nil {
			t.Errorf("expected validation error for %+v", cfg)
		}
	}
}

func TestComputeWorkZeroIsSquare(t *testing.T) {
	for _, v := range []int{0, 1, 7, 100, 99999} {
		if got := Compute(v, 0); got != v*v {
			t.Errorf("Compute(%d, 0) = %d, want %d", v, got, v*v)
		}
	}
}
