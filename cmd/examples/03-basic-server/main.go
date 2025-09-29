package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
)

func createBasicRouter() lokstra.Router {
	r := lokstra.NewRouter("basic-router")

	// incoming request logging middleware
	r.Use(func(c *lokstra.RequestContext) error {
		fmt.Println("[Incoming request]", c.R.Method, c.R.URL.Path)
		// proceed to next middleware or handler
		return c.Next()
	})

	r.GET("/ping", func(c *lokstra.RequestContext) error {
		return c.Ok("pong")
	}, route.WithNameOption("ping-route"))
	return r
}

func createAnotherRouter() lokstra.Router {
	r := lokstra.NewRouter("another-router")
	r.GET("/hello", func(c *lokstra.RequestContext) error {
		return c.Ok("Hello, World!")
	})
	return r
}

func createAdminRouter() lokstra.Router {
	r := lokstra.NewRouter("admin-router")
	r.POST("/status", func(c *lokstra.RequestContext) error {
		return c.Ok("Server is running")
	})
	return r
}

func main() {
	basicRouter := createBasicRouter()
	anotherRouter := createAnotherRouter()
	adminRouter := createAdminRouter()

	// Create multiple apps, some sharing the same address
	app1 := lokstra.NewApp("basic-app", ":8080", basicRouter)
	app2 := lokstra.NewApp("another-app", ":8080", anotherRouter)
	appAdmin := lokstra.NewApp("admin-app", ":8081", adminRouter)

	// Create server with multiple apps
	svr := lokstra.NewServer("multi-app-server", app1, app2, appAdmin)
	svr.PrintStartInfo()

	// Run server with 5 seconds graceful shutdown timeout
	svr.Run(5 * time.Second)
}
