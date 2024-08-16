const CACHE_NAME = 'Rehla-v2';
const urlsToCache = [
  '/',
  '/courses',
  '/progress',
  '/materials',
  '/payments',
  '/mycourses',
  '/myprofile',
  '/privacy_policy',
  '/contact',
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

// Install event: Cache initial resources
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(urlsToCache);
      })
  );
});

// Fetch event: Implement Stale-While-Revalidate strategy for all routes
//self.addEventListener('fetch', (event) => {
  //console.log('Fetch request for:', event.request.url);
//
  //event.respondWith(
    //caches.match(event.request).then((cachedResponse) => {
      //const fetchPromise = fetch(event.request).then((networkResponse) => {
        //console.log('Network response for:', event.request.url);
        //if (networkResponse && networkResponse.ok) {
          //caches.open(CACHE_NAME).then((cache) => {
            //console.log('Updating cache with:', event.request.url);
            //cache.put(event.request, networkResponse.clone());
          //});
        //}
        //return networkResponse;
      //}).catch(() => {
        //console.log('Network fetch failed for:', event.request.url);
        //return cachedResponse;
      //});
//
      //console.log('Returning cached response or fetch promise for:', event.request.url);
      //return cachedResponse || fetchPromise;
    //})
  //);
//});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    (async () => {
      const cachedResponse = await caches.match(event.request);
      const fetchPromise = fetch(event.request)
        .then(async (networkResponse) => {
          // Check if the response is valid
          if (networkResponse.status === 200) {
            const cache = await caches.open(CACHE_NAME);
            cache.put(event.request, networkResponse.clone());
          }
          return networkResponse;
        })
        .catch(() => {
          // Network fetch failed, return cached response
          return cachedResponse;
        });

      // Return cached response immediately, if available, or wait for the network fetch
      return cachedResponse || fetchPromise;
    })()
  );
});


// Clean up old caches during the activation phase
self.addEventListener('activate', (event) => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (!cacheWhitelist.includes(cacheName)) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

