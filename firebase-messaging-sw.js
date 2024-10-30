importScripts('https://www.gstatic.com/firebasejs/9.2.0/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/9.2.0/firebase-messaging-compat.js');

 // Initialize the Firebase app in the service worker by passing in
 // your app's Firebase config object.
 // https://firebase.google.com/docs/web/setup#config-object
const firebaseConfig = {
};
  firebase.initializeApp({
  apiKey: "AIzaSyA01nP4J1tbaEm7Buf3efG4J28KNLDPgtg",
  authDomain: "rehla-74745.firebaseapp.com",
  projectId: "rehla-74745",
  storageBucket: "rehla-74745.appspot.com",
  messagingSenderId: "818251494320",
  appId: "1:818251494320:web:24b46403b90eaa3c46dc75",
  measurementId: "G-6Q2KGCR684",
 });

 // Retrieve an instance of Firebase Messaging so that it can handle background
 // messages.
 const messaging = firebase.messaging();

messaging.onBackgroundMessage(function(payload) {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);
  // Customize notification here

  self.registration.showNotification();
});
