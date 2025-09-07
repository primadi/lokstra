package main

import (
	"net/http"

	"github.com/primadi/lokstra"
)

// This example demonstrates how to serve static files in Lokstra:
// 1. Mounting a static directory
// 2. Serving files from different directories
// 3. Static files with different prefixes
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "static-files-app", ":8080")

	// Mount static files from a directory
	// This will serve files from ./static directory at /static/* routes
	app.MountStatic("/static", http.Dir("./static"))

	// Mount static files with custom prefix
	// Serve files from ./assets directory at /assets/* routes
	// app.MountStatic("/assets", http.Dir("./assets"))
	app.Group("/assets").MountStatic("/", http.Dir("./assets"))

	// Mount files from public directory
	// Serve files from ./public directory at /public/* routes
	app.MountStatic("/public", http.Dir("./public"))

	// API routes (these should be defined before conflicting static routes)
	app.GET("/api/info", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"app":     "Static Files Example",
			"version": "1.0.0",
			"static_mounts": []map[string]string{
				{"path": "/static", "dir": "./static"},
				{"path": "/assets", "dir": "./assets"},
				{"path": "/public", "dir": "./public"},
			},
		})
	})

	// Health check endpoint
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok("OK")
	})

	// Root endpoint
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Static Files Server",
			"endpoints": []string{
				"/static/* - Files from ./static",
				"/assets/* - Files from ./assets",
				"/public/* - Files from ./public",
				"/api/info - Server information",
				"/health - Health check",
			},
		})
	})

	lokstra.Logger.Infof("Static files server started on :8080")
	lokstra.Logger.Infof("Static file endpoints:")
	lokstra.Logger.Infof("  /static/*      - Files from ./static directory")
	lokstra.Logger.Infof("  /assets/*      - Files from ./assets directory")
	lokstra.Logger.Infof("  /public/*      - Files from ./public directory")
	lokstra.Logger.Infof("  /api/info      - API endpoint (shows mount info)")
	lokstra.Logger.Infof("  /health        - Health check")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try accessing:")
	lokstra.Logger.Infof("  http://localhost:8080/")
	lokstra.Logger.Infof("  http://localhost:8080/static/sample.txt")
	lokstra.Logger.Infof("  http://localhost:8080/assets/style.css")
	lokstra.Logger.Infof("  http://localhost:8080/public/index.html")

	app.Start()
}
