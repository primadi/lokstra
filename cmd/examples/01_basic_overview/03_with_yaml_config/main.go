package main

import (
	"fmt"

	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/recovery"

	"github.com/primadi/lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	registerAllComponents(ctx)

	server := newServerFormConfig(ctx, "configs/dev")
	server.Start()
}

func registerAllComponents(ctx *lokstra.GlobalContext) {
	// Register hardcoded modules, services, middleware, and handlers if needed
	ctx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Auth check passed")
			return next(ctx)
		}
	})

	// Simulate an admin-only middleware
	ctx.RegisterMiddlewareFunc("admin_only", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Admin check passed")
			return next(ctx)
		}
	})

	// tips: to know middlware name, hover over .GetModule() function below
	ctx.RegisterMiddlewareModule(recovery.GetModule())
	ctx.RegisterMiddlewareModule(cors.GetModule())

	ctx.RegisterHandler("user.profile", func(c *lokstra.Context) error {
		return c.Ok("User Profile Handler")
	})

	ctx.RegisterHandler("user.list", func(c *lokstra.Context) error {
		return c.Ok("User List Handler")
	})
}

func newServerFormConfig(ctx *lokstra.GlobalContext, dir string) *lokstra.Server {
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
