package main

import (
	"lokstra"
)

func createApp1(ctx lokstra.ComponentContext, port int) *lokstra.App {
	app1 := lokstra.NewApp(ctx, "app1", port)

	app1.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("App1 Pong from anonymous handler")
	})

	return app1
}

func createApp2(ctx lokstra.ComponentContext, port int) *lokstra.App {
	app2 := lokstra.NewApp(ctx, "app2", port)

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

	ctx.RegisterHandler("pingHandler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})
	ctx.RegisterHandler("ping2Handler", func(ctx *lokstra.Context) error {
		return ctx.Ok("App2 Pong from named handler")
	})

	server.Start()
}
