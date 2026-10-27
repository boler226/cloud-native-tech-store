// Middleware №2: спільна валідація вхідних даних для кількох endpoint-ів.
// Некоректний запит зупиняється ТУТ і ніколи не доходить до controller/service/repository.
const { validateUserInput } = require('../models/user');
const { ValidationError, UnsupportedMediaTypeError } = require('../errors/app-errors');

// Перевіряє :id у шляху і кладе число в res.locals.id.
function validateIdParam(req, res, next) {
  const raw = req.params.id;
  if (!/^[1-9]\d*$/.test(raw) || !Number.isSafeInteger(Number(raw))) {
    throw new ValidationError('invalid_id', 'Invalid identifier: must be a positive integer');
  }
  res.locals.id = Number(raw);
  next();
}

// Перевіряє тіло POST/PUT: Content-Type, що це JSON-об'єкт, обов'язкові поля, типи, порожні значення.
// Результат кладе в res.locals.input.
function validateUserBody(req, res, next) {
  // req.is(): false — тіло є, але не JSON; null — тіла немає; рядок — тип збігся.
  if (req.is('application/json') === false) {
    throw new UnsupportedMediaTypeError('unsupported_media_type', 'Content-Type must be application/json');
  }
  const errors = validateUserInput(req.body);
  if (errors.length > 0) {
    throw new ValidationError('validation_failed', 'Request validation failed', errors);
  }
  res.locals.input = req.body;
  next();
}

module.exports = { validateIdParam, validateUserBody };
