package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates the simplest possible Lokstra application.
// It shows how to create an app, add a route, and start the server.
//
// Learning Objectives:
// - Understand basic Lokstra app structure
// - Learn how to create a registration context
// - Add simple routes with handlers
// - Start the application
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/getting-started.md
func main() {
	// Create global registration context for dependency injection
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Create a new application
	// Parameters: context, app name, address
	app := lokstra.NewApp(regCtx, "minimal-app", ":8080")

	// Add a simple GET route
	app.GET("/hello", func(ctx *lokstra.Context) error {
		return ctx.Ok("Hello, Lokstra!")
	})

	// Add a route with path parameter
	app.GET("/hello/:name", func(ctx *lokstra.Context) error {
		name := ctx.GetPathParam("name")
		return ctx.Ok("Hello, " + name + "!")
	})

	lokstra.Logger.Infof("Minimal Lokstra Application started on :8080")

	// Start the application (blocks until shutdown)
	app.Start()
}

// Test this example:
// 1. Run: go run main.go
// 2. Test: curl http://localhost:8080/hello
// 3. Test: curl http://localhost:8080/hello/world
//
// Expected responses:
// - GET /hello -> {"success": true, "message": "OK", "data": "Hello, Lokstra!"}
// - GET /hello/world -> {"success": true, "message": "OK", "data": "Hello, world!"}
