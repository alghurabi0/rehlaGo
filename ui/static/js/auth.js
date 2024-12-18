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
var userId = "";
const signup_form = document.getElementById("signup_form");
const inputs = signup_form.getElementsByTagName("input");
const select = signup_form.querySelector("select");

function checkFields(event) {
  if (event) event.preventDefault();

  // TODO
  // validation
  for (let i = 0; i < inputs.length; i++) {
    if (!inputs[i].value) {
      alert("يرجى ملئ جميع الحقول");
      return;
    }
  }
  if (!select.value) {
    alert("يرجى ملئ جميع الحقول");
    return;
  }
  return createUser();
}

async function createUser() {
  // send data to backend
  for (let i = 0; i < inputs.length; i++) {
    formData[inputs[i].name] = inputs[i].value;
  }
  formData[select.name] = select.value;
  // Send data to backend
  try {
    const response = await fetch("/signup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(formData),
    });

    if (response.status === 202) {
      const userid = await response.text();
      await sendOTP(userid); // Ensure sendOTP is awaited if it's async
    } else if (response.status === 409) {
      alert("يوجد حساب بهذا الرقم, يرجى استعمال رقم اخر او تسجيل الدخول");
      console.log("user exists");
    } else {
      const errorText = await response.text();
      alert(errorText);
      console.log("errors", errorText);
    }
  } catch (error) {
    console.error("Error during user creation:", error);
  }
}

async function sendOTP(user_id) {
  userId = user_id;
  const phone = "+964" + formData["phone_number"].slice(1);
  try {
    const confirmationResult = await signInWithPhoneNumber(
      auth,
      phone,
      window.recaptchaVerifier,
    );
    window.confirmationResult = confirmationResult;
    document.getElementById("signup_form").classList.remove("grid");
    document.getElementById("signup_form").classList.add("hidden");
    document.getElementById("verify_form").classList.remove("hidden");
    document.getElementById("verify_form").classList.add("grid");
  } catch (error) {
    console.log("firease error", error);
  }
}

async function verifyOTP() {
  try {
    const otp = document.getElementById("otp").value;

    // Validate OTP input
    if (!otp) {
      alert("يرجى ادخال رمز التأكيد");
      return;
    }

    // Verify the OTP
    await window.confirmationResult.confirm(otp);
    console.log("OTP verified");

    // Send a request to verify signup
    const response = await fetch("/verify_signup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(userId),
    });

    // Handle response
    if (response.status === 200) {
      alert("تم أنشاء الحساب بنجاح");
      window.location.href = "/";
    } else {
      console.log("Error", response);
      alert("حدث خطأ, يرجى التواصل مع الدعم");
    }
  } catch (error) {
    console.log("Fetch error or unknown error", error);
  }
}

const verify_button = document.getElementById("verify_button");
if (signup_form) {
  signup_form.addEventListener("submit", checkFields);
}
if (verify_button) {
  verify_button.addEventListener("click", verifyOTP);
}
