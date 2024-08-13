const CACHE_NAME = 'Rehla';
const urlsToCache = [
  '/',
  '/courses',
  '/progress',
  '/materials',
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

//self.addEventListener('fetch', (event) => {
//event.respondWith(
      //   caches.match(event.request)
//.then((response) => {
//       if (response) {
//         return response;
//       }
//
//       // Attempt to fetch the resource from the network if not cached
//       return fetch(event.request).catch(() => {
//         // If the request fails (e.g., offline), show a fallback page or resource
//         if (event.request.mode === 'navigate') {
//           // Return the cached home page for navigation requests
//           return caches.match('/');
//         }
//
//         // Optionally return a generic fallback for other requests (like images)
//         return caches.match('/static/icons/home.png'); // Example fallback
//       });
//     })
// );
//);
