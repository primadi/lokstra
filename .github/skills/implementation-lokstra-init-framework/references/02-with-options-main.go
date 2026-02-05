package main

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/gzipcompression"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/eventbus"
	"github.com/primadi/lokstra/services/kvstore"
)

// Advanced main.go with configuration options
func main() {
	// 1. Register middleware (order matters!)
	recovery.Register()
	request_logger.Register()
	cors.Register([]string{"*"}) // Allow all origins
	gzipcompression.Register()

	// 2. Register infrastructure services
	dbpool_pg.Register()
	eventbus.Register()
	kvstore.Register()

	// 3. Bootstrap with options
	if err := lokstra_init.BootstrapAndRun(
		// Set log level (default: Info)
		lokstra_init.WithLogLevel(logger.LogLevelDebug),

		// Enable database migrations (default: false)
		lokstra_init.WithDbMigrations(true, "migrations"),

		// Enable PgSyncMap for distributed config (default: false)
		lokstra_init.WithPgSyncMap(true, "db_main"),

		// Custom server initialization hook
		lokstra_init.WithServerInitFunc(func() error {
			fmt.Println("╔═══════════════════════════════════════╗")
			fmt.Println("║   Custom Server Initialization        ║")
			fmt.Println("╚═══════════════════════════════════════╝")

			// Register custom services, load extra configs, etc.
			return nil
		}),
	); err != nil {
		panic("Failed to initialize: " + err.Error())
	}
}
