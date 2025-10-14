package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/primadi/lokstra"
)

// Request Context - The Foundation
//
// Every Lokstra handler receives *lokstra.RequestContext:
//
//	type RequestContext struct {
//	    context.Context              // Standard Go context
//
//	    // Helpers (use these 95% of the time)
//	    Req  *RequestHelper          // Request data access
//	    Resp *Response               // Response building
//	    Api  *ApiHelper              // API responses
//
//	    // Primitives (advanced usage)
//	    R *http.Request              // Raw HTTP request
//	    W http.ResponseWriter        // Raw response writer
//	}
//
// Helper Layers:
//
// 1. c.Req - Unopinionated request access
//    - QueryParam, PathParam, HeaderParam, RawRequestBody
//    - BindQuery, BindPath, BindHeader, BindBody, BindAll
//    - BindBodyAuto, BindAllAuto
//
// 2. c.Resp - Flexible responses
//    - Json, Html, Text, Stream
//    - Full control over structure
//
// 3. c.Api - Opinionated API responses
//    - Ok, Created, NotFound, BadRequest
//    - Consistent response format
//
// Context Storage:
//    - c.Set(key, value) - Store data
//    - c.Value(key) - Retrieve data
//    - Share data between middleware and handlers
//      Helper : request.GetContextValue[T any](ctx, key, defaultValue T) T
//
// Middleware Chain:
//    - c.Next() - Execute next in chain
//    - Only call in middleware, NOT in handlers
//
// Run: go run .
// Test: See test.http or curl examples below

func main() {
	r := lokstra.NewRouter("context-demo")

	// === 1. c.Req - Request data access ===

	// Query parameters
	r.GET("/api/hello", func(c *lokstra.RequestContext) error {
		name := c.Req.QueryParam("name", "Guest")
		return c.Api.Ok(map[string]string{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Path parameters
	r.GET("/api/user/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")
		return c.Api.Ok(map[string]string{
			"user_id": id,
			"name":    "User " + id,
		})
	})

	// Body binding with validation
	r.POST("/api/create", func(c *lokstra.RequestContext) error {
		type Input struct {
			Name string `json:"name" validate:"required,min=3"`
			Age  int    `json:"age" validate:"required,gte=0,lte=120"`
		}

		var input Input
		if err := c.Req.BindBody(&input); err != nil {
			return c.Api.BadRequest("BIND_ERROR", err.Error())
		}

		return c.Api.Created(map[string]any{
			"id":   "new-123",
			"name": input.Name,
			"age":  input.Age,
		}, "User created")
	})

	// Smart binding - combines path + query + header + body
	r.PUT("/api/update/:id", func(c *lokstra.RequestContext) error {
		type Input struct {
			ID    string `path:"id"`                           // from URL path
			Page  int    `query:"page"`                        // from query string
			Token string `header:"Authorization"`              // from headers
			Name  string `json:"name" validate:"required"`     // from body
			Age   int    `json:"age" validate:"gte=0,lte=120"` // from body
		}

		var input Input
		if err := c.Req.BindAll(&input); err != nil {
			return c.Api.BadRequest("BIND_ERROR", err.Error())
		}

		return c.Api.Ok(map[string]any{
			"id":    input.ID,
			"page":  input.Page,
			"token": input.Token,
			"name":  input.Name,
			"age":   input.Age,
		})
	})

	// === 2. c.Resp - Flexible responses (non-API) ===

	r.GET("/", func(c *lokstra.RequestContext) error {
		return c.Resp.Json(map[string]any{
			"service": "lokstra-demo",
			"version": "1.0",
			"time":    time.Now(),
		})
	})

	r.GET("/admin", func(c *lokstra.RequestContext) error {
		html := `<!DOCTYPE html>
<html><head><title>Admin</title></head>
<body><h1>Admin Panel</h1></body></html>`
		return c.Resp.Html(html)
	})

	// === 3. c.Api - API responses with consistent format ===

	r.GET("/api/users", func(c *lokstra.RequestContext) error {
		users := []map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		}
		return c.Api.Ok(users)
	})

	r.GET("/api/error", func(c *lokstra.RequestContext) error {
		errorType := c.Req.QueryParam("type", "notfound")

		switch errorType {
		case "badrequest":
			return c.Api.BadRequest("INVALID_INPUT", "Missing required field")
		case "unauthorized":
			return c.Api.Unauthorized("Invalid token")
		case "forbidden":
			return c.Api.Forbidden("Access denied")
		case "notfound":
			return c.Api.NotFound("Resource not found")
		default:
			return c.Api.InternalError("Something went wrong")
		}
	})

	// === 4. Middleware + Context Storage ===

	// Middleware: authenticate and store user data
	authMiddleware := func(c *lokstra.RequestContext) error {
		token := c.Req.HeaderParam("Authorization", "")
		if token == "" {
			return c.Api.Unauthorized("Missing token")
		}

		// Store authenticated user info
		c.Set("user_id", "user-123")
		c.Set("username", "john")
		c.Set("role", "admin")

		return c.Next() // Continue to next middleware or handler
	}

	// Handler: read user data from context
	profileHandler := func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"user_id":  c.Get("user_id"),
			"username": c.Get("username"),
			"role":     c.Get("role"),
		})
	}

	// Apply: handler first, then middleware(s)
	r.GET("/api/profile", profileHandler, authMiddleware)

	// === 5. Middleware Chain ===

	// Logging middleware
	loggingMiddleware := func(c *lokstra.RequestContext) error {
		start := time.Now()
		log.Printf("→ %s %s", c.R.Method, c.R.URL.Path)

		err := c.Next() // Execute next in chain

		log.Printf("← %s %s (%v)", c.R.Method, c.R.URL.Path, time.Since(start))
		return err
	}

	// Handler first, middleware(s) after
	r.GET("/api/slow", func(c *lokstra.RequestContext) error {
		time.Sleep(100 * time.Millisecond)
		return c.Api.Ok(map[string]string{"status": "done"})
	}, loggingMiddleware)

	// === 6. c.R and c.W - Direct HTTP access ===

	r.GET("/api/info", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]string{
			"method":     c.R.Method,
			"path":       c.R.URL.Path,
			"remote_ip":  c.R.RemoteAddr,
			"user_agent": c.R.Header.Get("User-Agent"),
		})
	})

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	fmt.Println("\nTest endpoints:")
	fmt.Println("  curl http://localhost:8080/api/hello?name=John")
	fmt.Println("  curl http://localhost:8080/api/user/123")
	fmt.Println("  curl -X POST http://localhost:8080/api/create -H 'Content-Type: application/json' -d '{\"name\":\"Alice\",\"age\":25}'")
	fmt.Println("  curl -X PUT http://localhost:8080/api/update/99?page=2 -H 'Authorization: token123' -H 'Content-Type: application/json' -d '{\"name\":\"Bob\",\"age\":30}'")
	fmt.Println("  curl http://localhost:8080/api/profile -H 'Authorization: token123'")
	fmt.Println("  curl http://localhost:8080/api/error?type=unauthorized")
	fmt.Println("  curl http://localhost:8080/api/slow")
	fmt.Println()

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
