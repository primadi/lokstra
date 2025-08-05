package main

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates various middleware usage patterns in Lokstra:
// 1. Global middleware applied to all routes
// 2. Route-specific middleware
// 3. Group-level middleware
// 4. Multiple middleware chaining
// 5. Middleware with configuration
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "middleware-app", ":8080")

	// Register middleware functions
	registerMiddlewares(ctx)

	// Apply global middleware to all routes
	app.Use("logging")
	app.Use("request_id")

	// Simple route without additional middleware
	app.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong - global middleware applied")
	})

	// Route with specific middleware
	app.GET("/protected", func(ctx *lokstra.Context) error {
		return ctx.Ok("This is a protected route")
	}, "auth", "rate_limit")

	// Route with multiple middleware
	app.POST("/admin/action", func(ctx *lokstra.Context) error {
		return ctx.Ok("Admin action executed")
	}, "auth", "admin_check", "audit_log")

	// Group with middleware
	apiGroup := app.Group("/api/v1", "cors", "json_middleware")

	apiGroup.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"users": []string{"user1", "user2", "user3"},
		})
	})

	apiGroup.POST("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "User created successfully",
		})
	})

	// Nested group with additional middleware
	adminGroup := apiGroup.Group("/admin", "admin_check")

	adminGroup.GET("/stats", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"stats": map[string]int{
				"total_users":   100,
				"active_users":  85,
				"pending_users": 15,
			},
		})
	})

	// Route with middleware override (bypasses global middleware)
	app.HandleOverrideMiddleware("GET", "/public", func(ctx *lokstra.Context) error {
		return ctx.Ok("Public route - no middleware applied")
	}, "public_only")

	lokstra.Logger.Infof("Middleware example server started on :8080")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  GET  /ping              - Simple route with global middleware")
	lokstra.Logger.Infof("  GET  /protected         - Route with auth and rate limiting")
	lokstra.Logger.Infof("  POST /admin/action      - Route with multiple middleware")
	lokstra.Logger.Infof("  GET  /api/v1/users      - Group route with CORS and JSON middleware")
	lokstra.Logger.Infof("  POST /api/v1/users      - Group route for creating users")
	lokstra.Logger.Infof("  GET  /api/v1/admin/stats - Admin route with nested group middleware")
	lokstra.Logger.Infof("  GET  /public            - Public route with middleware override")

	app.Start()
}

// Context keys for storing values
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	requestIDKey contextKey = "request_id"
)

// registerMiddlewares registers all middleware functions used in this example
func registerMiddlewares(ctx lokstra.RegistrationContext) {
	// Logging middleware - logs request details
	ctx.RegisterMiddlewareFunc("logging", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			start := time.Now()
			fmt.Printf("[LOG] %s %s - Started\n", ctx.Request.Method, ctx.Request.URL.Path)

			err := next(ctx)

			duration := time.Since(start)
			fmt.Printf("[LOG] %s %s - Completed in %v\n", ctx.Request.Method, ctx.Request.URL.Path, duration)
			return err
		}
	})

	// Request ID middleware - adds unique request ID
	ctx.RegisterMiddlewareFunc("request_id", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
			// Store in context
			ctx.Context = context.WithValue(ctx.Context, requestIDKey, requestID)
			ctx.WithHeader("X-Request-ID", requestID)
			fmt.Printf("[REQUEST-ID] %s assigned to %s %s\n", requestID, ctx.Request.Method, ctx.Request.URL.Path)
			return next(ctx)
		}
	})

	// Auth middleware - simulates authentication
	ctx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			authHeader := ctx.Request.Header.Get("Authorization")
			if authHeader == "" {
				fmt.Println("[AUTH] No authorization header found")
				return ctx.ErrorBadRequest("Authorization header required")
			}

			// Simulate token validation
			if authHeader != "Bearer valid-token" {
				fmt.Println("[AUTH] Invalid token provided")
				return ctx.ErrorBadRequest("Invalid token")
			}

			fmt.Println("[AUTH] Authentication successful")
			// Store user ID in context
			ctx.Context = context.WithValue(ctx.Context, userIDKey, "user123")
			return next(ctx)
		}
	})

	// Rate limiting middleware
	ctx.RegisterMiddlewareFunc("rate_limit", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			// Simulate rate limiting check
			fmt.Println("[RATE-LIMIT] Rate limit check passed")
			ctx.WithHeader("X-RateLimit-Remaining", "99")
			return next(ctx)
		}
	})

	// Admin check middleware
	ctx.RegisterMiddlewareFunc("admin_check", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			userID := ctx.Value(userIDKey)
			if userID == nil {
				fmt.Println("[ADMIN] No user ID found")
				return ctx.ErrorBadRequest("Authentication required")
			}

			// Simulate admin role check
			if userID.(string) != "user123" { // In real app, check admin role
				fmt.Println("[ADMIN] User is not admin")
				return ctx.ErrorBadRequest("Admin access required")
			}

			fmt.Println("[ADMIN] Admin access granted")
			return next(ctx)
		}
	})

	// Audit log middleware
	ctx.RegisterMiddlewareFunc("audit_log", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			userID := ctx.Value(userIDKey)
			fmt.Printf("[AUDIT] User %v performed %s %s\n", userID, ctx.Request.Method, ctx.Request.URL.Path)
			return next(ctx)
		}
	})

	// CORS middleware
	ctx.RegisterMiddlewareFunc("cors", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			ctx.WithHeader("Access-Control-Allow-Origin", "*")
			ctx.WithHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.WithHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight requests
			if ctx.Request.Method == "OPTIONS" {
				ctx.StatusCode = 204
				return nil
			}

			fmt.Println("[CORS] CORS headers applied")
			return next(ctx)
		}
	})

	// JSON middleware - sets content type for API responses
	ctx.RegisterMiddlewareFunc("json_middleware", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			ctx.WithHeader("Content-Type", "application/json")
			fmt.Println("[JSON] JSON content type set")
			return next(ctx)
		}
	})

	// Public only middleware - for routes that bypass global middleware
	ctx.RegisterMiddlewareFunc("public_only", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[PUBLIC] Public route accessed")
			return next(ctx)
		}
	})
}
