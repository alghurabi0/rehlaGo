import {
  getStorage,
  ref,
  uploadBytesResumable,
} from "https://www.gstatic.com/firebasejs/10.11.1/firebase-storage.js";
import { app } from "/static/js/base.js";

const storage = getStorage(app);

const inputEl = document.querySelector("#file_upload");
const filename = document.querySelector("#filename");
if (inputEl && filename) {
  inputEl.addEventListener("change", handleFile, "false");
}
function handleFile() {
  if (!inputEl || !filename) {
    return;
  }
  const fileList = this.files;
  const file = fileList[0];
  filename.innerText = file.name;
}
const cancelDiv = document.querySelector("#cancel_upload");
if (cancelDiv) {
  cancelDiv.addEventListener("click", cancelUpload);
}
function cancelUpload() {
  inputEl.value = "";
  filename.innerText = "لم تقم باختيار ملف بعد";
}
