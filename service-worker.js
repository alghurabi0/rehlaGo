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
  const url = new URL(request.url);
  return url.pathname === '/login';
}, async ({ request }) => {
  const cache = await caches.open('homepage');
  await cache.keys().then(keys => {
    console.log('cache keys: ', keys)
    keys.forEach(key => cache.delete(key))
  });
  console.log("cleared homepage cache");
  const response = await fetch(request.clone());
  console.log(response);
  return response;
}, 'POST');

// Register routes
registerRoute(imageRoute);
registerRoute(scriptsRoute);
registerRoute(stylesRoute);
registerRoute(homepageRoute);