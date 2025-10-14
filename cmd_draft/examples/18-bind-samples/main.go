package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra"
)

func main() {
	createMainRouter()
}

func createMainRouter() {
	mainRouter := lokstra.NewRouter("main")

	// get param using QueryParam
	mainRouter.GET("/hello", func(c *lokstra.RequestContext) error {
		name := c.Req.QueryParam("name", "world")
		return c.Api.Ok("Hello, " + name + "!")
	})

	type BindParams struct {
		Name string `query:"name"`
	}

	// get params using bind
	mainRouter.GET("/hello_bind", func(c *lokstra.RequestContext) error {
		var params BindParams
		if err := c.Req.BindAll(&params); err != nil {
			return c.Api.BadRequest("BiND_ERROR", err.Error())
		}
		return c.Api.Ok("Hello, " + params.Name + "!")
	})

	// get params using smart bind
	mainRouter.GET("/hello_smart_bind", func(c *lokstra.RequestContext,
		params *BindParams) error {
		return c.Api.Ok("Hello, " + params.Name + "!")
	})

	fmt.Println("🚀 Starting Lokstra Bind Samples Demo at http://localhost:8080")
	mainRouter.PrintRoutes()
	http.ListenAndServe(":8080", mainRouter)
}
