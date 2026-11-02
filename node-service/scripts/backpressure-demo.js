// npm run demo:backpressure   (спершу: npm run gen:data)
// Порівнює: (A) ігнорування write()===false  vs  (B) pipe().
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const SlowFileWritable = require('../src/streams/slow-file.writable');
const { STORAGE_DIR } = require('../src/config');

const INPUT = path.join(STORAGE_DIR, 'input', process.argv[2] || 'big.csv');
const RUN_MS = 3000;

function run(mode) {
  return new Promise((resolve) => {
    const readable = fs.createReadStream(INPUT, { highWaterMark: 64 * 1024 });
    const slow = new SlowFileWritable(os.devNull, { delayMs: 20, highWaterMark: 16 * 1024 });
    let peak = 0;
    let pauses = 0;
    let writeFalse = 0;
    readable.on('pause', () => { pauses += 1; });

    if (mode === 'unsafe') {
      // Помилка: ігноруємо повернене значення write() -> буфер Writable росте необмежено.
      readable.on('data', (chunk) => { if (!slow.write(chunk)) writeFalse += 1; });
    } else {
      readable.pipe(slow); // pipe сам реагує на write()===false та 'drain'
    }

    const timer = setInterval(() => { peak = Math.max(peak, slow.writableLength); }, 50);
    setTimeout(() => {
      clearInterval(timer);
      readable.destroy();
      slow.destroy();
      resolve({
        mode,
        peakBufferedKB: Math.round(peak / 1024),
        readablePauses: pauses,
        writeReturnedFalse: writeFalse,
        rssMB: Math.round(process.memoryUsage().rss / 1024 / 1024),
      });
    }, RUN_MS);
  });
}

(async () => {
  console.table([await run('unsafe'), await run('pipe')]);
})();
