package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

// Basic main.go - Minimal setup for a Lokstra application
// This is the simplest working configuration
func main() {
	// 1. Register middleware (order matters!)
	recovery.Register()       // Panic recovery - always first
	request_logger.Register() // Request logging

	// 2. Register infrastructure services
	dbpool_pg.Register() // PostgreSQL connection pooling

	// 3. Bootstrap and run
	// - Auto-discovers @Handler and @Service annotations
	// - Loads config.yaml from configs/ folder
	// - Creates services in dependency order
	// - Mounts HTTP routes
	// - Starts server
	lokstra_init.BootstrapAndRun()
}
