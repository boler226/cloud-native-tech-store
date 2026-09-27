const HttpError = require('../errors/HttpError');

// Перетворює параметр шляху на додатне ціле число, інакше 400 Bad Request.
function parseId(raw) {
  const id = Number(raw);
  if (!/^[1-9]\d*$/.test(String(raw)) || !Number.isSafeInteger(id)) {
    throw new HttpError(400, 'Invalid product id: must be a positive integer');
  }
  return id;
}

module.exports = parseId;
