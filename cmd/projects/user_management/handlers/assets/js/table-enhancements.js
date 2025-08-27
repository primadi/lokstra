// Table Enhancement JavaScript
// Embedded JavaScript for enhanced table functionality

function initializeTableEnhancements() {
  // Add enhanced hover effects
  const tableRows = document.querySelectorAll("table tbody tr");

  tableRows.forEach((row) => {
    row.addEventListener("mouseenter", function () {
      this.style.transform = "translateY(-1px)";
      this.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.3)";
      this.style.transition = "all 0.2s ease";
    });

    row.addEventListener("mouseleave", function () {
      this.style.transform = "translateY(0)";
      this.style.boxShadow = "none";
    });
  });

  // Enhanced action buttons
  const actionButtons = document.querySelectorAll(
    ".action-button, button[title]"
  );

  actionButtons.forEach((button) => {
    button.addEventListener("mouseenter", function () {
      this.style.transform = "scale(1.05)";
      this.style.transition = "transform 0.2s ease";
    });

    button.addEventListener("mouseleave", function () {
      this.style.transform = "scale(1)";
    });
  });

  // Delete confirmation animation
  const deleteButtons = document.querySelectorAll("button[hx-delete]");

  deleteButtons.forEach((button) => {
    button.addEventListener("click", function (e) {
      // Add shake animation for confirmation
      this.classList.add("delete-confirm");
      setTimeout(() => {
        this.classList.remove("delete-confirm");
      }, 500);
    });
  });

  console.log("Table enhancements initialized");
}

function initializeSearchFilter() {
  const searchInput = document.querySelector('input[placeholder*="Search"]');

  if (searchInput) {
    let searchTimeout;

    searchInput.addEventListener("input", function () {
      clearTimeout(searchTimeout);
      const searchTerm = this.value.toLowerCase();

      // Debounce search for better performance
      searchTimeout = setTimeout(() => {
        filterTableRows(searchTerm);
      }, 300);
    });
  }
}

function filterTableRows(searchTerm) {
  const tableRows = document.querySelectorAll("table tbody tr");

  tableRows.forEach((row) => {
    const rowText = row.textContent.toLowerCase();
    const shouldShow = rowText.includes(searchTerm);

    if (shouldShow) {
      row.style.display = "";
      row.style.opacity = "1";
    } else {
      row.style.opacity = "0.3";
      setTimeout(() => {
        if (row.style.opacity === "0.3") {
          row.style.display = "none";
        }
      }, 200);
    }
  });
}

// Auto-initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", function () {
    initializeTableEnhancements();
    initializeSearchFilter();
  });
} else {
  initializeTableEnhancements();
  initializeSearchFilter();
}
