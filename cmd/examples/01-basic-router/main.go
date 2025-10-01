package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
)

func createV1Group(r lokstra.Router) {
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

func createV2Group(r lokstra.Router) {
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

func main() {
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

	r.GET("/hello", func(c *lokstra.RequestContext) error {
		name := c.Req.QueryParam("name", "stranger")
		return c.Api.Ok("Hello, " + name + "!")
	}, route.WithNameOption("hello-route"))

	createV1Group(r)
	createV2Group(r)

	r2 := createR2Router()

	// chain r2 to r with /r2 prefix
	r.SetNextChainWithPrefix(r2, "/r2")

	fmt.Println("starting server at :8080")
	r.PrintRoutes()

	http.ListenAndServe(":8080", r)
}
