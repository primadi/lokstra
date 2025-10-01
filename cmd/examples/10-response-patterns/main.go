package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

// Example User model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{1, "John Doe", "john@example.com"},
	{2, "Jane Smith", "jane@example.com"},
}

// Example 1: Using Resp (Base response - no wrapping)
func GetUsersRaw(c *request.Context) error {
	// Direct JSON response without ApiResponse wrapper
	return c.Resp.WithStatus(200).Json(users)
}

// Example 2: Using Api (Opinionated response - wrapped in ApiResponse)
func GetUsersApi(c *request.Context) error {
	// Wrapped in ApiResponse structure
	return c.Api.Ok(users)
}

// Example 3: Using Resp for custom structure
func GetUsersCustom(c *request.Context) error {
	customResp := map[string]any{
		"users":     users,
		"count":     len(users),
		"timestamp": "2024-10-01T10:00:00Z",
	}
	return c.Resp.WithStatus(200).Json(customResp)
}

// Example 4: Using Api with message
func GetUsersApiWithMessage(c *request.Context) error {
	return c.Api.OkWithMessage(users, "Users retrieved successfully")
}

// Example 5: Using Api for list with pagination
func GetUsersApiPaginated(c *request.Context) error {
	meta := api_formatter.CalculateListMeta(1, 10, len(users))
	return c.Api.OkList(users, meta)
}

// Example 6: Using Resp for plain text response
func GetHealthRaw(c *request.Context) error {
	return c.Resp.WithStatus(200).Text("OK")
}

// Example 7: Using Api for structured health check
func GetHealthApi(c *request.Context) error {
	health := map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	}
	return c.Api.Ok(health)
}

// Example 8: Error handling comparison
func GetUserByIdRaw(c *request.Context) error {
	id := c.Req.PathParam("id", "")
	if id == "" {
		// Custom error structure
		errorResp := map[string]string{
			"error": "ID is required",
			"code":  "MISSING_ID",
		}
		return c.Resp.WithStatus(400).Json(errorResp)
	}

	// User not found
	return c.Resp.WithStatus(404).Json(map[string]string{
		"error": "User not found",
	})
}

func GetUserByIdApi(c *request.Context) error {
	id := c.Req.PathParam("id", "")
	if id == "" {
		// Structured error with ApiResponse wrapper
		return c.Api.BadRequest("MISSING_ID", "ID is required")
	}

	// User not found
	return c.Api.NotFound("User not found")
}

// Example 9: Validation errors
func CreateUserRaw(c *request.Context) error {
	var user User
	if err := c.Req.BindBody(&user); err != nil {
		// Custom validation error structure
		errorResp := map[string]any{
			"error":   "Invalid input",
			"details": err.Error(),
		}
		return c.Resp.WithStatus(400).Json(errorResp)
	}

	// Success
	user.ID = len(users) + 1
	users = append(users, user)

	return c.Resp.WithStatus(201).Json(user)
}

func CreateUserApi(c *request.Context) error {
	var user User
	if err := c.Req.BindBody(&user); err != nil {
		// Structured validation error
		fields := []api_formatter.FieldError{
			{
				Field:   "body",
				Code:    "INVALID_JSON",
				Message: err.Error(),
			},
		}
		return c.Api.ValidationError("Invalid input", fields)
	}

	// Success with message
	user.ID = len(users) + 1
	users = append(users, user)

	return c.Api.Created(user, "User created successfully")
}

func main() {
	r := router.New("example-response-patterns")

	// Raw responses (no ApiResponse wrapper)
	r.GET("/raw/users", GetUsersRaw)
	r.GET("/raw/users/custom", GetUsersCustom)
	r.GET("/raw/health", GetHealthRaw)
	r.GET("/raw/users/{id}", GetUserByIdRaw)
	r.POST("/raw/users", CreateUserRaw)

	// API responses (wrapped in ApiResponse)
	r.GET("/api/users", GetUsersApi)
	r.GET("/api/users/message", GetUsersApiWithMessage)
	r.GET("/api/users/paginated", GetUsersApiPaginated)
	r.GET("/api/health", GetHealthApi)
	r.GET("/api/users/{id}", GetUserByIdApi)
	r.POST("/api/users", CreateUserApi)

	fmt.Println("🚀 Example Response Patterns Server")
	fmt.Println()
	fmt.Println("📋 Raw Response Examples (c.Resp):")
	fmt.Println("GET  /raw/users          - Direct JSON array")
	fmt.Println("GET  /raw/users/custom   - Custom JSON structure")
	fmt.Println("GET  /raw/health         - Plain text response")
	fmt.Println("GET  /raw/users/123      - Custom error structure")
	fmt.Println("POST /raw/users          - Direct user object")
	fmt.Println()
	fmt.Println("🎯 API Response Examples (c.Api):")
	fmt.Println("GET  /api/users          - Wrapped in ApiResponse")
	fmt.Println("GET  /api/users/message  - ApiResponse with message")
	fmt.Println("GET  /api/users/paginated- ListResponse with pagination")
	fmt.Println("GET  /api/health         - Structured health check")
	fmt.Println("GET  /api/users/123      - Structured error response")
	fmt.Println("POST /api/users          - ApiResponse with validation")

	printResponseExamples()
}

func printResponseExamples() {
	fmt.Println("\n📄 Response Format Examples:")

	fmt.Println("\n1. Raw Response (c.Resp):")
	fmt.Println(`[
  {"id": 1, "name": "John Doe", "email": "john@example.com"},
  {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
]`)

	fmt.Println("\n2. API Response (c.Api):")
	fmt.Println(`{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
  ]
}`)

	fmt.Println("\n3. API Response with Message:")
	fmt.Println(`{
  "status": "success",
  "message": "Users retrieved successfully",
  "data": [...]
}`)

	fmt.Println("\n4. API List Response with Pagination:")
	fmt.Println(`{
  "status": "success",
  "data": [...],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total": 2,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}`)

	fmt.Println("\n5. Raw Error:")
	fmt.Println(`{"error": "User not found"}`)

	fmt.Println("\n6. API Structured Error:")
	fmt.Println(`{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}`)

	fmt.Println("\n✨ Key Benefits:")
	fmt.Println("• c.Resp: Full control, no opinions, direct JSON")
	fmt.Println("• c.Api:  Consistent structure, validation, pagination support")
	fmt.Println("• Both:   Can be mixed in same application as needed")
}
