const CACHE_NAME = 'Rehla';
const urlsToCache = [
  '/',
  '/static/manifest.json',
  '/static/css/main.css',
  '/static/css/tailwind.css',
  '/static/js/auth.js',
  '/static/js/base.js',
  '/static/js/exam.js',
  '/static/icons/Video.png',
  '/static/icons/home.png',
  '/static/icons/sub.png'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(urlsToCache);
      })
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        return response || fetch(event.request);
      })
  );
});

