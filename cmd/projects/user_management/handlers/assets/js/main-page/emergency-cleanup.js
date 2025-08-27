// Emergency Cleanup - User Management Application
// Handles emergency cleanup for loading indicators and other UI states

// Emergency cleanup - force hide any lingering indicators
function startEmergencyCleanup() {
  setInterval(function () {
    if (!document.body.classList.contains("htmx-request")) {
      const loadingIndicator = document.getElementById("loading-indicator");
      if (
        loadingIndicator &&
        window.getComputedStyle(loadingIndicator).display !== "none"
      ) {
        forceHideLoadingIndicator("EMERGENCY CLEANUP");
      }
    }
  }, 50); // Reduced interval untuk lebih responsif
}

// Initialize emergency cleanup when DOM is ready
document.addEventListener("DOMContentLoaded", function () {
  startEmergencyCleanup();
});
