importScripts(
  'https://storage.googleapis.com/workbox-cdn/releases/6.4.1/workbox-sw.js'
);

const { registerRoute, Route, setCatchHandler, setDefaultHandler } = workbox.routing;
const { CacheFirst, StaleWhileRevalidate, NetworkOnly } = workbox.strategies;
const { offlineFallback } = workbox.recipes;
//const {CacheableResponse} = workbox.cacheableResponse;

const pageFallback = '/static/offline.html';
const imageFallback = false;
const fontFallback = false;

const imageCache = 'images';
const scriptCache = 'scripts';
const styleCache = 'styles';
const homepageCache = 'homepage';
const courseCache = 'courses';
const materialCache = 'materials';
const progressCache = 'progress';

// Handle images:
const imageRoute = new Route(({ request }) => {
  return request.destination === 'image'
}, new StaleWhileRevalidate({
  cacheName: imageCache
}));

// Handle scripts:
const scriptsRoute = new Route(({ request }) => {
  return request.destination === 'script';
}, new CacheFirst({
  cacheName: scriptCache
}));

// Handle styles:
const stylesRoute = new Route(({ request }) => {
  return request.destination === 'style';
}, new CacheFirst({
  cacheName: styleCache
}));

// Handle main pages
const homepageRoute = new Route(({ request }) => {
  const url = new URL(request.url)
  return request.method === 'GET' &&
    (url.pathname === '/' ||
      url.pathname === '/courses' ||
      url.pathname === '/materials' ||
      url.pathname === '/progress') &&
    (request.headers.get('HX-Request') === 'true' ||
      request.headers.get('HX-Request') !== 'true');
}, new StaleWhileRevalidate({
  cacheName: homepageCache
}));

// Handle course pages
registerRoute(/\/courses\/([^\/]+)$/,
  new StaleWhileRevalidate({
    cacheName: courseCache
  }));

// Handle free materials
const freeMaterials = new Route(({ request, url }) => {
  return url.pathname === '/free';
}, new StaleWhileRevalidate({
  cacheName: materialCache
}));

// Handle materials page
registerRoute(/\/materials\/([^\/]+)$/,
  new StaleWhileRevalidate({
    cacheName: materialCache
  }));

// Handle progress page
registerRoute(/\/progress\/([^\/]+)$/,
  new StaleWhileRevalidate({
    cacheName: progressCache
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
    const swap = originalResponse.headers.get('HX-Reswap');
    const target = originalResponse.headers.get('HX-Retarget');
    if (swap && target) {
      newHeaders.set('HX-Reswap', swap);
      newHeaders.set('HX-Retarget', target);
    } else {
      newHeaders.set('HX-Redirect', '/'); // Example header
    }

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
registerRoute(freeMaterials);

self.addEventListener('install', event => {
  const files = [pageFallback];
  if (imageFallback) {
    files.push(imageFallback);
  }
  if (fontFallback) {
    files.push(fontFallback);
  }

  event.waitUntil(
    self.caches
      .open('workbox-offline-fallbacks')
      .then(cache => cache.addAll(files)),
    self.skipWaiting()
  );
});

self.addEventListener('activate', (event) => {
  console.log('[Service Worker] Activating');

  event.waitUntil(
    (async () => {
      // Claim all clients immediately
      await self.clients.claim();

      // Get a list of all cache names
      const cacheNames = await caches.keys();

      // Delete all caches
      await Promise.all(
        cacheNames.map((cacheName) => {
          console.log('[Service Worker] Deleting cache:', cacheName);
          return caches.delete(cacheName);
        })
      );

      console.log('[Service Worker] All caches deleted');
    })()
  );
});

setDefaultHandler(new NetworkOnly());

const handler = async options => {
  const dest = options.request.destination;
  const cache = await self.caches.open('workbox-offline-fallbacks');

  if (dest === 'document' || options.request.headers.get('HX-Request') === 'true') {
    return (await cache.match(pageFallback)) || Response.error();
  }

  if (dest === 'image' && imageFallback !== false) {
    return (await cache.match(imageFallback)) || Response.error();
  }

  if (dest === 'font' && fontFallback !== false) {
    return (await cache.match(fontFallback)) || Response.error();
  }

  return Response.error();
};

setCatchHandler(handler);