package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/primadi/lokstra"
)

// Shared HTTP client for internal API calls to prevent connection buildup
var (
	internalAPIClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Simple cache to prevent rapid repeated requests
	lastUsersCache struct {
		data      []User
		timestamp time.Time
	}
	cacheTimeout = 2 * time.Second
)

// Helper functions

// getSidebarHTML returns the sidebar HTML for all pages using template and dynamic data
func getSidebarHTML(currentPage string) string {
	// Get menu data based on current page
	menuData := getMenuData(currentPage)

	// Use the modern template rendering function
	return renderSidebar(menuData)
}

// Modern Handlers using PageHandler - consistent behavior across full page and HTMX loads

// CreateDashboardHandler creates a handler for dashboard that works consistently with both full page and HTMX requests
func CreateDashboardHandler() lokstra.HandlerFunc {
	return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
		// Prepare dashboard data
		dashboardData := struct {
			TotalUsers    int
			ActiveUsers   int
			InactiveUsers int
			RecentUsers   []struct {
				Username    string
				Email       string
				Initial     string
				AvatarColor string
				Status      string
				StatusColor string
			}
		}{
			TotalUsers:    142,
			ActiveUsers:   128,
			InactiveUsers: 14,
			RecentUsers: []struct {
				Username    string
				Email       string
				Initial     string
				AvatarColor string
				Status      string
				StatusColor string
			}{
				{
					Username:    "admin",
					Email:       "admin@example.com",
					Initial:     "A",
					AvatarColor: "bg-blue-600",
					Status:      "Active",
					StatusColor: "bg-green-600 text-green-100",
				},
				{
					Username:    "user1",
					Email:       "user1@example.com",
					Initial:     "U",
					AvatarColor: "bg-green-600",
					Status:      "Active",
					StatusColor: "bg-green-600 text-green-100",
				},
				{
					Username:    "user2",
					Email:       "user2@example.com",
					Initial:     "U",
					AvatarColor: "bg-yellow-600",
					Status:      "Inactive",
					StatusColor: "bg-red-600 text-red-100",
				},
			},
		}

		// Render content using template
		content := renderPageContent("dashboard", dashboardData)

		return &PageContent{
			HTML:        content,
			Title:       "Dashboard",
			CurrentPage: "dashboard",
		}, nil
	})
}

// CreateUsersHandler creates a handler for users page with page-specific assets that work consistently in both full page and HTMX requests
func CreateUsersHandler() lokstra.HandlerFunc {
	return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
		// Call the user.list API endpoint to get real data
		users, err := getUsersFromAPI(c)
		if err != nil {
			// Fallback to empty table if API call fails
			users = []User{}
		}

		// Prepare users data for template
		usersData := struct {
			Users []User
		}{
			Users: users,
		}

		// Render content using template
		content := renderPageContent("users", usersData)

		return &PageContent{
			HTML:        content,
			Title:       "User Management",
			CurrentPage: "users",
			// Page-specific embedded JavaScript
			EmbeddedScripts: []string{
				"table-enhancements",
				"navigation-enhancements",
			},
			// Note: Page-specific CSS is now automatically loaded via loadPageCSS
		}, nil
	})
}

// User represents the user data structure from API response
type User struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt *string                `json:"created_at"`
	UpdatedAt *string                `json:"updated_at"`
	LastLogin *string                `json:"last_login"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// APIResponse represents the actual API response structure
type APIResponse struct {
	Code    string `json:"code"`
	Success bool   `json:"success"`
	Data    struct {
		Data    []User `json:"data"`
		Filters struct {
			Applied map[string]interface{} `json:"applied"`
		} `json:"filters"`
		Pagination struct {
			HasNext    bool `json:"has_next"`
			HasPrev    bool `json:"has_prev"`
			Page       int  `json:"page"`
			PageSize   int  `json:"page_size"`
			Total      int  `json:"total"`
			TotalPages int  `json:"total_pages"`
		} `json:"pagination"`
	} `json:"data"`
}

// getUsersFromAPI calls the internal user.list API to get real user data
func getUsersFromAPI(c *lokstra.Context) ([]User, error) {
	_ = c

	// Check cache first to prevent rapid repeated requests
	if time.Since(lastUsersCache.timestamp) < cacheTimeout && lastUsersCache.data != nil {
		fmt.Printf("[DEBUG] Using cached users data (age: %v)\n", time.Since(lastUsersCache.timestamp))
		return lastUsersCache.data, nil
	}

	fmt.Printf("[DEBUG] Fetching users from API...\n")

	// Create request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Make internal HTTP request to /api/v1/users
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/api/v1/users", nil)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create request: %v\n", err)
		return nil, err
	}

	// Copy important headers from original request
	req.Header.Set("Content-Type", "application/json")

	resp, err := internalAPIClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] API request failed: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[ERROR] API returned status %d\n", resp.StatusCode)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		fmt.Printf("[ERROR] Failed to decode API response: %v\n", err)
		return nil, err
	}

	// Check if API call was successful
	if !apiResp.Success {
		fmt.Printf("[ERROR] API call failed with code: %s\n", apiResp.Code)
		return nil, fmt.Errorf("API call failed with code: %s", apiResp.Code)
	}

	// Update cache
	lastUsersCache.data = apiResp.Data.Data
	lastUsersCache.timestamp = time.Now()

	fmt.Printf("[DEBUG] Successfully fetched %d users\n", len(apiResp.Data.Data))
	return apiResp.Data.Data, nil
}

// CreateRolesHandler creates a handler for roles page that works consistently with both full page and HTMX requests
func CreateRolesHandler() lokstra.HandlerFunc {
	return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
		content := `
			<div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700">
				<div class="p-6 border-b border-gray-700 flex justify-between items-center">
					<h2 class="text-xl font-bold text-gray-100">Roles & Permissions</h2>
					<button class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors">
						Create Role
					</button>
				</div>
				<div class="p-6">
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						<div class="bg-gray-700 p-4 rounded-lg border border-gray-600">
							<h3 class="text-lg font-semibold text-blue-400 mb-2">Administrator</h3>
							<p class="text-gray-300 text-sm mb-3">Full system access</p>
							<div class="flex space-x-2">
								<button class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">Edit</button>
								<button class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm">Delete</button>
							</div>
						</div>
						<div class="bg-gray-700 p-4 rounded-lg border border-gray-600">
							<h3 class="text-lg font-semibold text-green-400 mb-2">User Manager</h3>
							<p class="text-gray-300 text-sm mb-3">User management only</p>
							<div class="flex space-x-2">
								<button class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">Edit</button>
								<button class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm">Delete</button>
							</div>
						</div>
						<div class="bg-gray-700 p-4 rounded-lg border border-gray-600">
							<h3 class="text-lg font-semibold text-yellow-400 mb-2">Viewer</h3>
							<p class="text-gray-300 text-sm mb-3">Read-only access</p>
							<div class="flex space-x-2">
								<button class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">Edit</button>
								<button class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm">Delete</button>
							</div>
						</div>
					</div>
				</div>
			</div>
		`

		return &PageContent{
			HTML:        content,
			Title:       "Roles & Permissions",
			CurrentPage: "roles",
			// Page-specific styles for role cards
			CustomCSS: `
				.role-card {
					transition: all 0.3s ease;
				}
				
				.role-card:hover {
					transform: translateY(-2px);
					box-shadow: 0 8px 25px rgba(0, 0, 0, 0.4);
				}
				
				.role-actions button {
					transition: all 0.2s ease;
				}
				
				.role-actions button:hover {
					transform: scale(1.05);
				}
			`,
		}, nil
	})
}

// CreateSettingsHandler creates a handler for settings page that works consistently with both full page and HTMX requests
func CreateSettingsHandler() lokstra.HandlerFunc {
	return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
		content := `
			<div class="space-y-6">
				<div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700">
					<div class="p-6 border-b border-gray-700">
						<h2 class="text-xl font-bold text-gray-100">General Settings</h2>
					</div>
					<div class="p-6 space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Application Name</label>
							<input type="text" value="User Management System" 
								   class="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Max Users per Page</label>
							<select class="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
								<option>10</option>
								<option selected>25</option>
								<option>50</option>
								<option>100</option>
							</select>
						</div>
						<div class="flex items-center">
							<input type="checkbox" id="email-notifications" class="mr-2 text-blue-600">
							<label for="email-notifications" class="text-gray-300">Enable email notifications</label>
						</div>
					</div>
				</div>
				
				<div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700">
					<div class="p-6 border-b border-gray-700">
						<h2 class="text-xl font-bold text-gray-100">Security Settings</h2>
					</div>
					<div class="p-6 space-y-4">
						<div class="flex items-center">
							<input type="checkbox" id="two-factor" class="mr-2 text-blue-600" checked>
							<label for="two-factor" class="text-gray-300">Require two-factor authentication</label>
						</div>
						<div class="flex items-center">
							<input type="checkbox" id="session-timeout" class="mr-2 text-blue-600">
							<label for="session-timeout" class="text-gray-300">Auto-logout after 30 minutes</label>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Password Minimum Length</label>
							<input type="number" value="8" min="6" max="32"
								   class="w-32 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
						</div>
					</div>
				</div>
			</div>
		`

		return &PageContent{
			HTML:        content,
			Title:       "Settings",
			CurrentPage: "settings",
			// Page-specific styles for settings forms
			CustomCSS: `
				.settings-section {
					transition: all 0.3s ease;
				}
				
				.settings-section:hover {
					transform: translateY(-1px);
					box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
				}
				
				.settings-input:focus {
					transform: scale(1.02);
					transition: transform 0.2s ease;
				}
				
				.checkbox-custom {
					transition: all 0.2s ease;
				}
				
				.checkbox-custom:hover {
					transform: scale(1.1);
				}
			`,
		}, nil
	})
}

// CreateUserFormPageHandler creates handler for user create/edit form using modern layout system
func CreateUserFormPageHandler() lokstra.HandlerFunc {
	return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
		userID := c.GetPathParam("id")
		isEdit := userID != ""

		title := "Create New User"
		buttonText := "Create User"
		formAction := "/users/create"
		usernameValue := ""
		emailValue := ""
		fullNameValue := ""
		statusValue := "active"

		// If editing, load user data
		if isEdit {
			title = "Edit User"
			buttonText = "Update User"
			formAction = "/users/update/" + userID

			// Load user data from API
			user, err := getUserByIDFromAPI(c, userID)
			if err == nil {
				usernameValue = user.Username
				emailValue = user.Email
				if user.Metadata != nil && user.Metadata["full_name"] != nil {
					if fullName, ok := user.Metadata["full_name"].(string); ok {
						fullNameValue = fullName
					}
				}
				if user.IsActive {
					statusValue = "active"
				} else {
					statusValue = "inactive"
				}
			}
		}

		// Generate form content
		content := `
			<div class="max-w-2xl mx-auto bg-gray-800 rounded-lg shadow-lg border border-gray-700">
				<div class="p-6 border-b border-gray-700">
					<h2 class="text-xl font-bold text-gray-100">` + title + `</h2>
				</div>
				<form class="p-6" 
					  hx-post="` + formAction + `" 
					  hx-target="#form-result"
					  hx-swap="innerHTML">
					<div class="grid grid-cols-1 gap-6">
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Username</label>
							<input type="text" name="username" required value="` + usernameValue + `"
								   class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								   placeholder="Enter username">
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Email Address</label>
							<input type="email" name="email" required value="` + emailValue + `"
								   class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								   placeholder="Enter email address">
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Full Name</label>
							<input type="text" name="full_name" required value="` + fullNameValue + `"
								   class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								   placeholder="Enter full name">
						</div>`

		// Only show password field for new users
		if !isEdit {
			content += `
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Password</label>
							<input type="password" name="password" required
								   class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								   placeholder="Enter password">
						</div>`
		}

		content += `
						<div>
							<label class="block text-sm font-medium text-gray-300 mb-2">Status</label>
							<select name="status" required
									class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
								<option value="active"` + getSelectedAttribute(statusValue, "active") + `>Active</option>
								<option value="inactive"` + getSelectedAttribute(statusValue, "inactive") + `>Inactive</option>
								<option value="suspended"` + getSelectedAttribute(statusValue, "suspended") + `>Suspended</option>
							</select>
						</div>
					</div>
					
					<div class="mt-6 flex justify-end space-x-3">
						<button type="button" 
								hx-get="/users" 
								hx-target="#main-content"
								hx-push-url="/users"
								hx-indicator="#loading-indicator"
								class="nav-page px-4 py-2 border border-gray-600 rounded-md text-gray-300 hover:bg-gray-700 transition-colors">
							Cancel
						</button>
						<button type="submit" 
								class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors">
							` + buttonText + `
						</button>
					</div>
				</form>
				<div id="form-result" class="p-6 pt-0"></div>
			</div>`

		return &PageContent{
			HTML:        content,
			Title:       title,
			CurrentPage: "user_form",
			// Page-specific embedded JavaScript
			EmbeddedScripts: []string{
				"user-form-validation",
				"navigation-enhancements",
			},
			// External scripts (if needed)
			Scripts: []string{
				"https://cdn.jsdelivr.net/npm/validator@13.7.0/validator.min.js",
			},
			// Page-specific styles for enhanced form UX
			CustomCSS: `
				.user-form-container {
					background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
					border-radius: 12px;
					box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
				}
				
				.form-field {
					transition: all 0.3s ease;
				}
				
				.form-field:focus-within {
					transform: translateY(-1px);
					box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
				}
				
				.form-input:focus {
					transform: scale(1.02);
					transition: transform 0.2s ease;
				}
				
				.form-button {
					transition: all 0.3s ease;
				}
				
				.form-button:hover {
					transform: translateY(-1px);
					box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
				}
				
				.validation-feedback {
					animation: fadeIn 0.3s ease;
				}
				
				.field-error input {
					border-color: #ef4444;
					box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
				}
				
				@keyframes fadeIn {
					from { opacity: 0; transform: translateY(-10px); }
					to { opacity: 1; transform: translateY(0); }
				}
			`,
		}, nil
	})
}

// Helper function to get selected attribute for select options
func getSelectedAttribute(currentValue, optionValue string) string {
	if currentValue == optionValue {
		return ` selected`
	}
	return ""
}

// getUserByIDFromAPI gets a specific user by ID from the API
func getUserByIDFromAPI(c *lokstra.Context, userID string) (*User, error) {
	_ = c
	// Create request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Make internal HTTP request to /api/v1/users/id/{id}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/api/v1/users/id/"+userID, nil)
	if err != nil {
		return nil, err
	}

	// Copy important headers from original request
	req.Header.Set("Content-Type", "application/json")

	resp, err := internalAPIClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Code    string `json:"code"`
		Success bool   `json:"success"`
		Data    User   `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Check if API call was successful
	if !apiResp.Success {
		return nil, fmt.Errorf("API call failed with code: %s", apiResp.Code)
	}

	return &apiResp.Data, nil
}

// CreateUserSubmitHandler handles form submission for creating new users
func CreateUserSubmitHandler() lokstra.HandlerFunc {
	return func(c *lokstra.Context) error {
		// Parse form data
		username := c.GetQueryParam("username")
		email := c.GetQueryParam("email")
		fullName := c.GetQueryParam("full_name")
		password := c.GetQueryParam("password")
		status := c.GetQueryParam("status")

		// Try to get form values from POST body if not in query params
		if username == "" {
			// Parse request body to get form data
			body, err := c.GetRawRequestBody()
			if err == nil {
				// Simple form parsing - look for username=value&email=value format
				bodyStr := string(body)
				formParams := parseFormData(bodyStr)
				username = formParams["username"]
				email = formParams["email"]
				fullName = formParams["full_name"]
				password = formParams["password"]
				status = formParams["status"]
			}
		}

		// Validate required fields
		if username == "" || email == "" || fullName == "" || password == "" {
			return c.HTML(200, `<div class="bg-red-600 text-white p-3 rounded mb-4">All fields are required</div>`)
		}

		// Create request payload
		payload := map[string]interface{}{
			"username":  username,
			"email":     email,
			"password":  password,
			"full_name": fullName,
			"is_active": status == "active",
		}

		payloadJSON, _ := json.Marshal(payload)

		// Create request with context timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Make API request
		req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8081/api/v1/users", bytes.NewBuffer(payloadJSON))
		if err != nil {
			return c.HTML(200, `<div class="bg-red-600 text-white p-3 rounded mb-4">Failed to create request</div>`)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := internalAPIClient.Do(req)
		if err != nil {
			return c.HTML(200, `<div class="bg-red-600 text-white p-3 rounded mb-4">Failed to call API</div>`)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			// Success - redirect to users list with HTMX
			return c.HTML(200, `
				<div class="bg-green-600 text-white p-3 rounded mb-4">User created successfully!</div>
				<script>
					setTimeout(function() {
						htmx.ajax('GET', '/users', {target:'#main-content'});
					}, 1000);
				</script>
			`)
		} else {
			// Error response
			return c.HTML(200, `<div class="bg-red-600 text-white p-3 rounded mb-4">Failed to create user. Please check your input.</div>`)
		}
	}
}

// UpdateUserSubmitHandler handles form submission for updating existing users
func UpdateUserSubmitHandler() lokstra.HandlerFunc {
	return func(c *lokstra.Context) error {
		// Simply call the update handler directly - it now supports smart binding
		updateHandler := CreateUpdateUserHandler()
		err := updateHandler(c)

		if err != nil {
			return c.HTML(200, `<div class="bg-red-600 text-white p-3 rounded mb-4">Failed to update user: `+err.Error()+`</div>`)
		}

		// Success - redirect to users list with HTMX
		return c.HTML(200, `
			<div class="bg-green-600 text-white p-3 rounded mb-4">User updated successfully!</div>
			<script>
				setTimeout(function() {
					htmx.ajax('GET', '/users', {target:'#main-content'});
				}, 1000);
			</script>
		`)
	}
}

// parseFormData is a simple form data parser for URL-encoded form data
func parseFormData(body string) map[string]string {
	params := make(map[string]string)

	// Split by & to get key=value pairs
	pairs := strings.Split(body, "&")
	for _, pair := range pairs {
		// Split by = to get key and value
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Simple URL decode for spaces and common characters
			value = strings.ReplaceAll(value, "+", " ")
			params[key] = value
		}
	}

	return params
}
