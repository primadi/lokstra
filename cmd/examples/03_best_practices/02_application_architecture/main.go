package main

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/handlers"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/repository"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/services"
	"github.com/primadi/lokstra/serviceapi"
)

func main() {
	// Create registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Register components before loading config
	registerComponents(regCtx)

	// Create server from YAML configuration
	server := newServerFromConfig(regCtx, "lokstra.yaml")

	// Start the server
	log.Println("Starting Application Architecture Example Server...")
	log.Println("Server will be available at http://localhost:8080")
	log.Println("API Documentation:")
	log.Println("  GET    /api/health        - Health check")
	log.Println("  GET    /api/users         - List all users")
	log.Println("  GET    /api/users/:id     - Get user by ID")
	log.Println("  POST   /api/users         - Create new user")
	log.Println("  PUT    /api/users/:id     - Update user")
	log.Println("  DELETE /api/users/:id     - Delete user")

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerComponents(regCtx lokstra.RegistrationContext) {
	// Register health check handler
	regCtx.RegisterHandler("health.check", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"status":      "healthy",
			"service":     "application_architecture_example",
			"version":     "1.0.0",
			"description": "Lokstra Application Architecture Best Practices Example",
		})
	})
}

func setupUserManagement(regCtx lokstra.RegistrationContext) {
	// Get required services
	dbPoolService, err := regCtx.GetService("db_pool")
	if err != nil {
		log.Fatalf("Failed to get db_pool service: %v", err)
	}

	dbPool, ok := dbPoolService.(*pgxpool.Pool)
	if !ok {
		log.Fatalf("db_pool service is not a *pgxpool.Pool")
	}

	loggerService, err := regCtx.GetService("logger")
	if err != nil {
		log.Fatalf("Failed to get logger service: %v", err)
	}

	// Create repository, service, and handler
	userRepo := repository.NewPostgresUserRepository(dbPool, "users")
	userService := services.NewUserService(userRepo, loggerService.(serviceapi.Logger), true)
	userHandler := handlers.NewUserHandler(userService)

	// Register user management handlers
	regCtx.RegisterHandler("user_management.list_users", userHandler.ListUsers)
	regCtx.RegisterHandler("user_management.get_user", userHandler.GetUser)
	regCtx.RegisterHandler("user_management.create_user", userHandler.CreateUser)
	regCtx.RegisterHandler("user_management.update_user", userHandler.UpdateUser)
	regCtx.RegisterHandler("user_management.delete_user", userHandler.DeleteUser)

	log.Println("User management handlers registered successfully")
}

func newServerFromConfig(regCtx lokstra.RegistrationContext, configDir string) *lokstra.Server {
	if configDir == "" {
		configDir = "."
	}
	cfg, err := lokstra.LoadConfigDir(configDir)
	if err != nil {
		log.Fatalf("Failed to load config from config directory %s: %v", configDir, err)
	}

	server, err := lokstra.NewServerFromConfig(regCtx, cfg)
	if err != nil {
		log.Fatalf("Failed to create server from config: %v", err)
	}

	// Setup user management after services are created
	setupUserManagement(regCtx)

	log.Printf("Config loaded successfully:")
	log.Printf("- Server: %+v", cfg.Server)
	log.Printf("- Apps: %d", len(cfg.Apps))
	log.Printf("- Services: %d", len(cfg.Services))

	return server
}
