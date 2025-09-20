package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

// This example demonstrates advanced middleware chaining patterns in Lokstra.
// It shows different middleware execution orders, conditional middleware,
// error handling in middleware chains, and middleware communication.
//
// Learning Objectives:
// - Understand middleware execution order
// - Learn conditional middleware patterns
// - Master middleware error handling
// - Explore middleware communication
// - See performance monitoring with middleware
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/middleware.md

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "middleware-chains-app", ":8080")

	// ===== Logger Service Setup =====
	regCtx.RegisterModule(logger.GetModule)
	regCtx.CreateService("lokstra.logger", "app-logger", true, "debug")

	// ===== Register Middleware Functions =====

	// 1. Request Logger Middleware
	regCtx.RegisterMiddlewareFunc("request_logger", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
			start := time.Now()

			logger.Infof("[REQUEST] %s %s from %s",
				ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.RemoteAddr)

			// Call next middleware/handler
			err := next(ctx)

			// Post-processing after response
			duration := time.Since(start)
			status := ctx.StatusCode
			if status == 0 {
				status = 200 // Default success status
			}

			logger.Infof("[RESPONSE] %s %s -> %d (%v)",
				ctx.Request.Method, ctx.Request.URL.Path, status, duration)

			return err
		}
	})

	// 2. CORS Middleware
	regCtx.RegisterMiddlewareFunc("cors", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

			// Add CORS headers
			ctx.WithHeader("Access-Control-Allow-Origin", "*")
			ctx.WithHeader("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			ctx.WithHeader("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key")

			// Handle preflight requests
			if ctx.Request.Method == "OPTIONS" {
				logger.Debugf("[CORS] Preflight request handled for %s", ctx.Request.URL.Path)
				ctx.Writer.WriteHeader(204)
				return nil
			}

			logger.Debugf("[CORS] Headers added for %s", ctx.Request.URL.Path)
			return next(ctx)
		}
	})

	// 3. Rate Limiting Middleware
	var requestCount int
	regCtx.RegisterMiddlewareFunc("rate_limit", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

			requestCount++

			// Simple rate limiting (max 100 requests per instance)
			if requestCount > 100 {
				logger.Warnf("[RATE_LIMIT] Request limit exceeded: %d", requestCount)
				ctx.SetStatusCode(429) // Too Many Requests
				return ctx.ErrorBadRequest("Rate limit exceeded")
			}

			logger.Debugf("[RATE_LIMIT] Request count: %d/100", requestCount)
			return next(ctx)
		}
	})

	// 4. Authentication Middleware
	regCtx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

			// Simple API key authentication
			apiKey := ctx.GetHeader("X-API-Key")
			if apiKey == "" {
				logger.Warnf("[AUTH] Missing API key for %s", ctx.Request.URL.Path)
				ctx.SetStatusCode(401) // Unauthorized
				return ctx.ErrorBadRequest("API key required")
			}

			// Validate API key (simple check for demo)
			if apiKey != "demo-api-key-123" && apiKey != "demo-api-key-123-admin" {
				logger.Warnf("[AUTH] Invalid API key: %s", apiKey)
				ctx.SetStatusCode(401) // Unauthorized
				return ctx.ErrorBadRequest("Invalid API key")
			}

			// Store user context (using standard context)
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(ctx.Request.Context(), userIDKey, "demo-user"))
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(ctx.Request.Context(), apiKeyKey, apiKey))

			logger.Infof("[AUTH] Authenticated user for %s", ctx.Request.URL.Path)
			return next(ctx)
		}
	})

	// 5. Admin Authorization Middleware
	regCtx.RegisterMiddlewareFunc("admin_auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

			// Check for admin privileges (based on API key suffix)
			apiKey := getContextValue(ctx, apiKeyKey)
			if !strings.HasSuffix(apiKey, "-admin") {
				logger.Warnf("[ADMIN_AUTH] Non-admin access attempt with key: %s", apiKey)
				ctx.SetStatusCode(403) // Forbidden
				return ctx.ErrorBadRequest("Admin privileges required")
			}

			logger.Infof("[ADMIN_AUTH] Admin access granted for %s", ctx.Request.URL.Path)
			return next(ctx)
		}
	})

	// ===== Apply Global Middleware =====
	app.Use("request_logger")
	app.Use("cors")
	app.Use("rate_limit")

	// ===== Route Handlers =====

	// Home endpoint - demonstrates global middleware chain
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Advanced Middleware Chains Example",
			"middleware_chain": []string{
				"Request Logger",
				"CORS Handler",
				"Rate Limiter",
			},
			"routes": map[string]any{
				"public": []string{
					"GET /",
					"GET /health",
					"POST /webhook",
				},
				"protected": []string{
					"GET /api/protected/profile",
					"POST /api/protected/data",
				},
				"admin": []string{
					"GET /api/protected/admin/users",
					"DELETE /api/protected/admin/users/:id",
				},
			},
		})
	})

	// Health check - minimal middleware processing
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status":     "healthy",
			"timestamp":  time.Now(),
			"middleware": "global chain executed",
		})
	})

	// Webhook endpoint with conditional middleware
	app.POST("/webhook", func(ctx *lokstra.Context) error {
		logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Conditional middleware - verify webhook signature
		signature := ctx.GetHeader("X-Webhook-Signature")
		if signature == "" {
			logger.Warnf("[WEBHOOK] Missing signature header")
			return ctx.ErrorBadRequest("Webhook signature required")
		}

		// Simulate signature verification
		if signature != "valid-signature-123" {
			logger.Warnf("[WEBHOOK] Invalid signature: %s", signature)
			ctx.SetStatusCode(401) // Unauthorized
			return ctx.ErrorBadRequest("Invalid webhook signature")
		}

		logger.Infof("[WEBHOOK] Valid webhook received")
		return ctx.Ok(map[string]any{
			"message": "Webhook processed",
			"status":  "success",
		})
	})

	// Protected routes group with authentication middleware
	protected := app.Group("/api/protected", "auth")

	// Protected profile endpoint
	protected.GET("/profile", func(ctx *lokstra.Context) error {
		userID := getContextValue(ctx, userIDKey)
		return ctx.Ok(map[string]any{
			"user_id": userID,
			"profile": map[string]any{
				"name":  "Demo User",
				"email": "demo@example.com",
				"role":  "user",
			},
			"middleware_chain": []string{
				"Request Logger",
				"CORS Handler",
				"Rate Limiter",
				"Authentication",
			},
		})
	})

	// Protected data endpoint with request validation
	protected.POST("/data", func(ctx *lokstra.Context, req *DataRequest) error {
		userID := getContextValue(ctx, userIDKey)
		logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

		logger.Infof("[DATA] Processing data for user: %s, type: %s", userID, req.Type)

		// Calculate data size safely
		dataSize := 0
		if req.Data != nil {
			if str, ok := req.Data.(string); ok {
				dataSize = len(str)
			} else if bytes, ok := req.Data.([]byte); ok {
				dataSize = len(bytes)
			}
		}

		return ctx.Ok(map[string]any{
			"message":   "Data processed successfully",
			"user_id":   userID,
			"data_type": req.Type,
			"data_size": dataSize,
			"processed": time.Now(),
		})
	})

	// Admin routes group with additional authorization
	admin := protected.Group("/admin", "admin_auth")

	// Admin users endpoint
	admin.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"users": []map[string]any{
				{"id": 1, "name": "Admin User", "role": "admin"},
				{"id": 2, "name": "Regular User", "role": "user"},
				{"id": 3, "name": "Demo User", "role": "user"},
			},
			"middleware_chain": []string{
				"Request Logger",
				"CORS Handler",
				"Rate Limiter",
				"Authentication",
				"Admin Authorization",
			},
			"admin_access": true,
		})
	})

	// Admin delete user endpoint with method-specific middleware
	admin.DELETE("/users/:id", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")
		logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Additional validation middleware for destructive operations
		confirmHeader := ctx.GetHeader("X-Confirm-Delete")
		if confirmHeader != "yes" {
			logger.Warnf("[DELETE] Deletion attempt without confirmation for user: %s", userID)
			return ctx.ErrorBadRequest("Deletion confirmation required (X-Confirm-Delete: yes)")
		}

		logger.Warnf("[DELETE] User deletion confirmed: %s", userID)

		return ctx.Ok(map[string]any{
			"message":    "User deleted successfully",
			"deleted_id": userID,
			"timestamp":  time.Now(),
			"admin_user": getContextValue(ctx, userIDKey),
		})
	})

	// Error demonstration endpoint
	app.GET("/error-demo", func(ctx *lokstra.Context) error {
		errorType := ctx.GetQueryParam("type")

		switch errorType {
		case "middleware":
			// This will be caught by error handling middleware
			return fmt.Errorf("simulated middleware error")
		case "panic":
			panic("simulated panic")
		case "timeout":
			time.Sleep(5 * time.Second)
			return ctx.Ok("delayed response")
		default:
			return ctx.Ok(map[string]any{
				"message": "Error demo endpoint",
				"types": []string{
					"?type=middleware - middleware error",
					"?type=panic - panic recovery",
					"?type=timeout - request timeout",
				},
			})
		}
	})

	// Middleware status endpoint
	app.GET("/middleware-status", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"middleware_statistics": map[string]any{
				"request_count":        requestCount,
				"rate_limit_max":       100,
				"rate_limit_remaining": 100 - requestCount,
			},
			"middleware_chain": map[string]any{
				"global": []string{
					"Request Logger",
					"CORS Handler",
					"Rate Limiter",
				},
				"protected": []string{
					"+ Authentication",
				},
				"admin": []string{
					"+ Admin Authorization",
				},
			},
			"authentication": map[string]any{
				"type":      "API Key",
				"header":    "X-API-Key",
				"demo_key":  "demo-api-key-123",
				"admin_key": "demo-api-key-123-admin",
			},
		})
	})

	lokstra.Logger.Infof("🚀 Advanced Middleware Chains Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Available Endpoints:")
	lokstra.Logger.Infof("  GET  /                          - Home with middleware info")
	lokstra.Logger.Infof("  GET  /health                    - Health check")
	lokstra.Logger.Infof("  POST /webhook                   - Webhook with signature verification")
	lokstra.Logger.Infof("  GET  /error-demo                - Error handling demonstration")
	lokstra.Logger.Infof("  GET  /middleware-status         - Middleware chain status")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Protected Endpoints (API Key required):")
	lokstra.Logger.Infof("  GET  /api/protected/profile     - User profile")
	lokstra.Logger.Infof("  POST /api/protected/data        - Data processing")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Admin Endpoints (Admin API Key required):")
	lokstra.Logger.Infof("  GET    /api/protected/admin/users      - List users")
	lokstra.Logger.Infof("  DELETE /api/protected/admin/users/:id  - Delete user")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Authentication:")
	lokstra.Logger.Infof("  Header: X-API-Key")
	lokstra.Logger.Infof("  User Key: demo-api-key-123")
	lokstra.Logger.Infof("  Admin Key: demo-api-key-123-admin")

	app.Start()
}

// ===== Context Keys =====

type contextKey string

const (
	userIDKey contextKey = "user_id"
	apiKeyKey contextKey = "api_key"
)

// ===== Helper Functions =====

// getContextValue safely retrieves a value from context
func getContextValue(ctx *lokstra.Context, key any) string {
	if value := ctx.Request.Context().Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// ===== Request Types =====

type DataRequest struct {
	Type string `json:"type" validate:"required,oneof=user content analytics"`
	Data any    `json:"data" validate:"required"`
}

// Middleware Chain Key Concepts:
//
// 1. Execution Order:
//    - Middleware executes in registration order (top to bottom)
//    - Post-processing happens in reverse order (bottom to top)
//    - ctx.Next() calls the next middleware in chain
//    - Return without ctx.Next() stops the chain
//
// 2. Middleware Types:
//    - Global: Applied to all routes
//    - Group: Applied to route groups
//    - Route-specific: Applied to individual routes
//    - Conditional: Applied based on request conditions
//
// 3. Error Handling:
//    - Middleware can return errors to stop chain
//    - Errors bubble up through middleware stack
//    - Error middleware can handle and transform errors
//    - Panic recovery can be implemented in middleware
//
// 4. Context Communication:
//    - Use ctx.Set/Get for middleware communication
//    - Pass data between middleware layers
//    - Store authentication, user data, request metadata
//    - Available throughout request lifecycle
//
// 5. Performance Considerations:
//    - Minimize middleware overhead
//    - Use conditional execution when possible
//    - Avoid blocking operations in middleware
//    - Consider middleware order for optimization

// Test Commands:
//
// # Start the server
// go run main.go
//
// # Test public endpoints
// curl http://localhost:8080/
// curl http://localhost:8080/health
// curl http://localhost:8080/middleware-status
//
// # Test webhook with signature
// curl -X POST http://localhost:8080/webhook \
//   -H "X-Webhook-Signature: valid-signature-123" \
//   -d '{"event":"test"}'
//
// # Test webhook without signature (should fail)
// curl -X POST http://localhost:8080/webhook \
//   -d '{"event":"test"}'
//
// # Test protected endpoints (should fail without API key)
// curl http://localhost:8080/api/protected/profile
//
// # Test with API key
// curl http://localhost:8080/api/protected/profile \
//   -H "X-API-Key: demo-api-key-123"
//
// # Test data endpoint
// curl -X POST http://localhost:8080/api/protected/data \
//   -H "X-API-Key: demo-api-key-123" \
//   -H "Content-Type: application/json" \
//   -d '{"type":"user","data":{"name":"test"}}'
//
// # Test admin endpoints (should fail with user key)
// curl http://localhost:8080/api/protected/admin/users \
//   -H "X-API-Key: demo-api-key-123"
//
// # Test with admin key
// curl http://localhost:8080/api/protected/admin/users \
//   -H "X-API-Key: demo-api-key-123-admin"
//
// # Test user deletion (requires confirmation)
// curl -X DELETE http://localhost:8080/api/protected/admin/users/123 \
//   -H "X-API-Key: demo-api-key-123-admin" \
//   -H "X-Confirm-Delete: yes"
//
// # Test error scenarios
// curl http://localhost:8080/error-demo?type=middleware
// curl http://localhost:8080/error-demo?type=panic
//
// # Test CORS preflight
// curl -X OPTIONS http://localhost:8080/api/protected/profile \
//   -H "Access-Control-Request-Method: GET" \
//   -H "Access-Control-Request-Headers: X-API-Key"
