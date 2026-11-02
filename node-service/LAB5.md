# Лабораторна 5: Streams, Backpressure, WebSockets

## Запуск
```bash
npm i
npm run gen:data -- 200 big.csv      # тестовий файл 200 МБ -> storage/input/big.csv
npm run demo:backpressure            # таблиця: unsafe vs pipe
npm start

node scripts/ws-client.js            # окремий термінал: слухає події всіх задач

curl -X POST localhost:3000/api/streams/jobs -H "Content-Type: application/json" \
  -d '{"input":"big.csv","pattern":"ERROR","chunkSizeKb":64,"writableHighWaterMarkKb":16,"writeDelayMs":5}'
curl localhost:3000/api/streams/jobs/<id>
curl -o result.csv localhost:3000/api/streams/jobs/<id>/result
```

## HTTP API
| Метод | Шлях | Опис |
|---|---|---|
| POST | /api/streams/jobs | запуск задачі (202) |
| GET | /api/streams/jobs | список задач |
| GET | /api/streams/jobs/:id | статус і статистика |
| GET | /api/streams/jobs/:id/result | стрімінг результату (pipeline -> res) |

## WebSocket (ws://localhost:3000/ws)
Клієнт -> сервер: `subscribe`, `unsubscribe`, `broadcast`, `ping`.
Сервер -> клієнт: `welcome`, `subscribed`, `job:started`, `job:progress`, `job:completed`, `job:failed`, `broadcast`, `pong`, `error`.

## Параметри, що впливають на розмір порцій
- `highWaterMark` у `fs.createReadStream` — розмір читання з диска (за замовчуванням 64 КБ).
- `start`/`end` — діапазон байтів; `encoding` — Buffer -> string.
- `highWaterMark` у Transform/Writable — поріг, після якого `write()` повертає `false`.

## Backpressure
Повільний `SlowFileWritable` -> буфер заповнюється -> `write()` = false -> `pipe()` ставить Readable на паузу ->
після `drain` — `resume()`. Лічильник `backpressureEvents` у статистиці задачі й таблиця демо це підтверджують.
Також: `ws.bufferedAmount` — дроп progress-подій для повільного WS-клієнта.
