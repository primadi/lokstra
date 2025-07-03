package main

import (
	"lokstra"
)

func createApp1(port int) *lokstra.AppMeta {
	// Create a new AppInfo instance with a name and port, and mount the router.
	app1 := lokstra.NewApp("app1", port).WithFastHttpListener() // Use FastHttpListenerType

	// Register an anonymous handler for the "/ping" route.
	app1.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("App1 Pong from anonymous handler")
	})

	// Register a named handler for the "/namedping" route.
	app1.GET("/namedping", "pingHandler")

	return app1
}

func createApp2(port int) *lokstra.AppMeta {
	// Create a new AppInfo instance with a name and port, and mount the router.
	app2 := lokstra.NewApp("app2", port) // Use default NetHttpListenerType

	// Register an anonymous handler for the "/ping" route.
	app2.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("App2 Pong from anonymous handler")
	})

	// Register a named handler for the "/namedping" route.
	app2.GET("/namedping", "pingHandler")

	return app2
}

func main() {
	// This example demonstrates how to create a Lokstra server with multiple applications,

	// Create a new ServerInfo instance with a name and add the app to it.
	server := lokstra.NewServer("my-server")

	// Add the apps to the server.
	server.AddApp(createApp1(8080))
	server.AddApp(createApp2(8081))

	// Register a named handler that can be used in the router.
	lokstra.RegisterHandler("pingHandler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})

	// Start the server, which will resolve the app and handlers into live runtime components.
	server.Start()
}
