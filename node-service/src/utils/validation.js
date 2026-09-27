// Валідація тіла запиту для POST/PUT /products.
// Повертає { value, errors }. Якщо errors не порожній — потрібно відповісти 400.
const isNonEmptyString = (v) => typeof v === 'string' && v.trim().length > 0;

function validateProductInput(body) {
  if (typeof body !== 'object' || body === null || Array.isArray(body)) {
    return { value: null, errors: ['request body must be a JSON object'] };
  }

  const errors = [];

  if (!isNonEmptyString(body.name) || body.name.trim().length > 100) {
    errors.push('name is required and must be a non-empty string (max 100 chars)');
  }
  if (!isNonEmptyString(body.category)) {
    errors.push('category is required and must be a non-empty string');
  }
  if (!isNonEmptyString(body.brand)) {
    errors.push('brand is required and must be a non-empty string');
  }
  if (typeof body.price !== 'number' || !Number.isFinite(body.price) || body.price <= 0) {
    errors.push('price is required and must be a number greater than 0');
  }
  if (!Number.isInteger(body.stock) || body.stock < 0) {
    errors.push('stock is required and must be an integer >= 0');
  }
  if (body.description !== undefined && typeof body.description !== 'string') {
    errors.push('description must be a string');
  }

  if (errors.length > 0) return { value: null, errors };

  return {
    value: {
      name: body.name.trim(),
      category: body.category.trim(),
      brand: body.brand.trim(),
      price: body.price,
      stock: body.stock,
      description: (body.description ?? '').trim(),
    },
    errors: [],
  };
}

module.exports = { validateProductInput };
