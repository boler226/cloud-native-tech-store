// Тести рівня Service. HTTP-сервер НЕ запускається — сервіс перевіряється напряму.
// Одні й ті самі тести проганяються на двох різних реалізаціях репозиторію.
const { describe, it } = require('node:test');
const assert = require('node:assert/strict');

const UserService = require('../src/services/user.service');
const { ValidationError, NotFoundError, ConflictError } = require('../src/errors/app-errors');
const { newInMemoryRepository, ArrayUserRepository, BrokenUserRepository } = require('./helpers');

const john = { name: 'John Doe', email: 'john@example.com' };

const implementations = {
  'InMemoryUserRepository (DataStore)': () => newInMemoryRepository(),
  'ArrayUserRepository (інша реалізація)': () => new ArrayUserRepository(),
};

for (const [implName, makeRepo] of Object.entries(implementations)) {
  describe(`UserService з ${implName}`, () => {
    const makeService = () => new UserService(makeRepo());

    describe('створення', () => {
      it('створює користувача й присвоює id', async () => {
        const user = await makeService().create(john);
        assert.equal(user.id, 1);
        assert.equal(user.name, 'John Doe');
        assert.equal(user.email, 'john@example.com');
      });

      it('нормалізує дані: обрізає пробіли в name, email -> нижній регістр', async () => {
        const user = await makeService().create({ name: '  Ann  ', email: '  ANN@Example.COM ' });
        assert.equal(user.name, 'Ann');
        assert.equal(user.email, 'ann@example.com');
      });

      it('кожен новий користувач отримує наступний id', async () => {
        const service = makeService();
        const a = await service.create(john);
        const b = await service.create({ name: 'Jane', email: 'jane@example.com' });
        assert.deepEqual([a.id, b.id], [1, 2]);
      });

      it('дублікат email -> ConflictError (email_already_exists)', async () => {
        const service = makeService();
        await service.create(john);
        await assert.rejects(
          () => service.create({ name: 'Other', email: 'john@example.com' }),
          (err) => err instanceof ConflictError && err.code === 'email_already_exists',
        );
      });

      it('бізнес-правило: унікальність email не залежить від регістру', async () => {
        const service = makeService();
        await service.create(john);
        await assert.rejects(
          () => service.create({ name: 'Other', email: 'JOHN@EXAMPLE.COM' }),
          ConflictError,
        );
      });

      it('некоректні дані -> ValidationError з переліком полів', async () => {
        await assert.rejects(
          () => makeService().create({ name: '   ', email: 'not-an-email' }),
          (err) => {
            assert.ok(err instanceof ValidationError);
            assert.equal(err.code, 'validation_failed');
            assert.deepEqual(err.details.map((d) => d.field), ['name', 'email']);
            return true;
          },
        );
      });

      it('відсутні поля та неправильні типи -> ValidationError', async () => {
        const service = makeService();
        await assert.rejects(() => service.create({}), ValidationError);
        await assert.rejects(() => service.create({ name: 5, email: 'a@b.co' }), ValidationError);
        await assert.rejects(() => service.create(null), ValidationError);
      });

      it('занадто довге ім’я -> ValidationError', async () => {
        await assert.rejects(
          () => makeService().create({ name: 'x'.repeat(101), email: 'a@b.co' }),
          ValidationError,
        );
      });
    });

    describe('пошук', () => {
      it('getById повертає створеного користувача', async () => {
        const service = makeService();
        const created = await service.create(john);
        assert.deepEqual(await service.getById(created.id), created);
      });

      it('getById неіснуючого -> NotFoundError (user_not_found)', async () => {
        await assert.rejects(
          () => makeService().getById(42),
          (err) =>
            err instanceof NotFoundError &&
            err.code === 'user_not_found' &&
            err.message === 'User with id 42 was not found',
        );
      });

      it('некоректний id -> ValidationError (invalid_id)', async () => {
        const service = makeService();
        for (const bad of [0, -1, 1.5, NaN, '1', null, undefined]) {
          await assert.rejects(
            () => service.getById(bad),
            (err) => err instanceof ValidationError && err.code === 'invalid_id',
            `id=${String(bad)}`,
          );
        }
      });

      it('list: порожній список, потім усі користувачі', async () => {
        const service = makeService();
        assert.deepEqual(await service.list(), []);
        await service.create(john);
        await service.create({ name: 'Jane', email: 'jane@example.com' });
        assert.equal((await service.list()).length, 2);
      });
    });

    describe('оновлення', () => {
      it('оновлює name та email', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        const updated = await service.update(id, { name: 'Johnny', email: 'johnny@example.com' });
        assert.equal(updated.id, id);
        assert.equal(updated.name, 'Johnny');
        assert.equal((await service.getById(id)).email, 'johnny@example.com');
      });

      it('можна залишити власний email без змін', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        const updated = await service.update(id, { name: 'New Name', email: 'JOHN@example.com' });
        assert.equal(updated.name, 'New Name');
      });

      it('email іншого користувача -> ConflictError', async () => {
        const service = makeService();
        await service.create(john);
        const jane = await service.create({ name: 'Jane', email: 'jane@example.com' });
        await assert.rejects(
          () => service.update(jane.id, { name: 'Jane', email: 'john@example.com' }),
          ConflictError,
        );
      });

      it('неіснуючий користувач -> NotFoundError', async () => {
        await assert.rejects(() => makeService().update(7, john), NotFoundError);
      });

      it('некоректні дані -> ValidationError, дані не змінюються', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        await assert.rejects(() => service.update(id, { name: '', email: 'x' }), ValidationError);
        assert.equal((await service.getById(id)).name, 'John Doe');
      });
    });

    describe('видалення', () => {
      it('видаляє користувача', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        await service.delete(id);
        await assert.rejects(() => service.getById(id), NotFoundError);
      });

      it('повторне видалення -> NotFoundError', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        await service.delete(id);
        await assert.rejects(() => service.delete(id), NotFoundError);
      });

      it('email звільняється після видалення', async () => {
        const service = makeService();
        const { id } = await service.create(john);
        await service.delete(id);
        const again = await service.create(john);
        assert.equal(again.email, 'john@example.com');
      });
    });
  });
}

describe('UserService: ізоляція від репозиторію', () => {
  it('відмовляється працювати з об’єктом, що не виконує контракт репозиторію', () => {
    assert.throws(() => new UserService({}), /UserRepository contract/);
    assert.throws(() => new UserService({ create() {} }), /missing/);
  });

  it('помилки сховища не маскуються під доменні (їх оброблять як 500)', async () => {
    const service = new UserService(new BrokenUserRepository());
    await assert.rejects(
      () => service.getById(1),
      (err) => !(err instanceof ValidationError) && !(err instanceof NotFoundError) && /db connection lost/.test(err.message),
    );
  });
});
