package main

import (
	"github.com/google/uuid"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
)

func registerRouters() {
	// Register manual routers (not generated from @RouterService)
	healthRouter := NewHealthRouter()
	lokstra_registry.RegisterRouter("health-router", healthRouter)
	logger.LogInfo("✅ Registered manual router: health-router")
}

func registerMiddlewareTypes() {
	// Register recovery middleware (built-in)
	recovery.Register()

	// Register custom middleware
	lokstra_registry.RegisterMiddlewareFactory("request-logger", requestLoggerFactory)
	lokstra_registry.RegisterMiddlewareFactory("simple-auth", simpleAuthFactory)
	lokstra_registry.RegisterMiddlewareFactory("mw-test", func(config map[string]any) request.HandlerFunc {
		return func(ctx *request.Context) error {
			logger.LogInfo("→ [mw-test] Before request | Param1: %v, Param2: %v", config["param1"], config["param2"])
			err := ctx.Next()
			logger.LogInfo("← [mw-test] After request")
			return err
		}
	})
}

func requestLoggerFactory(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Before request
		reqID := uuid.New().String()
		logger.LogInfo("→ [%s] %s %s", reqID, ctx.R.Method, ctx.R.URL.Path)
		ctx.Set("request_id", reqID)

		// Process request
		err := ctx.Next()

		// After request
		if err != nil {
			logger.LogInfo("← [%s] ERROR: %v", reqID, err)
		} else {
			logger.LogInfo("← [%s] SUCCESS (status: %d)", reqID, ctx.Resp.RespStatusCode)
		}

		return err
	}
}

// simpleAuthFactory creates a simple authentication middleware
// Checks for "Authorization" header with Bearer token
// For demo purposes, accepts any token that starts with "demo-"
func simpleAuthFactory(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Get Authorization header
		authHeader := ctx.R.Header.Get("Authorization")

		// Check if Authorization header exists
		if authHeader == "" {
			logger.LogInfo("🔒 [simple-auth] Missing Authorization header")
			return ctx.Api.Unauthorized("Missing Authorization header")
		}

		// Check Bearer token format
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			logger.LogInfo("🔒 [simple-auth] Invalid Authorization format")
			return ctx.Api.Unauthorized("Invalid Authorization format. Use 'Bearer <token>'")
		}

		// Extract token
		token := authHeader[len(bearerPrefix):]

		// Simple validation: accept tokens starting with "demo-"
		// In production, validate against database or JWT
		if len(token) < 5 || token[:5] != "demo-" {
			logger.LogInfo("🔒 [simple-auth] Invalid token: %s", token)
			return ctx.Api.Unauthorized("Invalid token")
		}

		// Token is valid - store user info in context
		userID := token[5:] // Extract user ID from "demo-{userID}"
		ctx.Set("user_id", userID)
		ctx.Set("authenticated", true)

		logger.LogInfo("✅ [simple-auth] Authenticated user: %s", userID)

		// Continue to next handler
		return ctx.Next()
	}
}
