// node scripts/ws-client.js [jobId|*]
const WebSocket = require('ws');

const ws = new WebSocket(`ws://localhost:${process.env.PORT || 3000}/ws`);
const target = process.argv[2] || '*';

ws.on('open', () => ws.send(JSON.stringify({ type: 'subscribe', jobId: target })));
ws.on('message', (data) => console.log(data.toString()));
ws.on('close', () => console.log('closed'));
