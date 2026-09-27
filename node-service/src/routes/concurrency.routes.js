const { Router } = require('express');
const methodNotAllowed = require('../middleware/methodNotAllowed');

function createConcurrencyRouter(controller) {
  const router = Router();

  router.route('/io').get(controller.io).all(methodNotAllowed(['GET']));
  router.route('/cpu').get(controller.cpu).all(methodNotAllowed(['GET']));

  return router;
}

module.exports = createConcurrencyRouter;
