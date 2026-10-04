package pipeline

import "time"

type aggregation struct {
	completed  int
	total      int64
	duplicates int
	invalid    int
	finishedAt time.Time
}

// aggregate отримує результати з results, доки канал не буде закрито:
// рахує завершені задачі, накопичує загальний результат, виявляє дублі
// та фіксує момент завершення обробки.
//
// Канал results закривається окремою горутиною після wg.Wait() — тобто коли
// всі воркери завершились, тож range тут завершується гарантовано.
func aggregate(results <-chan Result, tasks int) aggregation {
	var a aggregation
	seen := make([]bool, tasks+1) // seen[JobID] — чи вже отримували результат цієї задачі

	for r := range results {
		a.completed++
		a.total += int64(r.Value)

		if r.JobID < 1 || r.JobID > tasks {
			a.invalid++
			continue
		}
		if seen[r.JobID] {
			a.duplicates++
			continue
		}
		seen[r.JobID] = true
	}

	a.finishedAt = time.Now()
	return a
}
