import {
  getAuth,
  RecaptchaVerifier,
  signInWithPhoneNumber,
} from "https://www.gstatic.com/firebasejs/10.11.1/firebase-auth.js";
import { app } from "/static/js/base.js";

const auth = getAuth(app);
render();
function render() {
  window.recaptchaVerifier = new RecaptchaVerifier(
    auth,
    "recaptcha_container",
    {}
  );
  recaptchaVerifier.render();
}

let formData = {};

function sendOTP(event) {
  if (event) event.preventDefault();
  const form = document.getElementById("signup_form");
  const inputs = form.getElementsByTagName("input");
  // TODO
  for (let i = 0; i < inputs.length; i++) {
    if (!inputs[i].value) {
      alert("Please fill all the fields");
      return;
    }
  }
  for (let i = 0; i < inputs.length; i++) {
    formData[inputs[i].name] = inputs[i].value;
  }
  fetch("/signup_validate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(formData),
  })
    .then((res) => {
      console.log(res);
      if (res.status == 202) {
        signInWithPhoneNumber(
          auth,
          formData["phone_number"],
          window.recaptchaVerifier
        )
          .then((confirmationResult) => {
            window.confirmationResult = confirmationResult;
            console.log("OTP is sent");
            document.getElementById("signup_form").classList.remove("grid");
            document.getElementById("signup_form").classList.add("hidden");
            document.getElementById("verify_form").classList.remove("hidden");
            document.getElementById("verify_form").classList.add("grid");
          })
          .catch((error) => {
            console.log("firebase error", error);
          });
      } else if (res.status == 409) {
        alert("User already exists");
      } else {
        console.log("Error", res);
      }
    })
    .catch((error) => {
      console.log("fetch error", error);
    });
}

function verifyOTP() {
  const otp = document.getElementById("otp").value;
  if (!otp) {
    alert("Please enter OTP");
    return;
  }
  window.confirmationResult
    .confirm(otp)
    .then(() => {
      console.log("OTP verified");
      fetch("/signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      })
        .then((res) => {
          console.log(res);
          if (res.status == 201) {
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

const signup_form = document.getElementById("signup_form");
const verify_button = document.getElementById("verify_button");
if (signup_button) {
  signup_form.addEventListener("submit", sendOTP);
}
if (verify_button) {
  verify_button.addEventListener("click", verifyOTP);
}
