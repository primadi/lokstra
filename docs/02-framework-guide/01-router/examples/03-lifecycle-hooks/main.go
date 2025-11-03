package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// Middleware untuk Before Hook
func beforeMiddleware(ctx *request.Context) error {
	fmt.Printf("[BEFORE] %s %s - Time: %s\n",
		ctx.R.Method,
		ctx.R.URL.Path,
		time.Now().Format("15:04:05"),
	)
	// Simpan start time untuk menghitung duration
	ctx.Set("start_time", time.Now())
	return ctx.Next()
}

// Middleware untuk After Hook
func afterMiddleware(ctx *request.Context) error {
	err := ctx.Next() // Call handler first

	// After handler executes
	if startTime, ok := ctx.Get("start_time").(time.Time); ok {
		duration := time.Since(startTime)
		fmt.Printf("[AFTER]  %s %s - Duration: %v\n",
			ctx.R.Method,
			ctx.R.URL.Path,
			duration,
		)
	}

	return err
}

// Logging middleware
func loggingMiddleware(ctx *request.Context) error {
	startTime := time.Now()

	// Before handler
	fmt.Printf("[LOG] → Request: %s %s\n", ctx.R.Method, ctx.R.URL.Path)

	// Execute handler
	err := ctx.Next()

	// After handler
	duration := time.Since(startTime)
	status := "success"
	if err != nil {
		status = "error"
	}
	fmt.Printf("[LOG] ← Response: %s (took %v)\n", status, duration)

	return err
}

// Auth middleware
func authMiddleware(ctx *request.Context) error {
	apiKey := ctx.Req.HeaderParam("X-API-Key", "")

	if apiKey == "" {
		return ctx.Api.Unauthorized("API key required")
	}

	if apiKey != "secret-key-123" {
		return ctx.Api.Forbidden("Invalid API key")
	}

	fmt.Println("[AUTH] ✓ API key validated")
	ctx.Set("api_key", apiKey)

	return ctx.Next()
}

// Request ID middleware
func requestIDMiddleware(ctx *request.Context) error {
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	ctx.Set("request_id", requestID)
	fmt.Printf("[REQUEST_ID] %s\n", requestID)
	return ctx.Next()
}

// Simple handlers
func handler1(ctx *request.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"message":    "Handler 1 - No middleware",
		"request_id": ctx.Get("request_id"),
	}, nil
}

func handler2(ctx *request.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"message":    "Handler 2 - With before/after middleware",
		"request_id": ctx.Get("request_id"),
	}, nil
}

func handler3(ctx *request.Context) (map[string]interface{}, error) {
	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	return map[string]interface{}{
		"message":    "Handler 3 - With logging middleware",
		"request_id": ctx.Get("request_id"),
	}, nil
}

func handler4(ctx *request.Context) (map[string]interface{}, error) {
	apiKey := ctx.Get("api_key")

	return map[string]interface{}{
		"message":    "Handler 4 - Protected with auth",
		"api_key":    apiKey,
		"request_id": ctx.Get("request_id"),
	}, nil
}

func handler5(ctx *request.Context) (map[string]interface{}, error) {
	// Simulate slow operation
	time.Sleep(100 * time.Millisecond)

	return map[string]interface{}{
		"message":    "Handler 5 - Multiple middleware chain",
		"api_key":    ctx.Get("api_key"),
		"request_id": ctx.Get("request_id"),
	}, nil
}

func main() {
	router := lokstra.NewRouter("lifecycle")

	// Global middleware - applies to all routes
	router.Use(requestIDMiddleware)

	// Route 1: No additional middleware
	router.GET("/h1", handler1)

	// Route 2: Before + After hooks
	router.GET("/h2", handler2, beforeMiddleware, afterMiddleware)

	// Route 3: With logging middleware
	router.GET("/h3", handler3, loggingMiddleware)

	// Route 4: Protected with auth
	router.GET("/h4", handler4, authMiddleware)

	// Route 5: Multiple middleware (auth + logging + before/after)
	router.GET("/h5", handler5,
		authMiddleware,
		loggingMiddleware,
		beforeMiddleware,
		afterMiddleware,
	)

	// Group with shared middleware
	apiGroup := router.AddGroup("/api")
	apiGroup.Use(authMiddleware, loggingMiddleware)

	apiGroup.GET("/users", func(ctx *request.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"message": "Users API - auto-protected by group middleware",
			"users":   []string{"Alice", "Bob", "Charlie"},
		}, nil
	})

	apiGroup.GET("/posts", func(ctx *request.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"message": "Posts API - auto-protected by group middleware",
			"posts":   []string{"Post 1", "Post 2", "Post 3"},
		}, nil
	})

	app := lokstra.NewApp("lifecycle-demo", ":3000", router)

	fmt.Println("🚀 Lifecycle Hooks & Middleware Demo")
	fmt.Println("📖 Server: http://localhost:3000")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET /h1              - No middleware")
	fmt.Println("  GET /h2              - Before + After hooks")
	fmt.Println("  GET /h3              - Logging middleware")
	fmt.Println("  GET /h4              - Auth middleware")
	fmt.Println("  GET /h5              - Multiple middleware")
	fmt.Println("  GET /api/users       - Group middleware (auth + log)")
	fmt.Println("  GET /api/posts       - Group middleware (auth + log)")
	fmt.Println("\nWatch console output to see middleware execution order!")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
