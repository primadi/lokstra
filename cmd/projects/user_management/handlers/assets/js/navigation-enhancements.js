// Navigation Enhancement JavaScript
// Embedded JavaScript for enhanced navigation and UI feedback

function initializeNavigationEnhancements() {
  // Enhanced loading indicators
  const navButtons = document.querySelectorAll(".nav-page, [hx-get]");

  navButtons.forEach((button) => {
    button.addEventListener("click", function () {
      // Add loading state
      const originalText = this.textContent;
      const originalHTML = this.innerHTML;

      // Show loading indicator
      this.innerHTML = `
                <svg class="animate-spin h-4 w-4 inline-block mr-2" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" opacity="0.25"></circle>
                    <path fill="currentColor" opacity="0.75" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Loading...
            `;
      this.disabled = true;

      // Reset after reasonable time or when HTMX completes
      setTimeout(() => {
        this.innerHTML = originalHTML;
        this.disabled = false;
      }, 2000);
    });
  });

  // Enhanced form feedback
  const forms = document.querySelectorAll("form[hx-post]");

  forms.forEach((form) => {
    form.addEventListener("submit", function () {
      const submitButton = this.querySelector('button[type="submit"]');

      if (submitButton) {
        const originalText = submitButton.textContent;
        submitButton.textContent = "Processing...";
        submitButton.disabled = true;

        // Reset after reasonable time
        setTimeout(() => {
          submitButton.textContent = originalText;
          submitButton.disabled = false;
        }, 3000);
      }
    });
  });

  console.log("Navigation enhancements initialized");
}

function initializeUIFeedback() {
  // Success/Error message auto-hide
  const messages = document.querySelectorAll(".bg-green-600, .bg-red-600");

  messages.forEach((message) => {
    if (
      message.textContent.includes("successfully") ||
      message.textContent.includes("error")
    ) {
      setTimeout(() => {
        message.style.opacity = "0";
        message.style.transform = "translateY(-20px)";
        message.style.transition = "all 0.3s ease";

        setTimeout(() => {
          message.remove();
        }, 300);
      }, 5000);
    }
  });

  // Enhanced tooltip behavior
  const tooltipElements = document.querySelectorAll("[title]");

  tooltipElements.forEach((element) => {
    element.addEventListener("mouseenter", function () {
      this.style.position = "relative";
    });
  });
}

// Keyboard shortcuts
function initializeKeyboardShortcuts() {
  document.addEventListener("keydown", function (e) {
    // Ctrl+/ or Cmd+/ to focus search
    if ((e.ctrlKey || e.metaKey) && e.key === "/") {
      e.preventDefault();
      const searchInput = document.querySelector(
        'input[placeholder*="Search"]'
      );
      if (searchInput) {
        searchInput.focus();
        searchInput.select();
      }
    }

    // Escape to clear search
    if (e.key === "Escape") {
      const searchInput = document.querySelector(
        'input[placeholder*="Search"]'
      );
      if (searchInput && searchInput === document.activeElement) {
        searchInput.value = "";
        searchInput.dispatchEvent(new Event("input"));
        searchInput.blur();
      }
    }
  });
}

// Auto-initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", function () {
    initializeNavigationEnhancements();
    initializeUIFeedback();
    initializeKeyboardShortcuts();
  });
} else {
  initializeNavigationEnhancements();
  initializeUIFeedback();
  initializeKeyboardShortcuts();
}
