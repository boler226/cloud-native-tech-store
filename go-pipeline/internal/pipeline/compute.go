package pipeline

import "sync"

const (
	mask = 0x7fffffff
	mulA = 1103515245
	incC = 12345
)

// Compute — обчислення однієї задачі.
// work == 0: result = value*value.
// work  > 0: після піднесення до квадрата виконується work ітерацій лінійного
// конгруентного генератора — це імітує "важчу" CPU-роботу, щоб накладні витрати
// на канали не домінували над самим обчисленням. Результат детермінований.
func Compute(value, work int) int {
	r := value * value
	for i := 0; i < work; i++ {
		r = ((r&mask)*mulA + incC) & mask
	}
	return r
}

type expectedKey struct{ tasks, work int }

var expectedCache sync.Map // expectedKey -> int64

// Expected послідовно (без горутин) рахує еталонну суму результатів для tasks задач
// (Job.ID = 1..tasks, Job.Value = ID). Результат кешується, бо при серії запусків
// з однаковими tasks/work він не змінюється.
func Expected(tasks, work int) int64 {
	key := expectedKey{tasks, work}
	if v, ok := expectedCache.Load(key); ok {
		return v.(int64)
	}
	var sum int64
	for i := 1; i <= tasks; i++ {
		sum += int64(Compute(i, work))
	}
	expectedCache.Store(key, sum)
	return sum
}
