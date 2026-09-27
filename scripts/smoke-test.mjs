// Смоук-тест: однакові перевірки для Node- і Go-сервісу.
// Запуск:  node scripts/smoke-test.mjs http://localhost:3000
//          node scripts/smoke-test.mjs http://localhost:8080
const base = (process.argv[2] || 'http://localhost:3000').replace(/\/$/, '');

let passed = 0;
let failed = 0;

async function call(method, path, body, rawBody) {
  const res = await fetch(base + path, {
    method,
    headers: body !== undefined || rawBody !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: rawBody !== undefined ? rawBody : body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let json = null;
  try { json = text ? JSON.parse(text) : null; } catch { /* не JSON */ }
  return { status: res.status, headers: res.headers, json };
}

function check(name, cond, extra = '') {
  if (cond) { passed++; console.log(`  ✓ ${name}`); }
  else { failed++; console.log(`  ✗ ${name} ${extra}`); }
}

const valid = {
  name: 'MacBook Air 13"', category: 'Laptops', brand: 'Apple',
  price: 44999.5, stock: 3, description: 'M3, 16GB, 512GB',
};

console.log(`Testing ${base}\n`);

let r = await call('GET', '/health');
check('GET /health -> 200', r.status === 200, `(got ${r.status})`);
check('GET /health body.status == "ok"', r.json?.status === 'ok');

r = await call('GET', '/products');
check('GET /products -> 200', r.status === 200, `(got ${r.status})`);
check('GET /products returns non-empty array', Array.isArray(r.json) && r.json.length > 0);

r = await call('GET', '/products?category=laptops');
check('GET /products?category=laptops filters (case-insensitive)',
  r.status === 200 && r.json.length > 0 && r.json.every((p) => p.category === 'Laptops'));

r = await call('POST', '/products', valid);
check('POST /products -> 201', r.status === 201, `(got ${r.status})`);
const id = r.json?.id;
check('POST returns created product with id', Number.isInteger(id) && r.json.name === valid.name);
check('POST sets Location header', r.headers.get('location') === `/products/${id}`);

r = await call('POST', '/products', { name: '', price: -5 });
check('POST invalid body -> 400', r.status === 400, `(got ${r.status})`);
check('POST invalid body has details[]', Array.isArray(r.json?.details) && r.json.details.length > 0);

r = await call('POST', '/products', undefined, '{ broken json');
check('POST malformed JSON -> 400', r.status === 400, `(got ${r.status})`);

r = await call('POST', '/products', { ...valid, stock: 1.5 });
check('POST non-integer stock -> 400', r.status === 400, `(got ${r.status})`);

r = await call('GET', `/products/${id}`);
check('GET /products/:id -> 200', r.status === 200 && r.json?.id === id, `(got ${r.status})`);

r = await call('GET', '/products/999999');
check('GET /products/999999 -> 404', r.status === 404, `(got ${r.status})`);

r = await call('GET', '/products/abc');
check('GET /products/abc -> 400', r.status === 400, `(got ${r.status})`);

r = await call('PUT', `/products/${id}`, { ...valid, price: 39999, stock: 10 });
check('PUT /products/:id -> 200', r.status === 200, `(got ${r.status})`);
check('PUT updated fields', r.json?.price === 39999 && r.json?.stock === 10);

r = await call('PUT', `/products/${id}`, { name: 'x' });
check('PUT invalid body -> 400', r.status === 400, `(got ${r.status})`);

r = await call('PUT', '/products/999999', valid);
check('PUT /products/999999 -> 404', r.status === 404, `(got ${r.status})`);

r = await call('DELETE', `/products/${id}`);
check('DELETE /products/:id -> 204', r.status === 204, `(got ${r.status})`);

r = await call('DELETE', `/products/${id}`);
check('DELETE again -> 404', r.status === 404, `(got ${r.status})`);

r = await call('GET', `/products/${id}`);
check('GET after DELETE -> 404', r.status === 404, `(got ${r.status})`);

r = await call('PATCH', '/products');
check('PATCH /products -> 405', r.status === 405, `(got ${r.status})`);
check('405 has Allow header', (r.headers.get('allow') || '').includes('GET'));

r = await call('GET', '/unknown');
check('GET /unknown -> 404', r.status === 404, `(got ${r.status})`);

console.log(`\n${passed} passed, ${failed} failed`);
process.exit(failed ? 1 : 0);
