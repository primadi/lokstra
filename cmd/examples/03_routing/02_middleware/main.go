package main

import (
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates middleware patterns and usage in Lokstra.
// It shows how to create, register, and chain middleware for cross-cutting concerns.
//
// Learning Objectives:
// - Understand middleware execution order
// - Learn to create custom middleware
// - Explore middleware chaining and composition
// - See common middleware patterns (logging, auth, CORS, etc.)
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/routing.md#middleware

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "middleware-app", ":8080")

	// ===== Global Middleware (applies to all routes) =====

	// 1. Request ID middleware
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		requestID := "req-" + time.Now().Format("20060102-150405-000000")
		lokstra.Logger.Infof("🆔 [%s] %s %s", requestID, ctx.Request.Method, ctx.Request.URL.Path)

		// Add request ID to response headers
		ctx.Response.WithHeader("X-Request-ID", requestID)

		return next(ctx)
	})

	// 2. Request timing middleware
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		start := time.Now()

		err := next(ctx)

		duration := time.Since(start)
		lokstra.Logger.Infof("⏱️  Request completed in %v", duration)
		ctx.Response.WithHeader("X-Response-Time", duration.String())

		return err
	})

	// 3. CORS middleware
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Set CORS headers
		ctx.Response.WithHeader("Access-Control-Allow-Origin", "*")
		ctx.Response.WithHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.WithHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if ctx.Request.Method == "OPTIONS" {
			return ctx.Ok("CORS preflight")
		}

		return next(ctx)
	})

	// ===== Custom Middleware Functions =====

	// Authentication middleware
	authMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		token := ctx.GetHeader("Authorization")

		if token == "" {
			return ctx.ErrorBadRequest("Authorization header required")
		}

		// Simple token validation (in real app, verify JWT or API key)
		if token != "Bearer valid-token" {
			return ctx.ErrorBadRequest("Invalid authorization token")
		}

		lokstra.Logger.Infof("🔐 Authentication successful")
		return next(ctx)
	}

	// Role-based authorization middleware
	adminMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// In real app, extract user role from token
		userRole := ctx.GetHeader("X-User-Role")

		if userRole != "admin" {
			return ctx.ErrorBadRequest("Admin access required")
		}

		lokstra.Logger.Infof("👑 Admin access granted")
		return next(ctx)
	}

	// Rate limiting middleware (simple version)
	rateLimitMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Simple rate limiting by IP (in real app, use Redis or similar)
		clientIP := ctx.Request.RemoteAddr

		// Simulate rate limit check
		lokstra.Logger.Infof("🚦 Rate limit check for %s", clientIP)

		// Add rate limit headers
		ctx.Response.WithHeader("X-RateLimit-Limit", "100")
		ctx.Response.WithHeader("X-RateLimit-Remaining", "95")
		ctx.Response.WithHeader("X-RateLimit-Reset", "3600")

		return next(ctx)
	}

	// Content validation middleware
	contentValidationMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" {
			contentType := ctx.GetHeader("Content-Type")

			if contentType != "application/json" {
				return ctx.ErrorBadRequest("Content-Type must be application/json")
			}
		}

		return next(ctx)
	}

	// ===== Routes with Different Middleware Combinations =====

	// Public routes (only global middleware)
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":    "Public endpoint",
			"middleware": []string{"requestID", "timing", "cors"},
		})
	})

	app.GET("/public", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Another public endpoint",
			"note":    "Only global middleware applied",
		})
	})

	// Protected routes (auth required)
	protectedGroup := app.Group("/protected")
	protectedGroup.Use(authMiddleware)

	protectedGroup.GET("/profile", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":    "User profile",
			"middleware": []string{"global", "auth"},
			"note":       "Requires valid Authorization header",
		})
	})

	protectedGroup.GET("/dashboard", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":    "User dashboard",
			"middleware": []string{"global", "auth"},
		})
	})

	// Admin routes (auth + admin role required)
	adminGroup := app.Group("/admin")
	adminGroup.Use(authMiddleware)
	adminGroup.Use(adminMiddleware)

	adminGroup.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":    "Admin: User list",
			"middleware": []string{"global", "auth", "admin"},
			"users":      []string{"user1", "user2", "user3"},
		})
	})

	adminGroup.DELETE("/users/:id", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")
		return ctx.Ok(map[string]any{
			"message":    "Admin: User deleted",
			"user_id":    userID,
			"middleware": []string{"global", "auth", "admin"},
		})
	})

	// API routes with rate limiting and content validation
	apiGroup := app.Group("/api")
	apiGroup.Use(rateLimitMiddleware)
	apiGroup.Use(contentValidationMiddleware)

	apiGroup.GET("/status", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status":     "operational",
			"middleware": []string{"global", "rateLimit", "contentValidation"},
		})
	})

	type CreateItemRequest struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
	}

	apiGroup.POST("/items", func(ctx *lokstra.Context, req *CreateItemRequest) error {
		return ctx.OkCreated(map[string]any{
			"message":    "Item created",
			"item":       req,
			"middleware": []string{"global", "rateLimit", "contentValidation"},
		})
	})

	// ===== Route-Specific Middleware =====

	// Single route with specific middleware
	app.GET("/special",
		rateLimitMiddleware,
		func(ctx *lokstra.Context) error {
			return ctx.Ok(map[string]any{
				"message":    "Special endpoint with route-specific middleware",
				"middleware": []string{"global", "routeSpecificRateLimit"},
			})
		})

	// Multiple route-specific middleware
	app.POST("/upload",
		authMiddleware,
		contentValidationMiddleware,
		func(ctx *lokstra.Context) error {
			return ctx.Ok(map[string]any{
				"message":    "File upload endpoint",
				"middleware": []string{"global", "auth", "contentValidation"},
			})
		})

	// ===== Conditional Middleware =====

	// Middleware that runs conditionally
	conditionalMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Only apply special handling for JSON requests
		if ctx.GetHeader("Content-Type") == "application/json" {
			lokstra.Logger.Infof("🔄 JSON request detected - applying special handling")
			ctx.Response.WithHeader("X-JSON-Processed", "true")
		}

		return next(ctx)
	}

	app.POST("/conditional", conditionalMiddleware, func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Conditional middleware example",
			"note":    "Check X-JSON-Processed header",
		})
	})

	// ===== Error Handling in Middleware =====

	// Middleware that handles errors gracefully
	errorHandlingMiddleware := func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		err := next(ctx)

		if err != nil {
			lokstra.Logger.Errorf("❌ Middleware caught error: %v", err)
			// Could transform or log error here
		}

		return err
	}

	app.GET("/error-test", errorHandlingMiddleware, func(ctx *lokstra.Context) error {
		errorType := ctx.GetQueryParam("type")

		switch errorType {
		case "panic":
			panic("Test panic")
		case "error":
			return ctx.ErrorInternal("Test error")
		default:
			return ctx.Ok(map[string]any{
				"message": "No error",
				"try":     "?type=panic or ?type=error",
			})
		}
	})

	lokstra.Logger.Infof("🚀 Middleware Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Middleware Examples:")
	lokstra.Logger.Infof("  Public Routes (global middleware only):")
	lokstra.Logger.Infof("    GET  /                    - Home page")
	lokstra.Logger.Infof("    GET  /public              - Public endpoint")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Protected Routes (auth required):")
	lokstra.Logger.Infof("    GET  /protected/profile   - User profile")
	lokstra.Logger.Infof("    GET  /protected/dashboard - User dashboard")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Admin Routes (auth + admin role):")
	lokstra.Logger.Infof("    GET    /admin/users       - Admin user list")
	lokstra.Logger.Infof("    DELETE /admin/users/123   - Admin delete user")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  API Routes (rate limit + content validation):")
	lokstra.Logger.Infof("    GET  /api/status          - API status")
	lokstra.Logger.Infof("    POST /api/items           - Create item (JSON required)")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Special Routes:")
	lokstra.Logger.Infof("    GET  /special             - Route-specific middleware")
	lokstra.Logger.Infof("    POST /upload              - Multiple route middleware")
	lokstra.Logger.Infof("    POST /conditional         - Conditional middleware")
	lokstra.Logger.Infof("    GET  /error-test?type=error - Error handling")

	app.Start()
}

// Middleware Key Concepts:
//
// 1. Execution Order:
//    - Global middleware runs first
//    - Group middleware runs next
//    - Route-specific middleware runs last
//    - Handler runs after all middleware
//
// 2. Middleware Function Signature:
//    - func(ctx *lokstra.Context, next func(*lokstra.Context) error) error
//    - Call next(ctx) to continue the chain
//    - Return error to stop execution
//
// 3. Common Middleware Patterns:
//    - Authentication: Verify user credentials
//    - Authorization: Check user permissions
//    - Logging: Log request/response details
//    - CORS: Handle cross-origin requests
//    - Rate Limiting: Prevent abuse
//    - Content Validation: Validate request format
//
// 4. Middleware Application:
//    - app.Use() for global middleware
//    - group.Use() for group middleware
//    - Inline for route-specific middleware
//
// 5. Error Handling:
//    - Middleware can catch and handle errors
//    - Early return stops middleware chain
//    - Errors bubble up through middleware stack

// Test Commands:
//
// # Public routes (no auth needed)
// curl http://localhost:8080/
// curl http://localhost:8080/public
//
// # Protected routes (auth required)
// curl -H "Authorization: Bearer valid-token" http://localhost:8080/protected/profile
// curl http://localhost:8080/protected/profile  # Should fail without auth
//
// # Admin routes (auth + admin role required)
// curl -H "Authorization: Bearer valid-token" -H "X-User-Role: admin" http://localhost:8080/admin/users
// curl -H "Authorization: Bearer valid-token" -H "X-User-Role: user" http://localhost:8080/admin/users  # Should fail
//
// # API routes (rate limit + content validation)
// curl http://localhost:8080/api/status
// curl -X POST http://localhost:8080/api/items -H "Content-Type: application/json" -d '{"name":"Test Item"}'
// curl -X POST http://localhost:8080/api/items -H "Content-Type: text/plain" -d 'test'  # Should fail
//
// # Special routes
// curl http://localhost:8080/special
// curl -X POST http://localhost:8080/conditional -H "Content-Type: application/json" -d '{}'
// curl -X POST http://localhost:8080/conditional -H "Content-Type: text/plain" -d 'test'
//
// # Error handling
// curl "http://localhost:8080/error-test?type=error"
