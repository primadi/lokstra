package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Lightweight middleware
func LightMiddleware(c *request.Context) error {
	return c.Next()
}

// Heavy middleware (simulates processing)
func HeavyMiddleware(c *request.Context) error {
	time.Sleep(10 * time.Millisecond)
	return c.Next()
}

func main() {
	router := lokstra.NewRouter("performance")

	// Endpoint with no middleware
	router.GET("/baseline", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Baseline - no middleware",
		})
	})

	// Endpoint with 1 lightweight middleware
	router.GET("/light", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Light middleware",
		})
	}, LightMiddleware)

	// Endpoint with 5 lightweight middleware
	router.GET("/light-5", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "5 light middleware",
		})
	}, LightMiddleware, LightMiddleware, LightMiddleware, LightMiddleware, LightMiddleware)

	// Endpoint with 1 heavy middleware
	router.GET("/heavy", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Heavy middleware",
		})
	}, HeavyMiddleware)

	// Home
	router.GET("/", func(c *request.Context) any {
		return response.NewHtmlResponse(`
		<html>
		<body>
			<h1>Middleware Performance Example</h1>
			<p>Compare performance with different middleware configurations</p>
			<ul>
				<li><a href="/baseline">Baseline</a> - No middleware</li>
				<li><a href="/light">Light</a> - 1 lightweight middleware</li>
				<li><a href="/light-5">Light x5</a> - 5 lightweight middleware</li>
				<li><a href="/heavy">Heavy</a> - 1 heavy middleware (10ms delay)</li>
			</ul>
			<p>Use ApacheBench or similar tools to benchmark:</p>
			<code>ab -n 1000 -c 10 http://localhost:3004/baseline</code>
		</body>
		</html>
		`)
	})

	log.Println("Server starting on :3004")
	app := lokstra.NewApp("performance", ":3004", router)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
