const express = require('express');

// --- products (лабораторна 1) ---
const ProductRepository = require('./repositories/product.repository');
const ProductService = require('./services/product.service');
const ProductController = require('./controllers/product.controller');
const createProductRouter = require('./routes/product.routes');
const seed = require('./data/seed');

// --- users (лабораторна 4: багаторівнева архітектура) ---
const InMemoryDataStore = require('./datastore/in-memory.datastore');
const InMemoryUserRepository = require('./repositories/in-memory-user.repository');
const UserService = require('./services/user.service');
const UserController = require('./controllers/user.controller');
const createUserRouter = require('./routes/user.routes');
const userSeed = require('./data/users.seed');

const healthRouter = require('./routes/health.routes');
const ConcurrencyController = require('./controllers/concurrency.controller');
const createConcurrencyRouter = require('./routes/concurrency.routes');
const createRequestLogger = require('./middleware/requestLogger');
const notFound = require('./middleware/notFound');
const errorHandler = require('./middleware/errorHandler');

/**
 * Composition root: єдине місце, де обираються конкретні реалізації.
 * Щоб замінити in-memory сховище на БД, достатньо передати інший userRepository —
 * UserService, контролер і маршрути не змінюються.
 *
 * @param {object}   [options]
 * @param {object}   [options.userRepository] реалізація контракту UserRepository
 * @param {Function} [options.logger]         функція виводу логів запитів (за замовч. console.log)
 */
function createApp({ userRepository, logger } = {}) {
  // Products: repository -> service -> controller
  const productController = new ProductController(new ProductService(new ProductRepository(seed)));

  // Users: DataStore -> Repository -> Service -> Controller
  const usersRepo =
    userRepository ?? new InMemoryUserRepository(new InMemoryDataStore(), { seed: userSeed });
  const userController = new UserController(new UserService(usersRepo));

  const concurrencyController = new ConcurrencyController();

  const app = express();
  app.disable('x-powered-by');

  app.use(createRequestLogger(logger));
  app.use(express.json({ limit: '1mb' }));

  app.use('/health', healthRouter);
  app.use('/products', createProductRouter(productController));
  app.use('/api/users', createUserRouter(userController));
  // /io та /cpu для лабораторної 2 (дослідження моделей конкурентності)
  app.use('/', createConcurrencyRouter(concurrencyController));

  app.use(notFound);
  app.use(errorHandler);

  return app;
}

module.exports = createApp;
