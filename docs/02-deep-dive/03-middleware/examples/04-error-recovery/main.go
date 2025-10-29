package main

import (
	"log"
	"runtime/debug"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Panic recovery middleware
func RecoveryMiddleware(c *request.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED: %v\n%s", r, debug.Stack())
			// Cannot send response here, just log
		}
	}()
	return c.Next()
}

// Error logging middleware
func ErrorLoggingMiddleware(c *request.Context) error {
	err := c.Next()
	if err != nil {
		log.Printf("ERROR: %v on %s %s", err, c.R.Method, c.R.URL.Path)
	}
	return err
}

func main() {
	router := lokstra.NewRouter("error-recovery")

	// Global recovery and error logging
	router.Use(RecoveryMiddleware)
	router.Use(ErrorLoggingMiddleware)

	// Normal endpoint
	router.GET("/ok", func(c *request.Context) any {
		return response.NewApiOk(map[string]any{
			"message": "Everything is fine",
		})
	})

	// Endpoint that panics
	router.GET("/panic", func(c *request.Context) any {
		panic("Something went wrong!")
	})

	// Endpoint that returns error
	router.GET("/error", func(c *request.Context) any {
		return response.NewApiInternalError("Intentional error")
	})

	// Home
	router.GET("/", func(c *request.Context) any {
		return response.NewHtmlResponse(`
		<html>
		<body>
			<h1>Error Recovery Example</h1>
			<ul>
				<li><a href="/ok">OK</a> - Normal endpoint</li>
				<li><a href="/panic">Panic</a> - Triggers panic (recovered)</li>
				<li><a href="/error">Error</a> - Returns error</li>
			</ul>
		</body>
		</html>
		`)
	})

	log.Println("Server starting on :3003")
	app := lokstra.NewApp("error-recovery", ":3003", router)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
