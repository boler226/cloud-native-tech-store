package pipeline

import "context"

// producer створює tasks задач і передає їх у канал jobs.
// Після завершення генерації (або скасування контексту) канал jobs закривається —
// саме це сигналізує воркерам, що вхідний потік закінчився.
func producer(ctx context.Context, tasks int, jobs chan<- Job) {
	defer close(jobs)

	for i := 1; i <= tasks; i++ {
		job := Job{ID: i, Value: i}
		select {
		case jobs <- job:
		case <-ctx.Done():
			return
		}
	}
}
