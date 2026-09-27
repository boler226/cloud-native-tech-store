// Шар доступу до даних. Дані живуть у пам'яті (Map) і зникають після перезапуску.
class ProductRepository {
  constructor(seed = []) {
    this.items = new Map();
    this.nextId = 1;
    seed.forEach((item) => this.create(item));
  }

  findAll(category) {
    const all = [...this.items.values()];
    if (!category) return all;
    const wanted = String(category).toLowerCase();
    return all.filter((p) => p.category.toLowerCase() === wanted);
  }

  findById(id) {
    return this.items.get(id) ?? null;
  }

  create(data) {
    const now = new Date().toISOString();
    const product = { id: this.nextId++, ...data, createdAt: now, updatedAt: now };
    this.items.set(product.id, product);
    return product;
  }

  update(id, data) {
    const existing = this.items.get(id);
    if (!existing) return null;
    const updated = {
      id: existing.id,
      ...data,
      createdAt: existing.createdAt,
      updatedAt: new Date().toISOString(),
    };
    this.items.set(id, updated);
    return updated;
  }

  delete(id) {
    return this.items.delete(id);
  }
}

module.exports = ProductRepository;
