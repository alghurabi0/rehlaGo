import { initializeApp } from "https://www.gstatic.com/firebasejs/10.11.1/firebase-app.js";
import {
  getMessaging,
  getToken,
  onMessage,
} from "https://www.gstatic.com/firebasejs/10.11.1/firebase-messaging.js";

const firebaseConfig = {
  apiKey: "AIzaSyA01nP4J1tbaEm7Buf3efG4J28KNLDPgtg",
  authDomain: "rehla-74745.firebaseapp.com",
  projectId: "rehla-74745",
  storageBucket: "rehla-74745.appspot.com",
  messagingSenderId: "818251494320",
  appId: "1:818251494320:web:24b46403b90eaa3c46dc75",
  measurementId: "G-6Q2KGCR684",
};

export const app = initializeApp(firebaseConfig);
export const messaging = getMessaging(app);
const vapidKey =
  "BE03NBCHDxocr72eaQka3A2Ttpsa7b4iF-VlfyAJ9_MzHRr7GVHxQKcj3Jh6IO3ku3-VNz4RjvCiwM6qn8W1YdA";
onMessage(messaging, (payload) => {
  console.log("Message received. ", payload);
});

function initializeNotifications() {
  if (Notification.permission === "granted" && isTokenSentToServer()) {
    console.log("Notification permission already granted.");
    retrieveToken();
  } else if (Notification.permission === "default") {
    requestPermission();
  } else {
    console.log("Notification permission denied or unavailable.");
    requestPermission();
  }
}

function retrieveToken() {
  getToken(messaging, vapidKey)
    .then((currentToken) => {
      if (currentToken) {
        sendTokenToServer(currentToken);
      } else {
        // Show permission request.
        console.log(
          "No registration token available. Request permission to generate one.",
        );
        // Show permission UI.
        setTokenSentToServer(false);
      }
    })
    .catch((err) => {
      console.log("An error occurred while retrieving token. ", err);
      setTokenSentToServer(false);
    });
}

function requestPermission() {
  console.log("Requesting permission...");
  Notification.requestPermission().then((permission) => {
    if (permission === "granted") {
      console.log("Notification permission granted.");
      // TODO(developer): Retrieve a registration token for use with FCM.
      // In many cases once an app has been granted notification permission,
      // it should update its UI reflecting this.
      retrieveToken();
    } else {
      console.log("Unable to get permission to notify.");
    }
  });
}

// Send the registration token your application server, so that it can:
// - send messages back to this app
// - subscribe/unsubscribe the token from topics
function sendTokenToServer(currentToken) {
  if (!isTokenSentToServer()) {
    console.log("Sending token to server...", currentToken);
    fetch(`/${currentToken}`);
    setTokenSentToServer(true);
  } else {
    console.log(
      "Token already sent to server so won't send it again unless it changes",
    );
  }
}

function isTokenSentToServer() {
  return window.localStorage.getItem("sentToServer") === "1";
}

function setTokenSentToServer(sent) {
  window.localStorage.setItem("sentToServer", sent ? "1" : "0");
}

initializeNotifications();

// ----------------------------------------------------------------
document.addEventListener("DOMContentLoaded", () => {
  const loginDialog = document.querySelector("#loginDialog");
  const subDialog = document.querySelector("#subDialog");
  const loginClose = document.querySelector("#loginClose");
  const subClose = document.querySelector("#subClose");
  const username = document.querySelector("#username");
  const tabUsername = document.querySelector("#tabUsername");
  const navDrawer = document.querySelector("#nav_drawer");

  if (loginDialog && loginClose) {
    loginClose.addEventListener("click", () => {
      loginDialog.classList.add("hidden");
      loginDialog.classList.remove("flex");
      loginDialog.close();
    });
  }

  if (subDialog && subClose) {
    subClose.addEventListener("click", () => {
      subDialog.classList.add("hidden");
      subDialog.classList.remove("flex");
      subDialog.close();
    });
  }

  document.addEventListener("htmx:responseError", (event) => {
    if (!event) {
      console.log("empty event");
      return;
    }

    if (event.detail.xhr.status == 401) {
      if (event.detail.xhr.responseText == "loginRequired") {
        console.log("login required");
        if (loginDialog) {
          loginDialog.classList.remove("hidden");
          loginDialog.classList.add("flex");
          loginDialog.showModal();
        }
      } else if (event.detail.xhr.responseText == "subRequired") {
        console.log("subscription required");
        if (subDialog) {
          subDialog.classList.remove("hidden");
          subDialog.classList.add("flex");
          subDialog.showModal();
        }
      }
    }
  });
  if (navDrawer && username && tabUsername) {
    const navDrawerUsername = navDrawer.querySelector("#nav_drawer_username");
    username.addEventListener("click", () => {
      if (navDrawer.classList.contains("hidden")) {
        navDrawer.classList.remove("hidden");
        navDrawer.classList.add("flex");
      } else if (navDrawer.classList.contains("flex")) {
        navDrawer.classList.remove("flex");
        navDrawer.classList.add("hidden");
      }
    });
    tabUsername.addEventListener("click", () => {
      if (navDrawer.classList.contains("hidden")) {
        navDrawer.classList.remove("hidden");
        navDrawer.classList.add("flex");
      } else if (navDrawer.classList.contains("flex")) {
        navDrawer.classList.remove("flex");
        navDrawer.classList.add("hidden");
      }
    });

    if (navDrawerUsername) {
      navDrawerUsername.addEventListener("click", () => {
        if (navDrawer.classList.contains("hidden")) {
          navDrawer.classList.remove("hidden");
          navDrawer.classList.add("flex");
        } else if (navDrawer.classList.contains("flex")) {
          navDrawer.classList.remove("flex");
          navDrawer.classList.add("hidden");
        }
      });
    }
  }
});
