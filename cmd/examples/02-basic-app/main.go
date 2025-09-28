package main

import (
	"fmt"

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

func main() {
	basicRouter := createBasicRouter()
	anotherRouter := createAnotherRouter()

	// app := lokstra.NewApp("basic-app", ":8080", r)
	app := lokstra.NewAppWithConfig("basic-app", ":8080", "fasthttp", nil, basicRouter, anotherRouter)

	app.PrintStartInfo()
	app.Start()
}
