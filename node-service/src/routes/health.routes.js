const { Router } = require('express');
const methodNotAllowed = require('../middleware/methodNotAllowed');
const { SERVICE_NAME } = require('../config');

const router = Router();

// GET /health -> 200 (використовується для перевірки, що сервер живий)
router
  .route('/')
  .get((req, res) => {
    res.status(200).json({
      status: 'ok',
      service: SERVICE_NAME,
      runtime: `node ${process.version}`,
      uptimeSeconds: Math.floor(process.uptime()),
      timestamp: new Date().toISOString(),
    });
  })
  .all(methodNotAllowed(['GET']));

module.exports = router;
