// Модель User і доменні правила валідації/нормалізації. Нічого не знає про HTTP чи сховище.
const NAME_MAX = 100;
const EMAIL_MAX = 254;
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

class User {
  constructor({ id, name, email }) {
    this.id = id;
    this.name = name;
    this.email = email;
  }

  toJSON() {
    return { id: this.id, name: this.name, email: this.email };
  }
}

const isPlainObject = (v) => typeof v === 'object' && v !== null && !Array.isArray(v);
const length = (s) => [...s].length; // рахуємо символи, а не UTF-16 одиниці

function checkField(input, field, rules) {
  const value = input[field];
  if (value === undefined || value === null) return [{ field, message: 'is required' }];
  if (typeof value !== 'string') return [{ field, message: 'must be a string' }];
  const trimmed = value.trim();
  if (trimmed === '') return [{ field, message: 'must not be empty' }];
  return rules(trimmed).map((message) => ({ field, message }));
}

/**
 * Перевіряє вхідні дані користувача: обов'язковість, типи, порожні значення, довжину, формат.
 * Повертає масив помилок полів (порожній, якщо все гаразд).
 */
function validateUserInput(input) {
  if (!isPlainObject(input)) {
    return [{ field: 'body', message: 'request body must be a JSON object' }];
  }
  return [
    ...checkField(input, 'name', (v) =>
      length(v) > NAME_MAX ? [`must be at most ${NAME_MAX} characters`] : []),
    ...checkField(input, 'email', (v) => {
      if (length(v) > EMAIL_MAX) return [`must be at most ${EMAIL_MAX} characters`];
      return EMAIL_RE.test(v) ? [] : ['must be a valid email address'];
    }),
  ];
}

/** Нормалізація — бізнес-правило: name без пробілів по краях, email у нижньому регістрі. */
function normalizeUserInput({ name, email }) {
  return { name: name.trim(), email: email.trim().toLowerCase() };
}

module.exports = { User, validateUserInput, normalizeUserInput, NAME_MAX, EMAIL_MAX };
