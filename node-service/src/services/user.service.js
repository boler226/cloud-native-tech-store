// Бізнес-логіка користувачів. НЕ залежить від HTTP/Express і від конкретного сховища:
// працює лише з контрактом UserRepository, який отримує через конструктор.
const { assertUserRepository } = require('../repositories/user.repository');
const { User, validateUserInput, normalizeUserInput } = require('../models/user');
const { ValidationError, userNotFound, emailTaken } = require('../errors/app-errors');

class UserService {
  #repository;

  constructor(repository) {
    this.#repository = assertUserRepository(repository);
  }

  async list() {
    return this.#repository.findAll();
  }

  async getById(id) {
    this.#assertId(id);
    const user = await this.#repository.findById(id);
    if (!user) throw userNotFound(id);
    return user;
  }

  async create(input) {
    const data = this.#prepare(input);
    // Бізнес-правило: email унікальний (без урахування регістру).
    if (await this.#repository.findByEmail(data.email)) throw emailTaken(data.email);
    return this.#repository.create(new User({ ...data }));
  }

  async update(id, input) {
    this.#assertId(id);
    const data = this.#prepare(input);

    if (!(await this.#repository.findById(id))) throw userNotFound(id);

    // Змінювати email на власний можна; на чужий — ні.
    const owner = await this.#repository.findByEmail(data.email);
    if (owner && owner.id !== id) throw emailTaken(data.email);

    const updated = await this.#repository.update(new User({ id, ...data }));
    if (!updated) throw userNotFound(id); // видалили між перевіркою та оновленням
    return updated;
  }

  async delete(id) {
    this.#assertId(id);
    if (!(await this.#repository.delete(id))) throw userNotFound(id);
  }

  // Сервіс захищає свої інваріанти й тоді, коли його викликають не через HTTP.
  #prepare(input) {
    const errors = validateUserInput(input);
    if (errors.length > 0) {
      throw new ValidationError('validation_failed', 'Request validation failed', errors);
    }
    return normalizeUserInput(input);
  }

  #assertId(id) {
    if (!Number.isSafeInteger(id) || id < 1) {
      throw new ValidationError('invalid_id', 'Invalid identifier: must be a positive integer');
    }
  }
}

module.exports = UserService;
