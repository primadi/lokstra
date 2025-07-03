package main

import (
	"fmt"
	"lokstra"
	"net/http"
)

func main() {
	// This example demonstrates how to create a basic server with a router
	// and register handlers both anonymously and by name.

	// NewRouter creates a new RouterInfo instance.
	router := lokstra.NewRouter()

	// Register an anonymous handler for the "/ping" route.
	router.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from anonymous handler")
	})

	// Register a named handler for the "/namedping" route.
	router.GET("/namedping", "pingHandler")

	// Register a named handler that can be used in the router.
	lokstra.RegisterHandler("pingHandler", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from named handler")
	})

	fmt.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", router.CreateNetHttpRouter()) // Start the server on port 8080 with net/http

	// Alternatively, you can use fasthttp for better performance.

	// Uncomment the following line to use fasthttp instead of net/http
	// fmt.Println("Starting server with fasthttp on port 8081...")
	// fasthttp.ListenAndServe(":8081", router.CreateFastHttpHandler()) // Start the server on port 8081 with fasthttp
}
