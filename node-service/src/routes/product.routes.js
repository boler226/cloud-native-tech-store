const { Router } = require('express');
const methodNotAllowed = require('../middleware/methodNotAllowed');

function createProductRouter(controller) {
  const router = Router();

  router
    .route('/')
    .get(controller.list)
    .post(controller.create)
    .all(methodNotAllowed(['GET', 'POST']));

  router
    .route('/:id')
    .get(controller.getById)
    .put(controller.update)
    .delete(controller.remove)
    .all(methodNotAllowed(['GET', 'PUT', 'DELETE']));

  return router;
}

module.exports = createProductRouter;
