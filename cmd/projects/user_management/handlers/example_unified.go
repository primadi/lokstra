package handlers

import (
	"github.com/primadi/lokstra"
)

// Example: Advanced User Form with Custom Assets using UnifiedPageHandler

func CreateAdvancedUserFormHandler() lokstra.HandlerFunc {
	return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
		userID := c.GetPathParam("id")
		isEdit := userID != ""

		// Load user data if editing
		var user *User
		if isEdit {
			var err error
			user, err = getUserByIDFromAPI(c, userID)
			if err != nil {
				return nil, err
			}
		}

		// Generate form HTML
		formHTML := generateUserFormHTML(user, isEdit)

		return &PageContent{
			HTML:        formHTML,
			Title:       getTitle(isEdit),
			CurrentPage: "user_form",

			// Page-specific scripts - CONSISTENT di full page DAN HTMX
			Scripts: []string{
				"https://cdn.jsdelivr.net/npm/validator@13.7.0/validator.min.js",
			},

			// Page-specific custom CSS - CONSISTENT di full page DAN HTMX
			CustomCSS: `
				.user-form-container {
					background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
					border-radius: 12px;
					box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
				}
				
				.form-field-focus {
					transform: scale(1.02);
					transition: transform 0.2s ease;
				}
				
				.validation-error {
					color: #ef4444;
					font-size: 0.875rem;
					margin-top: 0.25rem;
				}
			`,

			// Page-specific meta tags - CONSISTENT di full page DAN HTMX
			MetaTags: map[string]string{
				"description": "User management form with advanced validation",
				"keywords":    "user, form, validation, management",
			},
		}, nil
	})
}

// Helper functions
func getTitle(isEdit bool) string {
	if isEdit {
		return "Edit User"
	}
	return "Create User"
}

func conditionalValue(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func generateUserFormHTML(user *User, isEdit bool) string {
	// Form generation logic with inline JavaScript for validation
	return `
		<div class="user-form-container max-w-2xl mx-auto p-8">
			<h2 class="text-2xl font-bold text-white mb-6">` + conditionalValue(isEdit, "Edit User", "Create New User") + `</h2>
			
			<form id="userForm" class="space-y-6">
				<div class="form-field">
					<label class="block text-sm font-medium text-gray-300 mb-2">Username</label>
					<input type="text" name="username" id="username" required
						   class="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white 
								  focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
								  transition-all duration-200"
						   placeholder="Enter username"
						   onFocus="this.classList.add('form-field-focus')"
						   onBlur="this.classList.remove('form-field-focus'); validateUsername()">
					<div id="username-error" class="validation-error hidden"></div>
				</div>
				
				<div class="form-field">
					<label class="block text-sm font-medium text-gray-300 mb-2">Email</label>
					<input type="email" name="email" id="email" required
						   class="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white 
								  focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
								  transition-all duration-200"
						   placeholder="Enter email address"
						   onFocus="this.classList.add('form-field-focus')"
						   onBlur="this.classList.remove('form-field-focus'); validateEmail()">
					<div id="email-error" class="validation-error hidden"></div>
				</div>
				
				<div class="form-field">
					<label class="block text-sm font-medium text-gray-300 mb-2">Full Name</label>
					<input type="text" name="full_name" id="full_name" required
						   class="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white 
								  focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
								  transition-all duration-200"
						   placeholder="Enter full name"
						   onFocus="this.classList.add('form-field-focus')"
						   onBlur="this.classList.remove('form-field-focus')">
				</div>
				
				<div class="flex justify-end space-x-4 pt-6">
					<button type="button" onclick="cancelForm()"
							class="px-6 py-3 border border-gray-600 rounded-lg text-gray-300 
								   hover:bg-gray-700 transition-colors">
						Cancel
					</button>
					<button type="submit"
							class="px-6 py-3 bg-blue-600 text-white rounded-lg 
								   hover:bg-blue-700 transition-colors">
						` + conditionalValue(isEdit, "Update User", "Create User") + `
					</button>
				</div>
			</form>
		</div>
		
		<!-- Page-specific JavaScript - WORKS IN BOTH full page AND HTMX -->
		<script>
			// Custom validation using validator.js library
			function validateUsername() {
				const username = document.getElementById('username').value;
				const errorDiv = document.getElementById('username-error');
				
				if (!username || username.length < 3) {
					showError(errorDiv, 'Username must be at least 3 characters long');
					return false;
				}
				
				if (!validator.isAlphanumeric(username)) {
					showError(errorDiv, 'Username can only contain letters and numbers');
					return false;
				}
				
				hideError(errorDiv);
				return true;
			}
			
			function validateEmail() {
				const email = document.getElementById('email').value;
				const errorDiv = document.getElementById('email-error');
				
				if (!email) {
					showError(errorDiv, 'Email is required');
					return false;
				}
				
				if (!validator.isEmail(email)) {
					showError(errorDiv, 'Please enter a valid email address');
					return false;
				}
				
				hideError(errorDiv);
				return true;
			}
			
			function showError(errorDiv, message) {
				errorDiv.textContent = message;
				errorDiv.classList.remove('hidden');
			}
			
			function hideError(errorDiv) {
				errorDiv.classList.add('hidden');
			}
			
			function cancelForm() {
				// Navigate back using HTMX
				htmx.ajax('GET', '/users', {target:'#main-content'});
			}
			
			// Form submission
			document.getElementById('userForm').addEventListener('submit', function(e) {
				e.preventDefault();
				
				// Validate all fields
				const isUsernameValid = validateUsername();
				const isEmailValid = validateEmail();
				
				if (isUsernameValid && isEmailValid) {
					// Submit form data
					const formData = new FormData(this);
					console.log('Form submitted with validation passed');
					// Handle form submission...
				}
			});
			
			console.log('Advanced user form initialized - works in both full page and HTMX!');
		</script>
	`
}
