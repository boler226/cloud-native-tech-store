const { WebSocketServer, WebSocket } = require('ws');
const crypto = require('node:crypto');

const HEARTBEAT_MS = 30_000;
const MAX_BUFFERED_BYTES = 1024 * 1024; // поріг backpressure для повільного WS-клієнта

/**
 * Підключається до існуючого http.Server через подію 'upgrade' — той самий порт, що й Express.
 *
 * Протокол (JSON):
 *  клієнт -> сервер: { type:'subscribe', jobId:'<id>'|'*' } | { type:'unsubscribe', jobId }
 *                    { type:'broadcast', text } | { type:'ping' }
 *  сервер -> клієнт: { type:'welcome', clientId } | { type:'subscribed', jobId } | { type:'pong' }
 *                    { type:'broadcast', from, text } | { type:'error', message }
 *                    { type:'job:started'|'job:progress'|'job:completed'|'job:failed', jobId, ... }
 */
function attachWebSocketServer(httpServer, eventBus, { path = '/ws' } = {}) {
  const wss = new WebSocketServer({ noServer: true, maxPayload: 64 * 1024 });
  const clients = new Map(); // ws -> { id, subs:Set<string> }

  httpServer.on('upgrade', (req, socket, head) => {
    const { pathname } = new URL(req.url, 'http://localhost');
    if (pathname !== path) {
      socket.write('HTTP/1.1 404 Not Found\r\nConnection: close\r\n\r\n');
      socket.destroy();
      return;
    }
    wss.handleUpgrade(req, socket, head, (ws) => wss.emit('connection', ws, req));
  });

  function send(ws, payload) {
    if (ws.readyState !== WebSocket.OPEN) return;
    // Backpressure на рівні WS: якщо клієнт не встигає читати, дропаємо "дрібні" progress-події,
    // щоб не накопичувати їх у пам'яті сервера. Важливі події (completed/failed) шлемо завжди.
    if (ws.bufferedAmount > MAX_BUFFERED_BYTES && payload.type === 'job:progress') return;
    ws.send(JSON.stringify(payload));
  }

  function broadcast(payload, filter = () => true) {
    for (const [ws, client] of clients) {
      if (filter(client)) send(ws, payload);
    }
  }

  wss.on('connection', (ws) => {
    const client = { id: crypto.randomUUID(), subs: new Set() };
    clients.set(ws, client);
    ws.isAlive = true;
    ws.on('pong', () => { ws.isAlive = true; });

    send(ws, { type: 'welcome', clientId: client.id, clients: clients.size });

    ws.on('message', (raw) => {
      let msg;
      try {
        msg = JSON.parse(raw.toString());
      } catch {
        return send(ws, { type: 'error', message: 'Invalid JSON' });
      }
      switch (msg?.type) {
        case 'subscribe':
          if (typeof msg.jobId !== 'string') return send(ws, { type: 'error', message: 'jobId must be a string' });
          client.subs.add(msg.jobId);
          return send(ws, { type: 'subscribed', jobId: msg.jobId });
        case 'unsubscribe':
          client.subs.delete(msg.jobId);
          return send(ws, { type: 'unsubscribed', jobId: msg.jobId });
        case 'broadcast':
          if (typeof msg.text !== 'string' || msg.text.length > 1000) {
            return send(ws, { type: 'error', message: 'text must be a string (max 1000 chars)' });
          }
          return broadcast({ type: 'broadcast', from: client.id, text: msg.text });
        case 'ping':
          return send(ws, { type: 'pong' });
        default:
          return send(ws, { type: 'error', message: 'Unknown message type' });
      }
    });

    ws.on('close', () => clients.delete(ws));
    ws.on('error', () => ws.terminate());
  });

  // Події з стрім-модуля -> підписані клієнти
  const onJobEvent = (event) => {
    broadcast(event, (c) => c.subs.has(event.jobId) || c.subs.has('*'));
  };
  eventBus.on('job:event', onJobEvent);

  // Heartbeat: відсікаємо "мертві" з'єднання (обрив мережі без close-фрейму)
  const heartbeat = setInterval(() => {
    for (const ws of clients.keys()) {
      if (!ws.isAlive) { ws.terminate(); continue; }
      ws.isAlive = false;
      ws.ping();
    }
  }, HEARTBEAT_MS);
  heartbeat.unref();

  return {
    wss,
    clientCount: () => clients.size,
    close() {
      clearInterval(heartbeat);
      eventBus.off('job:event', onJobEvent);
      for (const ws of clients.keys()) ws.terminate();
      wss.close();
    },
  };
}

module.exports = attachWebSocketServer;
