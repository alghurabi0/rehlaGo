const SESSION_CACHE_NAME = "SESSION_CACHE";

// Install event: Cache initial resources
self.addEventListener("install", (event) => {
  console.log("installed service worker");
});

// Handle fetch requests
self.addEventListener("fetch", (event) => {
  const request = event.request;

  // Only cache GET requests
  if (request.method !== "GET") return;

  event.respondWith(
    caches.open(SESSION_CACHE_NAME).then((cache) => {
      return cache.match(request).then((cachedResponse) => {
        if (cachedResponse) {
          // Serve from cache
          console.log("[Service Worker] Serving from cache:", request.url);
          return cachedResponse;
        }

        // Otherwise, fetch from network and cache the response
        return fetch(request).then((networkResponse) => {
          if (networkResponse && networkResponse.status === 200) {
            cache.put(request, networkResponse.clone());
            console.log("[Service Worker] Cached:", request.url);
          }
          return networkResponse;
        });
      });
    }),
  );
});

// Clear cache when no tabs (clients) are active
async function clearCacheIfNoClients() {
  const clientsList = await self.clients.matchAll({ type: "window" });
  if (clientsList.length === 0) {
    console.log("[Service Worker] No active clients, clearing session cache.");
    await caches.delete(SESSION_CACHE_NAME);
  } else {
    console.log("the are active clients");
  }
}

// Listen for activate event (cleanup when necessary)
self.addEventListener("activate", (event) => {
  console.log("service worker activated");
});

// Listen for messages from the main thread
self.addEventListener("message", (event) => {
  console.log("message event");
  if (event.data && event.data.action === "clear-cache") {
    clearCacheIfNoClients();
    console.log("deleted cache");
  }
});

// Cache First Strategy
async function cacheFirst(request) {
  const cache = await caches.open("my-cache");
  const cachedResponse = await cache.match(request);

  // If there's a cached response, return it, otherwise fetch from the network
  return cachedResponse || fetch(request);
}

// Stale-While-Revalidate Strategy
async function staleWhileRevalidate(request) {
  const cache = await caches.open("my-cache");
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
    const cache = await caches.open("my-cache");
    cache.put(request, networkResponse.clone()); // Cache the network response
    return networkResponse;
  } catch (error) {
    // If network fails, try to return the cached response
    const cache = await caches.open("my-cache");
    return await cache.match(request);
  }
}
