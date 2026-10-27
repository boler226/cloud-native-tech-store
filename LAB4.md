# Лабораторна робота 4. Багаторівнева архітектура серверного застосунку

Продовження [лабораторної 1](README.md). До обох сервісів (Node.js і Go) додано ресурс **користувачі**
(`/api/users`), побудований за багаторівневою архітектурою:

```
HTTP Request
     │
     ▼
 ┌────────┐   endpoint-и, HTTP-методи, підключення middleware. Без бізнес-логіки.
 │ Router │
 └───┬────┘
     ▼
 ┌────────────┐  Middleware: logging, validation (id + тіло запиту)
 └───┬────────┘  Некоректний запит зупиняється тут.
     ▼
 ┌────────────┐   HTTP <-> бізнес-логіка: бере перевірені дані, викликає Service,
 │ Controller │   формує відповідь. Не працює зі сховищем.
 └───┬────────┘
     ▼
 ┌─────────┐      Бізнес-логіка (унікальність email, нормалізація, інваріанти).
 │ Service │      Не знає про HTTP і про конкретне сховище.
 └───┬─────┘
     │  абстракція (interface / контракт)
     ▼
 ┌────────────┐   Доступ до даних. Ховає деталі сховища.
 │ Repository │
 └───┬────────┘
     ▼
 ┌───────────┐    Найнижчий шар: "сирі" записи (тут — in-memory).
 │ DataStore │
 └───────────┘
```

## Що було, а що додано

| | Лабораторна 1 (товари, `/products`) | Лабораторна 4 (користувачі, `/api/users`) |
|---|---|---|
| Розділення Router/Controller/Service/Repository | є | є |
| Шар **DataStore** | немає (Map усередині репозиторію) | **є** |
| Service залежить від **абстракції** репозиторію | ні (від конкретного класу/типу) | **так** (контракт у Node, `interface` у Go) |
| Коди помилок `user_not_found`, `email_already_exists` тощо | ні | **так** |
| Центральна обробка помилок | частково | **так**, єдина точка в обох сервісах |
| Middleware валідації | ні | **так** |
| Тести Service та API | ні | **так** |

Код товарів не переписувався: він працює як раніше. Єдина зміна для нього — формат тіла помилки (див. нижче).

## Де що лежить

### Node.js (`node-service/`)

| Рівень | Файли |
|---|---|
| Router | `src/routes/user.routes.js` |
| Middleware | `src/middleware/requestLogger.js` (№1), `validation.js` (№2), `errorHandler.js` |
| Controller | `src/controllers/user.controller.js` |
| Service | `src/services/user.service.js` |
| Repository: контракт | `src/repositories/user.repository.js` |
| Repository: реалізація | `src/repositories/in-memory-user.repository.js` |
| DataStore | `src/datastore/in-memory.datastore.js` |
| Model + правила валідації | `src/models/user.js` |
| Помилки | `src/errors/app-errors.js`, `src/errors/error-mapper.js` |
| Composition root | `src/app.js` |
| Тести | `tests/` |

### Go (`go-service/`)

| Рівень | Файли |
|---|---|
| Router | `internal/router/router.go` |
| Middleware | `internal/middleware/middleware.go` (logging, recover), `validation.go` |
| Controller | `internal/controller/user.go` |
| Service + `interface UserRepository` | `internal/service/user.go` |
| Repository: реалізація | `internal/repository/user.go` |
| DataStore | `internal/datastore/memory.go` |
| Model + доменні помилки | `internal/model/user.go`, `errors.go` |
| Центральна обробка помилок | `internal/httpx/errors.go` |
| Composition root | `internal/app/app.go` |
| Тести | `tests/` |

## Відповідальності рівнів

**Router** лише пов'язує шлях і метод із контролером та підключає middleware, наприклад
`PUT /api/users/:id → validateId → validateBody → controller.update`.

**Middleware валідації** перевіряє `id` (додатне ціле), `Content-Type`, що тіло є коректним JSON-об'єктом,
обов'язкові поля, типи, порожні значення, довжину та формат email. Некоректний запит отримує відповідь
тут і **не доходить** ні до контролера, ні до сервісу, ні до репозиторію (є тест, що сховище лишається порожнім).

**Controller** бере вже перевірені дані, викликає сервіс і формує відповідь (201 + `Location`, 204 тощо).
Помилки сервісу він не форматує: вони потрапляють у **єдине** місце перетворення на HTTP-статус
(Node: `errorHandler` + `error-mapper`; Go: `httpx.Wrap` + `httpx.ErrorResponse`). Так немає дублювання
логіки помилок у кожному endpoint.

**Service** містить бізнес-правила й жодного HTTP:
- email унікальний, **без урахування регістру**;
- `name` обрізається по краях, `email` приводиться до нижнього регістру;
- змінювати email на власний можна, на чужий — ні (409);
- сервіс сам перевіряє свої інваріанти, навіть якщо його викликали в обхід HTTP.

**Repository** ховає деталі сховища; **DataStore** — "сирі" записи (`Map` / `map` + mutex), наскрізний лічильник id.
Унікальність email додатково гарантується атомарно на рівні репозиторію (аналог `UNIQUE`-індексу в БД).

## Ізоляція Service від Repository

Сервіс отримує репозиторій через конструктор і працює лише з абстракцією.

- **Go.** `service.UserRepository` — це `interface` (оголошений на боці споживача, у пакеті `service`).
  Пакет `repository` **не імпортує** `service`. `InMemoryUserRepository` просто реалізує інтерфейс. У тестах є
  друга, зовсім інша реалізація (`sliceRepo` на зрізі) і перевірка на етапі компіляції:
  `var _ service.UserRepository = (*repository.InMemoryUserRepository)(nil)`.
- **Node.js.** Інтерфейсів немає, тому контракт описано в `user.repository.js` (базовий клас + функція
  `assertUserRepository`, що перевіряє наявність усіх методів). `UserService` викликає її в конструкторі.
  У тестах є друга реалізація (`ArrayUserRepository`) і сервіс гоняється з обома.

Заміна `InMemoryRepository` на `DatabaseRepository` — це нова реалізація того самого контракту та один рядок
у composition root (`createApp({ userRepository })` у Node, `app.Options{UserRepository: ...}` у Go).
`UserService`, контролер і маршрути не змінюються. Для Go це виглядає так:

```go
type SQLUserRepository struct{ db *sql.DB }

func (r *SQLUserRepository) FindByID(id int) (model.User, error) {
    // SELECT id, name, email FROM users WHERE id = $1
    // sql.ErrNoRows -> return model.User{}, &model.UserNotFoundError{ID: id}
}
// ... решта методів інтерфейсу service.UserRepository
```

> Важливо для будь-якої нової реалізації: при "не знайдено" повертати `*model.UserNotFoundError`, а при порушенні
> унікальності email — `*model.EmailTakenError` (у Node — `ConflictError`), бо саме за цими типами
> центральний обробник визначає статус.

## API

| Метод | Шлях | Успіх | Помилки |
|---|---|---|---|
| GET | `/health` | 200 | — |
| GET | `/api/users` | 200 | — |
| GET | `/api/users/:id` | 200 | 400 `invalid_id`, 404 `user_not_found` |
| POST | `/api/users` | 201 + `Location` | 400, 409, 413, 415 |
| PUT | `/api/users/:id` | 200 | 400, 404, 409, 413, 415 |
| DELETE | `/api/users/:id` | 204 | 400 `invalid_id`, 404 |

Приклад:

```http
POST /api/users
Content-Type: application/json

{ "name": "John Doe", "email": "john@example.com" }
```
```json
{ "id": 3, "name": "John Doe", "email": "john@example.com" }
```

При старті вже є 2 користувачі (`john@example.com`, `jane@example.com`), тому на `GET /api/users` одразу є що показати.

### Валідація

| Що перевіряється | Приклад помилки |
|---|---|
| Обов'язковість полів | `{ "field": "name", "message": "is required" }` |
| Типи даних | `name: 123` → `must be a string` |
| Порожні значення | `"   "` → `must not be empty` |
| Довжина / формат | name ≤ 100 символів, email ≤ 254 і `must be a valid email address` |
| Коректність id | `abc`, `0`, `-1`, `1.5`, `01` → `invalid_id` |
| Дублювання | email зайнятий (будь-який регістр) → 409 |

Усі помилки полів повертаються **одразу**, а не лише перша.

## Помилки

Єдиний формат для всього API:

```json
{ "error": "user_not_found", "message": "User with id 42 was not found" }
```
```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [ { "field": "name", "message": "is required" } ]
}
```

| Статус | Коди `error` | Коли |
|---|---|---|
| 400 | `validation_failed`, `invalid_id`, `invalid_json` | некоректні дані, id, зламаний JSON |
| 404 | `user_not_found`, `route_not_found` | немає користувача / маршруту |
| 405 | `method_not_allowed` (+ заголовок `Allow`) | метод не підтримується |
| 409 | `email_already_exists` | email уже зайнятий |
| 413 | `payload_too_large` | тіло більше 1 MB |
| 415 | `unsupported_media_type` | тіло не JSON |
| 500 | `internal_error` | непередбачена помилка |

Для 500 клієнту віддається лише загальне повідомлення; деталі (наприклад, текст помилки БД) потрапляють тільки в
лог — це перевіряє окремий тест.

> **Зміна для товарів.** Раніше помилки `/products` мали вигляд `{ "error": "Product not found" }`. Тепер вони в тому ж
> форматі `{ "error": "not_found", "message": "Product not found" }`. Статус-коди не змінились, `smoke-test.mjs`
> лабораторної 1 далі проходить (25/25).

## Middleware

1. **Request Logging** (`requestLogger.js` / `middleware.go`): метод, шлях, статус, тривалість.
   Формат: `GET /api/users/10 404 0.7ms`.
2. **Request Validation** (`validation.js` / `validation.go`): `validateIdParam` і `validateUserBody` — спільна
   логіка для кількох endpoint-ів (id використовується в GET/PUT/DELETE, тіло — у POST/PUT).

Додатково в Go є `Recover` (panic → 500), а в Node — центральний `errorHandler`.

## Тести

**Service-тести працюють без HTTP-сервера.** Одні й ті самі тести проганяються на двох різних реалізаціях репозиторію.

```powershell
# Node.js: вбудований test runner, нових залежностей немає
cd node-service
npm install
npm test

# Go
cd go-service
go test ./...
go test -race ./...     # потрібні cgo і gcc, як і в лабораторній 3
```

| Рівень | Що перевіряється |
|---|---|
| **Service** | створення (id, нормалізація), пошук, список, оновлення, видалення; помилки (not found, дублікат email, валідація, некоректний id); бізнес-правила (регістр email, свій email при оновленні, звільнення email після видалення); помилка сховища не маскується під доменну; відмова працювати з об'єктом, що не виконує контракт |
| **API** | успішні GET (список і один), POST (201, `Location`, нормалізація), некоректний POST (поля, типи, зламаний JSON, порожнє тіло, масив, 415, 413), GET неіснуючого (404), некоректні id (400), PUT (200/404/400/409), DELETE (204 → 404), 405 + `Allow`, 404 маршруту, єдиний формат помилок, 500 без витоку деталей, робота з іншою реалізацією репозиторію, формат логів |

Результат у середовищі розробки: **Node.js — 70 тестів**, **Go — 32 тести** (54 разом із підтестами), усі проходять,
Go — також під `-race`. Крім того, перевірено, що тести справді ловлять помилки: при навмисному псуванні коду
(наприклад, 409 → 400 або вилучення валідації з маршруту) відповідні тести падають.

### Перевірка обох сервісів однаковими запитами

```powershell
# запустіть обидва сервіси, потім:
node scripts/users-smoke-test.mjs http://localhost:3000
node scripts/users-smoke-test.mjs http://localhost:8080
```

Очікується `24 passed, 0 failed` для кожного. Це доказ, що Node і Go реалізують **однаковий** API.
Для ручної перевірки є `http-requests/users.http` (перемикач оточень `node` / `go` зверху файлу).

## Запуск

```powershell
cd node-service
npm install
npm start               # http://localhost:3000

cd go-service
go run ./cmd/server     # http://localhost:8080
```

## Нові та змінені файли

```
Node.js
  node-service/src/errors/app-errors.js                  (нов.)
  node-service/src/errors/error-mapper.js                (нов.)
  node-service/src/models/user.js                        (нов.)
  node-service/src/datastore/in-memory.datastore.js      (нов.)
  node-service/src/repositories/user.repository.js       (нов.)
  node-service/src/repositories/in-memory-user.repository.js  (нов.)
  node-service/src/services/user.service.js              (нов.)
  node-service/src/controllers/user.controller.js        (нов.)
  node-service/src/middleware/validation.js              (нов.)
  node-service/src/routes/user.routes.js                 (нов.)
  node-service/src/data/users.seed.js                    (нов.)
  node-service/tests/helpers.js, user.service.test.js, user.api.test.js  (нов.)
  node-service/README.md                                 (нов.)
  node-service/src/app.js                                (змін.)
  node-service/src/middleware/errorHandler.js            (змін.)
  node-service/src/middleware/notFound.js                (змін.)
  node-service/src/middleware/methodNotAllowed.js        (змін.)
  node-service/src/middleware/requestLogger.js           (змін.)
  node-service/package.json                              (змін.: скрипт "test")

Go
  go-service/internal/model/user.go, errors.go           (нов.)
  go-service/internal/datastore/memory.go                (нов.)
  go-service/internal/repository/user.go                 (нов.)
  go-service/internal/service/user.go                    (нов.)
  go-service/internal/controller/user.go                 (нов.)
  go-service/internal/httpx/errors.go, context.go        (нов.)
  go-service/internal/middleware/validation.go           (нов.)
  go-service/internal/app/app.go                         (нов.)
  go-service/internal/seed/users.go                      (нов.)
  go-service/tests/helpers_test.go, user_service_test.go, user_api_test.go  (нов.)
  go-service/README.md                                   (нов.)
  go-service/cmd/server/main.go                          (змін.: збірка винесена в app.New)
  go-service/internal/router/router.go                   (змін.: маршрути /api/users)
  go-service/internal/httpx/json.go                      (змін.: єдиний формат помилки)
  go-service/internal/middleware/middleware.go           (змін.: формат логу)

Спільне
  http-requests/users.http                               (нов.)
  scripts/users-smoke-test.mjs                           (нов.)
  LAB4.md                                                (нов., цей файл)
  README.md                                              (змін.: посилання на LAB4.md)
```
