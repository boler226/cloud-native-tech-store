const fs = require('node:fs');
const LineFilterTransform = require('./line-filter.transform');
const SlowFileWritable = require('./slow-file.writable');

/**
 * Зв'язує File -> Readable -> Transform -> Writable через pipe().
 *
 * Параметри, що впливають на розмір порцій:
 *  - chunkSize (createReadStream.highWaterMark): розмір одного читання з диска.
 *    За замовчуванням для fs — 64 КБ. Більше => менше системних викликів, але більше RAM на чанк.
 *  - start / end: читання лише діапазону байтів файлу.
 *  - encoding: якщо задати, чанки будуть рядками, а не Buffer (тут не задаємо —
 *    інакше можна розірвати багатобайтовий символ; декодуємо у Transform).
 *  - writableHighWaterMark: поріг, після якого write() повертає false (backpressure).
 *
 * pipe() НЕ пробрасує помилки між потоками — тому вішаємо error-обробники вручну
 * (альтернатива — stream.pipeline, що робить це сам).
 */
async function processFile({
  inputPath,
  outputPath,
  pattern,
  chunkSize = 64 * 1024,
  writableHighWaterMark = 16 * 1024,
  writeDelayMs = 0,
  onProgress = () => {},
}) {
  const { size } = await fs.promises.stat(inputPath);
  const startedAt = Date.now();

  const readable = fs.createReadStream(inputPath, { highWaterMark: chunkSize });
  const transform = new LineFilterTransform({ pattern, highWaterMark: writableHighWaterMark });
  const writable = new SlowFileWritable(outputPath, {
    delayMs: writeDelayMs,
    highWaterMark: writableHighWaterMark,
  });

  // --- метрики backpressure ---
  // pipe() викликає readable.pause(), коли dest.write() повернув false,
  // і readable.resume() після 'drain'. Рахуємо паузи — це і є backpressure-події.
  let pauseCount = 0;
  let peakWritableLength = 0;
  let peakRss = 0;
  readable.on('pause', () => { pauseCount += 1; });

  const sampler = setInterval(() => {
    peakWritableLength = Math.max(peakWritableLength, writable.writableLength, transform.writableLength);
    peakRss = Math.max(peakRss, process.memoryUsage().rss);
    onProgress({
      bytesRead: readable.bytesRead,
      totalBytes: size,
      percent: size ? Math.round((readable.bytesRead / size) * 1000) / 10 : 100,
      linesTotal: transform.stats.linesTotal,
      linesMatched: transform.stats.linesMatched,
      backpressureEvents: pauseCount,
    });
  }, 250);

  try {
    await new Promise((resolve, reject) => {
      const fail = (err) => {
        readable.destroy();
        transform.destroy();
        writable.destroy();
        reject(err);
      };
      readable.on('error', fail);
      transform.on('error', fail);
      writable.on('error', fail);
      writable.on('finish', resolve);

      readable.pipe(transform).pipe(writable);
    });
  } finally {
    clearInterval(sampler);
  }

  return {
    ...transform.stats,
    inputBytes: size,
    chunkSize,
    writableHighWaterMark,
    writeDelayMs,
    backpressureEvents: pauseCount,
    peakWritableLengthBytes: peakWritableLength,
    peakRssMb: Math.round(peakRss / 1024 / 1024),
    durationMs: Date.now() - startedAt,
  };
}

module.exports = { processFile };
