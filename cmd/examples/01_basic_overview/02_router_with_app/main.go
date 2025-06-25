// Lokstra Example 02: Router With App
// ------------------------------------
// This example demonstrates the use of `App` in Lokstra.
// The `App` wraps a router and defines port and optional middleware.
// This is the recommended structure for building actual services.
//
// Run this file using:
//   go run main.go
// Then access:
//   http://localhost:8080/ping

package main

import (
	"lokstra/core"
)

func main() {
	// Step 1: Create a new App named "default" on port 8080
	app := core.NewApp("default", 8080)

	// Step 2: Register a simple GET route to respond with JSON
	app.GET("/ping", func(ctx *core.RequestContext) error {
		return ctx.WithMessage("pong").Ok(nil)
	})

	// Step 3: Start the App
	// Internally, it wraps the router and starts the HTTP server
	println("Server is running at http://localhost:8080")
	app.Start()
}
