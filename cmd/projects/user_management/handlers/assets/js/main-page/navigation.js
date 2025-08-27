// Navigation Utilities - User Management Application
// Handles HTMX navigation, loading indicators, and sidebar management

// Global variables
let navigationInProgress = false;
let navigationDebounceTimeout = null;
let isChrome = false;

// Initialize navigation system
function initializeNavigation() {
  // Detect Chrome and apply specific fixes
  isChrome =
    /Chrome/.test(navigator.userAgent) && /Google Inc/.test(navigator.vendor);
  if (isChrome) {
    console.log("🌐 Chrome detected - applying Chrome-specific fixes");
  }

  // Force clear loading indicator on page load/reload
  forceHideLoadingIndicator("NAVIGATION INIT");

  // Configure HTMX
  htmx.config.timeout = 10000;
  htmx.config.historyCacheSize = 3;
  htmx.config.defaultSwapStyle = "innerHTML";
  htmx.config.globalViewTransitions = true;

  // Chrome-specific: Disable transitions that might interfere
  if (isChrome) {
    htmx.config.globalViewTransitions = false;
    console.log("🌐 Chrome: Disabled view transitions");
  }

  // Set global defaults for all HTMX requests
  document.body.addEventListener("htmx:configRequest", function (evt) {
    if (!evt.detail.headers["hx-indicator"]) {
      evt.detail.indicator = "#loading-indicator";
    }

    if (
      evt.detail.verb === "get" &&
      evt.detail.target &&
      evt.detail.target.id === "main-content" &&
      !evt.detail.headers["hx-push-url"]
    ) {
      evt.detail.headers["hx-push-url"] = evt.detail.path;
    }
  });

  // Setup HTMX event handlers
  setupHTMXEventHandlers();

  // Setup navigation elements
  setupNavigationElements();

  // Initialize sidebar
  updateSidebarActiveMenu();

  // Listen for browser history changes
  window.addEventListener("popstate", function (evt) {
    console.log("🔙 Popstate event - browser back/forward button pressed");

    // Force clear any pending navigation state
    navigationInProgress = false;

    // Clear any debounce timeout
    if (navigationDebounceTimeout) {
      clearTimeout(navigationDebounceTimeout);
      navigationDebounceTimeout = null;
    }

    // Force hide loading indicator immediately
    forceHideLoadingIndicator("POPSTATE - Browser Back/Forward");

    // Update sidebar menu after a brief delay
    setTimeout(() => {
      updateSidebarActiveMenu();
    }, 50);
  });

  // Handle browser navigation controls (back/forward buttons, refresh, etc)
  window.addEventListener("beforeunload", function (evt) {
    console.log("🔄 Before unload - clearing navigation state");
    navigationInProgress = false;
    forceHideLoadingIndicator("BEFORE UNLOAD");
  });

  // Additional cleanup for mobile/modern browsers
  window.addEventListener("pagehide", function (evt) {
    console.log("🔄 Page hide - clearing navigation state");
    navigationInProgress = false;
    forceHideLoadingIndicator("PAGE HIDE");
  });

  // Cleanup when window regains focus (user returns to tab)
  window.addEventListener("focus", function (evt) {
    console.log("🎯 Window focus - ensuring clean state");
    navigationInProgress = false;
    forceHideLoadingIndicator("WINDOW FOCUS");
  });

  // Periodic cleanup as fallback (every 2 seconds)
  setInterval(function () {
    if (!navigationInProgress) {
      forceHideLoadingIndicator("PERIODIC CLEANUP");
    }
  }, 2000);
}

// Setup all HTMX event handlers
function setupHTMXEventHandlers() {
  // Navigation protection dengan logging yang detail
  htmx.on("htmx:beforeRequest", function (evt) {
    console.log("🔵 BEFORE REQUEST:", evt.detail.pathInfo.requestPath);

    if (navigationInProgress) {
      console.log("❌ Navigation blocked - request in progress");
      evt.preventDefault();
      return false;
    }

    if (navigationDebounceTimeout) {
      clearTimeout(navigationDebounceTimeout);
    }

    navigationInProgress = true;
    console.log("🚀 Navigation started:", evt.detail.pathInfo.requestPath);
  });

  // Immediate loading indicator control
  htmx.on("htmx:afterRequest", function (evt) {
    console.log("🟢 AFTER REQUEST:", evt.detail.pathInfo.requestPath);
    navigationInProgress = false;
    console.log("✅ Navigation lock removed");

    // Update sidebar menu first, THEN hide loading indicator
    setTimeout(() => {
      // Pass the target path from HTMX event instead of window.location
      const targetPath = evt.detail.pathInfo.requestPath;
      updateSidebarActiveMenu(targetPath);

      // HIDE loading indicator AFTER sidebar update completes
      if (isChrome) {
        // Use requestAnimationFrame for Chrome
        requestAnimationFrame(() => {
          forceHideLoadingIndicator("AFTER REQUEST - Chrome RAF");
        });
      } else {
        forceHideLoadingIndicator("AFTER REQUEST");
      }
    }, 10); // Reduced delay to minimum
  });

  // Double-check: Ensure loading indicator is hidden after swap
  htmx.on("htmx:afterSwap", function (evt) {
    console.log("🔄 AFTER SWAP:", evt.detail.pathInfo.requestPath);
    forceHideLoadingIndicator("AFTER SWAP");

    // Re-setup navigation elements after dynamic content is loaded
    setupNavigationElements();
  });

  // Triple-check: Final cleanup
  htmx.on("htmx:afterSettle", function (evt) {
    console.log("🏁 AFTER SETTLE - All done");
    forceHideLoadingIndicator("AFTER SETTLE");
  });

  // Error handling
  htmx.on("htmx:responseError", function (evt) {
    navigationInProgress = false;
    console.warn("Navigation error:", evt.detail);
  });

  htmx.on("htmx:timeout", function (evt) {
    navigationInProgress = false;
    console.warn("Navigation timeout");
  });
}

// Ultra-aggressive loading indicator hiding function
function forceHideLoadingIndicator(context) {
  const loadingIndicator = document.getElementById("loading-indicator");
  if (!loadingIndicator) {
    console.log("🟨", context, "- No loading indicator found");
    return;
  }

  // Remove all possible classes that might show it
  loadingIndicator.classList.remove(
    "htmx-indicator",
    "htmx-request",
    "show",
    "visible",
    "active"
  );

  // Force multiple CSS properties
  loadingIndicator.style.setProperty("display", "none", "important");
  loadingIndicator.style.setProperty("visibility", "hidden", "important");
  loadingIndicator.style.setProperty("opacity", "0", "important");
  loadingIndicator.style.setProperty("pointer-events", "none", "important");
  loadingIndicator.style.setProperty("z-index", "-9999", "important");

  // Force Chrome-specific hiding
  if (isChrome) {
    loadingIndicator.style.setProperty(
      "transform",
      "translateX(-9999px)",
      "important"
    );
    loadingIndicator.classList.add("chrome-force-hide");
  }

  // Force reflow to ensure immediate update
  loadingIndicator.offsetHeight;

  console.log("🚫", context, "- Loading indicator force hidden");
}

// Auto-setup navigation elements
function setupNavigationElements() {
  document.querySelectorAll(".nav-page[hx-get]").forEach(function (element) {
    const href = element.getAttribute("hx-get");

    if (!element.hasAttribute("hx-target")) {
      element.setAttribute("hx-target", "#main-content");
    }

    if (!element.hasAttribute("hx-swap")) {
      element.setAttribute("hx-swap", "innerHTML");
    }

    if (!element.hasAttribute("hx-indicator")) {
      element.setAttribute("hx-indicator", "#loading-indicator");
    }

    if (!element.hasAttribute("hx-push-url")) {
      element.setAttribute("hx-push-url", href);
    }

    console.log("Auto-setup navigation for:", href);
  });

  // Special handling for cancel buttons in forms
  document
    .querySelectorAll('button[type="button"][hx-get]')
    .forEach(function (button) {
      button.addEventListener("click", function () {
        console.log(
          "🔴 Cancel button clicked - ensuring loading indicator cleanup"
        );
        // Force immediate loading indicator hide for cancel actions
        setTimeout(() => {
          forceHideLoadingIndicator("CANCEL BUTTON CLICK");
        }, 100);
      });
    });
}
