package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Middleware composition example

// Chain combines multiple middleware into one
func Chain(middlewares ...func(*request.Context) error) func(*request.Context) error {
	return func(c *request.Context) error {
		for _, middleware := range middlewares {
			if err := middleware(c); err != nil {
				return err
			}
		}
		return c.Next()
	}
}

// Conditional middleware - executes only if condition is met
func When(condition func(*request.Context) bool, middleware func(*request.Context) error) func(*request.Context) error {
	return func(c *request.Context) error {
		if condition(c) {
			return middleware(c)
		}
		return c.Next()
	}
}

// Unless - opposite of When
func Unless(condition func(*request.Context) bool, middleware func(*request.Context) error) func(*request.Context) error {
	return func(c *request.Context) error {
		if !condition(c) {
			return middleware(c)
		}
		return c.Next()
	}
}

// Logging middleware
func LoggerMiddleware(c *request.Context) error {
	log.Printf("[LOG] %s %s", c.R.Method, c.R.URL.Path)
	return c.Next()
}

// Auth middleware
func AuthMiddleware(c *request.Context) error {
	auth := c.R.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return fmt.Errorf("unauthorized")
	}
	c.Set("authenticated", true)
	return c.Next()
}

// Admin check middleware
func AdminMiddleware(c *request.Context) error {
	auth := c.R.Header.Get("Authorization")
	if !strings.Contains(auth, "admin") {
		return fmt.Errorf("admin access required")
	}
	c.Set("is_admin", true)
	return c.Next()
}

// Timing middleware
func TimingMiddleware(c *request.Context) error {
	start := time.Now()
	err := c.Next()
	duration := time.Since(start)
	c.Set("duration", duration)
	log.Printf("[TIMING] %s took %v", c.R.URL.Path, duration)
	return err
}

// CORS middleware
func CORSMiddleware(c *request.Context) error {
	c.Set("cors", "enabled")
	return c.Next()
}

func main() {
	router := lokstra.NewRouter("middleware-composition")

	// Global middleware chain
	router.Use(LoggerMiddleware)
	router.Use(TimingMiddleware)

	// Public endpoint - no extra middleware
	router.GET("/public", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Public endpoint",
		})
	})

	// Authenticated endpoint - requires auth
	router.GET("/user", func(c *request.Context) any {
		auth := c.Get("authenticated")
		return response.NewApiOk(map[string]any{
			"message":       "User endpoint",
			"authenticated": auth,
		})
	}, AuthMiddleware)

	// Admin endpoint - requires auth + admin
	router.GET("/admin", func(c *request.Context) any {
		isAdmin := c.Get("is_admin")
		return response.NewApiOk(map[string]any{
			"message":  "Admin endpoint",
			"is_admin": isAdmin,
		})
	}, Chain(AuthMiddleware, AdminMiddleware))

	// Conditional middleware - CORS only for /api/* paths
	router.GET("/api/data", func(c *request.Context) any {
		cors := c.Get("cors")
		return response.NewApiOk(map[string]any{
			"message": "API data",
			"cors":    cors,
		})
	}, When(func(c *request.Context) bool {
		return strings.HasPrefix(c.R.URL.Path, "/api/")
	}, CORSMiddleware))

	// Unless example - skip logging for /health
	router.GET("/health", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"status": "healthy",
		})
	}, Unless(func(c *request.Context) bool {
		return c.R.URL.Path == "/health"
	}, LoggerMiddleware))

	// Composed middleware stack
	apiMiddleware := Chain(
		CORSMiddleware,
		AuthMiddleware,
	)

	router.GET("/api/protected", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Protected API endpoint",
		})
	}, apiMiddleware)

	// Home
	router.GET("/", func(c *request.Context) any {
		html := `
		<html>
		<body>
			<h1>Middleware Composition Example</h1>
			<ul>
				<li><a href="/public">Public</a> - No extra middleware</li>
				<li><a href="/user">User</a> - Auth required</li>
				<li><a href="/admin">Admin</a> - Auth + Admin required</li>
				<li><a href="/api/data">API Data</a> - Conditional CORS</li>
				<li><a href="/health">Health</a> - Skips logging</li>
				<li><a href="/api/protected">Protected API</a> - Composed middleware</li>
			</ul>
		</body>
		</html>
		`
		return response.NewHtmlResponse(html)
	})

	log.Println("Server starting on :3001")
	app := lokstra.NewApp("middleware-composition", ":3001", router)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
