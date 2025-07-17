package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates how to create a basic Lokstra application with a simple route.
// It sets up an HTTP server that listens on port 8080 and responds to a GET request at the "/ping" endpoint.
func main() {
	// Create a global context for the application.
	// This context is used to manage components and services within the application.
	ctx := lokstra.NewGlobalContext()

	// Create a new Lokstra application with the specified context, name, and port.
	// The application will use the default listener (net/http) and router engine (httprouter).
	app := lokstra.NewApp(ctx, "app1", ":8080")

	// Uncomment the following line to create an application with a Unix socket listener.
	// To test this: curl --unix-socket /tmp/lokstra.sock http://localhost/ping
	// app := lokstra.NewApp(ctx, "app1", "unix:///tmp/lokstra.sock")

	// Uncomment the following line to create an application with FastHTTP listener.
	// app := lokstra.NewAppFastHTTP(ctx, "app1", ":8080")

	// Uncomment the following line to create an application with a Unix socket listener using FastHTTP.
	// To test this: curl --unix-socket /tmp/lokstra.sock http://localhost/ping
	// app := lokstra.NewAppFastHTTP(ctx, "app1", "unix:///tmp/lokstra.sock")

	// Uncomment the following line to create a secure application with TLS.
	// generate a self-signed certificate and key using:
	// openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365
	// To test this: curl --insecure https://localhost:8080/ping
	// app := lokstra.NewAppSecure(ctx, "app1", ":8080",
	// 	"certs/cert.pem", "certs/key.pem", "")

	// Uncomment the following line to create a secure application with TLS using a Unix socket.
	// To test this: curl -k --unix-socket /tmp/lokstra.sock https://localhost/ping
	// app := lokstra.NewAppSecure(ctx, "app1", "unix:///tmp/lokstra.sock",
	// 	"certs/cert.pem", "certs/key.pem", "")

	// Uncomment the following line to create an application with HTTP/3 support.
	// Tot test this: curl --http3 -k https://localhost:8080/ping
	// app := lokstra.NewAppHttp3(ctx, "app1", ":8080",
	// 	"certs/cert.pem", "certs/key.pem", "")

	// Uncomment the following line to use a custom listener and router engine.
	// app := lokstra.NewAppCustom(ctx, "app1", ":8080",
	// 	lokstra.LISTENER_FASTHTTP, lokstra.ROUTER_ENGINE_SERVEMUX, nil)

	app.GET("/ping", func(ctx *lokstra.Context) error {
		return ctx.Ok("Pong from anonymous handler")
	})

	lokstra.Logger.Infof("Lokstra Application started")

	app.Start()
}
