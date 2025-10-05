package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra"
)

// This example demonstrates the most basic Lokstra concept: the Router.
// A Router is the foundation of Lokstra - it handles HTTP requests and routes them to handlers.
//
// Key Concepts:
// 1. Create a router with lokstra.NewRouter(name)
// 2. Define routes with HTTP methods: GET, POST, PUT, DELETE, etc.
// 3. Use http.ListenAndServe to start the server
//
// Run: go run main.go
// Test: curl http://localhost:8080/hello

func main() {
	// Step 1: Create a new router with a name
	// The name is useful for debugging and logging
	router := lokstra.NewRouter("my-first-router")

	// Step 2: Define routes
	// Routes consist of: HTTP Method + Path + Handler Function

	// Simple GET route
	router.GET("/hello", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("Hello, World!")
	})

	// GET route with path parameter
	router.GET("/hello/:name", func(c *lokstra.RequestContext) error {
		name := c.Req.PathParam("name", "Guest")
		return c.Api.Ok(fmt.Sprintf("Hello, %s!", name))
	})

	// POST route
	router.POST("/greet", func(c *lokstra.RequestContext) error {
		return c.Api.Created(map[string]string{
			"message": "Greeting created!",
		}, "Success")
	})

	// PUT route
	router.PUT("/update/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("Updated item with ID: %s", id))
	})

	// DELETE route
	router.DELETE("/delete/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("Deleted item with ID: %s", id))
	})

	// Step 3: Print registered routes (helpful for debugging)
	fmt.Println("📋 Registered Routes:")
	router.PrintRoutes()

	// Step 4: Start the HTTP server
	// The router implements http.Handler, so it can be used directly with http.ListenAndServe
	fmt.Println("\n🚀 Server starting on http://localhost:8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  - GET  http://localhost:8080/hello")
	fmt.Println("  - GET  http://localhost:8080/hello/John")
	fmt.Println("  - POST http://localhost:8080/greet")
	fmt.Println("  - PUT  http://localhost:8080/update/123")
	fmt.Println("  - DELETE http://localhost:8080/delete/123")

	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
