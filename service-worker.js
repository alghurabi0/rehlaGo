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

// Handle everything else
const homepageRoute = new Route(({ request }) => {
  const url = new URL(request.url)
  return request.method === 'GET' &&
    (url.pathname === '/' ||
      url.pathname === '/courses' ||
      url.pathname === '/materials' ||
      url.pathname === '/progress') &&
    request.headers.get('HX-Request') === 'true';
}, new StaleWhileRevalidate({
  cacheName: 'homepage'
}));

// Handle deleting cache on login
registerRoute(({ request }) => {
  const url = new URL(request.url);
  return url.pathname === '/login' || url.pathname === '/logout';
}, async ({ request }) => {
  try {
    // Clear the 'homepage' cache
    const cache = await caches.open('homepage');
    await cache.keys().then((keys) => {
      console.log('cache keys: ', keys);
      keys.forEach((key) => cache.delete(key));
    });
    console.log('cleared homepage cache');

    // Fetch the original response
    const originalResponse = await fetch(request.clone());

    // Create a new Headers object
    const newHeaders = new Headers(originalResponse.headers);

    // Add your custom header
    newHeaders.set('HX-Redirect', '/'); // Example header

    // Create a new Response object with the modified headers
    const newResponse = new Response(originalResponse.body, {
      status: originalResponse.status,
      statusText: originalResponse.statusText,
      headers: newHeaders,
    });

    return newResponse;
  } catch (error) {
    console.error('[Service Worker] Error handling login request:', error);
    return new Response('Login request failed', { status: 500 });
  }
}, 'POST');

// Register routes
registerRoute(imageRoute);
registerRoute(scriptsRoute);
registerRoute(stylesRoute);
registerRoute(homepageRoute);