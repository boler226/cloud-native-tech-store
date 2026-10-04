// Лабораторна робота 3. Конкурентний pipeline обробки даних у Go.
//
//	Producer -> jobs -> Worker Pool -> results -> Aggregator -> Output
//
// Приклади:
//
//	go run main.go -tasks=10000 -workers=4 -buffer=100
//	go run -race main.go -tasks=10000 -workers=8 -buffer=0
//	go run main.go -bench -tasks-list=10000 -repeat=5
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"pipeline-go/internal/bench"
	"pipeline-go/internal/pipeline"
)

func main() {
	var (
		tasks   = flag.Int("tasks", 10000, "кількість задач (напр. 100, 1000, 10000, 100000)")
		workers = flag.Int("workers", 4, "кількість воркерів (напр. 1, 2, 4, 8, 16)")
		buffer  = flag.Int("buffer", 0, "ємність каналів jobs/results (0 = unbuffered, 10, 100, 1000)")
		work    = flag.Int("work", 1000, "складність обчислення однієї задачі (0 = лише value*value)")
		timeout = flag.Duration("timeout", 0, "скасувати обробку через цей час, напр. 50ms (0 = без ліміту)")
		verify  = flag.Bool("verify", true, "звірити підсумок із послідовним еталонним обчисленням")

		benchMode   = flag.Bool("bench", false, "режим серії запусків: перебір усіх комбінацій і таблиця результатів")
		tasksList   = flag.String("tasks-list", "10000", "[bench] список Tasks через кому")
		workersList = flag.String("workers-list", "1,2,4,8,16", "[bench] список Workers через кому")
		buffersList = flag.String("buffers-list", "0,10,100,1000", "[bench] список Buffer через кому")
		repeat      = flag.Int("repeat", 5, "[bench] скільки разів повторювати кожну комбінацію (береться медіана)")
	)
	flag.Parse()

	// Ctrl+C скасовує контекст -> усі горутини pipeline завершуються коректно.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	fmt.Printf("%s, CPU cores: %d, GOMAXPROCS: %d\n\n", runtime.Version(), runtime.NumCPU(), runtime.GOMAXPROCS(0))

	if *benchMode {
		os.Exit(runBench(ctx, *tasksList, *workersList, *buffersList, *work, *repeat))
	}
	os.Exit(runSingle(ctx, pipeline.Config{
		Tasks: *tasks, Workers: *workers, Buffer: *buffer, Work: *work, Verify: *verify,
	}))
}

func runSingle(ctx context.Context, cfg pipeline.Config) int {
	stats, err := pipeline.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid config:", err)
		return 2
	}

	fmt.Printf("Pipeline: tasks=%d workers=%d buffer=%d work=%d\n", cfg.Tasks, cfg.Workers, cfg.Buffer, cfg.Work)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Completed:    %d / %d\n", stats.Completed, cfg.Tasks)
	fmt.Printf("Total result: %d\n", stats.Total)
	if stats.Verified {
		verdict := "OK"
		if stats.Total != stats.Expected {
			verdict = "MISMATCH"
		}
		fmt.Printf("Expected:     %d (%s)\n", stats.Expected, verdict)
	}
	fmt.Printf("Duplicates:   %d, invalid: %d\n", stats.Duplicates, stats.Invalid)
	fmt.Printf("Elapsed:      %.2f ms\n", float64(stats.Elapsed.Microseconds())/1000)
	fmt.Printf("Throughput:   %.0f jobs/s\n", stats.Throughput())
	fmt.Printf("Finished at:  %s\n", stats.FinishedAt.Format(time.TimeOnly+".000"))

	switch {
	case stats.Cancelled:
		fmt.Println("Status:       CANCELLED (context done) — оброблено лише частину задач")
		return 1
	case !stats.OK():
		fmt.Println("Status:       FAILED (результат не збігається з еталоном)")
		return 1
	}
	fmt.Println("Status:       OK")
	return 0
}

func runBench(ctx context.Context, tasksList, workersList, buffersList string, work, repeat int) int {
	tl, err1 := parseInts(tasksList)
	wl, err2 := parseInts(workersList)
	bl, err3 := parseInts(buffersList)
	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Fprintln(os.Stderr, "invalid list:", firstErr(err1, err2, err3))
		return 2
	}

	fmt.Printf("Bench: tasks=%v workers=%v buffers=%v work=%d repeat=%d (median)\n\n", tl, wl, bl, work, repeat)

	rows, err := bench.Run(ctx, bench.Options{
		TasksList: tl, WorkersList: wl, BuffersList: bl, Work: work, Repeat: repeat,
	}, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bench interrupted:", err)
		return 1
	}

	fmt.Println()
	fmt.Print(bench.Markdown(rows))
	return 0
}

func parseInts(s string) ([]int, error) {
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("%q is not a non-negative integer", part)
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return out, nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
