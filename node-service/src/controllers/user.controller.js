// HTTP-шар користувачів: бере вже ПЕРЕВІРЕНІ дані з middleware (res.locals), викликає сервіс
// і формує відповідь. Помилки сервісу не обробляються тут — вони йдуть у центральний errorHandler
// (Express 5 сам передає відхилені промізи в error-middleware).
class UserController {
  constructor(service) {
    this.service = service;
  }

  // GET /api/users -> 200
  list = async (req, res) => {
    res.status(200).json(await this.service.list());
  };

  // GET /api/users/:id -> 200 | 400 | 404
  getById = async (req, res) => {
    res.status(200).json(await this.service.getById(res.locals.id));
  };

  // POST /api/users -> 201 + Location | 400 | 409
  create = async (req, res) => {
    const user = await this.service.create(res.locals.input);
    res.status(201).location(`/api/users/${user.id}`).json(user);
  };

  // PUT /api/users/:id -> 200 | 400 | 404 | 409
  update = async (req, res) => {
    res.status(200).json(await this.service.update(res.locals.id, res.locals.input));
  };

  // DELETE /api/users/:id -> 204 | 400 | 404
  remove = async (req, res) => {
    await this.service.delete(res.locals.id);
    res.status(204).end();
  };
}

module.exports = UserController;
