// Lokstra Example 01: Minimal Router Only
// ----------------------------------------
// This example demonstrates the simplest usage of Lokstra: using only the Router.
// No App, Server, or Service is required. Suitable for lightweight routing scenarios.
//
// Run this file using:
//   go run main.go
// Then access:
//   http://localhost:8080/hello

package main

import (
	"lokstra/core"
	"net/http"
)

func main() {
	// Step 1: Create a new router instance using default engine (e.g., HttpRouter).
	router := core.NewRouter()

	// Step 2: Register a basic GET route using Lokstra RequestContext.
	router.GET("/hello", func(ctx *core.RequestContext) error {
		return ctx.WithMessage("Hello, World from Lokstra Router!").Ok(nil)
	})

	// Step 3: Start the HTTP server with the router's handler.
	println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
