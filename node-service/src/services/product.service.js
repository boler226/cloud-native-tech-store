// Бізнес-логіка: валідація, пошук, вибір HTTP-помилок. Про Express тут нічого не відомо.
const HttpError = require('../errors/HttpError');
const { validateProductInput } = require('../utils/validation');

class ProductService {
  constructor(repository) {
    this.repository = repository;
  }

  list(category) {
    return this.repository.findAll(category);
  }

  getById(id) {
    const product = this.repository.findById(id);
    if (!product) throw new HttpError(404, 'Product not found');
    return product;
  }

  create(body) {
    const { value, errors } = validateProductInput(body);
    if (errors.length > 0) throw new HttpError(400, 'Validation failed', errors);
    return this.repository.create(value);
  }

  update(id, body) {
    const { value, errors } = validateProductInput(body);
    if (errors.length > 0) throw new HttpError(400, 'Validation failed', errors);
    const updated = this.repository.update(id, value);
    if (!updated) throw new HttpError(404, 'Product not found');
    return updated;
  }

  remove(id) {
    if (!this.repository.delete(id)) throw new HttpError(404, 'Product not found');
  }
}

module.exports = ProductService;
