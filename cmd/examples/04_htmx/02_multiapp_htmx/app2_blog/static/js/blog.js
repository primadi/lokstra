// TechBlog JavaScript Functions

// Search functionality
function openSearch() {
  document.getElementById("search-overlay").style.display = "flex"
  document.querySelector(".search-input").focus()
}

function closeSearch() {
  document.getElementById("search-overlay").style.display = "none"
}

// Theme toggle functionality
function toggleTheme() {
  const body = document.body
  const currentTheme = body.getAttribute("data-theme")
  const newTheme = currentTheme === "dark" ? "light" : "dark"

  body.setAttribute("data-theme", newTheme)
  localStorage.setItem("blog-theme", newTheme)

  // Update theme toggle icon
  const themeToggle = document.querySelector(".theme-toggle")
  if (themeToggle) {
    themeToggle.textContent = newTheme === "dark" ? "☀️" : "🌙"
  }
}

// Load saved theme
function loadTheme() {
  const savedTheme = localStorage.getItem("blog-theme") || "light"
  document.body.setAttribute("data-theme", savedTheme)

  const themeToggle = document.querySelector(".theme-toggle")
  if (themeToggle) {
    themeToggle.textContent = savedTheme === "dark" ? "☀️" : "🌙"
  }
}

// Share article functionality
function shareArticle(title, url) {
  if (navigator.share) {
    navigator
      .share({
        title: title,
        url: url,
      })
      .catch(console.error)
  } else {
    // Fallback to copying URL to clipboard
    navigator.clipboard
      .writeText(url)
      .then(() => {
        showNotification("Article URL copied to clipboard!", "success")
      })
      .catch(() => {
        // Fallback to manual copy
        prompt("Copy this URL:", url)
      })
  }
}

// Keyboard shortcuts
document.addEventListener("keydown", function (e) {
  // Search shortcut (Ctrl+K or Cmd+K)
  if ((e.ctrlKey || e.metaKey) && e.key === "k") {
    e.preventDefault()
    openSearch()
  }

  // Theme toggle (Ctrl+Shift+T)
  if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "T") {
    e.preventDefault()
    toggleTheme()
  }

  // Close search on Escape
  if (e.key === "Escape") {
    closeSearch()
  }
})

// HTMX event handlers for blog
document.body.addEventListener("htmx:beforeRequest", function (evt) {
  console.log("Blog API request starting:", evt.detail.pathInfo.requestPath)

  // Add loading state to clickable elements
  const trigger = evt.detail.elt
  if (
    trigger &&
    (trigger.tagName === "BUTTON" || trigger.classList.contains("article-card"))
  ) {
    trigger.classList.add("loading")
  }
})

document.body.addEventListener("htmx:afterRequest", function (evt) {
  console.log("Blog API request completed:", evt.detail.pathInfo.requestPath)

  // Remove loading state
  const trigger = evt.detail.elt
  if (trigger) {
    trigger.classList.remove("loading")
  }

  if (!evt.detail.successful) {
    console.error("Blog API request failed:", evt.detail)
    showNotification("Request failed. Please try again.", "error")
  }
})

// Navigation active state management
document.addEventListener("htmx:afterRequest", function (evt) {
  if (evt.detail.successful && evt.detail.target.tagName === "MAIN") {
    // Update active navigation based on current URL
    const currentPath = window.location.pathname
    const navLinks = document.querySelectorAll(".main-nav .nav-link")

    navLinks.forEach((link) => {
      link.classList.remove("active")
      if (link.getAttribute("href") === currentPath) {
        link.classList.add("active")
      }
    })
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

// Auto-refresh functionality for dynamic content
function startAutoRefresh() {
  // Refresh activity feed every 60 seconds
  setInterval(() => {
    const activityFeed = document.querySelector(".activity-feed")
    if (activityFeed && !document.hidden) {
      htmx.trigger(activityFeed, "refresh")
    }
  }, 60000)

  // Refresh sidebar widgets every 5 minutes
  setInterval(() => {
    const widgets = document.querySelectorAll(".widget [hx-get]")
    widgets.forEach((widget) => {
      if (!document.hidden) {
        htmx.trigger(widget, "refresh")
      }
    })
  }, 300000)
}

// Infinite scroll for articles
function setupInfiniteScroll() {
  const loadMoreBtn = document.querySelector(".load-more-btn")
  if (!loadMoreBtn) return

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (
          entry.isIntersecting &&
          !entry.target.classList.contains("loading")
        ) {
          entry.target.click()
        }
      })
    },
    {
      threshold: 0.1,
      rootMargin: "50px",
    }
  )

  observer.observe(loadMoreBtn)
}

// Reading progress indicator
function setupReadingProgress() {
  const progressBar = document.createElement("div")
  progressBar.className = "reading-progress"
  progressBar.innerHTML = '<div class="reading-progress-bar"></div>'
  document.body.appendChild(progressBar)

  window.addEventListener("scroll", () => {
    const winScroll =
      document.body.scrollTop || document.documentElement.scrollTop
    const height =
      document.documentElement.scrollHeight -
      document.documentElement.clientHeight
    const scrolled = (winScroll / height) * 100

    const progressBarFill = document.querySelector(".reading-progress-bar")
    if (progressBarFill) {
      progressBarFill.style.width = scrolled + "%"
    }
  })
}

// Form validation
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

// Newsletter form handling
document.addEventListener("submit", function (e) {
  if (e.target.classList.contains("newsletter-form")) {
    const email = e.target.querySelector('input[type="email"]')
    if (email && !isValidEmail(email.value)) {
      e.preventDefault()
      showNotification("Please enter a valid email address.", "error")
      email.classList.add("error")
    }
  }
})

function isValidEmail(email) {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// Like button animation
document.addEventListener("click", function (e) {
  if (e.target.classList.contains("like-btn")) {
    e.target.classList.add("liked")
    setTimeout(() => e.target.classList.remove("liked"), 300)
  }
})

// Smooth scroll for anchor links
document.addEventListener("click", function (e) {
  if (e.target.matches('a[href^="#"]')) {
    e.preventDefault()
    const target = document.querySelector(e.target.getAttribute("href"))
    if (target) {
      target.scrollIntoView({
        behavior: "smooth",
        block: "start",
      })
    }
  }
})

// Initialize everything when DOM is loaded
document.addEventListener("DOMContentLoaded", function () {
  loadTheme()
  startAutoRefresh()
  setupInfiniteScroll()
  setupReadingProgress()

  console.log("TechBlog JavaScript initialized successfully")
})

// Add notification and reading progress styles if not already present
if (!document.querySelector("#blog-dynamic-styles")) {
  const dynamicStyles = document.createElement("style")
  dynamicStyles.id = "blog-dynamic-styles"
  dynamicStyles.textContent = `
        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
            border-left: 4px solid #10b981;
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
            color: #9ca3af;
            cursor: pointer;
            font-size: 1.125rem;
        }
        
        .loading {
            opacity: 0.7;
            pointer-events: none;
        }
        
        .form-input.error,
        .form-textarea.error,
        input.error {
            border-color: #ef4444;
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
        }
        
        .like-btn.liked {
            transform: scale(1.1);
            background: #fecaca !important;
            color: #dc2626 !important;
        }
        
        .reading-progress {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 3px;
            background: rgba(16, 185, 129, 0.2);
            z-index: 9999;
        }
        
        .reading-progress-bar {
            height: 100%;
            background: #10b981;
            width: 0%;
            transition: width 0.2s ease;
        }
        
        [data-theme="dark"] {
            background-color: #1f2937;
            color: #f9fafb;
        }
        
        [data-theme="dark"] .blog-header {
            background: linear-gradient(135deg, #064e3b 0%, #065f46 100%);
        }
        
        [data-theme="dark"] .widget,
        [data-theme="dark"] .article-card,
        [data-theme="dark"] .category-card,
        [data-theme="dark"] .about-hero,
        [data-theme="dark"] .hero-section {
            background: #374151;
            color: #f9fafb;
        }
        
        [data-theme="dark"] .article-content h3,
        [data-theme="dark"] .category-info h3 {
            color: #f9fafb;
        }
    `
  document.head.appendChild(dynamicStyles)
}

// Service worker registration for PWA features (optional)
if ("serviceWorker" in navigator) {
  window.addEventListener("load", function () {
    navigator.serviceWorker
      .register("/sw.js")
      .then(function (registration) {
        console.log("ServiceWorker registration successful")
      })
      .catch(function (err) {
        console.log("ServiceWorker registration failed: ", err)
      })
  })
}
