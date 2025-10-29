package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Adapter for third-party middleware pattern
func AdaptMiddleware(thirdPartyMw func(string, string) bool) func(*request.Context) error {
	return func(c *request.Context) error {
		method := c.R.Method
		path := c.R.URL.Path

		if !thirdPartyMw(method, path) {
			return fmt.Errorf("third-party middleware rejected request")
		}

		return c.Next()
	}
}

// Simulated third-party middleware
func ThirdPartyAuth(method, path string) bool {
	log.Printf("[ThirdParty] Checking %s %s", method, path)
	// Only allow GET requests
	return method == "GET"
}

// Simulated CORS middleware
func ThirdPartyCORS(method, path string) bool {
	log.Printf("[ThirdParty CORS] Processing %s %s", method, path)
	// Allow all requests with /api prefix
	return strings.HasPrefix(path, "/api")
}

func main() {
	router := lokstra.NewRouter("integration")

	// Use adapted third-party middleware
	router.GET("/api/data", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Data from third-party protected endpoint",
		})
	}, AdaptMiddleware(ThirdPartyAuth), AdaptMiddleware(ThirdPartyCORS))

	// Another endpoint
	router.GET("/public", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Public endpoint",
		})
	})

	// Home
	router.GET("/", func(c *request.Context) any {
		return response.NewHtmlResponse(`
		<html>
		<body>
			<h1>Third-Party Middleware Integration</h1>
			<p>Demonstrates adapting external middleware to Lokstra</p>
			<ul>
				<li><a href="/public">Public</a> - No third-party middleware</li>
				<li><a href="/api/data">API Data</a> - With third-party auth & CORS</li>
			</ul>
		</body>
		</html>
		`)
	})

	log.Println("Server starting on :3005")
	app := lokstra.NewApp("integration", ":3005", router)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
