package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates how to run multiple applications within a single server.
// It shows how to organize different services/APIs into separate apps with different ports.
//
// Learning Objectives:
// - Understand multi-app architecture
// - Learn how to create a server with multiple applications
// - See different apps serving different purposes (API, Admin, WebSocket)
// - Understand graceful shutdown handling
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/getting-started.md#multi-app-architecture

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Create the main server
	server := lokstra.NewServer(regCtx, "multi-app-server")

	// ===== App 1: Public API =====
	// Serves public REST API endpoints
	publicAPI := lokstra.NewApp(regCtx, "public-api", ":8080")

	publicAPI.GET("/api/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"service": "Public API",
			"status":  "healthy",
		})
	})

	publicAPI.GET("/api/users", func(ctx *lokstra.Context) error {
		users := []map[string]any{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		}
		return ctx.OkList(users, map[string]any{
			"total": len(users),
			"page":  1,
		})
	})

	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	publicAPI.POST("/api/users", func(ctx *lokstra.Context, req *CreateUserRequest) error {
		user := map[string]any{
			"id":    3,
			"name":  req.Name,
			"email": req.Email,
		}
		return ctx.OkCreated(user)
	})

	// Add the public API to the server
	server.AddApp(publicAPI)

	// ===== App 2: Admin Panel =====
	// Serves admin dashboard and management endpoints
	adminAPI := lokstra.NewApp(regCtx, "admin-api", ":8081")

	adminAPI.GET("/admin/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"service": "Admin API",
			"status":  "healthy",
		})
	})

	adminAPI.GET("/admin/dashboard", func(ctx *lokstra.Context) error {
		stats := map[string]any{
			"total_users":    150,
			"active_users":   120,
			"revenue":        25000.50,
			"pending_orders": 12,
		}
		return ctx.Ok(stats)
	})

	adminAPI.GET("/admin/users", func(ctx *lokstra.Context) error {
		// Admin sees more detailed user information
		users := []map[string]any{
			{
				"id":         1,
				"name":       "John Doe",
				"email":      "john@example.com",
				"created_at": "2024-01-01T10:00:00Z",
				"status":     "active",
				"role":       "user",
			},
			{
				"id":         2,
				"name":       "Jane Smith",
				"email":      "jane@example.com",
				"created_at": "2024-01-02T11:00:00Z",
				"status":     "active",
				"role":       "admin",
			},
		}
		return ctx.OkList(users, map[string]any{
			"total":      len(users),
			"page":       1,
			"per_page":   10,
			"can_delete": true,
			"can_edit":   true,
		})
	})

	type BanUserRequest struct {
		UserID int    `path:"id"`
		Reason string `json:"reason" validate:"required"`
	}

	adminAPI.POST("/admin/users/:id/ban", func(ctx *lokstra.Context, req *BanUserRequest) error {
		return ctx.Ok(map[string]any{
			"message": "User banned successfully",
			"user_id": req.UserID,
			"reason":  req.Reason,
		})
	})

	// Add the admin API to the server
	server.AddApp(adminAPI)

	// ===== App 3: Metrics/Monitoring =====
	// Serves application metrics and monitoring endpoints
	metricsAPI := lokstra.NewApp(regCtx, "metrics-api", ":8082")

	metricsAPI.GET("/metrics", func(ctx *lokstra.Context) error {
		metrics := map[string]any{
			"requests_per_second": 45.2,
			"response_time_avg":   120,
			"error_rate":          0.02,
			"uptime_seconds":      86400,
			"memory_usage_mb":     256,
			"cpu_usage_percent":   15.5,
		}
		return ctx.Ok(metrics)
	})

	metricsAPI.GET("/metrics/detailed", func(ctx *lokstra.Context) error {
		detailed := map[string]any{
			"http_requests": map[string]any{
				"total":   10540,
				"success": 10330,
				"errors":  210,
			},
			"endpoints": map[string]any{
				"/api/users": map[string]any{
					"requests": 5420,
					"avg_time": 89,
				},
				"/api/health": map[string]any{
					"requests": 3210,
					"avg_time": 12,
				},
			},
			"database": map[string]any{
				"connections_active": 8,
				"connections_total":  25,
				"query_time_avg":     45,
			},
		}
		return ctx.Ok(detailed)
	})

	// Add the metrics API to the server
	server.AddApp(metricsAPI)

	// Start all applications
	lokstra.Logger.Infof("Multi-App Server Configuration:")
	lokstra.Logger.Infof("  📡 Public API:  http://localhost:8080")
	lokstra.Logger.Infof("     GET /api/health")
	lokstra.Logger.Infof("     GET /api/users")
	lokstra.Logger.Infof("     POST /api/users")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  🔐 Admin API:   http://localhost:8081")
	lokstra.Logger.Infof("     GET /admin/health")
	lokstra.Logger.Infof("     GET /admin/dashboard")
	lokstra.Logger.Infof("     GET /admin/users")
	lokstra.Logger.Infof("     POST /admin/users/:id/ban")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  📊 Metrics API: http://localhost:8082")
	lokstra.Logger.Infof("     GET /metrics")
	lokstra.Logger.Infof("     GET /metrics/detailed")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Starting multi-app server...")

	// Start the server (this will start all apps concurrently)
	if err := server.Start(); err != nil {
		lokstra.Logger.Errorf("Server failed to start: %v", err)
	}
}

// Multi-App Architecture Benefits:
//
// 1. Separation of Concerns:
//    - Each app has a specific purpose and responsibility
//    - Different ports for different services
//    - Easier to maintain and scale
//
// 2. Independent Scaling:
//    - Can scale each app independently based on load
//    - Different resource requirements per app
//    - Targeted performance optimization
//
// 3. Security Isolation:
//    - Admin endpoints on separate port
//    - Different middleware and authentication per app
//    - Network-level access controls
//
// 4. Development Benefits:
//    - Team can work on different apps independently
//    - Easier testing and deployment
//    - Clear API boundaries
//
// 5. Operational Benefits:
//    - Health checks per service
//    - Independent monitoring and logging
//    - Graceful shutdown coordination

// Test Commands:
//
// Public API:
//   curl http://localhost:8080/api/health
//   curl http://localhost:8080/api/users
//   curl -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"name":"Test User","email":"test@example.com"}'
//
// Admin API:
//   curl http://localhost:8081/admin/health
//   curl http://localhost:8081/admin/dashboard
//   curl http://localhost:8081/admin/users
//   curl -X POST http://localhost:8081/admin/users/1/ban -H "Content-Type: application/json" -d '{"reason":"Violation of terms"}'
//
// Metrics API:
//   curl http://localhost:8082/metrics
//   curl http://localhost:8082/metrics/detailed
