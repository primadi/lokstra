package main

import (
	"fmt"

	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/recovery"

	"github.com/primadi/lokstra"
)

func main() {
	// 1. Setup registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// 2. Register all components
	registerAllComponents(regCtx)

	// 3. Create server from config
	server := newServerFromConfig(regCtx, "configs/dev")

	// 4. Start Server
	server.Start(true)
}

func registerAllComponents(regCtx lokstra.RegistrationContext) {
	// Register hardcoded modules, services, middleware, and handlers if needed
	regCtx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Auth check passed")
			return next(ctx)
		}
	})

	// Simulate an admin-only middleware
	regCtx.RegisterMiddlewareFunc("admin_only", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Admin check passed")
			return next(ctx)
		}
	})

	recovery.GetModule().Register(regCtx)
	cors.GetModule().Register(regCtx)

	regCtx.RegisterHandler("user.profile", func(c *lokstra.Context) error {
		return c.Ok("User Profile Handler")
	})

	regCtx.RegisterHandler("user.list", func(c *lokstra.Context) error {
		return c.Ok("User List Handler")
	})
}

func newServerFromConfig(ctx lokstra.RegistrationContext, dir string) *lokstra.Server {
	cfg, err := lokstra.LoadConfigDir(dir)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", dir, err))
	}

	server, err := lokstra.NewServerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create server from config: %v", err))
	}

	fmt.Println("Config loaded successfully:")
	fmt.Printf("- Server: %+v\n", cfg.Server)
	fmt.Printf("- Apps: %d\n", len(cfg.Apps))
	fmt.Printf("- Services: %d\n", len(cfg.Services))
	fmt.Printf("- Modules: %d\n", len(cfg.Modules))

	return server
}
