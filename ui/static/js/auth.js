import {
  getAuth,
  RecaptchaVerifier,
  signInWithPhoneNumber,
} from "https://www.gstatic.com/firebasejs/10.11.1/firebase-auth.js";
import { app } from "/static/js/base.js";
console.log("auth file");

const auth = getAuth(app);
render();
function render() {
  window.recaptchaVerifier = new RecaptchaVerifier(
    auth,
    "recaptcha_container",
    {},
  );
  recaptchaVerifier.render();
}

let formData = {};
// TODO
var userId = "";
const signup_form = document.getElementById("signup_form");

function sendOTP(event) {
  if (event) event.preventDefault();
  const inputs = signup_form.getElementsByTagName("input");
  const select = signup_form.querySelector("select");

  // TODO
  // validation
  for (let i = 0; i < inputs.length; i++) {
    if (!inputs[i].value) {
      alert("Please fill all the fields");
      return;
    }
  }
  if (!select.value) {
      alert("Please fill all the fields");
      return;
  }

  // send data to backend
  for (let i = 0; i < inputs.length; i++) {
    formData[inputs[i].name] = inputs[i].value;
    formData[select.name] = select.value;
  }
  fetch("/signup", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(formData),
  })
    .then((res) => {
      if (res.status == 202) {
        return res.text();
      } else if (res.status == 409) {
        // TODO
        alert("User already exists");
        return Promise.reject("User exists");
      } else {
        alert(res);
        console.log("Error", res);
      }
    })
    .then((userid) => {
      userId = userid;
      const phone = "+964" + formData["phone_number"].slice(1);
      return signInWithPhoneNumber(auth, phone, window.recaptchaVerifier);
    })
    .then((confirmationResult) => {
      window.confirmationResult = confirmationResult;
      document.getElementById("signup_form").classList.remove("grid");
      document.getElementById("signup_form").classList.add("hidden");
      document.getElementById("verify_form").classList.remove("hidden");
      document.getElementById("verify_form").classList.add("grid");
    })
    .catch((error) => {
      console.log("firebase error", error);
    });
}

function verifyOTP() {
  const otp = document.getElementById("otp").value;
  if (!otp) {
    alert("Please enter OTP");
    return;
  }
  fetch("/verify_signup", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: userId,
  })
    .then((res) => {
      console.log(res);
      if (res.status == 200) {
        alert("User created successfully");
        window.location.href = "/";
      } else {
        // TODO
        console.log("Error", res);
      }
    })
    .catch((error) => {
      console.log("fetch error", error);
    });
  window.confirmationResult
    .confirm(otp)
    .then(() => {
      console.log("OTP verified");
      fetch("/verify_signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userId),
      })
        .then((res) => {
          console.log(res);
          if (res.status == 200) {
            alert("User created successfully");
            window.location.href = "/";
          } else {
            // TODO
            console.log("Error", res);
          }
        })
        .catch((error) => {
          console.log("fetch error", error);
        });
    })
    .catch((error) => {
      console.log("firebase verify error", error);
    });
}

const verify_button = document.getElementById("verify_button");
if (signup_form) {
  signup_form.addEventListener("submit", sendOTP);
}
if (verify_button) {
  verify_button.addEventListener("click", verifyOTP);
}
