package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates how to serve Single Page Applications (SPA) in Lokstra:
// 1. Mounting SPA with fallback handling
// 2. API routes coexisting with SPA routes
// 3. Proper routing for client-side routing applications
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "spa-mount-app", ":8080")

	// API routes - these should be defined BEFORE SPA mounting
	// to ensure API endpoints take precedence over SPA fallback
	apiGroup := app.Group("/api")

	apiGroup.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok([]map[string]any{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		})
	})

	apiGroup.GET("/users/:id", func(ctx *lokstra.Context) error {
		id := ctx.GetPathParam("id")
		return ctx.Ok(map[string]any{
			"id":    id,
			"name":  "User " + id,
			"email": "user" + id + "@example.com",
		})
	})

	apiGroup.POST("/users", func(ctx *lokstra.Context) error {
		return ctx.OkCreated(map[string]any{
			"message": "User created successfully",
			"id":      123,
		})
	})

	// Health check endpoint
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok("SPA Server is healthy")
	})

	// Server info endpoint
	app.GET("/server-info", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"app":         "SPA Mount Example",
			"version":     "1.0.0",
			"spa_path":    "/",
			"fallback":    "./spa/index.html",
			"api_prefix":  "/api",
			"description": "Single Page Application with API backend",
		})
	})

	// Mount SPA - this should be the LAST route definition
	// All unmatched routes will fallback to serving the SPA's index.html
	// This enables client-side routing to work properly
	app.MountSPA("/", "./spa/index.html")

	// Create sample SPA files
	createSampleSPA()

	lokstra.Logger.Infof("SPA server started on :8080")
	lokstra.Logger.Infof("SPA Configuration:")
	lokstra.Logger.Infof("  SPA Mount:     / (fallback to ./spa/index.html)")
	lokstra.Logger.Infof("  API Prefix:    /api/*")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Available endpoints:")
	lokstra.Logger.Infof("  /                    - SPA root (React/Vue/Angular app)")
	lokstra.Logger.Infof("  /about               - SPA route (handled by client)")
	lokstra.Logger.Infof("  /dashboard           - SPA route (handled by client)")
	lokstra.Logger.Infof("  /users/123           - SPA route (handled by client)")
	lokstra.Logger.Infof("  /api/users           - API endpoint (server-side)")
	lokstra.Logger.Infof("  /api/users/123       - API endpoint (server-side)")
	lokstra.Logger.Infof("  /health              - Health check")
	lokstra.Logger.Infof("  /server-info         - Server information")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Visit http://localhost:8080 to see the SPA")

	app.Start()
}

// createSampleSPA creates a sample SPA structure
func createSampleSPA() {
	lokstra.Logger.Infof("Creating sample SPA files...")
	lokstra.Logger.Infof("SPA directory ./spa should exist with:")
	lokstra.Logger.Infof("  - index.html (main SPA entry point)")
	lokstra.Logger.Infof("  - static assets (CSS, JS, images)")
	lokstra.Logger.Infof("Sample SPA files will be created...")
}
