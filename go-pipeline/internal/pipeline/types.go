// Package pipeline реалізує конкурентний pipeline обробки числових задач:
//
//	Producer -> jobs -> Worker Pool -> results -> Aggregator -> Output
package pipeline

import (
	"errors"
	"fmt"
	"time"
)

// Job — вхідна задача.
type Job struct {
	ID    int
	Value int
}

// Result — результат обробки задачі.
type Result struct {
	JobID int
	Value int
}

// Config — параметри одного запуску pipeline.
type Config struct {
	Tasks   int  // кількість задач, які створює Producer (100, 1000, 10000, 100000...)
	Workers int  // розмір пулу воркерів (1, 2, 4, 8, 16...)
	Buffer  int  // ємність каналів jobs і results (0 = unbuffered, 10, 100, 1000...)
	Work    int  // "складність" обчислення однієї задачі (0 = просто value*value)
	Verify  bool // чи звіряти підсумок із послідовним еталонним обчисленням
}

// Validate перевіряє параметри конфігурації.
func (c Config) Validate() error {
	var errs []error
	if c.Tasks < 1 {
		errs = append(errs, fmt.Errorf("tasks must be >= 1 (got %d)", c.Tasks))
	}
	if c.Workers < 1 {
		errs = append(errs, fmt.Errorf("workers must be >= 1 (got %d)", c.Workers))
	}
	if c.Buffer < 0 {
		errs = append(errs, fmt.Errorf("buffer must be >= 0 (got %d)", c.Buffer))
	}
	if c.Work < 0 {
		errs = append(errs, fmt.Errorf("work must be >= 0 (got %d)", c.Work))
	}
	return errors.Join(errs...)
}

// Stats — підсумкова статистика, яку формує Aggregator.
type Stats struct {
	Config Config

	Completed  int           // скільки результатів отримано
	Total      int64         // сума всіх отриманих значень Result.Value
	Duplicates int           // скільки разів JobID приходив повторно (має бути 0)
	Invalid    int           // результати з JobID поза діапазоном (має бути 0)
	Elapsed    time.Duration // від старту Producer до отримання останнього результату
	FinishedAt time.Time     // момент завершення обробки
	Cancelled  bool          // обробку перервано через context

	Verified bool  // чи виконувалась звірка з еталоном
	Expected int64 // еталонна сума (якщо Verified)
}

// OK повідомляє, чи всі задачі оброблено рівно по одному разу й підсумок збігається з еталоном.
func (s Stats) OK() bool {
	if s.Cancelled || s.Completed != s.Config.Tasks || s.Duplicates != 0 || s.Invalid != 0 {
		return false
	}
	return !s.Verified || s.Total == s.Expected
}

// Throughput — кількість оброблених задач за секунду.
func (s Stats) Throughput() float64 {
	if s.Elapsed <= 0 {
		return 0
	}
	return float64(s.Completed) / s.Elapsed.Seconds()
}
