const { Router } = require('express');
const methodNotAllowed = require('../middleware/methodNotAllowed');

function createStreamRouter(controller) {
  const router = Router();

  router.route('/jobs')
    .post(controller.createJob)
    .get(controller.listJobs)
    .all(methodNotAllowed(['GET', 'POST']));
  router.route('/jobs/:id').get(controller.getJob).all(methodNotAllowed(['GET']));
  router.route('/jobs/:id/result').get(controller.downloadResult).all(methodNotAllowed(['GET']));

  return router;
}

module.exports = createStreamRouter;
