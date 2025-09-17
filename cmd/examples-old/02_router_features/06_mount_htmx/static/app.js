// Lokstra HTMX Demo JavaScript
document.addEventListener("DOMContentLoaded", function () {
  console.log("Lokstra HTMX Demo App loaded")

  // Handle form submissions
  document.body.addEventListener("htmx:responseError", function (evt) {
    console.error("HTMX request error:", evt.detail)

    // Show error message
    showMessage("An error occurred. Please try again.", "error")
  })

  // Custom functions
  window.showMessage = function (text, type = "info") {
    const messageDiv = document.createElement("div")
    messageDiv.className = `message ${type} fade-in`
    messageDiv.textContent = text

    // Find a good place to insert the message
    const target = document.querySelector(".page-content") || document.body
    target.insertBefore(messageDiv, target.firstChild)

    // Auto-remove after 5 seconds
    setTimeout(() => {
      messageDiv.remove()
    }, 5000)
  }

  // Product interactions
  window.loadProductDetails = function (productId) {
    console.log("Loading product details for:", productId)
  }

  // Contact form enhancements
  const contactForm = document.querySelector("#contact-form")
  if (contactForm) {
    contactForm.addEventListener("htmx:afterRequest", function (evt) {
      if (evt.detail.successful) {
        // Clear form after successful submission
        this.reset()
        showMessage("Thank you for your message!", "success")
      }
    })
  }
})

// Navigation helpers
function navigateTo(path) {
  // Use HTMX to navigate while preserving the layout
  htmx.ajax("GET", path, {
    target: "main",
    swap: "innerHTML",
  })
}

// Utility functions
function formatCurrency(amount) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(amount)
}

function formatDate(date) {
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(new Date(date))
}

// Theme toggler (example of additional functionality)
function toggleTheme() {
  document.body.classList.toggle("dark-theme")
  localStorage.setItem(
    "theme",
    document.body.classList.contains("dark-theme") ? "dark" : "light"
  )
}

// Load saved theme
const savedTheme = localStorage.getItem("theme")
if (savedTheme === "dark") {
  document.body.classList.add("dark-theme")
}
