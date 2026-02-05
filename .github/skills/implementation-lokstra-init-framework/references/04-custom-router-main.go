package main

import (
	"fmt"

	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

// Advanced: Custom router registration with ServerInitFunc
// Use this when you need to:
// - Register routers programmatically
// - Add custom middleware types
// - Configure services before server starts
func main() {
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	if err := lokstra_init.BootstrapAndRun(
		lokstra_init.WithServerInitFunc(func() error {
			fmt.Println("Registering custom routers...")

			// Register router types
			registerRouters()

			// Register middleware types
			registerMiddlewareTypes()

			return nil
		}),
	); err != nil {
		panic("❌ Failed to initialize: " + err.Error())
	}
}

func registerRouters() {
	// Register custom routers in the registry
	// These are referenced in config.yaml deployment definitions
	lokstra_registry.RegisterRouter("api-router" /* router instance */)
	lokstra_registry.RegisterRouter("admin-router" /* router instance */)
}

func registerMiddlewareTypes() {
	// Register custom middleware types
	lokstra_registry.RegisterMiddlewareType("auth" /* middleware constructor */)
	lokstra_registry.RegisterMiddlewareType("rate-limit" /* middleware constructor */)
}
