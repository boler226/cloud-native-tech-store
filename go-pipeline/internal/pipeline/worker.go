package pipeline

import (
	"context"
	"sync"
)

// worker читає задачі з jobs, обробляє їх і відправляє результат у results.
// Завершується, коли:
//   - канал jobs закрито й вичерпано (нормальне завершення), або
//   - контекст скасовано (ctx.Done()).
//
// Обидві блокуючі операції (читання задачі та запис результату) загорнуті в select
// з ctx.Done(), тому воркер ніколи не "застряє" назавжди, якщо Aggregator чи
// Producer вже зупинились.
func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result, work int) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return // jobs закрито: задач більше немає
			}
			res := Result{JobID: job.ID, Value: Compute(job.Value, work)}

			select {
			case results <- res:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return // скасування: припиняємо обробку
		}
	}
}
