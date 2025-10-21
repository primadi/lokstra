package main

import (
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	// Create router
	r := lokstra.NewRouterWithEngine("api", "chi")

	// Simple routes
	r.GET("/", func() string {
		return "Hello, Lokstra!"
	})

	r.GETPrefix("/", func() string {
		return "Hello, any!"
	})

	r.GET("/ping", func() string {
		return "pong"
	})

	r.GET("/time", func() map[string]any {
		return map[string]any{
			"timestamp": time.Now().Unix(),
			"datetime":  time.Now().Format(time.RFC3339),
		}
	})

	// Create app and run
	app := lokstra.NewApp("hello", ":3000", r)

	app.PrintStartInfo()
	if err := app.Run(30 * time.Second); err != nil {
		panic(err) // Or use log.Fatal(err)
	}
}
