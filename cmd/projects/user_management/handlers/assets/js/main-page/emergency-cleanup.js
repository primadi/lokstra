// Emergency Cleanup - User Management Application
// Handles emergency cleanup for loading indicators and other UI states

// Emergency cleanup - force hide any lingering indicators
function startEmergencyCleanup() {
  setInterval(function () {
    if (!document.body.classList.contains("htmx-request")) {
      const loadingIndicator = document.getElementById("loading-indicator");
      if (loadingIndicator) {
        const currentOpacity =
          window.getComputedStyle(loadingIndicator).opacity;
        if (currentOpacity !== "0") {
          console.log("🆘 Emergency cleanup - forcing hide loading indicator");
          loadingIndicator.style.opacity = "0";
          loadingIndicator.style.pointerEvents = "none";
        }
      }
    }
  }, 100); // Reduced interval untuk lebih responsif
}

// Initialize emergency cleanup when DOM is ready
document.addEventListener("DOMContentLoaded", function () {
  startEmergencyCleanup();
});
