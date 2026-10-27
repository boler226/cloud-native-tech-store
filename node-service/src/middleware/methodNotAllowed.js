const { MethodNotAllowedError } = require('../errors/app-errors');

// 405 Method Not Allowed. Заголовок Allow виставляє центральний errorHandler.
module.exports = (allowed) => (req, res, next) => {
  next(new MethodNotAllowedError(allowed));
};
