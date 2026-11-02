const { Transform } = require('node:stream');
const { StringDecoder } = require('node:string_decoder');

/**
 * Transform: Buffer-чанки -> рядки CSV. Пропускає лише рядки, що містять `pattern`,
 * і веде статистику. Коректно склеює рядки, розірвані межею чанків.
 * Підстрока (includes), а не RegExp — щоб користувач не міг подати ReDoS-шаблон.
 */
class LineFilterTransform extends Transform {
  constructor({ pattern, highWaterMark } = {}) {
    super({ highWaterMark }); // буфер на вхід і вихід (writable/readable HWM)
    this.pattern = pattern;
    this.decoder = new StringDecoder('utf8'); // не ламає символи на межі чанків
    this.leftover = '';
    this.stats = { chunksIn: 0, bytesIn: 0, bytesOut: 0, linesTotal: 0, linesMatched: 0 };
  }

  _processText(text, isLast) {
    const lines = text.split('\n');
    // Останній елемент — потенційно неповний рядок; відкладаємо до наступного чанка.
    this.leftover = isLast ? '' : lines.pop();
    let out = '';
    for (const line of lines) {
      if (line === '' && isLast) continue;
      this.stats.linesTotal += 1;
      if (line.includes(this.pattern)) {
        this.stats.linesMatched += 1;
        out += `${line}\n`;
      }
    }
    return out;
  }

  _transform(chunk, _encoding, callback) {
    this.stats.chunksIn += 1;
    this.stats.bytesIn += chunk.length;
    const out = this._processText(this.leftover + this.decoder.write(chunk), false);
    if (out) {
      this.stats.bytesOut += Buffer.byteLength(out);
      this.push(out); // якщо push() повертає false — Node сам призупинить upstream
    }
    callback();
  }

  _flush(callback) {
    const out = this._processText(this.leftover + this.decoder.end(), true);
    if (out) {
      this.stats.bytesOut += Buffer.byteLength(out);
      this.push(out);
    }
    callback();
  }
}

module.exports = LineFilterTransform;
