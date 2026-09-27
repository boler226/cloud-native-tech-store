const express = require('express');

const ProductRepository = require('./repositories/product.repository');
const ProductService = require('./services/product.service');
const ProductController = require('./controllers/product.controller');
const createProductRouter = require('./routes/product.routes');
const healthRouter = require('./routes/health.routes');
const ConcurrencyController = require('./controllers/concurrency.controller');
const createConcurrencyRouter = require('./routes/concurrency.routes');
const requestLogger = require('./middleware/requestLogger');
const notFound = require('./middleware/notFound');
const errorHandler = require('./middleware/errorHandler');
const seed = require('./data/seed');

// Збираємо застосунок: repository -> service -> controller -> router.
function createApp() {
  const repository = new ProductRepository(seed);
  const service = new ProductService(repository);
  const controller = new ProductController(service);
  const concurrencyController = new ConcurrencyController();

  const app = express();
  app.disable('x-powered-by');

  app.use(requestLogger);
  app.use(express.json({ limit: '1mb' }));

  app.use('/health', healthRouter);
  app.use('/products', createProductRouter(controller));
  // /io та /cpu для лабораторної 2 (дослідження моделей конкурентності)
  app.use('/', createConcurrencyRouter(concurrencyController));

  app.use(notFound);
  app.use(errorHandler);

  return app;
}

module.exports = createApp;
