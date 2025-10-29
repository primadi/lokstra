package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
)

func main() {
	router := lokstra.NewRouter("testing-example")

	// ============================================
	// Example Handlers to Test
	// ============================================
	router.GET("/users/:id", GetUser)
	router.POST("/users", CreateUser)
	router.GET("/health", HealthCheck)
	router.GET("/", Home)

	app := lokstra.NewApp("testing-example", ":3000", router)

	fmt.Println("🚀 Testing Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Endpoints:")
	fmt.Println("   GET  /users/:id   - Get user")
	fmt.Println("   POST /users       - Create user")
	fmt.Println("   GET  /health      - Health check")
	fmt.Println("\n🧪 Run tests: go test -v")
	fmt.Println("📊 Coverage: go test -cover")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// ============================================
// Handlers
// ============================================

type UserIDParam struct {
	ID string `path:"id"`
}

func GetUser(params UserIDParam) *response.ApiHelper {
	// Simulate database lookup
	if params.ID == "999" {
		return response.NewApiNotFound("User not found")
	}

	if params.ID == "" {
		return response.NewApiBadRequest("INVALID_ID", "User ID is required")
	}

	return response.NewApiOk(map[string]any{
		"id":    params.ID,
		"name":  fmt.Sprintf("User %s", params.ID),
		"email": fmt.Sprintf("user%s@example.com", params.ID),
	})
}

type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=3"`
	Age   int    `json:"age" validate:"required,gte=18,lte=100"`
}

func CreateUser(req CreateUserRequest) *response.ApiHelper {
	// Check for duplicate (example)
	if req.Email == "duplicate@example.com" {
		return response.NewApiBadRequest("DUPLICATE_EMAIL", "Email already exists")
	}

	user := map[string]any{
		"id":    "new-123",
		"email": req.Email,
		"name":  req.Name,
		"age":   req.Age,
	}

	return response.NewApiCreated(user, "User created successfully")
}

func HealthCheck() map[string]any {
	return map[string]any{
		"status": "healthy",
		"uptime": "100%",
	}
}

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Testing Example</title>
    <style>
        body { font-family: Arial; max-width: 1000px; margin: 40px auto; padding: 20px; }
        h1 { color: #333; }
        .section { margin: 30px 0; padding: 20px; background: #f5f5f5; border-radius: 8px; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        .command { background: #2d2d2d; color: #f8f8f2; padding: 15px; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>🧪 Testing Example</h1>
    <p>Demonstrates testing strategies for Lokstra handlers</p>

    <div class="section">
        <h2>Endpoints</h2>
        <ul>
            <li><a href="/users/123">GET /users/123</a> - Get user</li>
            <li>POST /users - Create user (use test.http)</li>
            <li><a href="/health">GET /health</a> - Health check</li>
        </ul>
    </div>

    <div class="section">
        <h2>Run Tests</h2>
        <div class="command">$ go test -v</div>
        <div class="command">$ go test -cover</div>
        <div class="command">$ go test -coverprofile=coverage.out</div>
        <div class="command">$ go tool cover -html=coverage.out</div>
    </div>

    <div class="section">
        <h2>Test Files</h2>
        <p>See <code>main_test.go</code> for test examples</p>
        <p>See <code>index.md</code> for testing patterns</p>
    </div>
</body>
</html>`

	return response.NewHtmlResponse(html)
}
