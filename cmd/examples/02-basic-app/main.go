package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
)

func main() {
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

	// app := lokstra.NewApp("basic-app", ":8080", r)
	app := lokstra.NewAppWithConfig("basic-app", ":8080", "fasthttp", nil, r)
	app.PrintStartInfo()
	app.Start()
}
