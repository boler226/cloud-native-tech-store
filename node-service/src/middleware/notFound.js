const { NotFoundError } = require('../errors/app-errors');

// 404 для маршрутів, яких не існує (відповідь формує центральний errorHandler).
module.exports = (req, res, next) => {
  next(new NotFoundError('route_not_found', `Route ${req.method} ${req.path} was not found`));
};
