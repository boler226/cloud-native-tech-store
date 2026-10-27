// Router: лише endpoint-и, HTTP-методи, зв'язок із контролером і підключення middleware.
// Жодної бізнес-логіки.
const { Router } = require('express');
const methodNotAllowed = require('../middleware/methodNotAllowed');
const { validateIdParam, validateUserBody } = require('../middleware/validation');

function createUserRouter(controller) {
  const router = Router();

  router
    .route('/')
    .get(controller.list)
    .post(validateUserBody, controller.create)
    .all(methodNotAllowed(['GET', 'POST']));

  router
    .route('/:id')
    .get(validateIdParam, controller.getById)
    .put(validateIdParam, validateUserBody, controller.update)
    .delete(validateIdParam, controller.remove)
    .all(methodNotAllowed(['GET', 'PUT', 'DELETE']));

  return router;
}

module.exports = createUserRouter;
