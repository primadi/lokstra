// User Form Validation JavaScript
// This file will be embedded into the Go binary

function validateUsername() {
  const username = document.getElementById("username").value;
  const errorDiv = document.getElementById("username-error");

  if (!username || username.length < 3) {
    showError(errorDiv, "Username must be at least 3 characters long");
    return false;
  }

  if (!validator.isAlphanumeric(username)) {
    showError(errorDiv, "Username can only contain letters and numbers");
    return false;
  }

  hideError(errorDiv);
  return true;
}

function validateEmail() {
  const email = document.getElementById("email").value;
  const errorDiv = document.getElementById("email-error");

  if (!email) {
    showError(errorDiv, "Email is required");
    return false;
  }

  if (!validator.isEmail(email)) {
    showError(errorDiv, "Please enter a valid email address");
    return false;
  }

  hideError(errorDiv);
  return true;
}

function showError(errorDiv, message) {
  errorDiv.textContent = message;
  errorDiv.classList.remove("hidden");
  errorDiv.parentElement.classList.add("field-error");
}

function hideError(errorDiv) {
  errorDiv.classList.add("hidden");
  errorDiv.parentElement.classList.remove("field-error");
}

function initializeFormValidation() {
  // Form submission handler
  const form = document.getElementById("userForm");
  if (form) {
    form.addEventListener("submit", function (e) {
      e.preventDefault();

      // Validate all fields
      const isUsernameValid = validateUsername();
      const isEmailValid = validateEmail();

      if (isUsernameValid && isEmailValid) {
        // Submit form data
        console.log("Form validation passed, submitting...");
        // Handle actual form submission here
        this.submit();
      } else {
        console.log("Form validation failed");
      }
    });
  }

  // Real-time validation
  const usernameField = document.getElementById("username");
  const emailField = document.getElementById("email");

  if (usernameField) {
    usernameField.addEventListener("blur", validateUsername);
    usernameField.addEventListener("focus", function () {
      this.classList.add("form-field-focus");
    });
    usernameField.addEventListener("blur", function () {
      this.classList.remove("form-field-focus");
    });
  }

  if (emailField) {
    emailField.addEventListener("blur", validateEmail);
    emailField.addEventListener("focus", function () {
      this.classList.add("form-field-focus");
    });
    emailField.addEventListener("blur", function () {
      this.classList.remove("form-field-focus");
    });
  }

  console.log("User form validation initialized");
}

// Auto-initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeFormValidation);
} else {
  initializeFormValidation();
}
