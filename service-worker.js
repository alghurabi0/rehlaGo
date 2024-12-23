importScripts(
  'https://storage.googleapis.com/workbox-cdn/releases/6.4.1/workbox-sw.js'
);

const { registerRoute, Route } = workbox.routing;
const { CacheFirst, StaleWhileRevalidate } = workbox.strategies;
//const {CacheableResponse} = workbox.cacheableResponse;

// Handle images:
const imageRoute = new Route(({ request }) => {
  return request.destination === 'image'
}, new StaleWhileRevalidate({
  cacheName: 'images'
}));

// Handle scripts:
const scriptsRoute = new Route(({ request }) => {
  return request.destination === 'script';
}, new CacheFirst({
  cacheName: 'scripts'
}));

// Handle styles:
const stylesRoute = new Route(({ request }) => {
  return request.destination === 'style';
}, new CacheFirst({
  cacheName: 'styles'
}));

// Handle homepage
const homepageRoute = new Route(({ request }) => {
  const url = new URL(request.url);
  return url.pathname === '/';
}, new CacheFirst({
  cacheName: 'homepage'
}));

// Handle deleting cache on login
self.addEventListener('fetch', (event) => {
  const request = event.request;
  const requestUrl = new URL(request.url);

  // Handle POST requests to /login
  if (request.method === 'POST' && requestUrl.pathname === '/login') {
    console.log('[Service Worker] Handling login POST request');

    event.respondWith(
      (async () => {
        try {
          const response = await fetch(request.clone()); // Clone to avoid consuming the body

          // Check for successful login (using cookie)
          if (response.headers.get('X-Login-Success') === 'true') {
            console.log('[Service Worker] Login successful, clearing cache and sending redirect message');

            // Clear the cache
            const cache = await caches.open('homepage');
            await cache.keys().then(keys => {
              keys.forEach(key => cache.delete(key))
            });

            console.log('[Service Worker] Cache cleared');

            const isHtmxRedirect = request.headers.get('Referer')?.includes('HX-Request'); // Adjust the logic as needed
            console.log(request.headers.get('Referer', isHtmxRedirect));
          } else {
            console.log('[Service Worker] No login success cookie found');
          }

          return response;
        } catch (error) {
          console.error('[Service Worker] Error handling login request:', error);
          return new Response('Login request failed', { status: 500 });
        }
      })()
    );
    return; // Important: Exit early for /login POST requests
  }
});

// Register routes
registerRoute(imageRoute);
registerRoute(scriptsRoute);
registerRoute(stylesRoute);
registerRoute(homepageRoute);