package pipeline

import (
	"context"
	"sync"
	"time"
)

// Run виконує весь pipeline:
//
//	Producer -> jobs -> Worker Pool (cfg.Workers) -> results -> Aggregator
//
// Канали jobs та results мають ємність cfg.Buffer (0 = unbuffered).
// Скасування parent-контексту (Ctrl+C, timeout) коректно зупиняє всі горутини.
func Run(parent context.Context, cfg Config) (Stats, error) {
	if err := cfg.Validate(); err != nil {
		return Stats{}, err
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	jobs := make(chan Job, cfg.Buffer)
	results := make(chan Result, cfg.Buffer)

	start := time.Now()

	// Producer
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		producer(ctx, cfg.Tasks, jobs)
	}()

	// Worker pool
	var wg sync.WaitGroup
	wg.Add(cfg.Workers)
	for i := 0; i < cfg.Workers; i++ {
		go worker(ctx, &wg, jobs, results, cfg.Work)
	}

	// results закриваємо лише коли ВСІ воркери завершились.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregator (працює в поточній горутині, тож Run повертається лише після нього).
	agg := aggregate(results, cfg.Tasks)
	<-producerDone

	stats := Stats{
		Config:     cfg,
		Completed:  agg.completed,
		Total:      agg.total,
		Duplicates: agg.duplicates,
		Invalid:    agg.invalid,
		Elapsed:    agg.finishedAt.Sub(start),
		FinishedAt: agg.finishedAt,
		Cancelled:  agg.completed < cfg.Tasks && parent.Err() != nil,
	}

	// Еталон рахується ПІСЛЯ зупинки таймера і не входить у Elapsed.
	if cfg.Verify && !stats.Cancelled {
		stats.Verified = true
		stats.Expected = Expected(cfg.Tasks, cfg.Work)
	}

	return stats, nil
}
