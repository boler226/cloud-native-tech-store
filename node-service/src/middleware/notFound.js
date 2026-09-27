// 404 для маршрутів, яких не існує.
module.exports = (req, res) => {
  res.status(404).json({ error: 'Route not found' });
};
