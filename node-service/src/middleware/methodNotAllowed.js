// 405 Method Not Allowed із заголовком Allow.
module.exports = (allowed) => (req, res) => {
  res.set('Allow', allowed.join(', '));
  res.status(405).json({ error: 'Method not allowed' });
};
