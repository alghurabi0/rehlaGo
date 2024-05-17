import { initializeApp } from "https://www.gstatic.com/firebasejs/10.11.1/firebase-app.js";

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

document.addEventListener('DOMContentLoaded', () => {
    const loginDialog = document.querySelector('#loginDialog');
    const subDialog = document.querySelector('#subDialog');
    const loginClose = document.querySelector('#loginClose');
    const subClose = document.querySelector('#subClose');
    if (loginDialog && loginClose) {
        loginClose.addEventListener('click', () => {
            loginDialog.close();
        }
    }

    if (subDialog && subClose) {
        subClose.addEventListener('click', () => {
            subDialog.close();
        }
    }

    document.addEventListener('htmx:responseError', (event) => {
        if (!event) {
            console.log("empty event");
            return
        }

        if (event.detail.xhr.status == 401) {

            if (event.detail.xhr.responseText == 'loginRequired') {
                if (loginDialog) {
                    loginDialog.showModal();
                }
            } else if (event.detail.xhr.responseText == 'subRequired`) {
                if (subDialog) {
                    subDialog.showModal();
                }
            }

        }
    }
});
