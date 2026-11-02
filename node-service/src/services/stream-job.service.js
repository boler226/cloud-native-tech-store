const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const HttpError = require('../errors/HttpError');
const { processFile: defaultProcessor } = require('../streams/stream-processor');

const INPUT_NAME_RE = /^[\w.-]{1,100}$/; // лише ім'я файлу: захист від path traversal

function intField(value, { name, def, min, max, errors }) {
  if (value === undefined) return def;
  if (!Number.isInteger(value) || value < min || value > max) {
    errors.push(`${name} must be an integer in [${min}, ${max}]`);
    return def;
  }
  return value;
}

function validateParams(body) {
  if (typeof body !== 'object' || body === null || Array.isArray(body)) {
    throw new HttpError(400, 'Validation failed', ['request body must be a JSON object']);
  }
  const errors = [];
  if (typeof body.input !== 'string' || !INPUT_NAME_RE.test(body.input) || body.input.includes('..')) {
    errors.push('input is required: file name from storage/input (letters, digits, _ . -)');
  }
  if (typeof body.pattern !== 'string' || body.pattern.length === 0 || body.pattern.length > 100) {
    errors.push('pattern is required: non-empty string (max 100 chars)');
  }
  const chunkSizeKb = intField(body.chunkSizeKb, { name: 'chunkSizeKb', def: 64, min: 1, max: 1024, errors });
  const writableHighWaterMarkKb = intField(body.writableHighWaterMarkKb, {
    name: 'writableHighWaterMarkKb', def: 16, min: 1, max: 1024, errors,
  });
  const writeDelayMs = intField(body.writeDelayMs, { name: 'writeDelayMs', def: 0, min: 0, max: 1000, errors });

  if (errors.length > 0) throw new HttpError(400, 'Validation failed', errors);
  return {
    input: body.input,
    pattern: body.pattern,
    chunkSize: chunkSizeKb * 1024,
    writableHighWaterMark: writableHighWaterMarkKb * 1024,
    writeDelayMs,
  };
}

class StreamJobService {
  constructor({ eventBus, storageDir, maxRunningJobs = 3, processor = defaultProcessor }) {
    this.eventBus = eventBus;
    this.storageDir = storageDir;
    this.maxRunningJobs = maxRunningJobs;
    this.processor = processor;
    this.jobs = new Map(); // id -> { job, outputPath }
  }

  list() {
    return [...this.jobs.values()].map((e) => e.job);
  }

  get(id) {
    const entry = this.jobs.get(id);
    if (!entry) throw new HttpError(404, 'Job not found');
    return entry.job;
  }

  getOutputPath(id) {
    this.get(id);
    return this.jobs.get(id).outputPath;
  }

  async start(body) {
    const params = validateParams(body ?? {});

    const running = this.list().filter((j) => j.status === 'running').length;
    if (running >= this.maxRunningJobs) {
      throw new HttpError(429, `Too many running jobs (max ${this.maxRunningJobs})`);
    }

    const inputPath = path.join(this.storageDir, 'input', params.input);
    try {
      await fs.promises.access(inputPath, fs.constants.R_OK);
    } catch {
      throw new HttpError(404, 'Input file not found in storage/input');
    }

    const id = crypto.randomUUID();
    const outputDir = path.join(this.storageDir, 'output');
    await fs.promises.mkdir(outputDir, { recursive: true });
    const outputPath = path.join(outputDir, `${id}.csv`);

    const job = {
      id,
      status: 'running',
      input: params.input,
      pattern: params.pattern,
      startedAt: new Date().toISOString(),
      finishedAt: null,
      stats: null,
      error: null,
    };
    this.jobs.set(id, { job, outputPath });
    this.eventBus.emit('job:event', { type: 'job:started', jobId: id, input: job.input, pattern: job.pattern });

    // Фонове виконання: HTTP-відповідь (202) не чекає на завершення стріму.
    this.processor({
      inputPath,
      outputPath,
      pattern: params.pattern,
      chunkSize: params.chunkSize,
      writableHighWaterMark: params.writableHighWaterMark,
      writeDelayMs: params.writeDelayMs,
      onProgress: (p) => this.eventBus.emit('job:event', { type: 'job:progress', jobId: id, ...p }),
    })
      .then((stats) => {
        job.status = 'completed';
        job.stats = stats;
        job.finishedAt = new Date().toISOString();
        this.eventBus.emit('job:event', { type: 'job:completed', jobId: id, stats });
      })
      .catch((err) => {
        job.status = 'failed';
        job.error = err.message;
        job.finishedAt = new Date().toISOString();
        this.eventBus.emit('job:event', { type: 'job:failed', jobId: id, error: err.message });
      });

    return job;
  }
}

module.exports = StreamJobService;
