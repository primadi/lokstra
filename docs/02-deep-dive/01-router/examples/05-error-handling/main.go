package main

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

// Custom error types
var (
	ErrNotFound          = errors.New("resource not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrDuplicateEntry    = errors.New("duplicate entry")
)

// Error codes
const (
	CodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	CodeDuplicateEmail    = "DUPLICATE_EMAIL"
	CodeOutOfStock        = "OUT_OF_STOCK"
	CodeQuotaExceeded     = "QUOTA_EXCEEDED"
)

func main() {
	router := lokstra.NewRouter("error-handling")

	// Add error recovery middleware
	router.Use(ErrorRecoveryMiddleware)

	// ============================================
	// Standard HTTP Errors
	// ============================================
	router.GET("/errors/400", Error400)
	router.GET("/errors/401", Error401)
	router.GET("/errors/403", Error403)
	router.GET("/errors/404", Error404)
	router.GET("/errors/422", Error422)
	router.GET("/errors/500", Error500)

	// ============================================
	// Custom Error Codes
	// ============================================
	router.GET("/errors/custom/insufficient-funds", ErrorInsufficientFunds)
	router.GET("/errors/custom/duplicate", ErrorDuplicate)
	router.GET("/errors/custom/quota", ErrorQuotaExceeded)

	// ============================================
	// Error Patterns
	// ============================================
	router.GET("/patterns/early-return", PatternEarlyReturn)
	router.GET("/patterns/error-wrapping", PatternErrorWrapping)
	router.GET("/patterns/validation", PatternValidation)
	router.POST("/patterns/validation", PatternValidationPost)

	// ============================================
	// Simulated Errors
	// ============================================
	router.GET("/simulate/db-error", SimulateDBError)
	router.GET("/simulate/external-api", SimulateExternalAPIError)
	router.GET("/simulate/panic", SimulatePanic)
	router.GET("/simulate/not-found", SimulateNotFound)

	// ============================================
	// Resource Operations (with errors)
	// ============================================
	router.GET("/users/:id", GetUser)
	router.POST("/users", CreateUser)
	router.DELETE("/users/:id", DeleteUser)
	router.POST("/transfer", Transfer)
	router.POST("/orders", CreateOrder)

	// Home
	router.GET("/", Home)

	app := lokstra.NewApp("error-handling", ":3000", router)

	fmt.Println("🚀 Error Handling Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Test error scenarios:")
	fmt.Println("   GET /errors/404         (Not Found)")
	fmt.Println("   GET /errors/500         (Internal Error)")
	fmt.Println("   GET /simulate/panic     (Panic Recovery)")
	fmt.Println("\n🧪 Open test.http for all examples")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// ============================================
// Middleware
// ============================================

func ErrorRecoveryMiddleware(c *request.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 PANIC RECOVERED: %v\n%s", r, debug.Stack())

			// Return structured error response
			api := response.NewApiHelper()
			api.Error(500, "PANIC_RECOVERED", fmt.Sprintf("Panic: %v", r))
		}
	}()

	return c.Next()
}

// ============================================
// Standard HTTP Errors
// ============================================

func Error400() *response.ApiHelper {
	return response.NewApiBadRequest("INVALID_INPUT", "Bad request example")
}

func Error401() *response.ApiHelper {
	return response.NewApiUnauthorized("Authentication required")
}

func Error403() *response.ApiHelper {
	return response.NewApiForbidden("Admin access required")
}

func Error404() *response.ApiHelper {
	return response.NewApiNotFound("Resource not found")
}

func Error422() *response.ApiHelper {
	fields := []api_formatter.FieldError{
		{
			Field:   "email",
			Code:    "INVALID_FORMAT",
			Message: "Email format is invalid",
		},
		{
			Field:   "age",
			Code:    "MIN_VALUE",
			Message: "Age must be at least 18",
		},
	}
	return response.NewApiValidationError("Validation failed", fields)
}

func Error500() *response.ApiHelper {
	return response.NewApiInternalError("Internal server error example")
}

// ============================================
// Custom Error Codes
// ============================================

func ErrorInsufficientFunds() *response.ApiHelper {
	return response.NewApiBadRequest(
		CodeInsufficientFunds,
		"Account balance is insufficient for this transaction",
	)
}

func ErrorDuplicate() *response.ApiHelper {
	return response.NewApiBadRequest(
		CodeDuplicateEmail,
		"Email address already exists",
	)
}

func ErrorQuotaExceeded() *response.ApiHelper {
	return response.NewApiError(
		429,
		CodeQuotaExceeded,
		"API quota exceeded. Please try again later.",
	)
}

// ============================================
// Error Patterns
// ============================================

func PatternEarlyReturn() *response.ApiHelper {
	// Simulate business logic with multiple validation points
	stock := 0 // out of stock

	// Early return for out of stock
	if stock == 0 {
		return response.NewApiBadRequest(CodeOutOfStock, "Product out of stock")
	}

	// Early return for insufficient funds
	balance := 50
	price := 100
	if balance < price {
		return response.NewApiBadRequest(CodeInsufficientFunds, "Insufficient funds")
	}

	// Success path
	return response.NewApiOk(map[string]any{
		"message": "Order created",
		"pattern": "early-return",
	})
}

func PatternErrorWrapping() *response.ApiHelper {
	// Simulate multiple operations
	userID := "123"

	// First operation
	user, err := simulateGetUser(userID)
	if err != nil {
		log.Printf("PatternErrorWrapping: failed to get user %s: %v", userID, err)
		return response.NewApiInternalError("Failed to fetch user")
	}

	// Second operation (non-critical)
	posts, err := simulateGetPosts(userID)
	if err != nil {
		log.Printf("PatternErrorWrapping: failed to get posts for user %s: %v", userID, err)
		// Continue with empty posts
		posts = []string{}
	}

	return response.NewApiOk(map[string]any{
		"user":    user,
		"posts":   posts,
		"pattern": "error-wrapping",
	})
}

func PatternValidation() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "Send POST request with JSON body to test validation",
		"example": map[string]any{
			"email": "test@example.com",
			"age":   25,
			"name":  "John Doe",
		},
	})
}

type ValidationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,gte=18,lte=100"`
	Name  string `json:"name" validate:"required,min=3"`
}

func PatternValidationPost(req ValidationRequest) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message":    "Validation passed",
		"email":      req.Email,
		"age":        req.Age,
		"name":       req.Name,
		"validation": "automatic",
	})
}

// ============================================
// Simulated Errors
// ============================================

func SimulateDBError() *response.ApiHelper {
	// Simulate database error
	err := errors.New("connection timeout")
	log.Printf("Database error: %v", err)

	return response.NewApiInternalError("Database connection failed")
}

func SimulateExternalAPIError() *response.ApiHelper {
	// Simulate external API failure
	err := errors.New("external API returned 503")
	log.Printf("External API error: %v", err)

	return response.NewApiError(503, "SERVICE_UNAVAILABLE", "External service temporarily unavailable")
}

func SimulatePanic() *response.ApiHelper {
	// This will trigger panic recovery middleware
	panic("intentional panic for testing")
}

func SimulateNotFound() *response.ApiHelper {
	return response.NewApiNotFound("The requested resource was not found")
}

// ============================================
// Resource Operations
// ============================================

type UserIDParam struct {
	ID string `path:"id"`
}

func GetUser(params UserIDParam) *response.ApiHelper {
	// Simulate database lookup
	if params.ID == "999" {
		return response.NewApiNotFound("User not found")
	}

	if params.ID == "error" {
		log.Printf("Database error fetching user %s", params.ID)
		return response.NewApiInternalError("Failed to fetch user")
	}

	return response.NewApiOk(map[string]any{
		"id":       params.ID,
		"username": fmt.Sprintf("user_%s", params.ID),
		"email":    fmt.Sprintf("user%s@example.com", params.ID),
	})
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3"`
	Age      int    `json:"age" validate:"required,gte=18"`
}

func CreateUser(req CreateUserRequest) *response.ApiHelper {
	// Check for duplicate
	if req.Email == "duplicate@example.com" {
		return response.NewApiBadRequest(CodeDuplicateEmail, "Email already exists")
	}

	user := map[string]any{
		"id":       "new-123",
		"email":    req.Email,
		"username": req.Username,
		"age":      req.Age,
	}

	return response.NewApiCreated(user, "User created successfully")
}

func DeleteUser(params UserIDParam) *response.ApiHelper {
	if params.ID == "admin" {
		return response.NewApiForbidden("Cannot delete admin user")
	}

	if params.ID == "999" {
		return response.NewApiNotFound("User not found")
	}

	return response.NewApiOk(map[string]any{
		"message": "User deleted successfully",
		"id":      params.ID,
	})
}

type TransferRequest struct {
	FromAccount string  `json:"from_account" validate:"required"`
	ToAccount   string  `json:"to_account" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
}

func Transfer(req TransferRequest) *response.ApiHelper {
	// Simulate balance check
	balance := 100.0
	if req.Amount > balance {
		return response.NewApiBadRequest(
			CodeInsufficientFunds,
			fmt.Sprintf("Insufficient funds. Balance: %.2f, Requested: %.2f", balance, req.Amount),
		)
	}

	return response.NewApiOk(map[string]any{
		"message":      "Transfer successful",
		"from_account": req.FromAccount,
		"to_account":   req.ToAccount,
		"amount":       req.Amount,
		"new_balance":  balance - req.Amount,
	})
}

type CreateOrderRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

func CreateOrder(req CreateOrderRequest) *response.ApiHelper {
	// Check stock
	stock := 5
	if req.Quantity > stock {
		return response.NewApiBadRequest(
			CodeOutOfStock,
			fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", stock, req.Quantity),
		)
	}

	return response.NewApiCreated(map[string]any{
		"order_id":   "order-123",
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"status":     "pending",
	}, "Order created successfully")
}

// ============================================
// Helpers
// ============================================

func simulateGetUser(id string) (map[string]any, error) {
	if id == "error" {
		return nil, errors.New("database error")
	}
	return map[string]any{
		"id":   id,
		"name": "John Doe",
	}, nil
}

func simulateGetPosts(userID string) ([]string, error) {
	if userID == "123" {
		return nil, errors.New("posts service unavailable")
	}
	return []string{"post1", "post2"}, nil
}

// ============================================
// Home
// ============================================

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Error Handling Example</title>
    <style>
        body { font-family: Arial; max-width: 1200px; margin: 40px auto; padding: 20px; }
        h1 { color: #333; }
        .section { margin: 30px 0; padding: 20px; background: #f5f5f5; border-radius: 8px; }
        .error { margin: 10px 0; padding: 10px; background: white; border-left: 4px solid #dc3545; }
        .error-400 { border-left-color: #ffc107; }
        .error-500 { border-left-color: #dc3545; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>🚨 Error Handling Example</h1>
    <p>Comprehensive error handling patterns in Lokstra</p>

    <div class="section">
        <h2>Standard HTTP Errors</h2>
        <div class="error error-400">
            <a href="/errors/400">400 Bad Request</a> - Invalid input
        </div>
        <div class="error error-400">
            <a href="/errors/401">401 Unauthorized</a> - Authentication required
        </div>
        <div class="error error-400">
            <a href="/errors/403">403 Forbidden</a> - Permission denied
        </div>
        <div class="error error-400">
            <a href="/errors/404">404 Not Found</a> - Resource not found
        </div>
        <div class="error error-400">
            <a href="/errors/422">422 Validation Error</a> - Field validation failed
        </div>
        <div class="error error-500">
            <a href="/errors/500">500 Internal Error</a> - Server error
        </div>
    </div>

    <div class="section">
        <h2>Custom Error Codes</h2>
        <div class="error">
            <a href="/errors/custom/insufficient-funds">INSUFFICIENT_FUNDS</a> - Not enough balance
        </div>
        <div class="error">
            <a href="/errors/custom/duplicate">DUPLICATE_EMAIL</a> - Email already exists
        </div>
        <div class="error">
            <a href="/errors/custom/quota">QUOTA_EXCEEDED</a> - API quota exceeded (429)
        </div>
    </div>

    <div class="section">
        <h2>Error Patterns</h2>
        <div class="error">
            <a href="/patterns/early-return">Early Return</a> - Multiple validation points
        </div>
        <div class="error">
            <a href="/patterns/error-wrapping">Error Wrapping</a> - Logging and handling
        </div>
        <div class="error">
            <a href="/patterns/validation">Validation Pattern</a> - POST to test
        </div>
    </div>

    <div class="section">
        <h2>Simulated Errors</h2>
        <div class="error error-500">
            <a href="/simulate/db-error">Database Error</a> - Connection failure
        </div>
        <div class="error error-500">
            <a href="/simulate/external-api">External API Error</a> - Service unavailable (503)
        </div>
        <div class="error error-500">
            <a href="/simulate/panic">Panic Recovery</a> - Middleware catches panic
        </div>
        <div class="error error-400">
            <a href="/simulate/not-found">Not Found</a> - 404 error
        </div>
    </div>

    <div class="section">
        <h2>Resource Operations</h2>
        <div class="error">
            <a href="/users/123">GET /users/123</a> - Fetch user (success)
        </div>
        <div class="error">
            <a href="/users/999">GET /users/999</a> - Not found (404)
        </div>
        <div class="error">
            <a href="/users/error">GET /users/error</a> - Database error (500)
        </div>
        <p>Use <code>test.http</code> to test POST, DELETE operations</p>
    </div>

    <div class="section">
        <h2>📖 Documentation</h2>
        <p>See <code>index</code> for detailed patterns</p>
        <p>Use <code>test.http</code> for all test cases</p>
    </div>
</body>
</html>`

	return response.NewHtmlResponse(html)
}
