# node-service (Node.js + Express)

Багаторівнева архітектура (лабораторна 4):

```
Router -> Middleware (validation) -> Controller -> Service -> Repository -> DataStore
```

| Рівень | Файли |
|---|---|
| Router | `src/routes/user.routes.js` |
| Middleware | `src/middleware/requestLogger.js`, `validation.js`, `errorHandler.js` |
| Controller | `src/controllers/user.controller.js` |
| Service | `src/services/user.service.js` |
| Repository | `src/repositories/user.repository.js` (контракт), `in-memory-user.repository.js` (реалізація) |
| DataStore | `src/datastore/in-memory.datastore.js` |
| Model | `src/models/user.js` |
| Помилки | `src/errors/app-errors.js`, `error-mapper.js` |
| Composition root | `src/app.js` |

```bash
npm install
npm start          # http://localhost:3000
npm test           # node --test: тести Service та API (без зовнішніх залежностей)
```

Повний опис — у [LAB4.md](../LAB4.md).
