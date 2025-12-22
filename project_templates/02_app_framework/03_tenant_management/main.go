package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
	recovery.Register()
	request_logger.Register()

	lokstra_init.BootstrapAndRun(
		lokstra_init.WithPgSyncMap(true, "db_core"),
		lokstra_init.WithDbPoolAutoSync(true),
	// lokstra_init.WithDbMigrations(true, "migrations"),
	)
}
