// Смоук-тест /api/users: однакові перевірки для Node- і Go-сервісу (лабораторна 4).
// Запуск (сервіс має бути вже запущений):
//   node scripts/users-smoke-test.mjs http://localhost:3000
//   node scripts/users-smoke-test.mjs http://localhost:8080
const base = (process.argv[2] || 'http://localhost:3000').replace(/\/$/, '');

let passed = 0;
let failed = 0;

async function call(method, path, { json, raw, contentType } = {}) {
  const headers = {};
  let body;
  if (json !== undefined) { body = JSON.stringify(json); headers['Content-Type'] = 'application/json'; }
  if (raw !== undefined) { body = raw; headers['Content-Type'] = contentType ?? 'application/json'; }
  const res = await fetch(base + path, { method, headers, body });
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { /* не JSON */ }
  return { status: res.status, headers: res.headers, body: data };
}

function check(name, cond, extra = '') {
  if (cond) { passed++; console.log(`  ✓ ${name}`); } else { failed++; console.log(`  ✗ ${name} ${extra}`); }
}
const isError = (r, status, code) =>
  r.status === status && r.body?.error === code && typeof r.body?.message === 'string';

// унікальний email, щоб повторні запуски не конфліктували
const email = `smoke-${Date.now()}@example.com`;
const user = { name: 'Smoke Tester', email };

console.log(`Testing ${base}\n`);

let r = await call('GET', '/health');
check('GET /health -> 200', r.status === 200, `(got ${r.status})`);

r = await call('GET', '/api/users');
check('GET /api/users -> 200 + array', r.status === 200 && Array.isArray(r.body), `(got ${r.status})`);

r = await call('POST', '/api/users', { json: user });
check('POST /api/users -> 201', r.status === 201, `(got ${r.status})`);
const id = r.body?.id;
check('POST returns { id, name, email }', Number.isInteger(id) && r.body.name === user.name && r.body.email === email);
check('POST sets Location header', r.headers.get('location') === `/api/users/${id}`);

r = await call('POST', '/api/users', { json: { name: 'Dup', email: email.toUpperCase() } });
check('POST duplicate email (any case) -> 409 email_already_exists', isError(r, 409, 'email_already_exists'), `(got ${r.status} ${JSON.stringify(r.body)})`);

r = await call('POST', '/api/users', { json: {} });
check('POST missing fields -> 400 validation_failed + details', isError(r, 400, 'validation_failed') && r.body.details?.length === 2);

r = await call('POST', '/api/users', { json: { name: 123, email: 'nope' } });
check('POST wrong types / bad email -> 400 validation_failed', isError(r, 400, 'validation_failed'));

r = await call('POST', '/api/users', { raw: '{ broken json' });
check('POST invalid JSON -> 400 invalid_json', isError(r, 400, 'invalid_json'), `(got ${r.status} ${JSON.stringify(r.body)})`);

r = await call('POST', '/api/users', { raw: 'name=x', contentType: 'text/plain' });
check('POST non-JSON content type -> 415', isError(r, 415, 'unsupported_media_type'), `(got ${r.status})`);

r = await call('GET', `/api/users/${id}`);
check('GET /api/users/:id -> 200', r.status === 200 && r.body?.id === id, `(got ${r.status})`);

r = await call('GET', '/api/users/999999');
check('GET unknown id -> 404 user_not_found', isError(r, 404, 'user_not_found') && r.body.message === 'User with id 999999 was not found');

for (const bad of ['abc', '0', '-1', '1.5']) {
  r = await call('GET', `/api/users/${bad}`);
  check(`GET /api/users/${bad} -> 400 invalid_id`, isError(r, 400, 'invalid_id'), `(got ${r.status})`);
}

r = await call('PUT', `/api/users/${id}`, { json: { name: 'Renamed', email } });
check('PUT /api/users/:id -> 200 + updated', r.status === 200 && r.body?.name === 'Renamed', `(got ${r.status})`);

r = await call('PUT', '/api/users/999999', { json: user });
check('PUT unknown id -> 404', isError(r, 404, 'user_not_found'), `(got ${r.status})`);

r = await call('PUT', `/api/users/${id}`, { json: { name: 'x', email: 'jane@example.com' } });
check("PUT someone else's email -> 409", isError(r, 409, 'email_already_exists'), `(got ${r.status})`);

r = await call('DELETE', `/api/users/${id}`);
check('DELETE /api/users/:id -> 204', r.status === 204, `(got ${r.status})`);

r = await call('DELETE', `/api/users/${id}`);
check('DELETE again -> 404', isError(r, 404, 'user_not_found'), `(got ${r.status})`);

r = await call('GET', `/api/users/${id}`);
check('GET after DELETE -> 404', r.status === 404, `(got ${r.status})`);

r = await call('PATCH', '/api/users');
check('PATCH /api/users -> 405 + Allow', isError(r, 405, 'method_not_allowed') && (r.headers.get('allow') || '').includes('GET'));

r = await call('GET', '/api/nope');
check('GET unknown route -> 404', r.status === 404 && typeof r.body?.error === 'string');

console.log(`\n${passed} passed, ${failed} failed`);
process.exit(failed ? 1 : 0);
