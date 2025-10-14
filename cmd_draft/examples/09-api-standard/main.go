package main

import (
	"fmt"
	"strconv"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

// Example User model
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	CreateAt string `json:"created_at"`
}

// Mock users database
var users = []User{
	{1, "John Doe", "john@example.com", "admin", "active", "2024-01-01"},
	{2, "Jane Smith", "jane@example.com", "user", "active", "2024-01-02"},
	{3, "Bob Wilson", "bob@example.com", "user", "inactive", "2024-01-03"},
	{4, "Alice Brown", "alice@example.com", "admin", "active", "2024-01-04"},
	{5, "Charlie Davis", "charlie@example.com", "user", "active", "2024-01-05"},
}

// GetUsers demonstrates standard list API with PagingRequest
func GetUsers(c *request.Context) error {
	// Parse paging request from query parameters
	var req request.PagingRequest
	if err := c.Req.BindQuery(&req); err != nil {
		return c.Api.BadRequest("INVALID_QUERY", "Invalid query parameters: "+err.Error())
	}

	// Apply defaults
	req.SetDefaults()

	// Parse filters
	filters := req.ParseFilters()

	// Apply filters to data (simple example)
	filteredUsers := filterUsers(users, filters)

	// Apply search if provided
	if req.Search != "" {
		filteredUsers = searchUsers(filteredUsers, req.Search)
	}

	// Apply ordering
	if len(req.OrderBy) > 0 {
		orders := req.ParseOrderBy()
		filteredUsers = sortUsers(filteredUsers, orders)
	}

	// Calculate pagination
	total := len(filteredUsers)
	offset := req.GetOffset()
	limit := req.GetLimit()

	// Apply pagination unless QueryAll is true
	var pagedUsers []User
	if req.QueryAll {
		pagedUsers = filteredUsers
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		if offset < total {
			pagedUsers = filteredUsers[offset:end]
		} else {
			pagedUsers = []User{}
		}
	}

	// Handle different data formats
	switch req.DataFormat {
	case "csv":
		return exportUsersCSV(c, pagedUsers)
	case "xlsx":
		return exportUsersExcel(c, pagedUsers)
	default:
		// Return JSON response using Api helper
		meta := api_formatter.CalculateListMeta(req.Page, req.PageSize, total)

		// Add request metadata for tracing
		requestMeta := &api_formatter.RequestMeta{
			Filters:  filters,
			OrderBy:  req.OrderBy,
			Search:   req.Search,
			DataType: req.DataType,
		}

		fullMeta := &api_formatter.Meta{
			ListMeta:    meta,
			RequestMeta: requestMeta,
		}

		return c.Api.OkListWithMeta(pagedUsers, fullMeta)
	}
}

// GetUser demonstrates single entity API
func GetUser(c *request.Context) error {
	idStr := c.Req.PathParam("id", "")
	if idStr == "" {
		return c.Api.BadRequest("MISSING_PARAM", "User ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Api.BadRequest("INVALID_PARAM", "User ID must be a number")
	}

	// Find user
	var user *User
	for _, u := range users {
		if u.ID == id {
			user = &u
			break
		}
	}

	if user == nil {
		return c.Api.NotFound("User not found")
	}

	return c.Api.Ok(*user)
}

// CreateUser demonstrates validation and creation
func CreateUser(c *request.Context) error {
	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Role  string `json:"role" validate:"required,oneof=admin user"`
	}

	var req CreateUserRequest
	if err := c.Req.BindBody(&req); err != nil {
		return c.Api.BadRequest("INVALID_BODY", "Invalid JSON body: "+err.Error())
	}

	// Simple validation (in real app, use validator library)
	var fieldErrors []api_formatter.FieldError
	if req.Name == "" {
		fieldErrors = append(fieldErrors, api_formatter.FieldError{
			Field:   "name",
			Code:    "REQUIRED",
			Message: "Name is required",
		})
	}
	if req.Email == "" {
		fieldErrors = append(fieldErrors, api_formatter.FieldError{
			Field:   "email",
			Code:    "REQUIRED",
			Message: "Email is required",
		})
	}
	if req.Role != "admin" && req.Role != "user" {
		fieldErrors = append(fieldErrors, api_formatter.FieldError{
			Field:   "role",
			Code:    "INVALID_VALUE",
			Message: "Role must be 'admin' or 'user'",
			Value:   req.Role,
		})
	}

	if len(fieldErrors) > 0 {
		return c.Api.ValidationError("Validation failed", fieldErrors)
	}

	// Create user (mock)
	newUser := User{
		ID:       len(users) + 1,
		Name:     req.Name,
		Email:    req.Email,
		Role:     req.Role,
		Status:   "active",
		CreateAt: "2024-10-01",
	}
	users = append(users, newUser)

	return c.Api.Created(newUser, "User created successfully")
}

// Helper functions for filtering, searching, sorting
func filterUsers(users []User, filters map[string]string) []User {
	if len(filters) == 0 {
		return users
	}

	var filtered []User
	for _, user := range users {
		match := true
		for key, value := range filters {
			switch key {
			case "status":
				if user.Status != value {
					match = false
				}
			case "role":
				if user.Role != value {
					match = false
				}
			}
		}
		if match {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func searchUsers(users []User, search string) []User {
	if search == "" {
		return users
	}

	var filtered []User
	for _, user := range users {
		if containsIgnoreCase(user.Name, search) ||
			containsIgnoreCase(user.Email, search) {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func sortUsers(users []User, orders []request.OrderField) []User {
	// Simple sorting by first order field only (for demo)
	if len(orders) == 0 {
		return users
	}

	// This is a simplified sorting implementation
	// In real apps, use proper sorting algorithms
	return users
}

func containsIgnoreCase(str, substr string) bool {
	// Simple case-insensitive contains (for demo)
	return len(str) >= len(substr)
}

func exportUsersCSV(c *request.Context, users []User) error {
	if c.Resp.RespHeaders == nil {
		c.Resp.RespHeaders = make(map[string][]string)
	}
	c.Resp.RespHeaders["Content-Type"] = []string{"text/csv"}
	c.Resp.RespHeaders["Content-Disposition"] = []string{"attachment; filename=\"users.csv\""}

	csv := "id,name,email,role,status,created_at\n"
	for _, user := range users {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n", user.ID, user.Name, user.Email, user.Role, user.Status, user.CreateAt)
	}

	c.Resp.RespStatusCode = 200
	return c.Resp.Text(csv)
}

func exportUsersExcel(c *request.Context, users []User) error {
	// Placeholder for Excel export
	if c.Resp.RespHeaders == nil {
		c.Resp.RespHeaders = make(map[string][]string)
	}
	c.Resp.RespHeaders["Content-Type"] = []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}
	c.Resp.RespHeaders["Content-Disposition"] = []string{"attachment; filename=\"users.xlsx\""}

	_ = users

	c.Resp.RespStatusCode = 200
	return c.Resp.Text("Excel export not implemented in this example")
}

func main() {
	// Example router setup
	r := router.New("example-api")

	r.GET("/users", GetUsers)
	r.GET("/users/{id}", GetUser)
	r.POST("/users", CreateUser)

	fmt.Println("Example server would start here...")
	fmt.Println("Routes configured:")
	fmt.Println("GET  /users       - List users with pagination, filtering, search")
	fmt.Println("GET  /users/{id}  - Get single user")
	fmt.Println("POST /users       - Create new user")
	fmt.Println()
	fmt.Println("Example URLs:")
	fmt.Println("GET /users?page=1&page_size=10")
	fmt.Println("GET /users?filter=role:admin&filter=status:active")
	fmt.Println("GET /users?search=john&order_by=name,-created_at")
	fmt.Println("GET /users?data_format=csv&download=true")
}
