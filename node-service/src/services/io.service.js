// I/O-сервіс: імітує асинхронну операцію (напр. звернення до БД чи іншого API).
// setTimeout ставить таймер у event loop і звільняє потік — інші запити
// обробляються, поки цей "чекає".
function wait(delayMs) {
  return new Promise((resolve) => setTimeout(resolve, delayMs));
}

module.exports = { wait };
