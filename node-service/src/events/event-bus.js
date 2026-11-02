const { EventEmitter } = require('node:events');

// Розв'язує зв'язок: стрім-модуль публікує події, WebSocket-модуль їх розсилає.
// Єдина подія 'job:event' з payload { type, jobId, ... }.
function createEventBus() {
  const bus = new EventEmitter();
  bus.setMaxListeners(50);
  return bus;
}

module.exports = createEventBus;
