package main

import (
	"context"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	// Create application with builtin services
	app := app.NewApp(regCtx, "Database Pool Example", ":8080")

	createDbServiceAndPing(regCtx)

	registerHandlers(regCtx)

	// Register routes that demonstrate database usage
	registerRoutes(app)

	// Start the application
	app.Start()
}

func createDbServiceAndPing(regCtx lokstra.RegistrationContext) {
	// Configure and register database pool service
	dbConfig := map[string]any{
		"connection_string": "postgres://localhost:5432/testdb?sslmode=disable",
		"max_connections":   10,
		"min_connections":   2,
		"max_idle_time":     "30m",
		"max_lifetime":      "1h",
	}

	// Register PostgreSQL database pool
	dbService, err := registration.CreateService[serviceapi.DbPool](
		regCtx, dbpool_pg.GetModule().Name(), "dbMain", false, dbConfig)
	if err != nil {
		panic("Failed to create database service")
	}

	// test connection
	ctx := context.Background()
	conn, err := dbService.Acquire(ctx, "")
	if err != nil {
		panic("Failed to acquire database connection")
	}
	err = conn.Ping(ctx)
	if err != nil {
		panic("Failed to ping database")
	}
	conn.Release()
}

func registerRoutes(app *app.App) {

	// Health check with database status
	app.GET("/health", "handleHealth")

	// Database operations
	app.GET("/users", "handleGetUsers")
	app.POST("/users", "handleCreateUser")
	app.GET("/users/:id", "handleGetUser")
	app.PUT("/users/:id", "handleUpdateUser")
	app.DELETE("/users/:id", "handleDeleteUser")

	// Database statistics
	app.GET("/db/stats", "handleDatabaseStats")
}

func registerHandlers(regCtx registration.Context) {
	log, err := serviceapi.GetService[serviceapi.Logger](regCtx, "logger")
	if err != nil {
		panic("Logger service not available")
	}

	dbService, err := serviceapi.GetService[serviceapi.DbPool](regCtx, "dbMain")
	if err != nil {
		log.Errorf("Database service not available: %v", err)
	}

	regCtx.RegisterHandler("handleHealth", func(ctx *request.Context) error {
		conn, err := dbService.Acquire(ctx, "")
		if err != nil {
			log.Errorf("Failed to acquire database connection: %v", err)
			return ctx.Ok(map[string]any{
				"status":   "unhealthy",
				"database": "connection failed",
				"error":    err.Error(),
			})
		}
		defer conn.Release()

		// Simple ping test
		err = conn.Ping(ctx)
		if err != nil {
			log.Errorf("Database ping failed: %v", err)
			return ctx.JSON(map[string]any{
				"status":   "unhealthy",
				"database": "ping failed",
				"error":    err.Error(),
			})
		}

		return ctx.JSON(map[string]any{
			"status":   "healthy",
			"database": "connected",
			"message":  "All systems operational",
		})
	})

	regCtx.RegisterHandler("handleGetUsers", func(ctx *request.Context) error {
		conn, err := dbService.Acquire(ctx, "")
		if err != nil {
			log.Errorf("Failed to acquire database connection: %v", err)
			return ctx.ErrorInternal("Database connection failed")
		}
		defer conn.Release()

		// Simulate getting users (in real app, you'd have actual SQL queries)
		log.Infof("Fetching users from database")

		// Mock data for demonstration
		users := []map[string]any{
			{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
			},
			{
				"id":    2,
				"name":  "Jane Smith",
				"email": "jane@example.com",
			},
		}

		return ctx.JSON(map[string]any{
			"users": users,
			"count": len(users),
		})
	})

	// Parse request body
	type CreateReq struct {
		Name  string `body:"name"`
		Email string `body:"email"`
	}

	regCtx.RegisterHandler("handleCreateUser",
		func(ctx *request.Context, userReq *CreateReq) error {
			// Validate input
			if userReq.Name == "" || userReq.Email == "" {
				return ctx.ErrorBadRequest("Name and email are required")
			}

			conn, err := dbService.Acquire(ctx.Request.Context(), "")
			if err != nil {
				log.Errorf("Failed to acquire database connection: %v", err)
				return ctx.ErrorInternal("Database connection failed")
			}
			defer conn.Release()

			// Simulate creating user
			log.Infof("Creating user: %s <%s>", userReq.Name, userReq.Email)

			// In real app, you'd execute INSERT SQL here
			// For demo, return mock created user
			createdUser := map[string]any{
				"id":    999, // mock ID
				"name":  userReq.Name,
				"email": userReq.Email,
			}

			return ctx.JSON(map[string]any{
				"message": "User created successfully",
				"user":    createdUser,
			})
		})

	regCtx.RegisterHandler("handleGetUser", func(ctx *request.Context) error {
		userID := ctx.GetPathParam("id")

		conn, err := dbService.Acquire(ctx, "")
		if err != nil {
			log.Errorf("Failed to acquire database connection: %v", err)
			return ctx.ErrorInternal("Database connection failed")
		}
		defer conn.Release()

		log.Infof("Fetching user with ID: %s", userID)

		// Mock user data for demonstration
		user := map[string]any{
			"id":    userID,
			"name":  "Sample User",
			"email": "user@example.com",
		}

		return ctx.JSON(map[string]any{
			"user": user,
		})
	})

	// Parse request body
	type updateRequest struct {
		UserID string `path:"id"`
		Name   string `body:"name"`
		Email  string `body:"email"`
	}

	regCtx.RegisterHandler("handleUpdateUser",
		func(ctx *request.Context, updateReq *updateRequest) error {
			conn, err := dbService.Acquire(ctx, "")
			if err != nil {
				log.Errorf("Failed to acquire database connection: %v", err)
				return ctx.ErrorInternal("Database connection failed")
			}
			defer conn.Release()

			log.Infof("Updating user %s: %s <%s>",
				updateReq.UserID, updateReq.Name, updateReq.Email)

			// Mock updated user
			updatedUser := map[string]any{
				"id":    updateReq.UserID,
				"name":  updateReq.Name,
				"email": updateReq.Email,
			}

			return ctx.JSON(map[string]any{
				"message": "User updated successfully",
				"user":    updatedUser,
			})
		})

	regCtx.RegisterHandler("handleDeleteUser", func(ctx *request.Context) error {
		userID := ctx.GetPathParam("id")

		conn, err := dbService.Acquire(ctx.Request.Context(), "")
		if err != nil {
			log.Errorf("Failed to acquire database connection: %v", err)
			return ctx.ErrorInternal("Database connection failed")
		}
		defer conn.Release()

		log.Infof("Deleting user with ID: %s", userID)

		return ctx.JSON(map[string]any{
			"message": "User deleted successfully",
			"id":      userID,
		})
	})

	regCtx.RegisterHandler("handleDatabaseStats",
		func(ctx *request.Context) error {
			// Get pool statistics (if available in your DbPool interface)
			stats := map[string]any{
				"status":             "connected",
				"active_connections": "N/A", // Would come from actual pool stats
				"idle_connections":   "N/A", // Would come from actual pool stats
				"max_connections":    10,    // From config
				"note":               "Statistics would come from actual pool implementation",
			}

			return ctx.JSON(map[string]any{
				"database_stats": stats,
			})
		})
}
