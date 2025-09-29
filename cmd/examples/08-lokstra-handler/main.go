package main

import (
	"net/http"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_handler"
)

// createHelloApp creates a simple application that responds to /hello and /admin/hello routes.
func createHelloApp(addr string) *lokstra.App {
	r := lokstra.NewRouter("hello-router")
	r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	r.GET("/admin/hello", func(c *lokstra.RequestContext) error {
		return c.Ok("Hello, Admin!")
	})

	app := lokstra.NewApp("hello-app", addr, r)
	return app
}

// createMainApp creates the main application with a reverse proxy to the hello service.
func createMainApp(addr string) *lokstra.App {
	r := lokstra.NewRouter("main-router")

	r.GET("/ping", func(c *lokstra.RequestContext) error {
		return c.Ok("pong")
	})

	// Reverse proxy for /say* to http://localhost:8081
	// Example:
	//   /say/hello          -> http://localhost:8081/hello
	//   /say/admin/hello    -> http://localhost:8081/admin/hello
	r.ANYPrefix("/say", lokstra_handler.
		MountReverseProxy("/say/", "http://localhost:8081"))

	app := lokstra.NewApp("main-app", addr, r)
	return app
}

func main() {
	helloApp := createHelloApp(":8081")
	mainApp := createMainApp(":8080")

	srv := lokstra.NewServer("main-server", mainApp, helloApp)
	srv.PrintStartInfo()
	srv.Start()
}
