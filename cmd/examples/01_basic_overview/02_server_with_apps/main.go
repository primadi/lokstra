package main

import (
	"lokstra"
)

func createApp1(ctx lokstra.ComponentContext, port int) *lokstra.App {
	app1 := lokstra.NewApp(ctx, "app1", port)

	// Register a handler using a named handler
	app1.GET("/ping", "ping1Handler")

	return app1
}

func createApp2(ctx lokstra.ComponentContext, port int) *lokstra.App {
	app2 := lokstra.NewApp(ctx, "app2", port)

	// Register a handler using an Handler function
	app2.GET("/ping2", func(ctx *lokstra.Context) error {
		return ctx.Ok("App2 Pong from anonymous handler")
	})

	return app2
}

func main() {
	ctx := lokstra.NewGlobalContext()
	server := lokstra.NewServer(ctx, "my-server")

	server.AddApp(createApp1(ctx, 8080))
	server.AddApp(createApp2(ctx, 8081))

	// Register a named handler
	ctx.RegisterHandler("ping1Handler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})

	server.Start()
}
