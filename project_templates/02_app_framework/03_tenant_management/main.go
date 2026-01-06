package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	lokstra_init.BootstrapAndRun(
	// lokstra_init.WithLogLevel(logger.LogLevelDebug),
	// lokstra_init.WithPgSyncMap(true, "db_auth"),
	// lokstra_init.WithDbMigrations(true, "migrations"),
	)
}
