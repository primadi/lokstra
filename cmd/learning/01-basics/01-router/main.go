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
// 3. Exact match vs Prefix match: GET vs GETPrefix
// 4. Path parameter syntax: :param or {param}
// 5. Use http.ListenAndServe to start the server
//
// Run: go run main.go
// Test: curl http://localhost:8080/hello

func main() {
	// Step 1: Create a new router with a name
	// The name is useful for debugging and logging
	router := lokstra.NewRouter("my-first-router")

	// Step 2: Define routes
	// Routes consist of: HTTP Method + Path + Handler Function

	// ============================================
	// Basic Routes
	// ============================================

	// Simple GET route
	router.GET("/hello", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("Hello, World!")
	})

	// ============================================
	// Path Parameters - Two Syntax Styles
	// ============================================

	// Style 1: Colon syntax :name
	router.GET("/hello/:name", func(c *lokstra.RequestContext) error {
		name := c.Req.PathParam("name", "Guest")
		return c.Api.Ok(fmt.Sprintf("Hello, %s! (using :name syntax)", name))
	})

	// Style 2: Curly braces syntax {id}
	router.GET("/user/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("User ID: %s (using {id} syntax)", id))
	})

	// Multiple parameters in one route
	router.GET("/user/{userId}/post/{postId}", func(c *lokstra.RequestContext) error {
		userId := c.Req.PathParam("userId", "0")
		postId := c.Req.PathParam("postId", "0")
		return c.Api.Ok(fmt.Sprintf("User %s, Post %s", userId, postId))
	})

	// ============================================
	// Exact Match vs Prefix Match
	// ============================================

	// GET: Exact match only
	// Will match: /api/users
	// Will NOT match: /api/users/123 or /api/users/search
	router.GET("/api/users", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("GET /api/users - EXACT match")
	})

	// GETPrefix: Matches path and all sub-paths
	// Will match: /api/products, /api/products/123, /api/products/search, etc.
	router.GETPrefix("/api/products", func(c *lokstra.RequestContext) error {
		path := c.R.URL.Path
		return c.Api.Ok(fmt.Sprintf("GETPrefix /api/products - matched: %s", path))
	})

	// ============================================
	// POST Methods: Exact vs Prefix
	// ============================================

	// POST: Exact match
	router.POST("/greet", func(c *lokstra.RequestContext) error {
		return c.Api.Created(map[string]string{
			"message": "Greeting created!",
		}, "Success")
	})

	// POSTPrefix: Match all sub-paths
	router.POSTPrefix("/api/submit", func(c *lokstra.RequestContext) error {
		path := c.R.URL.Path
		return c.Api.Created(map[string]string{
			"message": "Created via POSTPrefix",
			"path":    path,
		}, "Success")
	})

	// ============================================
	// PUT Methods: Exact vs Prefix
	// ============================================

	// PUT: Exact match
	router.PUT("/update/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("PUT exact: Updated item %s", id))
	})

	// PUTPrefix: Match all sub-paths
	router.PUTPrefix("/api/update", func(c *lokstra.RequestContext) error {
		path := c.R.URL.Path
		return c.Api.Ok(fmt.Sprintf("PUTPrefix: Updated via %s", path))
	})

	// ============================================
	// DELETE Methods: Exact vs Prefix
	// ============================================

	// DELETE: Exact match
	router.DELETE("/delete/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("DELETE exact: Deleted item %s", id))
	})

	// DELETEPrefix: Match all sub-paths
	router.DELETEPrefix("/api/remove", func(c *lokstra.RequestContext) error {
		path := c.R.URL.Path
		return c.Api.Ok(fmt.Sprintf("DELETEPrefix: Removed via %s", path))
	})

	// ============================================
	// PATCH Method (also has Prefix variant)
	// ============================================

	router.PATCH("/patch/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(fmt.Sprintf("PATCH exact: Patched item %s", id))
	})

	router.PATCHPrefix("/api/patch", func(c *lokstra.RequestContext) error {
		path := c.R.URL.Path
		return c.Api.Ok(fmt.Sprintf("PATCHPrefix: Patched via %s", path))
	})

	// ============================================
	// ANY Method - Matches ALL HTTP Methods
	// ============================================

	// ANY: Exact match for all HTTP methods
	// Will handle GET, POST, PUT, DELETE, PATCH, etc.
	router.ANY("/api/flexible", func(c *lokstra.RequestContext) error {
		method := c.R.Method
		return c.Api.Ok(fmt.Sprintf("ANY exact: Handled %s request", method))
	})

	// ANYPrefix: Matches all HTTP methods and all sub-paths
	// Will handle any method on /api/wildcard, /api/wildcard/anything, etc.
	router.ANYPrefix("/api/wildcard", func(c *lokstra.RequestContext) error {
		method := c.R.Method
		path := c.R.URL.Path
		return c.Api.Ok(fmt.Sprintf("ANYPrefix: Handled %s request on %s", method, path))
	})

	// Step 3: Print registered routes (helpful for debugging)
	fmt.Println("📋 Registered Routes:")
	router.PrintRoutes()

	// Step 4: Start the HTTP server
	// The router implements http.Handler, so it can be used directly with http.ListenAndServe
	fmt.Println("\n🚀 Server starting on http://localhost:8080")
	fmt.Println("\n📖 Endpoint Examples:")
	fmt.Println("\n  Basic Routes:")
	fmt.Println("    - GET  http://localhost:8080/hello")
	fmt.Println("\n  Path Parameters (two syntax styles):")
	fmt.Println("    - GET  http://localhost:8080/hello/John           (:name syntax)")
	fmt.Println("    - GET  http://localhost:8080/user/123             ({id} syntax)")
	fmt.Println("    - GET  http://localhost:8080/user/10/post/20      (multiple params)")
	fmt.Println("\n  Exact Match vs Prefix Match:")
	fmt.Println("    - GET  http://localhost:8080/api/users            (exact only)")
	fmt.Println("    - GET  http://localhost:8080/api/users/123        (won't match)")
	fmt.Println("    - GET  http://localhost:8080/api/products         (prefix - matches)")
	fmt.Println("    - GET  http://localhost:8080/api/products/123     (prefix - matches)")
	fmt.Println("    - GET  http://localhost:8080/api/products/search  (prefix - matches)")
	fmt.Println("\n  Other HTTP Methods:")
	fmt.Println("    - POST   http://localhost:8080/greet")
	fmt.Println("    - POST   http://localhost:8080/api/submit/anything")
	fmt.Println("    - PUT    http://localhost:8080/update/123")
	fmt.Println("    - PUT    http://localhost:8080/api/update/anything")
	fmt.Println("    - DELETE http://localhost:8080/delete/123")
	fmt.Println("    - DELETE http://localhost:8080/api/remove/anything")
	fmt.Println("    - PATCH  http://localhost:8080/patch/123")
	fmt.Println("    - PATCH  http://localhost:8080/api/patch/anything")
	fmt.Println("\n  ANY Method (accepts all HTTP methods):")
	fmt.Println("    - ANY    http://localhost:8080/api/flexible        (try with GET, POST, PUT, etc.)")
	fmt.Println("    - ANY    http://localhost:8080/api/wildcard/any    (try with any method & path)")

	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
