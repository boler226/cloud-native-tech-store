// ЄДИНЕ місце, де помилки перетворюються на HTTP-відповідь.
// Повертає { status, body, headers } — чиста функція, легко тестується.
const HttpError = require('./HttpError');
const {
  AppError,
  ValidationError,
  NotFoundError,
  ConflictError,
  UnsupportedMediaTypeError,
  MethodNotAllowedError,
} = require('./app-errors');

const CODE_BY_STATUS = {
  400: 'bad_request',
  404: 'not_found',
  405: 'method_not_allowed',
  409: 'conflict',
  413: 'payload_too_large',
  415: 'unsupported_media_type',
  500: 'internal_error',
};

const codeForStatus = (status) => CODE_BY_STATUS[status] ?? (status >= 500 ? 'internal_error' : 'bad_request');

function response(status, code, message, details, headers) {
  const body = { error: code, message };
  if (details !== undefined) body.details = details;
  return { status, body, headers: headers ?? {} };
}

function toErrorResponse(err) {
  // 1) Типізовані прикладні помилки (нова архітектура)
  if (err instanceof MethodNotAllowedError) {
    return response(405, err.code, err.message, undefined, { Allow: err.allowed.join(', ') });
  }
  if (err instanceof ValidationError) return response(400, err.code, err.message, err.details);
  if (err instanceof NotFoundError) return response(404, err.code, err.message, err.details);
  if (err instanceof ConflictError) return response(409, err.code, err.message, err.details);
  if (err instanceof UnsupportedMediaTypeError) return response(415, err.code, err.message, err.details);
  if (err instanceof AppError) return response(500, 'internal_error', 'Internal server error');

  // 2) Помилки з лабораторної 1 (products) — приводимо до того ж формату
  if (err instanceof HttpError) {
    const code = err.status === 400 && err.details ? 'validation_failed' : codeForStatus(err.status);
    return response(err.status, code, err.message, err.details);
  }

  // 3) Помилки розбору тіла запиту (body-parser, який використовує express.json)
  if (err && err.type === 'entity.parse.failed') {
    return response(400, 'invalid_json', 'Request body contains invalid JSON');
  }
  if (err && err.type === 'entity.too.large') {
    return response(413, 'payload_too_large', 'Request body is too large');
  }
  if (err && Number.isInteger(err.status) && err.status >= 400 && err.status < 500) {
    return response(err.status, codeForStatus(err.status), 'Request could not be processed');
  }

  // 4) Усе інше — непередбачена помилка. Деталі НЕ віддаємо клієнту.
  return response(500, 'internal_error', 'Internal server error');
}

module.exports = { toErrorResponse, codeForStatus };
