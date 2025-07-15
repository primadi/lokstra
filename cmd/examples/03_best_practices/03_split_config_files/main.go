package main

import (
	"fmt"
	"lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	registerComponents(ctx)

	server := newServerFromConfig(ctx, "configs/production")
	server.Start()
}

func registerComponents(ctx *lokstra.GlobalContext) {
	ctx.RegisterHandler("api.health", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"status": "healthy",
			"time":   "2025-01-15T09:00:00Z",
		})
	})

	ctx.RegisterHandler("api.users.list", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"users": []map[string]any{
				{"id": 1, "name": "John Doe"},
				{"id": 2, "name": "Jane Smith"},
			},
		})
	})

	ctx.RegisterHandler("api.products.list", func(c *lokstra.Context) error {
		return c.Ok(map[string]any{
			"products": []map[string]any{
				{"id": 1, "name": "Product A"},
				{"id": 2, "name": "Product B"},
			},
		})
	})

	ctx.RegisterMiddlewareFunc("api_auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			apiKey := ctx.Headers.Get("X-API-Key")
			if apiKey != "secret-api-key" {
				return ctx.ErrorUnauthorized("Invalid API key")
			}
			return next(ctx)
		}
	})
}

func newServerFromConfig(ctx *lokstra.GlobalContext, dir string) *lokstra.Server {
	cfg, err := lokstra.LoadConfigDir(dir)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", dir, err))
	}

	server, err := lokstra.NewServerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create server from config: %v", err))
	}

	fmt.Println("Split config loaded successfully:")
	fmt.Printf("- Server: %+v\n", cfg.Server)
	fmt.Printf("- Apps: %d\n", len(cfg.Apps))
	fmt.Printf("- Services: %d\n", len(cfg.Services))
	fmt.Printf("- Modules: %d\n", len(cfg.Modules))

	return server
}
