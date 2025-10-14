package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	r := lokstra.NewRouter("basic-router")

	// Middleware example: simple request logger
	r.Use(func(c *lokstra.RequestContext) error {
		start := time.Now()
		// Process the next handler in the chain
		err := c.Next()
		// Log the request path and duration
		fmt.Println("Request", c.R.URL.Path, "took:", time.Since(start))
		return err
	})

	r.GET("/ping", func() (string, error) {
		return "pong", nil
	})

	type helloParams struct {
		Name string `query:"name"`
	}

	// Route with query string parameter
	r.GET("/hello", func(req *helloParams) (string, error) {
		return fmt.Sprintf("Hello, %s!", req.Name), nil
	})

	app := lokstra.NewApp("basic-app", ":8080", r)
	app.PrintStartInfo()
	app.Run(0)
}
