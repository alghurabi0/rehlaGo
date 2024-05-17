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
const sendDiv = document.querySelector("#send_file");
if (sendDiv) {
  sendDiv.addEventListener("click", sendFile);
}
function sendFile() {
  if (inputEl.files.length === 0) {
    console.log("no file selected");
    return;
  }
  const file = inputEl.files[0];
  const fileSizeMB = file.size / (1024 * 1024);
  if (fileSizeMB > 10) {
    console.log("file size is more than 10 mbs");
    return;
  }
  if (file.type !== "application/pdf") {
    console.log("file type is not pdf");
    return;
  }
  const examInfo = document.querySelector("#examInfo");
  // TODO - userId
  const storageRef = ref(storage, `/users/userId/courses/${examInfo.dataset.courseId}/answers/${examInfo.dataset.examId}.pdf`);
  const uploadFile = uploadBytesResumable(storageRef, file);

  uploadFile.on(
    "state_changed",
    (snapshot) => {
      console.log("file is uploading");
    },
    (error) => {
      switch (error.code) {
        case "storage/unauthorized":
          console.log("User doesn't have permission to access the object");
          break;
        case "storage/canceled":
          console.log("User canceled the upload");
          break;

        //...

        case "storage/unknown":
          console.log("Unknown error occurred, inspect error.serverResponse");
          break;
      }
    },
    () => {
      console.log("file uploaded successfully");
      const successEvent = new CustomEvent("successfulUpload", { detail: { filename: file.name } });
      examInfo.dispatchEvent(successEvent);
    }
  );
}
