package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_handler"
)

// Example 1: Pure code-based reverse proxy (no config)
func example1_CodeBased() {
	r := lokstra.NewRouter("api-gateway")

	// Mount reverse proxy to backend API
	r.ANYPrefix("/api", lokstra_handler.MountReverseProxy("/api", "http://localhost:9000", nil))

	app := lokstra.NewApp("gateway", ":8080", r)

	log.Println("Starting code-based reverse proxy on :8080")
	log.Println("  /api/* -> http://localhost:9000/*")

	if err := app.Run(5 * time.Second); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Choose which example to run
	example1_CodeBased()

	// Or use config-based approach:
	// example2_ConfigBased()
}
