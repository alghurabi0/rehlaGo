const CACHE_NAME = 'Rehla-v3';
const urlsToCache = [
  '/',
];

// Install event: Cache initial resources
self.addEventListener('install', (event) => {
  event.waitUntil(
      // prefetch
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(urlsToCache);
      })
  );
});

self.addEventListener('activate', (event) => {
  // Clean up old caches if necessary
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cache) => {
          if (cache !== CACHE_NAME) {
            return caches.delete(cache);
          }
        })
      );
    })
  );
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
    event.respondWith(networkFirst(event.request));
  }
);

// Cache First Strategy
async function cacheFirst(request) {
  const cache = await caches.open('my-cache');
  const cachedResponse = await cache.match(request);

  // If there's a cached response, return it, otherwise fetch from the network
  return cachedResponse || fetch(request);
}

// Stale-While-Revalidate Strategy
async function staleWhileRevalidate(request) {
  const cache = await caches.open('my-cache');
  const cachedResponse = await cache.match(request);

  // Fetch new data in the background
  const fetchPromise = fetch(request).then((networkResponse) => {
    cache.put(request, networkResponse.clone()); // Update the cache with the fresh response
    return networkResponse;
  });

  // Return cached response immediately, then update cache with network response in the background
  return cachedResponse || fetchPromise;
}

// Network First Strategy (Fallback)
async function networkFirst(request) {
  try {
    const networkResponse = await fetch(request);
    const cache = await caches.open('my-cache');
    cache.put(request, networkResponse.clone()); // Cache the network response
    return networkResponse;
  } catch (error) {
    // If network fails, try to return the cached response
    const cache = await caches.open('my-cache');
    return await cache.match(request);
  }
}
