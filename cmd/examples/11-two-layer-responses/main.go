package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
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

// Layer 1: Base Response (Unopinionated - Full Control)
func GetUsersBase(c *request.Context) error {
	return c.Resp.Json(users)
}

func CreateUserBase(c *request.Context) error {
	var user User
	if err := c.Req.BindBody(&user); err != nil {
		errorResponse := map[string]any{
			"error":   "Invalid JSON",
			"details": err.Error(),
			"code":    "PARSE_ERROR",
		}
		return c.Resp.WithStatus(400).Json(errorResponse)
	}

	user.ID = len(users) + 1
	users = append(users, user)

	return c.Resp.WithStatus(201).Json(map[string]any{
		"user":    user,
		"message": "User created",
	})
}

func GetHealthBase(c *request.Context) error {
	return c.Resp.Text("OK")
}

// Layer 2: Configurable API Response (Registry-based formatters)
func GetUsersApi(c *request.Context) error {
	return c.Api.Ok(users)
}

func CreateUserApi(c *request.Context) error {
	var user User
	if err := c.Req.BindBody(&user); err != nil {
		fields := []api_formatter.FieldError{
			{Field: "body", Code: "INVALID_JSON", Message: err.Error()},
		}
		return c.Api.ValidationError("Invalid request body", fields)
	}

	user.ID = len(users) + 1
	users = append(users, user)

	return c.Api.Created(user, "User created successfully")
}

func GetUserNotFoundApi(c *request.Context) error {
	return c.Api.NotFound("User not found")
}

func GetHealthApi(c *request.Context) error {
	health := map[string]string{"status": "healthy", "version": "1.0"}
	return c.Api.Ok(health)
}

// Paginated list example
func GetUsersPaginatedApi(c *request.Context) error {
	meta := api_formatter.CalculateListMeta(1, 10, len(users))
	return c.Api.OkList(users, meta)
}

func main() {
	r := router.New("two-layer-example")

	// Layer 1: Base Response (full control, any content type)
	r.GET("/base/users", GetUsersBase)
	r.POST("/base/users", CreateUserBase)
	r.GET("/base/health", GetHealthBase)

	// Layer 2: API Response (configurable formatters)
	r.GET("/api/users", GetUsersApi)
	r.POST("/api/users", CreateUserApi)
	r.GET("/api/users/404", GetUserNotFoundApi)
	r.GET("/api/health", GetHealthApi)
	r.GET("/api/users/paginated", GetUsersPaginatedApi)

	// Demonstrate different response formats
	r.GET("/api/simple", func(c *request.Context) error {
		// Switch to simple formatter
		response.SetApiResponseFormatterByName("simple")
		return c.Api.Ok(users)
	})

	r.GET("/api/legacy", func(c *request.Context) error {
		// Switch to legacy formatter
		response.SetApiResponseFormatterByName("legacy")
		return c.Api.Ok(users)
	})

	r.GET("/api/structured", func(c *request.Context) error {
		// Switch back to structured formatter
		response.SetApiResponseFormatterByName("api")
		return c.Api.Ok(users)
	})

	fmt.Println("🎯 Two-Layer Response Pattern Example")
	fmt.Println()
	fmt.Println("📍 Layer 1: Base Response (c.Resp - Unopinionated)")
	fmt.Println("GET  /base/users   - Direct JSON array")
	fmt.Println("POST /base/users   - Custom success structure")
	fmt.Println("GET  /base/health  - Plain text response")
	fmt.Println()
	fmt.Println("📍 Layer 2: API Response (c.Api - Configurable formatters)")
	fmt.Println("GET  /api/users           - Using default 'api' formatter")
	fmt.Println("POST /api/users           - Structured creation response")
	fmt.Println("GET  /api/users/404       - Structured error response")
	fmt.Println("GET  /api/health          - Structured health check")
	fmt.Println("GET  /api/users/paginated - Paginated list response")
	fmt.Println()
	fmt.Println("🔧 Dynamic Formatter Switching:")
	fmt.Println("GET  /api/simple     - Switch to 'simple' formatter")
	fmt.Println("GET  /api/legacy     - Switch to 'legacy' formatter")
	fmt.Println("GET  /api/structured - Switch to 'api' formatter")

	printResponseComparison()
}

func printResponseComparison() {
	fmt.Println("\n📄 RESPONSE OUTPUT COMPARISON:")

	fmt.Println("\n🔹 Layer 1 - Base Response (GET /base/users):")
	fmt.Println(`[
  {"id": 1, "name": "John Doe", "email": "john@example.com"},
  {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
]`)

	fmt.Println("\n🔹 Layer 2 - API Response 'api' formatter (GET /api/users):")
	fmt.Println(`{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
  ]
}`)

	fmt.Println("\n🔹 Layer 2 - API Response 'simple' formatter (GET /api/simple):")
	fmt.Println(`[
  {"id": 1, "name": "John Doe", "email": "john@example.com"},
  {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
]`)

	fmt.Println("\n🔹 Layer 2 - API Response 'legacy' formatter (GET /api/legacy):")
	fmt.Println(`{
  "success": true,
  "result": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
  ]
}`)

	fmt.Println("\n❌ ERROR HANDLING COMPARISON:")

	fmt.Println("\n🔹 Base Response (Custom Format):")
	fmt.Println(`{
  "error": "Invalid JSON",
  "details": "unexpected token...",
  "code": "PARSE_ERROR"
}`)

	fmt.Println("\n🔹 API Response 'api' formatter (Structured Format):")
	fmt.Println(`{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}`)

	fmt.Println("\n🔹 API Response 'simple' formatter:")
	fmt.Println(`{
  "error": "User not found",
  "code": "NOT_FOUND"
}`)

	fmt.Println("\n🔹 API Response 'legacy' formatter:")
	fmt.Println(`{
  "success": false,
  "errorCode": "NOT_FOUND",
  "errorMsg": "User not found"
}`)

	fmt.Println("\n✨ TWO-LAYER BENEFITS:")
	fmt.Println("• Layer 1: Maximum flexibility, any content type, full control")
	fmt.Println("• Layer 2: Consistent structure, configurable formats via registry")
	fmt.Println("• Registry Pattern: Switch formatters at startup or runtime")
	fmt.Println("• Built-in Formatters: 'api', 'simple', 'legacy'")
	fmt.Println("• Custom Formatters: Register your own response formats")
	fmt.Println()
	fmt.Println("🎚️ CHOOSE YOUR APPROACH:")
	fmt.Println("• Need full control? → Use Layer 1 (c.Resp.Json/Text/Raw)")
	fmt.Println("• Building API? → Use Layer 2 (c.Api.Ok/Created/ValidationError)")
	fmt.Println("• Legacy system? → Register custom formatter for compatibility")
}
