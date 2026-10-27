// Спільні утиліти для тестів. Файл не закінчується на .test.js, тож node --test його не запускає.
const createApp = require('../src/app');
const InMemoryDataStore = require('../src/datastore/in-memory.datastore');
const InMemoryUserRepository = require('../src/repositories/in-memory-user.repository');
const { UserRepository } = require('../src/repositories/user.repository');

const newInMemoryRepository = (seed = []) => new InMemoryUserRepository(new InMemoryDataStore(), { seed });

/**
 * Друга, зовсім інша реалізація контракту UserRepository (проста, без DataStore).
 * Нею доводимо, що UserService працює з будь-якою реалізацією без змін.
 */
class ArrayUserRepository extends UserRepository {
  constructor() {
    super();
    this.rows = [];
    this.seq = 0;
  }

  async create(user) {
    const row = { id: ++this.seq, name: user.name, email: user.email };
    this.rows.push(row);
    return { ...row };
  }

  async findById(id) {
    const row = this.rows.find((r) => r.id === id);
    return row ? { ...row } : null;
  }

  async findByEmail(email) {
    const row = this.rows.find((r) => r.email.toLowerCase() === email.toLowerCase());
    return row ? { ...row } : null;
  }

  async findAll() {
    return this.rows.map((r) => ({ ...r }));
  }

  async update(user) {
    const i = this.rows.findIndex((r) => r.id === user.id);
    if (i === -1) return null;
    this.rows[i] = { id: user.id, name: user.name, email: user.email };
    return { ...this.rows[i] };
  }

  async delete(id) {
    const before = this.rows.length;
    this.rows = this.rows.filter((r) => r.id !== id);
    return this.rows.length < before;
  }
}

/** Репозиторій, який завжди падає — для перевірки обробки непередбачених помилок (500). */
class BrokenUserRepository extends UserRepository {
  async findAll() { throw new Error('db connection lost: secret-host:5432'); }
  async findById() { throw new Error('db connection lost: secret-host:5432'); }
  async findByEmail() { throw new Error('db connection lost: secret-host:5432'); }
  async create() { throw new Error('db connection lost: secret-host:5432'); }
  async update() { throw new Error('db connection lost: secret-host:5432'); }
  async delete() { throw new Error('db connection lost: secret-host:5432'); }
}

/** Піднімає застосунок на вільному порту й повертає зручний HTTP-клієнт. */
async function startApp({ userRepository = newInMemoryRepository(), logger = () => {} } = {}) {
  const app = createApp({ userRepository, logger });
  const server = await new Promise((resolve, reject) => {
    const s = app.listen(0, '127.0.0.1', (err) => (err ? reject(err) : resolve(s)));
  });
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  async function request(method, path, { json, body, headers } = {}) {
    const init = { method, headers: { ...headers } };
    if (json !== undefined) {
      init.body = JSON.stringify(json);
      init.headers['Content-Type'] = 'application/json';
    } else if (body !== undefined) {
      init.body = body;
    }
    const res = await fetch(baseUrl + path, init);
    const text = await res.text();
    let data = null;
    try { data = text ? JSON.parse(text) : null; } catch { /* не JSON */ }
    return { status: res.status, headers: res.headers, body: data, text };
  }

  const close = () =>
    new Promise((resolve) => {
      server.close(() => resolve());
      server.closeAllConnections();
    });

  return { baseUrl, request, close };
}

module.exports = {
  newInMemoryRepository,
  ArrayUserRepository,
  BrokenUserRepository,
  startApp,
};
