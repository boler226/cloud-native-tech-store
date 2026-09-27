// Навантажувальне тестування для лабораторної 2.
// Піднімає Node- і Go-сервіси по черзі, б'є по /io та /cpu (у Go — /cpu/sequential
// і /cpu/parallel) через autocannon, паралельно семплює CPU% і RAM процесу-сервера
// (через pidusage) і в кінці друкує зведену Markdown-таблицю.
//
// Використання (з папки scripts/, після `npm install`):
//   node bench-all.mjs
//   node bench-all.mjs --duration=10 --connections=50
//
// Параметри одного запиту в /cpu підбираються під ваше залізо через
// --cpu-iterations і --cpu-passes (див. нижче), інакше кожен /cpu-запит
// триває надто довго чи надто коротко для показового навантаження.

import autocannon from 'autocannon';
import pidusage from 'pidusage';
import { spawn } from 'node:child_process';
import { setTimeout as sleep } from 'node:timers/promises';
import path from 'node:path';
import os from 'node:os';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '..');

// /tmp не існує в Windows — беремо крос-платформну тимчасову папку,
// і виконуваний файл Go на Windows має мати розширення .exe.
const GO_BINARY_PATH = path.join(
  os.tmpdir(),
  process.platform === 'win32' ? 'lab2-go-server.exe' : 'lab2-go-server',
);

const args = Object.fromEntries(
  process.argv.slice(2).map((a) => {
    const [k, v] = a.replace(/^--/, '').split('=');
    return [k, v ?? true];
  }),
);

// I/O — навантаження на конкурентність (багато паралельних очікувань), тому
// беремо високу к-сть з'єднань і коротший timeout.
const IO_DURATION_S = Number(args['io-duration'] || args.duration || 10);
const IO_CONNECTIONS = Number(args['io-connections'] || args.connections || 50);

// CPU — навантаження звантажує процесор, а не мережу: занадто велика
// конкурентність лише змушує запити чекати в черзі, нічого не прискорюючи.
// Тому тут значно менше з'єднань, довший timeout і триваліше вікно виміру,
// щоб встигло завершитись достатньо запитів для статистики.
const CPU_DURATION_S = Number(args['cpu-duration'] || 15);
const CPU_CONNECTIONS = Number(args['cpu-connections'] || 4);
const CPU_TIMEOUT_S = Number(args['cpu-timeout'] || 30);

const IO_DELAY_MS = Number(args['io-delay'] || 200);
const CPU_ITERATIONS = Number(args['cpu-iterations'] || 300_000_000);
const CPU_PASSES = Number(args['cpu-passes'] || 4);
const SAMPLE_INTERVAL_MS = 200;

const NODE_PORT = 3100;
const GO_PORT = 8180;

function log(...parts) {
  console.log(...parts);
}

// Запускає команду, чекає, поки порт відповість на /health, повертає {proc, pid}.
async function startServer(label, command, cmdArgs, cwd, env, port) {
  log(`\n[${label}] starting: ${command} ${cmdArgs.join(' ')} (port ${port})`);
  const proc = spawn(command, cmdArgs, {
    cwd,
    env: { ...process.env, ...env, PORT: String(port) },
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  proc.stdout.on('data', () => {});
  proc.stderr.on('data', (d) => process.stderr.write(`[${label}] ${d}`));

  const base = `http://localhost:${port}`;
  const deadline = Date.now() + 15_000;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${base}/health`);
      if (res.ok) {
        log(`[${label}] ready, pid=${proc.pid}`);
        return { proc, pid: proc.pid, base };
      }
    } catch {
      // сервер ще не піднявся
    }
    // eslint-disable-next-line no-await-in-loop
    await sleep(150);
  }
  throw new Error(`${label}: server did not become ready on ${base}`);
}

// Надсилає SIGTERM і чекає на реальне завершення процесу (з fallback у SIGKILL),
// щоб порт гарантовано звільнився до старту наступного сервера.
async function stopServer({ proc }) {
  if (proc.exitCode !== null || proc.signalCode !== null) return;
  const exited = new Promise((resolve) => proc.once('exit', resolve));
  proc.kill('SIGTERM');
  const timedOut = await Promise.race([exited.then(() => false), sleep(6000).then(() => true)]);
  if (timedOut) {
    proc.kill('SIGKILL');
    await exited;
  }
}

// Семплить CPU%/RAM процесу кожні SAMPLE_INTERVAL_MS, поки не викликано stop().
function startSampling(pid) {
  const samples = [];
  const timer = setInterval(async () => {
    try {
      const stat = await pidusage(pid);
      samples.push({ cpu: stat.cpu, memMB: stat.memory / (1024 * 1024) });
    } catch {
      // процес міг щойно завершитись між семплами — ігноруємо
    }
  }, SAMPLE_INTERVAL_MS);
  return {
    stop() {
      clearInterval(timer);
      if (samples.length === 0) return { avgCpu: 0, avgMemMB: 0, peakMemMB: 0 };
      const avgCpu = samples.reduce((s, x) => s + x.cpu, 0) / samples.length;
      const avgMemMB = samples.reduce((s, x) => s + x.memMB, 0) / samples.length;
      const peakMemMB = Math.max(...samples.map((x) => x.memMB));
      return { avgCpu, avgMemMB, peakMemMB };
    },
  };
}

async function runWorkload({ workloadLabel, runtimeLabel, url, pid, connections, duration, timeoutS }) {
  log(`\n=== ${workloadLabel} / ${runtimeLabel} ===`);
  log(`GET ${url}  connections=${connections} duration=${duration}s timeout=${timeoutS ?? 10}s`);

  const sampler = startSampling(pid);
  const result = await autocannon({
    url,
    connections,
    duration,
    ...(timeoutS ? { timeout: timeoutS } : {}),
  });
  const { avgCpu, avgMemMB, peakMemMB } = sampler.stop();

  const row = {
    workload: workloadLabel,
    runtime: runtimeLabel,
    rps: Math.round(result.requests.average),
    avgMs: Number(result.latency.average.toFixed(1)),
    p95Ms: result.latency.p97_5, // autocannon не дає p95 напряму, p97_5 — найближчий перцентиль
    p99Ms: result.latency.p99,
    errors: result.errors + result.timeouts + result.non2xx,
    cpuPercent: Math.round(avgCpu),
    ramMB: Math.round(avgMemMB),
    ramPeakMB: Math.round(peakMemMB),
  };
  log(
    `RPS=${row.rps} avg=${row.avgMs}ms p95~=${row.p95Ms}ms p99=${row.p99Ms}ms ` +
    `CPU=${row.cpuPercent}% RAM(avg)=${row.ramMB}MB errors=${row.errors}`,
  );
  return row;
}

function toMarkdownTable(rows) {
  const header = '| Workload | Runtime | RPS | Avg (ms) | p95 (ms) | p99 (ms) | CPU (%) | RAM avg (MB) |';
  const sep = '|---|---|---|---|---|---|---|---|';
  const lines = rows.map(
    (r) => `| ${r.workload} | ${r.runtime} | ${r.rps} | ${r.avgMs} | ${r.p95Ms} | ${r.p99Ms} | ${r.cpuPercent} | ${r.ramMB} |`,
  );
  return [header, sep, ...lines].join('\n');
}

async function main() {
  log('Параметри запуску:');
  log(`  I/O:  duration=${IO_DURATION_S}s, connections=${IO_CONNECTIONS}, delayMs=${IO_DELAY_MS}`);
  log(`  CPU:  duration=${CPU_DURATION_S}s, connections=${CPU_CONNECTIONS}, iterations=${CPU_ITERATIONS}, passes=${CPU_PASSES}`);

  const rows = [];

  // --- Node.js ---
  const node = await startServer(
    'node',
    'node',
    ['src/server.js'],
    path.join(repoRoot, 'node-service'),
    {},
    NODE_PORT,
  );
  await sleep(300);
  rows.push(
    await runWorkload({
      workloadLabel: 'I/O',
      runtimeLabel: 'Node.js',
      url: `${node.base}/io?delayMs=${IO_DELAY_MS}`,
      pid: node.pid,
      connections: IO_CONNECTIONS,
      duration: IO_DURATION_S,
    }),
  );
  rows.push(
    await runWorkload({
      workloadLabel: 'CPU',
      runtimeLabel: 'Node.js',
      url: `${node.base}/cpu?iterations=${CPU_ITERATIONS}`,
      pid: node.pid,
      connections: CPU_CONNECTIONS,
      duration: CPU_DURATION_S,
      timeoutS: CPU_TIMEOUT_S,
    }),
  );
  await stopServer(node);
  await sleep(500);

  // --- Go ---
  log('\nBuilding Go binary...');
  await new Promise((resolve, reject) => {
    const build = spawn('go', ['build', '-o', GO_BINARY_PATH, './cmd/server'], {
      cwd: path.join(repoRoot, 'go-service'),
      stdio: 'inherit',
    });
    build.on('exit', (code) => (code === 0 ? resolve() : reject(new Error('go build failed'))));
  });

  const go = await startServer('go', GO_BINARY_PATH, [], repoRoot, {}, GO_PORT);
  await sleep(300);
  rows.push(
    await runWorkload({
      workloadLabel: 'I/O',
      runtimeLabel: 'Go',
      url: `${go.base}/io?delayMs=${IO_DELAY_MS}`,
      pid: go.pid,
      connections: IO_CONNECTIONS,
      duration: IO_DURATION_S,
    }),
  );
  rows.push(
    await runWorkload({
      workloadLabel: 'CPU (sequential)',
      runtimeLabel: 'Go',
      url: `${go.base}/cpu/sequential?iterations=${CPU_ITERATIONS}&passes=${CPU_PASSES}`,
      pid: go.pid,
      connections: CPU_CONNECTIONS,
      duration: CPU_DURATION_S,
      timeoutS: CPU_TIMEOUT_S,
    }),
  );
  rows.push(
    await runWorkload({
      workloadLabel: 'CPU (parallel)',
      runtimeLabel: 'Go',
      url: `${go.base}/cpu/parallel?iterations=${CPU_ITERATIONS}&passes=${CPU_PASSES}`,
      pid: go.pid,
      connections: CPU_CONNECTIONS,
      duration: CPU_DURATION_S,
      timeoutS: CPU_TIMEOUT_S,
    }),
  );
  await stopServer(go);

  log('\n\n########## Результат (Markdown) ##########\n');
  console.log(toMarkdownTable(rows));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
