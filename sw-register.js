if ("serviceWorker" in navigator) {
  navigator.serviceWorker
    .register("service-worker.js")
    .then((reg) => {
      console.log("Service Worker Registered");

      // Function to send the clear-cache message
      function sendClearCacheMessage() {
        if (navigator.serviceWorker.controller) {
          navigator.serviceWorker.controller.postMessage({
            action: "clear-cache",
          });
        }
      }

      // Visibility change event (for tab/window changes)
      let visibilityChange;
      if (typeof document.hidden !== "undefined") {
        visibilityChange = "visibilitychange";
      } else if (typeof document.msHidden !== "undefined") {
        visibilityChange = "msvisibilitychange";
      } else if (typeof document.webkitHidden !== "undefined") {
        visibilityChange = "webkitvisibilitychange";
      }

      // Page lifecycle event (for mobile app backgrounding)
      window.addEventListener("pagehide", (event) => {
        // On mobile, assume the app is being closed or backgrounded
        sendClearCacheMessage();
      });

      if (visibilityChange) {
        document.addEventListener(
          visibilityChange,
          () => {
            if (document.hidden) {
              // Tab is hidden, might be closing
              sendClearCacheMessage();
            }
          },
          false,
        );
      }
    })
    .catch((err) => console.log("Service Worker Registration Failed", err));
}
