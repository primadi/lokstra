package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// ============================================
// Custom Middleware Examples
// ============================================

// 1. Request Logger Middleware
func RequestLoggerMiddleware(c *request.Context) error {
	start := time.Now()
	method := c.R.Method
	path := c.R.URL.Path

	fmt.Printf("→ [%s] %s\n", method, path)

	// Call next handler
	err := c.Next()

	duration := time.Since(start)
	fmt.Printf("← [%s] %s (%v)\n", method, path, duration)

	return err
}

// 2. Request ID Middleware
func RequestIDMiddleware(c *request.Context) error {
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Store in context
	c.Set("request_id", requestID)

	fmt.Printf("📋 Request ID: %s\n", requestID)

	return c.Next()
}

// 3. Auth Middleware
func AuthMiddleware(c *request.Context) error {
	token := c.R.Header.Get("Authorization")

	if token == "" {
		fmt.Println("✗ Auth failed: Missing token")
		// Return error - framework will handle it
		return fmt.Errorf("missing authorization token")
	}

	// Simple token validation
	if token != "Bearer secret-token" {
		fmt.Println("✗ Auth failed: Invalid token")
		return fmt.Errorf("invalid authorization token")
	}

	// Store user info in context
	c.Set("user_id", 123)
	c.Set("username", "john_doe")

	fmt.Println("✓ Auth successful")

	return c.Next()
}

// 4. Rate Limiter Middleware (simple in-memory)
func RateLimiterMiddleware(maxRequests int, window time.Duration) func(*request.Context) error {
	requests := make(map[string][]time.Time)

	return func(c *request.Context) error {
		ip := c.R.RemoteAddr

		now := time.Now()

		// Clean old entries
		if times, exists := requests[ip]; exists {
			var validTimes []time.Time
			for _, t := range times {
				if now.Sub(t) < window {
					validTimes = append(validTimes, t)
				}
			}
			requests[ip] = validTimes
		}

		// Check limit
		if len(requests[ip]) >= maxRequests {
			fmt.Printf("✗ Rate limit exceeded for %s\n", ip)
			return fmt.Errorf("rate limit exceeded: max %d requests per %v", maxRequests, window)
		}

		// Add current request
		requests[ip] = append(requests[ip], now)

		return c.Next()
	}
}

// 5. Timing Middleware
func TimingMiddleware(c *request.Context) error {
	start := time.Now()
	c.Set("request_start", start)

	err := c.Next()

	elapsed := time.Since(start)
	fmt.Printf("⏱ Request took %v\n", elapsed)

	return err
}

// ============================================
// HTTP Handlers
// ============================================

func PublicEndpoint() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "This is a public endpoint",
		"access":  "public",
	})
}

func ProtectedEndpoint(c *request.Context) *response.ApiHelper {
	userID := c.Get("user_id")
	username := c.Get("username")
	requestID := c.Get("request_id")

	return response.NewApiOk(map[string]any{
		"message":    "This is a protected endpoint",
		"access":     "protected",
		"user_id":    userID,
		"username":   username,
		"request_id": requestID,
	})
}

func RateLimitedEndpoint() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "Rate limited endpoint",
		"note":    "Max 5 requests per minute",
	})
}

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Custom Middleware Example</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { display: inline-block; padding: 3px 8px; border-radius: 3px; font-weight: bold; color: white; }
        .get { background: #61affe; }
        code { background: #eee; padding: 2px 6px; border-radius: 3px; }
        .middleware { color: #666; font-size: 0.9em; margin-top: 5px; }
    </style>
</head>
<body>
    <h1>🔧 Custom Middleware Example</h1>
    
    <p>This example demonstrates various custom middleware patterns:</p>
    <ul>
        <li><strong>Request Logger</strong> - Logs all requests with timing</li>
        <li><strong>Request ID</strong> - Generates unique request IDs</li>
        <li><strong>Timing</strong> - Tracks request processing time</li>
        <li><strong>Auth</strong> - Token-based authentication</li>
        <li><strong>Rate Limiter</strong> - Limits requests per IP</li>
    </ul>

    <h2>Test Endpoints</h2>

    <div class="endpoint">
        <span class="method get">GET</span>
        <code>/public</code> - Public endpoint
        <div class="middleware">Middleware: Logger, RequestID, Timing</div>
    </div>

    <div class="endpoint">
        <span class="method get">GET</span>
        <code>/protected</code> - Protected endpoint (requires auth)
        <div class="middleware">Middleware: Logger, RequestID, Auth</div>
    </div>

    <div class="endpoint">
        <span class="method get">GET</span>
        <code>/rate-limited</code> - Rate limited endpoint (5 req/min)
        <div class="middleware">Middleware: Logger, RateLimiter</div>
    </div>

    <h2>📖 Documentation</h2>
    <p>See <code>index</code> for detailed explanation</p>
    <p>Use <code>test.http</code> for API testing</p>
</body>
</html>`

	return response.NewHtmlResponse(html)
}

// ============================================
// Main
// ============================================

func main() {
	router := lokstra.NewRouter("custom-middleware")

	// Global middleware (applies to all routes)
	router.Use(RequestLoggerMiddleware)
	router.Use(RequestIDMiddleware)

	// Routes
	router.GET("/", Home)

	// Public endpoint with timing
	router.GET("/public", PublicEndpoint,
		TimingMiddleware,
	)

	// Protected endpoint with auth
	router.GET("/protected", ProtectedEndpoint,
		AuthMiddleware,
	)

	// Rate limited endpoint
	router.GET("/rate-limited", RateLimitedEndpoint,
		RateLimiterMiddleware(5, 1*time.Minute),
	)

	app := lokstra.NewApp("custom-middleware", ":3000", router)

	fmt.Println("🚀 Custom Middleware Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("   GET /public        - Public (timing)")
	fmt.Println("   GET /protected     - Protected (auth required)")
	fmt.Println("   GET /rate-limited  - Rate limited (5 req/min)")
	fmt.Println("\n🔑 Auth token: Bearer secret-token")
	fmt.Println("\n🧪 Open test.http for API testing")

	if err := app.Run(0); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
