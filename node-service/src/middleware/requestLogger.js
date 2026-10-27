// Middleware №1: логування запитів. Формат: "GET /api/users/10 200 4.2ms"
// (метод, шлях, статус-код, тривалість). Фабрика приймає функцію виводу — у тестах її підміняють.
module.exports = function createRequestLogger(log = console.log) {
  return (req, res, next) => {
    const start = process.hrtime.bigint();
    res.on('finish', () => {
      const ms = Number(process.hrtime.bigint() - start) / 1e6;
      log(`${req.method} ${req.originalUrl} ${res.statusCode} ${ms.toFixed(1)}ms`);
    });
    next();
  };
};
