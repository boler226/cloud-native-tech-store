// Прикладні (доменні) помилки. Вони НЕ знають про HTTP: жодних статус-кодів тут немає.
// Перетворення "тип помилки -> HTTP-статус" відбувається в одному місці — errors/error-mapper.js.
class AppError extends Error {
  /**
   * @param {string} code      машинозчитуваний код, напр. 'user_not_found'
   * @param {string} message   повідомлення для людини
   * @param {Array}  [details] додаткові деталі (напр. помилки полів)
   */
  constructor(code, message, details) {
    super(message);
    this.name = this.constructor.name;
    this.code = code;
    if (details !== undefined) this.details = details;
  }
}

class ValidationError extends AppError {}      // некоректні вхідні дані  -> 400
class NotFoundError extends AppError {}        // сутність не знайдена    -> 404
class ConflictError extends AppError {}        // порушення унікальності  -> 409
class UnsupportedMediaTypeError extends AppError {} // не JSON            -> 415

class MethodNotAllowedError extends AppError {
  constructor(allowed) {
    super('method_not_allowed', 'Method not allowed');
    this.allowed = allowed;
  }
}

// Фабрики, щоб повідомлення про типові помилки були в одному місці.
const userNotFound = (id) => new NotFoundError('user_not_found', `User with id ${id} was not found`);
const emailTaken = (email) =>
  new ConflictError('email_already_exists', `User with email ${email} already exists`);

module.exports = {
  AppError,
  ValidationError,
  NotFoundError,
  ConflictError,
  UnsupportedMediaTypeError,
  MethodNotAllowedError,
  userNotFound,
  emailTaken,
};
