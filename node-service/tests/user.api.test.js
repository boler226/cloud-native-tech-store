// Тести рівня API: повний ланцюжок Router -> Middleware -> Controller -> Service -> Repository
// через реальні HTTP-запити до застосунку на вільному порту.
const { describe, it, before, after, beforeEach, afterEach } = require('node:test');
const assert = require('node:assert/strict');

const { startApp, newInMemoryRepository, BrokenUserRepository } = require('./helpers');

const john = { name: 'John Doe', email: 'john@example.com' };

describe('API /api/users', () => {
  let app;
  beforeEach(async () => { app = await startApp(); });
  afterEach(async () => { await app.close(); });

  describe('GET /health', () => {
    it('200 і status ok', async () => {
      const res = await app.request('GET', '/health');
      assert.equal(res.status, 200);
      assert.equal(res.body.status, 'ok');
    });
  });

  describe('GET /api/users', () => {
    it('200 і порожній масив, якщо користувачів немає', async () => {
      const res = await app.request('GET', '/api/users');
      assert.equal(res.status, 200);
      assert.deepEqual(res.body, []);
    });

    it('200 і список створених користувачів', async () => {
      await app.request('POST', '/api/users', { json: john });
      const res = await app.request('GET', '/api/users');
      assert.equal(res.status, 200);
      assert.deepEqual(res.body, [{ id: 1, ...john }]);
    });
  });

  describe('POST /api/users', () => {
    it('201, Location і тіло { id, name, email }', async () => {
      const res = await app.request('POST', '/api/users', { json: john });
      assert.equal(res.status, 201);
      assert.equal(res.headers.get('location'), '/api/users/1');
      assert.deepEqual(res.body, { id: 1, name: 'John Doe', email: 'john@example.com' });
    });

    it('нормалізує дані (trim, email у нижньому регістрі)', async () => {
      const res = await app.request('POST', '/api/users', { json: { name: ' Ann ', email: 'ANN@Example.com' } });
      assert.equal(res.status, 201);
      assert.deepEqual(res.body, { id: 1, name: 'Ann', email: 'ann@example.com' });
    });

    it('400 validation_failed: відсутні поля', async () => {
      const res = await app.request('POST', '/api/users', { json: {} });
      assert.equal(res.status, 400);
      assert.equal(res.body.error, 'validation_failed');
      assert.deepEqual(res.body.details, [
        { field: 'name', message: 'is required' },
        { field: 'email', message: 'is required' },
      ]);
    });

    it('400 validation_failed: неправильні типи, порожні значення, формат email', async () => {
      const res = await app.request('POST', '/api/users', { json: { name: 123, email: 'nope' } });
      assert.equal(res.status, 400);
      assert.deepEqual(res.body.details, [
        { field: 'name', message: 'must be a string' },
        { field: 'email', message: 'must be a valid email address' },
      ]);

      const blank = await app.request('POST', '/api/users', { json: { name: '   ', email: 'a@b.co' } });
      assert.equal(blank.status, 400);
      assert.deepEqual(blank.body.details, [{ field: 'name', message: 'must not be empty' }]);
    });

    it('400 invalid_json: зламаний JSON', async () => {
      const res = await app.request('POST', '/api/users', {
        body: '{ "name": ',
        headers: { 'Content-Type': 'application/json' },
      });
      assert.equal(res.status, 400);
      assert.equal(res.body.error, 'invalid_json');
    });

    it('400: порожнє тіло та тіло-масив', async () => {
      const empty = await app.request('POST', '/api/users', { headers: { 'Content-Type': 'application/json' } });
      assert.equal(empty.status, 400);
      assert.equal(empty.body.error, 'validation_failed');

      const array = await app.request('POST', '/api/users', { json: [1, 2] });
      assert.equal(array.status, 400);
      assert.equal(array.body.error, 'validation_failed');
    });

    it('415 unsupported_media_type: тіло не JSON', async () => {
      const res = await app.request('POST', '/api/users', {
        body: 'name=John',
        headers: { 'Content-Type': 'text/plain' },
      });
      assert.equal(res.status, 415);
      assert.equal(res.body.error, 'unsupported_media_type');
    });

    it('409 email_already_exists: дублікат email (без урахування регістру)', async () => {
      await app.request('POST', '/api/users', { json: john });
      const res = await app.request('POST', '/api/users', { json: { name: 'Other', email: 'JOHN@example.com' } });
      assert.equal(res.status, 409);
      assert.equal(res.body.error, 'email_already_exists');
      assert.match(res.body.message, /already exists/);
    });

    it('некоректний запит не доходить до сховища', async () => {
      await app.request('POST', '/api/users', { json: { name: '', email: 'x' } });
      const list = await app.request('GET', '/api/users');
      assert.deepEqual(list.body, []);
    });
  });

  describe('GET /api/users/:id', () => {
    it('200 для існуючого користувача', async () => {
      await app.request('POST', '/api/users', { json: john });
      const res = await app.request('GET', '/api/users/1');
      assert.equal(res.status, 200);
      assert.deepEqual(res.body, { id: 1, ...john });
    });

    it('404 user_not_found для неіснуючого', async () => {
      const res = await app.request('GET', '/api/users/42');
      assert.equal(res.status, 404);
      assert.deepEqual(res.body, {
        error: 'user_not_found',
        message: 'User with id 42 was not found',
      });
    });

    it('400 invalid_id для некоректного ідентифікатора', async () => {
      for (const bad of ['abc', '0', '-1', '1.5', '01', '99999999999999999999']) {
        const res = await app.request('GET', `/api/users/${bad}`);
        assert.equal(res.status, 400, `id=${bad}`);
        assert.equal(res.body.error, 'invalid_id', `id=${bad}`);
      }
    });
  });

  describe('PUT /api/users/:id', () => {
    it('200 і оновлені дані', async () => {
      await app.request('POST', '/api/users', { json: john });
      const res = await app.request('PUT', '/api/users/1', { json: { name: 'Johnny', email: 'johnny@example.com' } });
      assert.equal(res.status, 200);
      assert.deepEqual(res.body, { id: 1, name: 'Johnny', email: 'johnny@example.com' });

      const check = await app.request('GET', '/api/users/1');
      assert.equal(check.body.name, 'Johnny');
    });

    it('404 для неіснуючого користувача', async () => {
      const res = await app.request('PUT', '/api/users/9', { json: john });
      assert.equal(res.status, 404);
      assert.equal(res.body.error, 'user_not_found');
    });

    it('400 для некоректного id та для некоректного тіла', async () => {
      const badId = await app.request('PUT', '/api/users/abc', { json: john });
      assert.equal(badId.status, 400);
      assert.equal(badId.body.error, 'invalid_id');

      await app.request('POST', '/api/users', { json: john });
      const badBody = await app.request('PUT', '/api/users/1', { json: { name: '' } });
      assert.equal(badBody.status, 400);
      assert.equal(badBody.body.error, 'validation_failed');
    });

    it('409 при спробі взяти email іншого користувача', async () => {
      await app.request('POST', '/api/users', { json: john });
      await app.request('POST', '/api/users', { json: { name: 'Jane', email: 'jane@example.com' } });
      const res = await app.request('PUT', '/api/users/2', { json: { name: 'Jane', email: 'john@example.com' } });
      assert.equal(res.status, 409);
      assert.equal(res.body.error, 'email_already_exists');
    });
  });

  describe('DELETE /api/users/:id', () => {
    it('204 без тіла, далі GET дає 404', async () => {
      await app.request('POST', '/api/users', { json: john });
      const del = await app.request('DELETE', '/api/users/1');
      assert.equal(del.status, 204);
      assert.equal(del.text, '');

      const get = await app.request('GET', '/api/users/1');
      assert.equal(get.status, 404);
    });

    it('404 для неіснуючого та 400 для некоректного id', async () => {
      assert.equal((await app.request('DELETE', '/api/users/5')).status, 404);
      assert.equal((await app.request('DELETE', '/api/users/xyz')).status, 400);
    });
  });

  describe('інші помилки', () => {
    it('405 + Allow для непідтримуваного методу', async () => {
      const res = await app.request('PATCH', '/api/users');
      assert.equal(res.status, 405);
      assert.equal(res.headers.get('allow'), 'GET, POST');
      assert.equal(res.body.error, 'method_not_allowed');

      const one = await app.request('POST', '/api/users/1');
      assert.equal(one.status, 405);
      assert.equal(one.headers.get('allow'), 'GET, PUT, DELETE');
    });

    it('404 route_not_found для невідомого маршруту', async () => {
      const res = await app.request('GET', '/api/unknown');
      assert.equal(res.status, 404);
      assert.equal(res.body.error, 'route_not_found');
    });

    it('кожна помилка має єдиний формат { error, message }', async () => {
      const responses = [
        await app.request('GET', '/api/users/42'),
        await app.request('GET', '/api/users/abc'),
        await app.request('POST', '/api/users', { json: {} }),
        await app.request('PATCH', '/api/users'),
        await app.request('GET', '/nope'),
      ];
      for (const res of responses) {
        assert.equal(typeof res.body.error, 'string');
        assert.equal(typeof res.body.message, 'string');
        assert.match(res.body.error, /^[a-z_]+$/);
      }
    });
  });
});

describe('API: 500 Internal Server Error', () => {
  let app;
  before(async () => { app = await startApp({ userRepository: new BrokenUserRepository() }); });
  after(async () => { await app.close(); });

  it('помилка сховища -> 500 internal_error без витоку деталей', async () => {
    // глушимо console.error, щоб не засмічувати вивід тестів очікуваною помилкою
    const original = console.error;
    console.error = () => {};
    try {
      const res = await app.request('GET', '/api/users');
      assert.equal(res.status, 500);
      assert.deepEqual(res.body, { error: 'internal_error', message: 'Internal server error' });
      assert.ok(!res.text.includes('secret-host'), 'деталі помилки не повинні потрапляти клієнту');
    } finally {
      console.error = original;
    }
  });
});

describe('API: заміна репозиторію без змін сервісу', () => {
  it('застосунок працює з довільною реалізацією контракту', async () => {
    const { ArrayUserRepository } = require('./helpers');
    const app = await startApp({ userRepository: new ArrayUserRepository() });
    try {
      const created = await app.request('POST', '/api/users', { json: john });
      assert.equal(created.status, 201);
      const found = await app.request('GET', `/api/users/${created.body.id}`);
      assert.deepEqual(found.body, created.body);
    } finally {
      await app.close();
    }
  });

  it('seed-дані доступні через API (newInMemoryRepository із seed)', async () => {
    const app = await startApp({ userRepository: newInMemoryRepository([john]) });
    try {
      const res = await app.request('GET', '/api/users/1');
      assert.equal(res.status, 200);
      assert.equal(res.body.email, 'john@example.com');
    } finally {
      await app.close();
    }
  });
});

describe('Middleware №1: логування запитів', () => {
  it('пише метод, шлях, статус і тривалість: "GET /api/users/10 404 3.2ms"', async () => {
    const lines = [];
    const app = await startApp({ logger: (line) => lines.push(line) });
    try {
      await app.request('GET', '/api/users/10');
      await app.request('POST', '/api/users', { json: john });
      // 'finish' спрацьовує після відправки відповіді — даємо подіям завершитись
      await new Promise((r) => setTimeout(r, 20));
    } finally {
      await app.close();
    }
    assert.match(lines[0], /^GET \/api\/users\/10 404 \d+\.\d{1}ms$/);
    assert.match(lines[1], /^POST \/api\/users 201 \d+\.\d{1}ms$/);
  });
});
