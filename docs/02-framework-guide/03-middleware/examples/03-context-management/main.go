package main

import (
	"log"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Context management example - storing and retrieving request-scoped data

func UserMiddleware(c *request.Context) error {
	// Simulate user authentication
	userID := c.R.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	// Repository in context
	c.Set("user_id", userID)
	c.Set("user_role", "user")

	log.Printf("User authenticated: %s", userID)
	return c.Next()
}

func RequestMetadataMiddleware(c *request.Context) error {
	// Repository request metadata
	c.Set("request_path", c.R.URL.Path)
	c.Set("request_method", c.R.Method)
	c.Set("client_ip", c.R.RemoteAddr)

	return c.Next()
}

func main() {
	router := lokstra.NewRouter("context-management")

	// Global middleware
	router.Use(UserMiddleware)
	router.Use(RequestMetadataMiddleware)

	// Endpoint that reads from context
	router.GET("/profile", func(c *request.Context) any {
		userID := c.Get("user_id")
		role := c.Get("user_role")
		path := c.Get("request_path")
		method := c.Get("request_method")
		ip := c.Get("client_ip")

		return response.NewApiOk(map[string]any{
			"user_id":        userID,
			"role":           role,
			"request_path":   path,
			"request_method": method,
			"client_ip":      ip,
		})
	})

	// Home
	router.GET("/", func(c *request.Context) any {
		return response.NewHtmlResponse(`
		<html>
		<body>
			<h1>Context Management Example</h1>
			<p>Demonstrates storing and retrieving request-scoped data</p>
			<ul>
				<li><a href="/profile">Profile</a> - View context data</li>
			</ul>
		</body>
		</html>
		`)
	})

	log.Println("Server starting on :3002")
	app := lokstra.NewApp("context-management", ":3002", router)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
