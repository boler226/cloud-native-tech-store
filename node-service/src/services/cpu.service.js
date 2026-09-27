// CPU-сервіс: синхронна, звантажена робота (проста лічилка, щоб виключити
// побічні ефекти на кшталт I/O чи алокацій). Виконується у головному потоці
// Node і блокує event loop, поки не завершиться.
function busyLoop(iterations) {
  let counter = 0;
  for (let i = 0; i < iterations; i += 1) {
    counter += 1;
  }
  return counter;
}

module.exports = { busyLoop };
