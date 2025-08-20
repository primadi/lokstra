package main

import (
	"fmt"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/services/dbpool_pg"
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
	regCtx.RegisterModule(dbpool_pg.GetModule)

	// For now, register simple handlers to test the setup
	regCtx.RegisterHandler("user.create", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"message": "User creation endpoint - implementation in progress",
			"status":  "success",
		})
	})

	regCtx.RegisterHandler("user.list", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"users":   []any{},
			"total":   0,
			"message": "User list endpoint - implementation in progress",
		})
	})

	regCtx.RegisterHandler("user.get", func(c *lokstra.Context) error {
		username := c.GetQueryParam("username")
		return c.Ok(map[string]any{
			"username": username,
			"message":  "User get endpoint - implementation in progress",
		})
	})

	regCtx.RegisterHandler("user.update", func(c *lokstra.Context) error {
		username := c.GetQueryParam("username")
		return c.Ok(map[string]any{
			"username": username,
			"message":  "User update endpoint - implementation in progress",
		})
	})

	regCtx.RegisterHandler("user.delete", func(c *lokstra.Context) error {
		username := c.GetQueryParam("username")
		return c.Ok(map[string]any{
			"username": username,
			"message":  "User delete endpoint - implementation in progress",
		})
	})

	regCtx.RegisterHandler("health.check", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"status":  "ok",
			"service": "user-management",
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
