const HttpError = require('../errors/HttpError');

// Читає query-параметр як ціле число з дефолтом і межами. Кидає 400, якщо значення некоректне.
function parseIntParam(raw, { name, def, min, max }) {
  if (raw === undefined || raw === '') return def;
  if (!/^-?\d+$/.test(String(raw))) {
    throw new HttpError(400, 'Validation failed', [`${name} must be an integer`]);
  }
  const n = Number(raw);
  if (!Number.isSafeInteger(n)) {
    throw new HttpError(400, 'Validation failed', [`${name} is out of safe integer range`]);
  }
  if (min !== undefined && n < min) {
    throw new HttpError(400, 'Validation failed', [`${name} must be >= ${min}`]);
  }
  if (max !== undefined && n > max) {
    throw new HttpError(400, 'Validation failed', [`${name} must be <= ${max}`]);
  }
  return n;
}

module.exports = parseIntParam;
