package main

import (
	"embed"
	"net/http"

	"github.com/primadi/lokstra"
)

//go:embed framework_assets/*
var frameworkAssets embed.FS

// This example demonstrates how to serve static files with fallback in Lokstra:
// 1. Serving files from multiple sources with fallback priority
// 2. Project assets override framework assets
// 3. Embed.FS fallback for bundled assets
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "static-fallback-app", ":8080")

	// Create sub-filesystem for framework_assets
	// frameworkSubFS, err := fs.Sub(frameworkAssets, "framework_assets")
	// if err != nil {
	// 	panic(err)
	// }

	// Serve static files with fallback:
	// 1. First try project/assets (project-specific overrides)
	// 2. Then try project/static (main project assets)
	// 3. Finally try frameworkAssets embed.FS (framework default assets)
	app.MountStaticWithFallback("/static",
		http.Dir("./project/assets"), // Project overrides (highest priority)
		http.Dir("./project/static"), // Project assets
		frameworkAssets,              // Framework default assets (lowest priority)
	)

	// Alternative: Using string paths instead of http.Dir
	app.MountStaticWithFallback("/assets",
		"./project/assets",   // Project overrides
		"./framework/assets", // Framework assets
		frameworkAssets,      // Embedded fallback
	)

	// API routes
	app.GET("/api/info", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"app":     "Static Files with Fallback Example",
			"version": "1.0.0",
			"static_mounts": []map[string]any{
				{
					"path": "/static",
					"sources": []string{
						"./project/assets (highest priority)",
						"./project/static",
						"frameworkAssets embed.FS (lowest priority)",
					},
				},
				{
					"path": "/assets",
					"sources": []string{
						"./project/assets (highest priority)",
						"./framework/assets",
						"frameworkAssets embed.FS (lowest priority)",
					},
				},
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
			"message":     "Static Files with Fallback Server",
			"description": "Files are served with fallback priority: project overrides → project assets → framework defaults",
			"endpoints": []string{
				"/static/* - Files with fallback from multiple sources",
				"/assets/* - Alternative static files with fallback",
				"/api/info - Server information",
				"/health - Health check",
			},
		})
	})

	// Create sample directories and files
	createSampleFiles()

	lokstra.Logger.Infof("Static files with fallback server started on :8080")
	lokstra.Logger.Infof("Static file endpoints with fallback:")
	lokstra.Logger.Infof("  /static/*      - Fallback: ./project/assets → ./project/static → embed.FS")
	lokstra.Logger.Infof("  /assets/*      - Fallback: ./project/assets → ./framework/assets → embed.FS")
	lokstra.Logger.Infof("  /api/info      - API endpoint (shows mount info)")
	lokstra.Logger.Infof("  /health        - Health check")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Fallback logic:")
	lokstra.Logger.Infof("  1. Check project-specific overrides first (highest priority)")
	lokstra.Logger.Infof("  2. Check main project assets")
	lokstra.Logger.Infof("  3. Check framework default assets (lowest priority)")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try accessing:")
	lokstra.Logger.Infof("  http://localhost:8080/")
	lokstra.Logger.Infof("  http://localhost:8080/static/app.js")
	lokstra.Logger.Infof("  http://localhost:8080/static/style.css")
	lokstra.Logger.Infof("  http://localhost:8080/assets/style.css")
	lokstra.Logger.Infof("  http://localhost:8080/static/default.js")

	app.Start()
}

// createSampleFiles creates sample static files for demonstration
func createSampleFiles() {
	lokstra.Logger.Infof("Creating sample directory structure...")

	// Create directories
	dirs := []string{
		"./project/assets",
		"./project/static",
		"./framework/assets",
		"./framework_assets",
	}
	for _, dir := range dirs {
		lokstra.Logger.Infof("Directory %s should exist with sample files", dir)
	}

	lokstra.Logger.Infof("Sample file priorities:")
	lokstra.Logger.Infof("  ./project/assets/app.js      - Project override (will be served first if exists)")
	lokstra.Logger.Infof("  ./project/static/app.js      - Project default (served if override doesn't exist)")
	lokstra.Logger.Infof("  ./framework_assets/app.js    - Framework fallback (embedded, served if others don't exist)")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("To test fallback:")
	lokstra.Logger.Infof("  1. Create ./project/assets/custom.js → will be served from project override")
	lokstra.Logger.Infof("  2. Create ./project/static/default.js → will be served from project default")
	lokstra.Logger.Infof("  3. Access non-existent files → will fallback to framework embed.FS or 404")

	lokstra.Logger.Infof("Sample files setup complete")
}
