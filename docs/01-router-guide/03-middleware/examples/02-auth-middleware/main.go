package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// Simple auth middleware - checks for API key in header
func authMiddleware(ctx *request.Context) error {
	apiKey := ctx.Req.HeaderParam("X-API-Key", "")

	if apiKey == "" {
		return ctx.Api.Unauthorized("API key required")
	}

	if apiKey != "secret-key-123" {
		return ctx.Api.Forbidden("Invalid API key")
	}

	// Authentication successful - continue to handler
	return ctx.Next()
}

// Admin-only middleware - checks for admin role
func adminMiddleware(ctx *request.Context) error {
	role := ctx.Req.HeaderParam("X-User-Role", "")

	if role != "admin" {
		return ctx.Api.Forbidden("Admin access required")
	}

	// Authorization successful - continue to handler
	return ctx.Next()
}

func main() {
	router := lokstra.NewRouter("api")

	// Public routes - no auth required
	router.GET("/public", func() map[string]any {
		return map[string]any{
			"message": "This is a public endpoint",
			"access":  "everyone",
		}
	})

	// Protected route - requires API key
	router.GET("/protected", func() map[string]any {
		return map[string]any{
			"message": "This is a protected endpoint",
			"access":  "authenticated users only",
		}
	}, authMiddleware)

	// Admin route - requires API key AND admin role
	router.GET("/admin", func() map[string]any {
		return map[string]any{
			"message": "This is an admin endpoint",
			"access":  "admins only",
		}
	}, authMiddleware, adminMiddleware)

	// Protected routes group - all require auth
	apiGroup := router.AddGroup("/api")
	apiGroup.Use(authMiddleware)

	apiGroup.GET("/users", func() map[string]any {
		return map[string]any{
			"users": []string{"Alice", "Bob", "Charlie"},
		}
	})

	apiGroup.GET("/orders", func() map[string]any {
		return map[string]any{
			"orders": []int{101, 102, 103},
		}
	})

	// Admin routes group - require both auth and admin role
	adminGroup := router.AddGroup("/api/admin")
	adminGroup.Use(authMiddleware, adminMiddleware)

	adminGroup.GET("/stats", func() map[string]any {
		return map[string]any{
			"total_users":  150,
			"total_orders": 1250,
		}
	})

	adminGroup.GET("/logs", func() map[string]any {
		return map[string]any{
			"logs": []string{"Event 1", "Event 2", "Event 3"},
		}
	})

	// Create app
	app := lokstra.NewApp("auth-demo", ":3000", router)

	log.Println("🚀 Authentication Middleware Demo")
	log.Println("🔒 Demonstrates auth and admin middleware")
	log.Println()
	log.Println("Endpoints:")
	log.Println("  Public:")
	log.Println("    GET /public               (no auth)")
	log.Println()
	log.Println("  Protected (requires X-API-Key: secret-key-123):")
	log.Println("    GET /protected")
	log.Println("    GET /api/users")
	log.Println("    GET /api/orders")
	log.Println()
	log.Println("  Admin (requires X-API-Key + X-User-Role: admin):")
	log.Println("    GET /admin")
	log.Println("    GET /api/admin/stats")
	log.Println("    GET /api/admin/logs")
	log.Println()
	log.Println("Server: http://localhost:3000")

	// Run
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
