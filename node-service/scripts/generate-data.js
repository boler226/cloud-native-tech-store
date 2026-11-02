// Використання: npm run gen:data -- [розмірMB=100] [ім'я=big.csv]
// Генерує CSV-лог. У пам'яті ніколи не тримається більше одного батчу (~64 КБ).
const fs = require('node:fs');
const path = require('node:path');
const { once } = require('node:events');
const { STORAGE_DIR } = require('../src/config');

const sizeMb = Number(process.argv[2]) || 100;
const name = process.argv[3] || 'big.csv';
const LEVELS = ['INFO', 'INFO', 'INFO', 'DEBUG', 'WARN', 'ERROR'];
const SERVICES = ['auth', 'orders', 'catalog', 'payments', 'search'];
const BATCH_BYTES = 64 * 1024;

async function main() {
  const dir = path.join(STORAGE_DIR, 'input');
  fs.mkdirSync(dir, { recursive: true });
  const target = sizeMb * 1024 * 1024;
  const out = fs.createWriteStream(path.join(dir, name));

  let written = 0;
  let id = 0;
  const startedAt = Date.now();

  while (written < target) {
    let batch = '';
    while (batch.length < BATCH_BYTES) {
      id += 1;
      const level = LEVELS[Math.floor(Math.random() * LEVELS.length)];
      const service = SERVICES[Math.floor(Math.random() * SERVICES.length)];
      batch += `${id},${new Date(startedAt + id * 7).toISOString()},${level},${service},request #${id} handled in ${Math.floor(Math.random() * 500)}ms\n`;
    }
    written += Buffer.byteLength(batch);
    // Backpressure: якщо внутрішній буфер повний — чекаємо 'drain'.
    if (!out.write(batch)) await once(out, 'drain');
  }

  out.end();
  await once(out, 'finish');
  console.log(`Готово: ${name}, ${(written / 1024 / 1024).toFixed(1)} MB, ${id} рядків, ${Date.now() - startedAt} мс`);
}

main().catch((e) => { console.error(e); process.exit(1); });
