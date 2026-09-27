// Package service (concurrency) — операції для лабораторної 2: дослідження
// моделей конкурентності Go порівняно з Node.js.
package service

import (
	"sync"
	"time"
)

type ConcurrencyService struct{}

func NewConcurrencyService() *ConcurrencyService {
	return &ConcurrencyService{}
}

// Wait імітує асинхронну I/O-операцію. time.Sleep паркує ГОРУТИНУ в рантаймі
// Go: поки вона "спить", її OS-потік вільний і може виконувати інші горутини
// (це concurrency-модель Go: багато горутин, кожна може незалежно блокуватись).
func (s *ConcurrencyService) Wait(delay time.Duration) {
	time.Sleep(delay)
}

// busyLoop — CPU-звантажена робота: інкремент лічильника iterations разів.
// Проста лічилка обрана навмисно, щоб виключити побічні ефекти (I/O, алокації).
func busyLoop(iterations int64) int64 {
	var counter int64
	for i := int64(0); i < iterations; i++ {
		counter++
	}
	return counter
}

// CPUSequential виконує passes важких проходів ОДИН ЗА ОДНИМ у поточній
// горутині (аналог того, як Node виконав би той самий обсяг роботи).
func (s *ConcurrencyService) CPUSequential(iterationsPerPass int64, passes int) int64 {
	var total int64
	for p := 0; p < passes; p++ {
		total += busyLoop(iterationsPerPass)
	}
	return total
}

// CPUParallel виконує ТОЙ САМИЙ сумарний обсяг роботи, що й CPUSequential,
// але кожен прохід запускається в окремій горутині. На системі з декількома
// ядрами (GOMAXPROCS > 1) горутини реально виконуються паралельно і сумарний
// час — ближче до часу ОДНОГО проходу, а не суми всіх.
func (s *ConcurrencyService) CPUParallel(iterationsPerPass int64, passes int) int64 {
	results := make([]int64, passes)

	var wg sync.WaitGroup
	wg.Add(passes)
	for p := 0; p < passes; p++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = busyLoop(iterationsPerPass)
		}(p)
	}
	wg.Wait()

	var total int64
	for _, r := range results {
		total += r
	}
	return total
}
