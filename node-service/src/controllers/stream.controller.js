const fs = require('node:fs');
const { pipeline } = require('node:stream');
const HttpError = require('../errors/HttpError');

class StreamController {
  constructor(service) {
    this.service = service;
  }

  // POST /api/streams/jobs  { input, pattern, chunkSizeKb?, writableHighWaterMarkKb?, writeDelayMs? } -> 202
  createJob = async (req, res) => {
    const job = await this.service.start(req.body);
    res.status(202).location(`/api/streams/jobs/${job.id}`).json(job);
  };

  // GET /api/streams/jobs
  listJobs = (req, res) => {
    res.status(200).json(this.service.list());
  };

  // GET /api/streams/jobs/:id
  getJob = (req, res) => {
    res.status(200).json(this.service.get(req.params.id));
  };

  // GET /api/streams/jobs/:id/result
  // HTTP-відповідь теж Writable: файл віддається потоком, backpressure від повільного
  // клієнта автоматично зупиняє читання з диска (pipeline це гарантує).
  downloadResult = (req, res, next) => {
    const job = this.service.get(req.params.id);
    if (job.status !== 'completed') {
      throw new HttpError(409, `Job is ${job.status}, result is not available`);
    }
    res.type('text/csv; charset=utf-8');
    pipeline(fs.createReadStream(this.service.getOutputPath(job.id)), res, (err) => {
      if (err && !res.headersSent) next(err);
    });
  };
}

module.exports = StreamController;
