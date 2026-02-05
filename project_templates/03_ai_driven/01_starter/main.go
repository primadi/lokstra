package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	// TODO: register any middleware, services, or database pools if needed
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	// Auto-generate code from @Handler, @Service annotations
	// This will detect changes in files with @Handler, @Service and regenerate code
	lokstra_init.BootstrapAndRun()
}
