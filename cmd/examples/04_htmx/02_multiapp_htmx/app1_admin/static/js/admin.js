// Admin Panel JavaScript Functions

// Search functionality
function openSearch() {
  document.getElementById("search-overlay").style.display = "flex"
  document.querySelector(".search-input").focus()
}

function closeSearch() {
  document.getElementById("search-overlay").style.display = "none"
}

// Modal functionality
function showModal() {
  document.getElementById("action-modal").style.display = "flex"
}

function hideModal() {
  document.getElementById("action-modal").style.display = "none"
}

// Keyboard shortcuts
document.addEventListener("keydown", function (e) {
  // Search shortcut (Ctrl+K or Cmd+K)
  if ((e.ctrlKey || e.metaKey) && e.key === "k") {
    e.preventDefault()
    openSearch()
  }

  // Close modal/search on Escape
  if (e.key === "Escape") {
    closeSearch()
    hideModal()
  }
})

// HTMX event handlers for admin panel
document.body.addEventListener("htmx:beforeRequest", function (evt) {
  console.log("Admin API request starting:", evt.detail.pathInfo.requestPath)

  // Add loading state to buttons
  const trigger = evt.detail.elt
  if (trigger && trigger.tagName === "BUTTON") {
    trigger.classList.add("loading")
    trigger.disabled = true
  }
})

document.body.addEventListener("htmx:afterRequest", function (evt) {
  console.log("Admin API request completed:", evt.detail.pathInfo.requestPath)

  // Remove loading state from buttons
  const trigger = evt.detail.elt
  if (trigger && trigger.tagName === "BUTTON") {
    trigger.classList.remove("loading")
    trigger.disabled = false
  }

  if (!evt.detail.successful) {
    console.error("Admin API request failed:", evt.detail)
    showNotification("Request failed. Please try again.", "error")
  }
})

// Notification system
function showNotification(message, type = "info") {
  const notification = document.createElement("div")
  notification.className = `notification notification-${type}`
  notification.innerHTML = `
        <div class="notification-content">
            <span class="notification-icon">${getNotificationIcon(type)}</span>
            <span class="notification-message">${message}</span>
            <button class="notification-close" onclick="this.parentElement.parentElement.remove()">✕</button>
        </div>
    `

  // Add to page
  document.body.appendChild(notification)

  // Auto-remove after 5 seconds
  setTimeout(() => {
    if (notification.parentElement) {
      notification.remove()
    }
  }, 5000)

  // Animate in
  setTimeout(() => notification.classList.add("show"), 100)
}

function getNotificationIcon(type) {
  switch (type) {
    case "success":
      return "✅"
    case "error":
      return "❌"
    case "warning":
      return "⚠️"
    default:
      return "ℹ️"
  }
}

// Navigation active state management
document.addEventListener("htmx:afterRequest", function (evt) {
  if (
    evt.detail.successful &&
    evt.detail.target &&
    evt.detail.target.tagName === "MAIN"
  ) {
    // Update active navigation based on current URL
    const currentPath = window.location.pathname
    const navLinks = document.querySelectorAll(".sidebar-nav .nav-link")

    navLinks.forEach((link) => {
      link.classList.remove("active")
      if (link.getAttribute("href") === currentPath) {
        link.classList.add("active")
      }
    })
  }
})

// Auto-refresh functionality for real-time data
function startAutoRefresh() {
  // Refresh activity feed every 30 seconds
  setInterval(() => {
    const activityFeed = document.getElementById("activity-feed")
    if (activityFeed && !document.hidden) {
      htmx.trigger(activityFeed, "refresh")
    }
  }, 30000)

  // Refresh real-time stats every 5 seconds
  setInterval(() => {
    const realtimeStats = document.querySelector(".realtime-stats")
    if (realtimeStats && !document.hidden) {
      htmx.trigger(realtimeStats, "refresh")
    }
  }, 5000)
}

// Initialize auto-refresh when page loads
document.addEventListener("DOMContentLoaded", startAutoRefresh)

// Form validation helpers
function validateForm(formElement) {
  const requiredFields = formElement.querySelectorAll("[required]")
  let isValid = true

  requiredFields.forEach((field) => {
    if (!field.value.trim()) {
      field.classList.add("error")
      isValid = false
    } else {
      field.classList.remove("error")
    }
  })

  return isValid
}

// Settings form handling
document.addEventListener("change", function (e) {
  if (e.target.closest(".settings-card")) {
    // Mark settings as modified
    const saveButton = document.querySelector(".btn-success")
    if (saveButton) {
      saveButton.textContent = "💾 Save Changes*"
      saveButton.classList.add("modified")
    }
  }
})

// Add notification styles if not already present
if (!document.querySelector("#notification-styles")) {
  const notificationStyles = document.createElement("style")
  notificationStyles.id = "notification-styles"
  notificationStyles.textContent = `
        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
            border-left: 4px solid #3b82f6;
            opacity: 0;
            transform: translateX(100%);
            transition: all 0.3s ease;
            z-index: 10000;
            min-width: 300px;
            max-width: 500px;
        }
        
        .notification.show {
            opacity: 1;
            transform: translateX(0);
        }
        
        .notification-content {
            padding: 1rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }
        
        .notification-success {
            border-left-color: #10b981;
        }
        
        .notification-error {
            border-left-color: #ef4444;
        }
        
        .notification-warning {
            border-left-color: #f59e0b;
        }
        
        .notification-message {
            flex: 1;
            font-size: 0.875rem;
            color: #374151;
        }
        
        .notification-close {
            background: none;
            border: none;
            color: #94a3b8;
            cursor: pointer;
            font-size: 1.125rem;
        }
        
        .btn.loading {
            opacity: 0.7;
            pointer-events: none;
        }
        
        .form-input.error,
        .form-textarea.error {
            border-color: #ef4444;
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
        }
        
        .btn.modified {
            background: #f59e0b !important;
        }
    `
  document.head.appendChild(notificationStyles)
}

console.log("Admin Panel JavaScript loaded successfully")
