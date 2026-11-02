const path = require('node:path');

// Конфігурація сервісу. Порт можна змінити через змінну оточення PORT.
module.exports = {
  PORT: Number(process.env.PORT) || 3000,
  SERVICE_NAME: 'tech-store-node',

  // Лабораторна 5
  STORAGE_DIR: process.env.STORAGE_DIR || path.join(__dirname, '..', 'storage'),
  WS_PATH: '/ws',
  MAX_RUNNING_JOBS: 3,
};
