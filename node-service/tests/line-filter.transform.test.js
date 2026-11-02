const { describe, it } = require('node:test');
const assert = require('node:assert/strict');
const { Readable } = require('node:stream');
const { pipeline } = require('node:stream/promises');
const LineFilterTransform = require('../src/streams/line-filter.transform');

async function run(chunks, pattern) {
  const t = new LineFilterTransform({ pattern });
  let out = '';
  await pipeline(Readable.from(chunks), t, async (source) => {
    for await (const c of source) out += c.toString();
  });
  return { out, stats: t.stats };
}

describe('LineFilterTransform', () => {
  it('склеює рядок, розірваний межею чанків', async () => {
    const { out, stats } = await run([Buffer.from('1,ERR'), Buffer.from('OR,a\n2,INFO,b\n')], 'ERROR');
    assert.equal(out, '1,ERROR,a\n');
    assert.deepEqual([stats.linesTotal, stats.linesMatched], [2, 1]);
  });

  it('не ламає багатобайтові символи на межі чанків', async () => {
    const buf = Buffer.from('1,ERROR,привіт\n');
    const { out } = await run([buf.subarray(0, 12), buf.subarray(12)], 'ERROR');
    assert.equal(out, '1,ERROR,привіт\n');
  });

  it('обробляє останній рядок без \\n (flush)', async () => {
    const { out } = await run([Buffer.from('a,ERROR\nb,ERROR')], 'ERROR');
    assert.equal(out, 'a,ERROR\nb,ERROR\n');
  });
});
