package main

import (
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

// CustomAuthMiddleware - Example of a custom middleware
func CustomAuthMiddleware(ctx *request.Context) error {
	// Simple API key check
	apiKey := ctx.R.Header.Get("X-API-Key")
	if apiKey == "" {
		return ctx.Api.Unauthorized("Missing API key")
	}

	// Accept both regular and admin keys
	validKeys := []string{"secret-key-123", "admin-key-456"}
	isValid := slices.Contains(validKeys, apiKey)

	if !isValid {
		return ctx.Api.Forbidden("Invalid API key")
	}

	// Store user info in context for later use
	ctx.Set("authenticated", true)
	ctx.Set("api_key", apiKey)

	return ctx.Next() // Continue to next middleware/handler
}

// RateLimitMiddleware - Example of custom middleware with state
func RateLimitMiddleware(maxRequests int, window time.Duration) request.HandlerFunc {
	requests := make(map[string][]time.Time)

	return func(ctx *request.Context) error {
		ip, _, err := net.SplitHostPort(ctx.R.RemoteAddr)
		if err != nil {
			return ctx.Api.Error(400, "BAD_REQUEST", "Invalid IP address")
		}

		// Clean old requests
		now := time.Now()
		cutoff := now.Add(-window)
		if times, ok := requests[ip]; ok {
			filtered := []time.Time{}
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			requests[ip] = filtered
		}

		// Check rate limit
		if len(requests[ip]) >= maxRequests {
			return ctx.Api.Error(429, "RATE_LIMIT_EXCEEDED",
				fmt.Sprintf("Rate limit exceeded. Max %d requests per %v", maxRequests, window))
		}

		// Add current request
		requests[ip] = append(requests[ip], now)
		return ctx.Next()
	}
}

// LoggingMiddleware - Custom request/response logger
func LoggingMiddleware(ctx *request.Context) error {
	start := time.Now()
	method := ctx.R.Method
	path := ctx.R.URL.Path

	log.Printf("→ %s %s", method, path)

	// Continue to next handler
	err := ctx.Next()

	// Log after handler completes
	duration := time.Since(start)
	status := ctx.W.StatusCode()
	log.Printf("← %s %s - %d (%v)", method, path, status, duration)

	return err
}

// ===== Example Handlers =====

func PublicHandler() string {
	return "This is a public endpoint - no auth required"
}

func ProtectedHandler(ctx *request.Context) string {
	apiKey := ctx.Get("api_key").(string)
	return fmt.Sprintf("Welcome! You're authenticated with key: %s", apiKey)
}

func AdminHandler(ctx *request.Context) map[string]any {
	return map[string]any{
		"message": "Admin access granted",
		"user":    "admin",
		"time":    time.Now().Format(time.RFC3339),
	}
}

func PanicHandler() string {
	// This will panic but recovery middleware will catch it
	panic("Something went wrong!")
}

func SlowHandler() string {
	time.Sleep(2 * time.Second) // Simulate slow operation
	return "This took a while..."
}

// AdminOnlyMiddleware - Check if user is admin
func AdminOnlyMiddleware(ctx *request.Context) error {
	// In real app, check actual user role from DB/JWT
	apiKey := ctx.Get("api_key")
	if apiKey != "admin-key-456" {
		return ctx.Api.Forbidden("Admin access required")
	}
	return ctx.Next()
}

func main() {
	// Register built-in middleware factories
	cors.Register()
	recovery.Register()
	request_logger.Register()

	// Register middleware instances with different configs
	lokstra_registry.RegisterMiddlewareName("cors-all", cors.CORS_TYPE, map[string]any{
		"allow_origins": []string{"*"},
	})

	lokstra_registry.RegisterMiddlewareName("recovery-prod", recovery.RECOVERY_TYPE, map[string]any{
		"enable_stack_trace": true,
		"enable_logging":     true,
	})

	lokstra_registry.RegisterMiddlewareName("logger-color", request_logger.REQUEST_LOGGER_TYPE, map[string]any{
		"enable_colors": true,
		"skip_paths":    []string{"/health"},
	})

	// Create router
	r := lokstra.NewRouter("api")

	// ===== GLOBAL MIDDLEWARES (Applied to all routes) =====
	// r.Use("recovery-prod")                      // Catch panics
	r.Use("cors-all")                           // CORS for all origins
	r.Use("logger-color")                       // Request logging
	r.Use(LoggingMiddleware)                    // Custom logging
	r.Use(RateLimitMiddleware(10, time.Minute)) // Max 10 req/min

	// ===== PUBLIC ROUTES (No auth required) =====
	r.GET("/", func() string {
		return "Lokstra Middleware Example - Try different endpoints!"
	})

	r.GET("/public", PublicHandler)

	r.GET("/health", func() map[string]any {
		return map[string]any{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		}
	}, route.WithOverrideParentMwOption(true))

	// ===== PROTECTED ROUTES (Auth required) =====
	// Apply middleware to specific route
	r.GET("/protected", ProtectedHandler, CustomAuthMiddleware)

	// Multiple routes with same middleware
	r.GET("/api/profile", func(ctx *request.Context) map[string]any {
		return map[string]any{
			"user":          "john_doe",
			"api_key":       ctx.Get("api_key"),
			"authenticated": ctx.Get("authenticated"),
		}
	}, CustomAuthMiddleware)

	r.GET("/api/data", func() []string {
		return []string{"item1", "item2", "item3"}
	}, CustomAuthMiddleware)

	// ===== ADMIN ROUTES (Auth + Admin check) =====
	r.GET("/api/admin/dashboard", AdminHandler, CustomAuthMiddleware, AdminOnlyMiddleware)

	r.GET("/api/admin/users", func() []string {
		return []string{"user1", "user2", "admin"}
	}, CustomAuthMiddleware, AdminOnlyMiddleware)

	// ===== TEST ROUTES =====
	r.GET("/panic", PanicHandler) // Test recovery middleware
	r.GET("/slow", SlowHandler)   // Test request logger with slow request

	// ===== MIDDLEWARE CHAIN EXAMPLE =====
	// Multiple middlewares executed in order
	r.GET("/chain",
		func() string {
			log.Println("Handler executed")
			return "All middlewares passed!"
		},
		func(ctx *request.Context) error {
			log.Println("Middleware 1: Before handler")
			ctx.Set("middleware1", "executed")
			err := ctx.Next()
			log.Println("Middleware 1: After handler")
			return err
		},
		func(ctx *request.Context) error {
			log.Println("Middleware 2: Before handler")
			ctx.Set("middleware2", "executed")
			err := ctx.Next()
			log.Println("Middleware 2: After handler")
			return err
		},
		func(ctx *request.Context) error {
			log.Println("Middleware 3: Checking previous middlewares")
			if ctx.Get("middleware1") == nil || ctx.Get("middleware2") == nil {
				return ctx.Api.Error(500, "MIDDLEWARE_CHAIN_BROKEN", "Middleware chain broken")
			}
			err := ctx.Next()
			log.Println("Middleware 3: After handler")
			return err
		},
	)

	// Print available routes
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 Lokstra Middleware Example")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n📋 Available Endpoints:")
	fmt.Println("\n  PUBLIC (No auth):")
	fmt.Println("    GET  http://localhost:3000/")
	fmt.Println("    GET  http://localhost:3000/public")
	fmt.Println("    GET  http://localhost:3000/health")
	fmt.Println("\n  PROTECTED (Requires X-API-Key: secret-key-123):")
	fmt.Println("    GET  http://localhost:3000/protected")
	fmt.Println("    GET  http://localhost:3000/api/profile")
	fmt.Println("    GET  http://localhost:3000/api/data")
	fmt.Println("\n  ADMIN (Requires X-API-Key: admin-key-456):")
	fmt.Println("    GET  http://localhost:3000/api/admin/dashboard")
	fmt.Println("    GET  http://localhost:3000/api/admin/users")
	fmt.Println("\n  TEST:")
	fmt.Println("    GET  http://localhost:3000/panic  (test recovery)")
	fmt.Println("    GET  http://localhost:3000/slow   (test logger)")
	fmt.Println("    GET  http://localhost:3000/chain  (middleware chain)")
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\n💡 Tips:")
	fmt.Println("  - Try without X-API-Key header → 401 Unauthorized")
	fmt.Println("  - Try with wrong key → 403 Forbidden")
	fmt.Println("  - Try /panic → Recovery middleware catches it")
	fmt.Println("  - Make 11+ requests → Rate limit kicks in")
	fmt.Println("  - Use test.http file for easy testing")
	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Start server
	app := lokstra.NewApp("middleware-example", ":3000", r)
	app.PrintStartInfo()
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
