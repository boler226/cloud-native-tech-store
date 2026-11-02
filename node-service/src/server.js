const createApp = require('./app');
const createEventBus = require('./events/event-bus');
const attachWebSocketServer = require('./ws/websocket.server');
const { PORT, SERVICE_NAME, WS_PATH } = require('./config');

// Одна шина подій на процес: stream-модуль публікує, WebSocket-модуль розсилає.
const eventBus = createEventBus();
const server = createApp({ eventBus }).listen(PORT, () => {
  console.log(`${SERVICE_NAME} listening on http://localhost:${PORT} (ws: ws://localhost:${PORT}${WS_PATH})`);
});
const ws = attachWebSocketServer(server, eventBus, { path: WS_PATH });

// Коректне завершення по Ctrl+C / SIGTERM.
function shutdown(signal) {
  console.log(`\n${signal} received, shutting down...`);
  ws.close(); // інакше відкриті WS-з'єднання не дадуть server.close() завершитись
  server.close(() => process.exit(0));
  setTimeout(() => process.exit(1), 5000).unref();
}
process.on('SIGINT', () => shutdown('SIGINT'));
process.on('SIGTERM', () => shutdown('SIGTERM'));
