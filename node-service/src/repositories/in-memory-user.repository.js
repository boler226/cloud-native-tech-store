// Реалізація контракту UserRepository поверх InMemoryDataStore.
const { UserRepository } = require('./user.repository');
const { User } = require('../models/user');
const { emailTaken } = require('../errors/app-errors');

class InMemoryUserRepository extends UserRepository {
  #store;

  /**
   * @param {InMemoryDataStore} store
   * @param {{ seed?: Array<{name: string, email: string}> }} [options] початкові дані
   */
  constructor(store, { seed = [] } = {}) {
    super();
    this.#store = store;
    for (const { name, email } of seed) {
      const id = store.nextId();
      store.set(id, { id, name, email });
    }
  }

  async create(user) {
    this.#assertEmailFree(user.email);
    const id = this.#store.nextId();
    const record = { id, name: user.name, email: user.email };
    this.#store.set(id, record);
    return new User(record);
  }

  async findById(id) {
    const record = this.#store.get(id);
    return record ? new User(record) : null;
  }

  async findByEmail(email) {
    const wanted = email.toLowerCase();
    const record = this.#store.values().find((r) => r.email.toLowerCase() === wanted);
    return record ? new User(record) : null;
  }

  async findAll() {
    return this.#store.values().map((r) => new User(r));
  }

  async update(user) {
    if (!this.#store.get(user.id)) return null;
    this.#assertEmailFree(user.email, user.id);
    const record = { id: user.id, name: user.name, email: user.email };
    this.#store.set(user.id, record);
    return new User(record);
  }

  async delete(id) {
    return this.#store.delete(id);
  }

  // Аналог UNIQUE-індексу в БД: перевірка й запис відбуваються синхронно, без await між ними,
  // тому два одночасні запити з однаковим email не можуть обидва пройти.
  #assertEmailFree(email, exceptId) {
    const wanted = email.toLowerCase();
    const clash = this.#store.values().find((r) => r.email.toLowerCase() === wanted && r.id !== exceptId);
    if (clash) throw emailTaken(email);
  }
}

module.exports = InMemoryUserRepository;
