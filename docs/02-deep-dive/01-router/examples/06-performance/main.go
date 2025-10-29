package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Cache for demonstration
var (
	cachedData = map[string]any{
		"id":      123,
		"name":    "Cached User",
		"cached":  true,
		"expires": time.Now().Add(1 * time.Hour).Unix(),
	}

	userCache  = make(map[string]map[string]any)
	cacheMutex sync.RWMutex
)

func main() {
	router := lokstra.NewRouter("performance")

	// ============================================
	// Handler Form Benchmarks
	// ============================================
	router.GET("/bench/no-input", BenchNoInput)
	router.GET("/bench/context-only", BenchContextOnly)
	router.GET("/bench/with-params/:id", BenchWithParams)
	router.POST("/bench/with-body", BenchWithBody)

	// ============================================
	// Optimization Examples
	// ============================================
	router.GET("/optimize/cached", OptimizedCached)
	router.GET("/optimize/helper", OptimizedHelper)
	router.GET("/optimize/user/:id", OptimizedUserCache)

	// ============================================
	// Comparison Routes
	// ============================================
	router.GET("/compare/slow", ComparisonSlow)
	router.GET("/compare/fast", ComparisonFast)

	// ============================================
	// Middleware Performance
	// ============================================
	router.GET("/middleware/none", NoMiddleware)
	router.GET("/middleware/multi/test", MiddlewareTest, TimingMiddleware, LoggingMiddleware, MetricsMiddleware)

	// Home
	router.GET("/", Home)

	app := lokstra.NewApp("performance", ":3000", router)

	fmt.Println("🚀 Performance Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n⚡ Benchmark routes:")
	fmt.Println("   GET /bench/no-input          (fastest)")
	fmt.Println("   GET /bench/context-only      (fast)")
	fmt.Println("   GET /bench/with-params/:id   (good)")
	fmt.Println("   POST /bench/with-body        (slower)")
	fmt.Println("\n🧪 Open test.http for all examples")

	if err := app.Run(0); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// ============================================
// Handler Form Benchmarks
// ============================================

func BenchNoInput() map[string]any {
	return map[string]any{
		"message": "No input, value return",
		"speed":   "⚡⚡⚡ Fastest",
		"note":    "Minimal overhead, no parameter binding",
	}
}

func BenchContextOnly(c *request.Context) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "Context only",
		"speed":   "⚡⚡⚡ Fast",
		"note":    "Context access is virtually free",
	})
}

type BenchParams struct {
	ID string `path:"id"`
}

func BenchWithParams(params BenchParams) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "With parameters",
		"speed":   "⚡⚡ Good",
		"id":      params.ID,
		"note":    "Parameter binding adds ~800ns overhead",
	})
}

type BenchRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func BenchWithBody(req BenchRequest) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "With body binding",
		"speed":   "⚡ Slower",
		"data":    req,
		"note":    "JSON parsing + validation adds ~8000ns",
	})
}

// ============================================
// Optimization Examples
// ============================================

func OptimizedCached() map[string]any {
	// Returns pre-allocated data (no allocations)
	return cachedData
}

func OptimizedHelper() *response.ApiHelper {
	// Uses helper constructor (faster than manual)
	return response.NewApiOk(map[string]any{
		"optimization": "helper-constructor",
		"benefit":      "Single allocation vs multiple",
	})
}

type UserIDParam struct {
	ID string `path:"id"`
}

func OptimizedUserCache(params UserIDParam) *response.ApiHelper {
	// Check cache first (read lock)
	cacheMutex.RLock()
	if user, found := userCache[params.ID]; found {
		cacheMutex.RUnlock()
		return response.NewApiOk(map[string]any{
			"user":   user,
			"cached": true,
			"speed":  "⚡⚡⚡ Cache hit",
		})
	}
	cacheMutex.RUnlock()

	// Simulate database fetch
	time.Sleep(10 * time.Millisecond)
	user := map[string]any{
		"id":   params.ID,
		"name": fmt.Sprintf("User %s", params.ID),
	}

	// Update cache (write lock)
	cacheMutex.Lock()
	userCache[params.ID] = user
	cacheMutex.Unlock()

	return response.NewApiOk(map[string]any{
		"user":   user,
		"cached": false,
		"speed":  "⚡ Database fetch",
	})
}

// ============================================
// Comparison Routes
// ============================================

func ComparisonSlow() *response.ApiHelper {
	// Slow: Multiple allocations
	api := response.NewApiHelper()
	data := make(map[string]any)
	data["message"] = "Slow version"
	data["allocations"] = "Multiple"

	// String concatenation (allocates)
	result := "Result: " + "slow" + " " + "version"
	data["result"] = result

	api.Ok(data)
	return api
}

func ComparisonFast() *response.ApiHelper {
	// Fast: Helper constructor, pre-allocated data
	return response.NewApiOk(map[string]any{
		"message":     "Fast version",
		"allocations": "Minimal",
		"result":      "Result: fast version", // Single allocation
	})
}

// ============================================
// Middleware Performance
// ============================================

func NoMiddleware() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message":    "No middleware",
		"overhead":   "0ns",
		"middleware": 0,
	})
}

func TimingMiddleware(c *request.Context) error {
	start := time.Now()
	err := c.Next()
	duration := time.Since(start)

	// Only log slow requests (avoid allocation for fast requests)
	if duration > 100*time.Millisecond {
		fmt.Printf("Slow request: %v\n", duration)
	}

	return err
}

func LoggingMiddleware(c *request.Context) error {
	// Minimal logging overhead
	return c.Next()
}

func MetricsMiddleware(c *request.Context) error {
	// Simulate metrics collection
	return c.Next()
}

func MiddlewareTest() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message":    "With middleware",
		"overhead":   "~200-600ns",
		"middleware": 3,
		"note":       "Each middleware adds ~100-200ns",
	})
}

// ============================================
// Home
// ============================================

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Performance Example</title>
    <style>
        body { font-family: Arial; max-width: 1200px; margin: 40px auto; padding: 20px; }
        h1 { color: #333; }
        .section { margin: 30px 0; padding: 20px; background: #f5f5f5; border-radius: 8px; }
        .route { margin: 10px 0; padding: 10px; background: white; border-left: 4px solid #28a745; }
        .fast { border-left-color: #28a745; }
        .medium { border-left-color: #ffc107; }
        .slow { border-left-color: #dc3545; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        .speed { font-weight: bold; }
    </style>
</head>
<body>
    <h1>⚡ Performance Example</h1>
    <p>Benchmark different handler patterns and optimization techniques</p>

    <div class="section">
        <h2>Handler Form Benchmarks</h2>
        <div class="route fast">
            <span class="speed">⚡⚡⚡ Fastest</span>
            <a href="/bench/no-input">No Input</a> - ~1050 ns/op
        </div>
        <div class="route fast">
            <span class="speed">⚡⚡⚡ Fast</span>
            <a href="/bench/context-only">Context Only</a> - ~1100 ns/op
        </div>
        <div class="route medium">
            <span class="speed">⚡⚡ Good</span>
            <a href="/bench/with-params/123">With Params</a> - ~1850 ns/op
        </div>
        <div class="route slow">
            <span class="speed">⚡ Slower</span>
            <code>POST /bench/with-body</code> - ~9500 ns/op (use test.http)
        </div>
    </div>

    <div class="section">
        <h2>Optimization Examples</h2>
        <div class="route fast">
            <a href="/optimize/cached">Cached Response</a> - Pre-allocated data
        </div>
        <div class="route fast">
            <a href="/optimize/helper">Helper Constructor</a> - Single allocation
        </div>
        <div class="route fast">
            <a href="/optimize/user/123">User Cache</a> - Cache with read/write locks
        </div>
    </div>

    <div class="section">
        <h2>Comparison (Slow vs Fast)</h2>
        <div class="route slow">
            <a href="/compare/slow">Slow Version</a> - Multiple allocations
        </div>
        <div class="route fast">
            <a href="/compare/fast">Fast Version</a> - Optimized allocations
        </div>
    </div>

    <div class="section">
        <h2>Middleware Performance</h2>
        <div class="route fast">
            <a href="/middleware/none">No Middleware</a> - Baseline (0ns overhead)
        </div>
        <div class="route medium">
            <a href="/middleware/multi/test">With 3 Middleware</a> - ~200-600ns overhead
        </div>
    </div>

    <div class="section">
        <h2>📖 Documentation</h2>
        <p>See <code>index.md</code> for optimization techniques</p>
        <p>Use <code>test.http</code> for API testing</p>
        <p><strong>Benchmark:</strong> <code>go test -bench=. -benchmem</code></p>
    </div>
</body>
</html>`

	return response.NewHtmlResponse(html)
}
