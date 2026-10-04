# Лабораторна робота 1. Базові HTTP-сервіси на Node.js та Go

> Лабораторна 2 (моделі конкурентності, `/io` та `/cpu`) описана окремо в [LAB2.md](LAB2.md).
> Лабораторна 3 (конкурентний pipeline на Go, `go-pipeline/`) описана окремо в [LAB3.md](LAB3.md).

Два **однакові** REST-сервіси для інтернет-магазину техніки з CRUD-операціями над сутністю `Product`
та дані в пам'яті (in-memory storage, після перезапуску скидаються до початкових).

| Сервіс | Технологія | Порт за замовчуванням |
|---|---|---|
| `node-service/` | Node.js + Express 5 | 3000 |
| `go-service/`   | Go 1.22+, лише стандартний `net/http` (без сторонніх залежностей) | 8080 |

## Структура проєкту

```
lab1-tech-store/
├── README.md
├── node-service/                    # Node.js + Express
│   ├── package.json
│   └── src/
│       ├── server.js                # точка входу: запуск + graceful shutdown
│       ├── app.js                   # збірка застосунку (repository -> service -> controller -> router)
│       ├── config.js
│       ├── routes/                  # URL -> методи контролера (product.routes.js, health.routes.js)
│       ├── controllers/             # HTTP-шар: запит/відповідь, статус-коди
│       ├── services/                # бізнес-логіка та валідація
│       ├── repositories/            # in-memory сховище (Map)
│       ├── middleware/              # логер, 404, 405, обробник помилок
│       ├── utils/                   # validation.js, parseId.js
│       ├── errors/HttpError.js
│       └── data/seed.js             # початкові товари
├── go-service/                      # Go, net/http
│   ├── go.mod
│   ├── cmd/server/main.go           # точка входу: запуск + graceful shutdown
│   └── internal/
│       ├── config/                  # налаштування
│       ├── router/                  # маршрути (ServeMux, Go 1.22 патерни "GET /products/{id}")
│       ├── handler/                 # HTTP-шар: запит/відповідь, статус-коди
│       ├── service/                 # бізнес-логіка та валідація
│       ├── repository/              # in-memory сховище (map + RWMutex)
│       ├── model/                   # структури Product / ProductInput
│       ├── middleware/              # логер, recover від panic
│       ├── httpx/                   # хелпери JSON-відповідей
│       └── seed/                    # початкові товари
├── http-requests/                   # готові запити для HTTP Client у WebStorm
│   ├── products.http
│   └── http-client.env.json
└── scripts/smoke-test.mjs           # автотест: однакові перевірки для обох сервісів
```

Архітектура в обох сервісах шарова: **routes/router → controller/handler → service → repository**.
Шар HTTP не знає, як зберігаються дані, а сховище не знає про HTTP.

## API (однаковий для обох сервісів)

Сутність `Product`:

```json
{
  "id": 1,
  "name": "iPhone 15 128GB",
  "category": "Smartphones",
  "brand": "Apple",
  "price": 32999,
  "stock": 12,
  "description": "Смартфон Apple з чипом A16 Bionic",
  "createdAt": "2026-09-21T14:00:00.000Z",
  "updatedAt": "2026-09-21T14:00:00.000Z"
}
```

| Метод | Шлях | Опис | Успіх | Помилки |
|---|---|---|---|---|
| GET | `/health` | Перевірка стану сервера | **200** | — |
| GET | `/products` | Список товарів, фільтр `?category=` | **200** | — |
| GET | `/products/{id}` | Один товар | **200** | 400 (некоректний id), 404 |
| POST | `/products` | Створити товар | **201** + `Location` | 400 (валідація / битий JSON) |
| PUT | `/products/{id}` | Повне оновлення | **200** | 400, 404 |
| DELETE | `/products/{id}` | Видалити | **204** (без тіла) | 400, 404 |

Додатково: **404** для невідомого маршруту, **405** (+ заголовок `Allow`) для непідтримуваного методу,
**413** для надто великого тіла, **500** для непередбачених помилок.

Правила валідації (POST/PUT): `name` (1–100 символів), `category`, `brand` — непорожні рядки;
`price` — число > 0; `stock` — ціле ≥ 0; `description` — необов'язковий рядок.

Формат помилки:

```json
{ "error": "Validation failed", "details": ["price is required and must be a number greater than 0"] }
```

## Запуск

### Node.js (потрібен Node.js 18+)

```bash
cd node-service
npm install
npm start          # http://localhost:3000
# npm run dev      # перезапуск при зміні файлів
```

### Go (потрібен Go 1.22+)

```bash
cd go-service
go run ./cmd/server        # http://localhost:8080
```

Змінити порт: `PORT=4000 npm start` (Linux/macOS), у PowerShell: `$env:PORT=4000; npm start`.

## Перевірка

1. **HTTP Client у WebStorm:** відкрийте `http-requests/products.http`, у верхньому правому куті виберіть
   оточення `node` або `go` і запускайте запити кнопкою ▶.
2. **Автотест** (обидва сервіси мають бути запущені):

```bash
node scripts/smoke-test.mjs http://localhost:3000
node scripts/smoke-test.mjs http://localhost:8080
```

Очікуваний результат для кожного: `25 passed, 0 failed`.

## Налаштування WebStorm

### Node.js
1. Встановіть Node.js LTS з https://nodejs.org
2. `File → Open` → виберіть папку `lab1-tech-store` (НЕ створюйте новий проєкт із шаблонів — код уже готовий).
3. `Settings → Languages & Frameworks → Node.js` — інтерпретатор має підхопитися автоматично.
4. Відкрийте термінал (`Alt+F12`): `cd node-service`, `npm install`, `npm start`.
   Або відкрийте `package.json` і натисніть ▶ біля скрипта `start`.

### Go
WebStorm не вміє Go «з коробки» — потрібен плагін:
1. `Settings → Plugins → Marketplace` → знайдіть **Go** (автор JetBrains) → `Install` → перезапустіть IDE.
2. `Settings → Go → GOROOT` → `+` → `Download…` і виберіть Go 1.22 або новішу
   (або встановіть Go з https://go.dev/dl і вкажіть його папку).
3. `Settings → Go → Go Modules` → увімкніть *Enable Go modules integration*.
4. Відкрийте `go-service/cmd/server/main.go` і натисніть зелений ▶ біля `func main()`.
5. Якщо WebStorm не бачить модуль (файли підсвічені червоним, бо `go.mod` лежить у підпапці), відкрийте
   `go-service` окремим вікном (`File → Open`) або запускайте з термінала: `go run ./cmd/server`.
