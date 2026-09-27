// Помилка з HTTP-статусом. Її ловить errorHandler і перетворює на JSON-відповідь.
class HttpError extends Error {
  constructor(status, message, details) {
    super(message);
    this.name = 'HttpError';
    this.status = status;
    this.details = details;
  }
}

module.exports = HttpError;
