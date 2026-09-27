const createApp = require('./app');
const { PORT, SERVICE_NAME } = require('./config');

const server = createApp().listen(PORT, () => {
  console.log(`${SERVICE_NAME} listening on http://localhost:${PORT}`);
});

// Коректне завершення по Ctrl+C / SIGTERM.
function shutdown(signal) {
  console.log(`\n${signal} received, shutting down...`);
  server.close(() => process.exit(0));
  setTimeout(() => process.exit(1), 5000).unref();
}
process.on('SIGINT', () => shutdown('SIGINT'));
process.on('SIGTERM', () => shutdown('SIGTERM'));
