package main

import (
	"lokstra"
)

// This example demonstrates how to create a Lokstra application with a router,
// and register handlers both anonymously and by name.

// App is a high-level structure that combines a router and an HTTP listener.
// It allows you to define routes and handlers, and then start the application.

func main() {
	// Create a new AppInfo instance with a name and port, and mount the router.
	app := lokstra.NewApp("app1", 8080) // use default NetHttpListenerType

	// uncomment the following line to use FastHttpListenerType instead of NetHttpListenerType.
	// app.WithFastHttpListener()

	// Register an anonymous handler for the "/ping" route.
	app.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from anonymous handler")
	})

	// Register a named handler for the "/namedping" route.
	app.GET("/namedping", "pingHandler")

	// Register a named handler that can be used in the router.
	lokstra.RegisterHandler("pingHandler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})

	// Start the app, which will resolve the router and handlers into live runtime components.
	app.Start()
}
