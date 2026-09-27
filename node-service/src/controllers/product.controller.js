// HTTP-шар: бере дані із запиту, викликає сервіс, формує відповідь з потрібним кодом.
const parseId = require('../utils/parseId');

class ProductController {
  constructor(service) {
    this.service = service;
  }

  // GET /products?category=Laptops -> 200
  list = (req, res) => {
    const category = typeof req.query.category === 'string' ? req.query.category : undefined;
    res.status(200).json(this.service.list(category));
  };

  // GET /products/:id -> 200 | 400 | 404
  getById = (req, res) => {
    res.status(200).json(this.service.getById(parseId(req.params.id)));
  };

  // POST /products -> 201 + Location | 400
  create = (req, res) => {
    const product = this.service.create(req.body ?? {});
    res.status(201).location(`/products/${product.id}`).json(product);
  };

  // PUT /products/:id -> 200 | 400 | 404
  update = (req, res) => {
    const id = parseId(req.params.id);
    res.status(200).json(this.service.update(id, req.body ?? {}));
  };

  // DELETE /products/:id -> 204 | 400 | 404
  remove = (req, res) => {
    this.service.remove(parseId(req.params.id));
    res.status(204).end();
  };
}

module.exports = ProductController;
