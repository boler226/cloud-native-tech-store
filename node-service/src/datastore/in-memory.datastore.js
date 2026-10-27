// DataStore — найнижчий шар: "сире" сховище записів за числовим id.
// Репозиторій ховає від сервісу, що саме за цим стоїть (Map, файл, SQL, MongoDB...).
class InMemoryDataStore {
  #items = new Map();
  #sequence = 0;

  /** Видає наступний унікальний id (аналог AUTO_INCREMENT). */
  nextId() {
    this.#sequence += 1;
    return this.#sequence;
  }

  get(id) {
    const record = this.#items.get(id);
    return record ? { ...record } : undefined; // копія — зовнішній код не змінить сховище
  }

  set(id, record) {
    this.#items.set(id, { ...record });
    if (id > this.#sequence) this.#sequence = id;
  }

  delete(id) {
    return this.#items.delete(id);
  }

  values() {
    return [...this.#items.values()].map((r) => ({ ...r }));
  }
}

module.exports = InMemoryDataStore;
