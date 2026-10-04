// Package bench запускає серію pipeline-запусків з різними комбінаціями
// параметрів і формує порівняльну таблицю.
package bench

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"pipeline-go/internal/pipeline"
)

// Options — матриця параметрів для серії запусків.
type Options struct {
	TasksList   []int
	WorkersList []int
	BuffersList []int
	Work        int
	Repeat      int // скільки разів повторювати кожну комбінацію (береться медіана)
}

// Row — один рядок підсумкової таблиці.
type Row struct {
	Tasks, Workers, Buffer int
	TimeMs                 float64 // медіана по Repeat запусках
	Speedup                float64 // відносно 1 worker з тим самим Tasks/Buffer
	HasSpeedup             bool
	AllOK                  bool // усі запуски пройшли перевірку коректності
}

type groupKey struct{ tasks, buffer int }

// Run виконує всі комбінації з opts і повертає рядки таблиці.
// Порядок: Tasks -> Buffer -> Workers (так таблиця читається зручніше).
func Run(ctx context.Context, opts Options, progress io.Writer) ([]Row, error) {
	if opts.Repeat < 1 {
		opts.Repeat = 1
	}

	// Прогрів: перший запуск у процесі завжди повільніший (планувальник, кеші, GC).
	if _, err := pipeline.Run(ctx, pipeline.Config{Tasks: 1000, Workers: 2, Work: opts.Work}); err != nil {
		return nil, err
	}

	var rows []Row
	for _, tasks := range opts.TasksList {
		for _, buffer := range opts.BuffersList {
			for _, workers := range opts.WorkersList {
				if progress != nil {
					fmt.Fprintf(progress, "running tasks=%d workers=%d buffer=%d (x%d)...\n",
						tasks, workers, buffer, opts.Repeat)
				}

				cfg := pipeline.Config{Tasks: tasks, Workers: workers, Buffer: buffer, Work: opts.Work, Verify: true}
				times := make([]float64, 0, opts.Repeat)
				allOK := true

				for i := 0; i < opts.Repeat; i++ {
					stats, err := pipeline.Run(ctx, cfg)
					if err != nil {
						return nil, err
					}
					if stats.Cancelled {
						return rows, ctx.Err()
					}
					allOK = allOK && stats.OK()
					times = append(times, float64(stats.Elapsed.Microseconds())/1000)
				}

				rows = append(rows, Row{
					Tasks: tasks, Workers: workers, Buffer: buffer,
					TimeMs: median(times), AllOK: allOK,
				})
			}
		}
	}

	fillSpeedup(rows)
	return rows, nil
}

func median(xs []float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// fillSpeedup рахує прискорення відносно запуску з 1 worker (якщо він є в матриці).
func fillSpeedup(rows []Row) {
	base := map[groupKey]float64{}
	for _, r := range rows {
		if r.Workers == 1 {
			base[groupKey{r.Tasks, r.Buffer}] = r.TimeMs
		}
	}
	for i := range rows {
		if b, ok := base[groupKey{rows[i].Tasks, rows[i].Buffer}]; ok && rows[i].TimeMs > 0 {
			rows[i].Speedup = b / rows[i].TimeMs
			rows[i].HasSpeedup = true
		}
	}
}

// Markdown форматує рядки у Markdown-таблицю.
func Markdown(rows []Row) string {
	var b strings.Builder
	b.WriteString("| Tasks | Workers | Buffer | Time, ms | Speedup |\n")
	b.WriteString("|---|---|---|---|---|\n")

	failed := false
	for _, r := range rows {
		speed := "—"
		if r.HasSpeedup {
			speed = fmt.Sprintf("%.2fx", r.Speedup)
		}
		mark := ""
		if !r.AllOK {
			mark = " ⚠"
			failed = true
		}
		fmt.Fprintf(&b, "| %d | %d | %d | %.2f%s | %s |\n", r.Tasks, r.Workers, r.Buffer, r.TimeMs, mark, speed)
	}
	if failed {
		b.WriteString("\n⚠ — для цього рядка перевірка коректності (сума/дублі) НЕ пройшла.\n")
	}
	return b.String()
}
