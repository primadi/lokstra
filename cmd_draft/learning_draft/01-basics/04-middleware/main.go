package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
)

// Middleware in Lokstra
//
// Middleware is a function that wraps handlers to add cross-cutting concerns:
// - Logging
// - Authentication
// - Rate limiting
// - Recovery from panics
// - Request/response modification
//
// Key Concepts:
// 1. Middleware signature: func(c *lokstra.RequestContext) error
// 2. Call c.Next() to continue to next middleware/handler
// 3. Return early to stop the chain
// 4. Code after c.Next() runs in reverse order (like onion layers)
// 5. Apply middleware: r.GET("/path", handler, middleware1, middleware2, ...)
//    Order: handler FIRST, then middleware(s)
//
// Run: go run .

func main() {
	r := lokstra.NewRouter("middleware-demo")

	// === 1. LOGGING MIDDLEWARE ===

	loggingMiddleware := func(c *lokstra.RequestContext) error {
		start := time.Now()
		method := c.R.Method
		path := c.R.URL.Path

		log.Printf("→ %s %s", method, path)

		err := c.Next() // Continue to next middleware or handler

		duration := time.Since(start)
		log.Printf("← %s %s [%v]", method, path, duration)

		return err
	}

	// === 2. AUTH MIDDLEWARE ===

	authMiddleware := func(c *lokstra.RequestContext) error {
		token := c.Req.HeaderParam("Authorization", "")

		// Check if token exists
		if token == "" {
			return c.Api.Unauthorized("Missing Authorization header")
		}

		// Simple validation (in real app, verify JWT/token)
		if !strings.HasPrefix(token, "Bearer ") {
			return c.Api.Unauthorized("Invalid token format")
		}

		// Extract and "validate" token
		tokenValue := strings.TrimPrefix(token, "Bearer ")
		if tokenValue == "invalid" {
			return c.Api.Unauthorized("Invalid token")
		}

		// Store user info in context
		c.Set("user_id", "user-123")
		c.Set("username", "john")
		c.Set("token", tokenValue)

		return c.Next()
	}

	// === 3. ROLE-BASED ACCESS MIDDLEWARE ===

	adminOnlyMiddleware := func(c *lokstra.RequestContext) error {
		role := c.Get("role")

		if role != "admin" {
			return c.Api.Forbidden("Admin access required")
		}

		return c.Next()
	}

	// Helper to set role (simulating auth that sets role)
	roleMiddleware := func(role string) func(*lokstra.RequestContext) error {
		return func(c *lokstra.RequestContext) error {
			c.Set("role", role)
			return c.Next()
		}
	}

	// === 4. RECOVERY MIDDLEWARE ===

	recoveryMiddleware := func(c *lokstra.RequestContext) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ PANIC RECOVERED: %v", r)
				// Write error response
				c.Api.InternalError("Internal server error")
			}
		}()

		return c.Next()
	}

	// === 5. RATE LIMITING MIDDLEWARE (Simplified) ===

	type rateLimiter struct {
		requests map[string][]time.Time
		limit    int
		window   time.Duration
	}

	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    5,
		window:   time.Minute,
	}

	rateLimitMiddleware := func(c *lokstra.RequestContext) error {
		ip := utils.ClientIP(c.R)
		now := time.Now()

		// Clean old requests
		requests := limiter.requests[ip]
		var validRequests []time.Time
		for _, t := range requests {
			if now.Sub(t) < limiter.window {
				validRequests = append(validRequests, t)
			}
		}

		// Check limit
		if len(validRequests) >= limiter.limit {
			return c.Api.Error(429, "RATE_LIMIT_EXCEEDED",
				fmt.Sprintf("Rate limit exceeded. Max %d requests per minute", limiter.limit))
		}

		// Add current request
		validRequests = append(validRequests, now)
		limiter.requests[ip] = validRequests

		return c.Next()
	}

	// === 6. CORS MIDDLEWARE ===

	corsMiddleware := func(c *lokstra.RequestContext) error {
		// Set CORS headers
		c.W.Header().Set("Access-Control-Allow-Origin", "*")
		c.W.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.W.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if c.R.Method == "OPTIONS" {
			c.W.WriteHeader(http.StatusOK)
			return nil
		}

		return c.Next()
	}

	// === EXAMPLE ROUTES ===

	// Public endpoint - no middleware
	r.GET("/", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"service": "middleware-demo",
			"endpoints": []string{
				"GET /",
				"GET /api/public",
				"GET /api/protected",
				"GET /api/admin",
				"GET /api/limited",
				"GET /api/panic",
				"GET /api/slow",
			},
		})
	})

	// Public with logging only
	r.GET("/api/public",
		func(c *lokstra.RequestContext) error {
			return c.Api.Ok(map[string]string{
				"message": "This is a public endpoint",
			})
		},
		loggingMiddleware,
	)

	// Protected endpoint - requires auth
	r.GET("/api/protected",
		func(c *lokstra.RequestContext) error {
			return c.Api.Ok(map[string]any{
				"message":  "Protected data",
				"user_id":  c.Get("user_id"),
				"username": c.Get("username"),
			})
		},
		loggingMiddleware,
		authMiddleware,
	)

	// Admin only - requires auth + admin role
	r.GET("/api/admin",
		func(c *lokstra.RequestContext) error {
			return c.Api.Ok(map[string]any{
				"message": "Admin panel",
				"user":    c.Get("username"),
			})
		},
		loggingMiddleware,
		authMiddleware,
		roleMiddleware("admin"),
		adminOnlyMiddleware,
	)

	// Rate limited endpoint
	r.GET("/api/limited",
		func(c *lokstra.RequestContext) error {
			return c.Api.Ok(map[string]any{
				"message": "This endpoint is rate limited",
				"limit":   "5 requests per minute",
			})
		},
		loggingMiddleware,
		rateLimitMiddleware,
	)

	// Endpoint that panics (recovered by middleware)
	r.GET("/api/panic",
		func(c *lokstra.RequestContext) error {
			shouldPanic := c.Req.QueryParam("panic", "false")
			if shouldPanic == "true" {
				panic("Intentional panic for testing recovery!")
			}
			return c.Api.Ok(map[string]string{"message": "No panic"})
		},
		loggingMiddleware,
		recoveryMiddleware,
	)

	// Slow endpoint to see logging with duration
	r.GET("/api/slow",
		func(c *lokstra.RequestContext) error {
			duration := c.Req.QueryParam("duration", "100")
			ms, _ := time.ParseDuration(duration + "ms")
			time.Sleep(ms)
			return c.Api.Ok(map[string]any{
				"message": "Completed",
				"slept":   ms.String(),
			})
		},
		loggingMiddleware,
	)

	// CORS example
	r.GET("/api/cors",
		func(c *lokstra.RequestContext) error {
			return c.Api.Ok(map[string]string{
				"message": "CORS headers are set",
			})
		},
		corsMiddleware,
		loggingMiddleware,
	)

	// Multiple middleware chaining
	r.POST("/api/secure",
		func(c *lokstra.RequestContext) error {
			type Input struct {
				Data string `json:"data" validate:"required"`
			}
			var input Input
			if err := c.Req.BindBody(&input); err != nil {
				return c.Api.BadRequest("INVALID_INPUT", err.Error())
			}
			return c.Api.Created(map[string]any{
				"received": input.Data,
				"user":     c.Get("username"),
			}, "Data saved")
		},
		recoveryMiddleware,  // Outermost - catches all panics
		loggingMiddleware,   // Logs request/response
		corsMiddleware,      // CORS headers
		rateLimitMiddleware, // Rate limiting
		authMiddleware,      // Authentication
	)

	// Start server
	fmt.Println("🚀 Middleware Demo Server")
	fmt.Println("=========================")
	fmt.Println("\n📖 Middleware Execution Order:")
	fmt.Println("   Request → MW1 → MW2 → MW3 → Handler → MW3 → MW2 → MW1 → Response")
	fmt.Println("   (Like onion layers - last middleware in, first out)")
	fmt.Println("\n🔧 Available Endpoints:")
	fmt.Println("   GET  /api/public")
	fmt.Println("   GET  /api/protected  (requires: Authorization: Bearer <token>)")
	fmt.Println("   GET  /api/admin      (requires: auth + admin role)")
	fmt.Println("   GET  /api/limited    (max 5 requests/minute)")
	fmt.Println("   GET  /api/panic?panic=true")
	fmt.Println("   GET  /api/slow?duration=500")
	fmt.Println("   POST /api/secure     (all middleware stacked)")
	fmt.Println("\n💡 Test Commands:")
	fmt.Println("   # Public")
	fmt.Println("   curl http://localhost:8080/api/public")
	fmt.Println()
	fmt.Println("   # Protected (no token - should fail)")
	fmt.Println("   curl http://localhost:8080/api/protected")
	fmt.Println()
	fmt.Println("   # Protected (with token - should work)")
	fmt.Println("   curl http://localhost:8080/api/protected -H 'Authorization: Bearer mytoken'")
	fmt.Println()
	fmt.Println("   # Rate limited (try 6+ times)")
	fmt.Println("   curl http://localhost:8080/api/limited")
	fmt.Println()
	fmt.Println("   # Panic recovery")
	fmt.Println("   curl http://localhost:8080/api/panic?panic=true")
	fmt.Println()
	fmt.Println("\n🚀 Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
