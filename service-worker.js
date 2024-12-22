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
registerRoute(({ request }) => {
  console.log('[Service Worker] loginRoute - request.url:', request.url);
  const url = new URL(request.url);
  console.log('[Service Worker] loginRoute - url.pathname:', url.pathname);
  return request.method === 'POST' && url.pathname === '/login';
}, async ({ event, request }) => {
  console.log("request to login");
  try {
    const response = await fetch(request.clone()); // Use clone() to avoid consuming the request body
    console.log("sent request");

    // Check for successful login (using cookie in this example)
    if (response.headers.get('Set-Cookie')?.includes('Login-Success=true')) {
      console.log('[Service Worker] Login successful, clearing homepage cache');
      const cache = await caches.open('homepage');
      await cache.keys().then(keys => {
        keys.forEach(key => cache.delete(key))
      });
    } else {
      console.log("no cookie");
    }

    return response;
  } catch (error) {
    console.error('[Service Worker] Error handling login request:', error);
    // Handle fetch errors if needed, e.g., return a fallback response
    return new Response('Login request failed', { status: 500 }); // Example fallback
  }
});

// Register routes
registerRoute(imageRoute);
registerRoute(scriptsRoute);
registerRoute(stylesRoute);
registerRoute(homepageRoute);