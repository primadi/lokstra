package main

import (
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	// Create router
	r := lokstra.NewRouter("api")

	// Simple routes
	r.GET("/", func() string {
		return "Hello, Lokstra!"
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
	app.Run(30 * time.Second)
}
