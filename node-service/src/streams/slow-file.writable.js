const fs = require('node:fs');
const { Writable } = require('node:stream');

/**
 * Writable, що пише у файл із штучною затримкою на кожен chunk (п. 5.1).
 * Поки _write не викликав callback, нові дані накопичуються у внутрішньому буфері;
 * коли він перевищує highWaterMark, write() повертає false -> pipe() ставить
 * Readable на паузу, а після 'drain' відновлює.
 */
class SlowFileWritable extends Writable {
  constructor(filePath, { delayMs = 0, highWaterMark = 16 * 1024 } = {}) {
    super({ highWaterMark });
    this.delayMs = delayMs;
    this.file = fs.createWriteStream(filePath);
    this.file.on('error', (err) => this.destroy(err));
  }

  _write(chunk, _encoding, callback) {
    const doWrite = () => this.file.write(chunk, callback);
    if (this.delayMs > 0) setTimeout(doWrite, this.delayMs);
    else doWrite();
  }

  _final(callback) {
    this.file.end(callback);
  }

  _destroy(err, callback) {
    this.file.destroy();
    callback(err);
  }
}

module.exports = SlowFileWritable;
