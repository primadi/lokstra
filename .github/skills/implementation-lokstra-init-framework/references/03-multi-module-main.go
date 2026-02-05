package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

// Multi-module application
// Module imports are AUTO-GENERATED in zz_lokstra_imports.go
// You DON'T need to manually import module packages!
//
// During bootstrap, Lokstra will:
// 1. Scan all directories for @Handler and @Service annotations
// 2. Generate zz_lokstra_imports.go with required imports
// 3. Auto-restart if code changes detected (dev mode)
func main() {
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	// BootstrapAndRun will:
	// 1. Detect @Handler and @Service annotations across all modules
	// 2. Generate route registration code
	// 3. Load configuration from configs/
	// 4. Create services in dependency order
	// 5. Start HTTP server
	lokstra_init.BootstrapAndRun()
}
