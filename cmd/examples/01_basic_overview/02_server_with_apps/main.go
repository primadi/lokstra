package main

import (
	"time"

	"github.com/primadi/lokstra"
)

func createApp1(ctx lokstra.RegistrationContext, addr string) *lokstra.App {
	app1 := lokstra.NewApp(ctx, "app1", addr)

	// Register a handler using a named handler
	app1.GET("/ping", "ping1Handler")

	return app1
}

func createApp2(ctx lokstra.RegistrationContext, addr string) *lokstra.App {
	app2 := lokstra.NewApp(ctx, "app2", addr)

	// Register a handler using an Handler function
	app2.GET("/ping2", func(ctx *lokstra.Context) error {
		return ctx.Ok("App2 Pong from anonymous handler")
	})

	return app2
}

func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	server := lokstra.NewServer(ctx, "my-server")

	server.AddApp(createApp1(ctx, ":8080"))
	server.AddApp(createApp2(ctx, ":8081"))

	// Register a named handler
	ctx.RegisterHandler("ping1Handler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})

	// Wait for shutdown signal with a timeout
	server.StartAndWaitForShutdown(5 * time.Second)
}

// func main() {
// 	ctx := lokstra.NewGlobalContext()

// 	app1 := createApp1(ctx, "127.0.0.1:8080")
// 	app2 := createApp2(ctx, "127.0.0.1:8081")

// 	app3 := lokstra.NewApp(ctx, "app3", ":80")

// 	app3.MountReverseProxy("/app1", "http://localhost:8080")
// 	app3.MountReverseProxy("/app2", "http://localhost:8081")

// 	server := lokstra.NewServer(ctx, "my-server")
// 	server.AddApp(app1)
// 	server.AddApp(app2)
// 	server.AddApp(app3)

// 	// 	// Register a named handler
// 	ctx.RegisterHandler("ping1Handler", func(ctx *lokstra.Context) error {
// 		return ctx.Ok("Pong from named handler")
// 	})

// 	server.Start()
// }
