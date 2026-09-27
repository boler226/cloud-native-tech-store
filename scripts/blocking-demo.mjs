// Демонструє, як важкий /cpu-запит впливає на паралельні /health запити.
// Спочатку йде /cpu (без очікування завершення), одразу за ним — декілька
// /health із невеликим інтервалом. Виводить, коли кожен запит пішов і коли
// повернулась відповідь — за цим видно, чи блокується сервер.
//
// Використання:
//   node scripts/blocking-demo.mjs <base> <cpuPath> [cpuQuery] [healthCount]
//
// Приклади:
//   node scripts/blocking-demo.mjs http://localhost:3000 /cpu
//   node scripts/blocking-demo.mjs http://localhost:8080 /cpu/sequential
//   node scripts/blocking-demo.mjs http://localhost:8080 /cpu/parallel

const base = (process.argv[2] || 'http://localhost:3000').replace(/\/$/, '');
const cpuPath = process.argv[3] || '/cpu';
const cpuQuery = process.argv[4] || '';
const healthCount = Number(process.argv[5] || 5);

const t0 = performance.now();

function timedFetch(label, path) {
  const sentAt = performance.now() - t0;
  return fetch(base + path).then(async (res) => {
    await res.json().catch(() => null);
    const doneAt = performance.now() - t0;
    return { label, status: res.status, sentAt, doneAt, durationMs: doneAt - sentAt };
  });
}

console.log(`Base: ${base}`);
console.log(`CPU endpoint: ${cpuPath}${cpuQuery}`);
console.log(`Health requests: ${healthCount}\n`);

// CPU-запит іде першим і НЕ очікується одразу (fire-and-collect-later),
// саме так, як описано у завданні: спочатку /cpu, потім декілька /health.
const cpuPromise = timedFetch('CPU', cpuPath + cpuQuery);

// Невелика затримка, щоб гарантувати порядок відправлення запитів.
await new Promise((r) => setTimeout(r, 20));

const healthPromises = [];
for (let i = 0; i < healthCount; i += 1) {
  healthPromises.push(timedFetch(`health#${i + 1}`, '/health'));
  // eslint-disable-next-line no-await-in-loop
  await new Promise((r) => setTimeout(r, 15));
}

const results = await Promise.all([cpuPromise, ...healthPromises]);
results.sort((a, b) => a.sentAt - b.sentAt);

console.log('label        sentAt(ms)  doneAt(ms)  duration(ms)  status');
for (const r of results) {
  console.log(
    `${r.label.padEnd(12)} ${r.sentAt.toFixed(0).padStart(9)}   ${r.doneAt.toFixed(0).padStart(9)}   ${r.durationMs.toFixed(0).padStart(10)}   ${r.status}`,
  );
}

console.log(
  '\nЯкщо доданки "health#N" завершуються майже одночасно ЗІ СПІЛЬНИМ моментом ' +
  'завершення CPU-запиту (усі "doneAt" ~ рівні doneAt CPU) — сервер БЛОКУЄТЬСЯ ' +
  'на час обчислення. Якщо health-запити завершуються швидко, незалежно від ' +
  'CPU — сервер їх обробляє паралельно.',
);
