// Ендпойнти для лабораторної 2: дослідження моделей конкурентності Node.js.
const { wait } = require('../services/io.service');
const { busyLoop } = require('../services/cpu.service');
const parseIntParam = require('../utils/parseIntParam');

// Підібрано так, щоб один запит тривав ~0.3-0.5с у типовому середовищі;
// під своє залізо підберіть власні значення через query-параметри.
const DEFAULT_IO_DELAY_MS = 200;
const MAX_IO_DELAY_MS = 30_000;
const DEFAULT_CPU_ITERATIONS = 500_000_000;
const MAX_CPU_ITERATIONS = 20_000_000_000;

class ConcurrencyController {
  // GET /io?delayMs=200 -> 200
  // Асинхронне очікування: поки триває, event loop вільний і обробляє інші запити.
  io = async (req, res) => {
    const delayMs = parseIntParam(req.query.delayMs, {
      name: 'delayMs', def: DEFAULT_IO_DELAY_MS, min: 0, max: MAX_IO_DELAY_MS,
    });
    const start = performance.now();
    await wait(delayMs);
    res.status(200).json({
      endpoint: 'io',
      runtime: 'node',
      delayMs,
      actualDurationMs: Math.round(performance.now() - start),
    });
  };

  // GET /cpu?iterations=N -> 200
  // Синхронний важкий цикл: під час виконання Node НЕ може обробляти жодні
  // інші запити (весь event loop блокується одним потоком).
  cpu = (req, res) => {
    const iterations = parseIntParam(req.query.iterations, {
      name: 'iterations', def: DEFAULT_CPU_ITERATIONS, min: 1, max: MAX_CPU_ITERATIONS,
    });
    const start = performance.now();
    const result = busyLoop(iterations);
    res.status(200).json({
      endpoint: 'cpu',
      runtime: 'node',
      iterations,
      result,
      durationMs: Math.round(performance.now() - start),
    });
  };
}

module.exports = ConcurrencyController;
