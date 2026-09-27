// Конфігурація сервісу. Порт можна змінити через змінну оточення PORT.
module.exports = {
  PORT: Number(process.env.PORT) || 3000,
  SERVICE_NAME: 'tech-store-node',
};
