// КОНТРАКТ репозиторію користувачів. Сервіс залежить лише від цього контракту,
// а не від конкретної реалізації (in-memory, SQL, MongoDB...).
//
// Усі методи асинхронні (як у справжньої БД), тому заміна реалізації не змінює сервіс.
//
//   create(user)      -> User             кидає ConflictError, якщо email зайнятий
//   findById(id)      -> User | null
//   findByEmail(mail) -> User | null
//   findAll()         -> User[]
//   update(user)      -> User | null      null, якщо користувача з таким id немає
//   delete(id)        -> boolean          true, якщо щось було видалено
const REQUIRED_METHODS = ['create', 'findById', 'findByEmail', 'findAll', 'update', 'delete'];

class UserRepository {
  async create() { throw new Error('UserRepository.create is not implemented'); }
  async findById() { throw new Error('UserRepository.findById is not implemented'); }
  async findByEmail() { throw new Error('UserRepository.findByEmail is not implemented'); }
  async findAll() { throw new Error('UserRepository.findAll is not implemented'); }
  async update() { throw new Error('UserRepository.update is not implemented'); }
  async delete() { throw new Error('UserRepository.delete is not implemented'); }
}

/** Перевіряє, що об'єкт відповідає контракту (duck typing — наслідування від класу не обов'язкове). */
function assertUserRepository(repository) {
  const missing = REQUIRED_METHODS.filter((m) => typeof repository?.[m] !== 'function');
  if (missing.length > 0) {
    throw new TypeError(`Object does not satisfy the UserRepository contract, missing: ${missing.join(', ')}`);
  }
  return repository;
}

module.exports = { UserRepository, assertUserRepository, REQUIRED_METHODS };
