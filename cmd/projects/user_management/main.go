package main

import (
	"fmt"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/cmd/projects/user_management/handlers"
	"github.com/primadi/lokstra/defaults"
)

func main() {
	// 1. Setup registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// 2. Register handlers
	registerComponents(regCtx)

	// 3. Load configuration
	configPath := "config/"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 4. Create server from config
	server := newServerFromConfig(regCtx, configPath)

	// 5. Start server and wait for shutdown signal
	server.StartAndWaitForShutdown(5 * time.Second)
}

func registerComponents(regCtx lokstra.RegistrationContext) {
	// Register Services
	defaults.RegisterAll(regCtx)

	// Register actual CRUD handlers
	regCtx.RegisterHandler("user.create", handlers.CreateNewUserHandler())
	regCtx.RegisterHandler("user.list", handlers.CreateListUserHandler())
	regCtx.RegisterHandler("user.get", handlers.CreateGetUserByIDHandler()) // GET by ID
	regCtx.RegisterHandler("user.update", handlers.CreateUpdateUserHandler())
	regCtx.RegisterHandler("user.delete", handlers.CreateDeleteUserHandler())
	regCtx.RegisterHandler("user.get_by_name", handlers.CreateGetUserByNameHandler())

	// Register Unified UI handlers (automatically handle both full page and HTMX partial requests with consistent behavior)
	regCtx.RegisterHandler("ui.dashboard", handlers.CreateDashboardHandler())
	regCtx.RegisterHandler("ui.users", handlers.CreateUsersHandler())
	regCtx.RegisterHandler("ui.user_form", handlers.CreateUserFormPageHandler())
	regCtx.RegisterHandler("ui.user_create", handlers.CreateUserSubmitHandler())
	regCtx.RegisterHandler("ui.user_update", handlers.UpdateUserSubmitHandler())
	regCtx.RegisterHandler("ui.roles", handlers.CreateRolesHandler())
	regCtx.RegisterHandler("ui.settings", handlers.CreateSettingsHandler())

	regCtx.RegisterHandler("auth.login", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"message": "Login endpoint - implementation in progress",
			"status":  "success",
		})
	})

	regCtx.RegisterHandler("auth.logout", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"message": "Logout endpoint - implementation in progress",
			"status":  "success",
		})
	})

	regCtx.RegisterHandler("auth.refresh", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"message": "Refresh token endpoint - implementation in progress",
			"status":  "success",
		})
	})

	regCtx.RegisterHandler("admin.user_stats", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"total_users":    0,
			"active_users":   0,
			"inactive_users": 0,
			"message":        "Admin user stats endpoint - implementation in progress",
		})
	})

	regCtx.RegisterHandler("admin.activate_user", func(c *lokstra.Context) error {
		userID := c.GetPathParam("id")
		return c.Ok(map[string]any{
			"user_id": userID,
			"message": "User activated successfully",
			"status":  "active",
		})
	})

	regCtx.RegisterHandler("admin.deactivate_user", func(c *lokstra.Context) error {
		userID := c.GetPathParam("id")
		return c.Ok(map[string]any{
			"user_id": userID,
			"message": "User deactivated successfully",
			"status":  "inactive",
		})
	})

	regCtx.RegisterHandler("health.check", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"status":    "ok",
			"service":   "user-management",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})
}

func newServerFromConfig(regCtx lokstra.RegistrationContext, configPath string) *lokstra.Server {
	cfg, err := lokstra.LoadConfigDir(configPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", configPath, err))
	}

	server, err := lokstra.NewServerFromConfig(regCtx, cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create server from config: %v", err))
	}

	fmt.Println("Config loaded successfully")
	return server
}
