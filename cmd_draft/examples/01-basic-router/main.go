package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
)

// This example demonstrates creating a basic router with route groups
// and chaining another router with a prefix.
//
// The main router has two groups: /v1 and /v2, each with their own routes.
// Additionally, another router (r2) is created and chained to the main router
// with the prefix /r2.
//
// The server listens on port 8080.
func main() {
	// create main router
	mainRouter := createBasicRouter()

	// add v1 and v2 groups
	addV1Group(mainRouter)
	addV2Group(mainRouter)

	// create another router r2
	r2 := createR2Router()

	// chain r2 to mainRouter with /r2 prefix
	mainRouter.SetNextChainWithPrefix(r2, "/r2")

	fmt.Println("starting server at :8080")
	mainRouter.PrintRoutes()

	http.ListenAndServe(":8080", mainRouter)
}

func addV1Group(r lokstra.Router) {
	// group v1, with its own routes
	r.Group("/v1", func(g lokstra.Router) {
		g.GET("/hello", func(c *lokstra.RequestContext) error {
			name := c.Req.QueryParam("name", "stranger")
			return c.Api.Ok("Hello v1, " + name + "!")
		}, route.WithOverrideParentMwOption(true))

		// nested group under v1, with its own routes
		g.Group("/admin", func(admin lokstra.Router) {
			admin.GET("/dashboard", func(c *lokstra.RequestContext) error {
				return c.Api.Ok("Admin dashboard v1")
			})
			admin.GET("/stats", func(c *lokstra.RequestContext) error {
				return c.Api.Ok("Admin stats v1")
			})
		})
	})
}

func addV2Group(r lokstra.Router) {
	// group v2, using AddGroup
	gv2 := r.AddGroup("/v2")
	gv2.GET("/hello", func(c *lokstra.RequestContext) error {
		name := c.Req.QueryParam("name", "friend")
		return c.Api.Ok("Hello v2, " + name + "!")
	})

	// nested group under v2
	gv2Admin := gv2.AddGroup("/admin")
	gv2Admin.GET("/dashboard", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("Admin dashboard v2")
	}, route.WithNameOption("dashboard"))
	gv2Admin.GET("/stats", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("Admin stats v2")
	}, route.WithNameOption("stats"))
}

func createR2Router() lokstra.Router {
	r2 := lokstra.NewRouter("r2-router")
	r2.GET("/status", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("r2 status ok")
	}, route.WithNameOption("r2-status-route"))
	r2.GET("/ping", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("r2 pong")
	})

	return r2
}

func createBasicRouter() lokstra.Router {
	r := lokstra.NewRouter("basic-router")

	// incoming request logging middleware
	r.Use(func(c *lokstra.RequestContext) error {
		fmt.Println("[Incoming request]", c.R.Method, c.R.URL.Path)
		// proceed to next middleware or handler
		return c.Next()
	})

	r.GET("/ping", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("pong")
	}, route.WithNameOption("ping-route"))

	r.GET("/ping2", func() (string, error) {
		return "pong2", nil
	}, route.WithNameOption("ping2-route"))

	r.GET("/hello", func(c *lokstra.RequestContext) error {
		name := c.Req.QueryParam("name", "stranger")
		return c.Api.Ok("Hello, " + name + "!")
	}, route.WithNameOption("hello-route"))

	return r
}
