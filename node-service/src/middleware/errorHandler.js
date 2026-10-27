const { toErrorResponse } = require('../errors/error-mapper');

// Єдина точка обробки помилок усього застосунку. Має 4 аргументи — так Express розуміє,
// що це error-middleware. Усі помилки (next(err), throw у middleware, відхилені промізи
// в async-контролерах) потрапляють сюди й перетворюються на { error, message, details? }.
// eslint-disable-next-line no-unused-vars
module.exports = (err, req, res, next) => {
  const { status, body, headers } = toErrorResponse(err);

  if (status >= 500) console.error(err); // деталі лише в лог, клієнту — загальне повідомлення

  res.set(headers);
  res.status(status).json(body);
};
