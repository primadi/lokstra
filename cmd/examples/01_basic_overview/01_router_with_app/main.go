package main

import (
	"lokstra"
)

// This example demonstrates how to create a basic Lokstra application with a simple route.
// It sets up an HTTP server that listens on port 8080 and responds to a GET request at the "/ping" endpoint.
func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "app1", 8080)
	// app := lokstra.NewAppCustom(ctx, "app1", 8080,
	// 	lokstra.LISTENER_FASTHTTP, lokstra.ROUTER_ENGINE_SERVEMUX)

	app.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from anonymous handler")
	})

	app.Start()
}
