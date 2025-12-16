package main

import (
	"github.com/primadi/lokstra/lokstra_init"
)

func main() {
	lokstra_init.BootstrapAndRun(
		lokstra_init.WithAnnotations(true),
		lokstra_init.WithYAMLConfigPath(true, "config"),
		lokstra_init.WithPgSyncMap(true, "db_main"),
		lokstra_init.WithDbPoolManager(true, true),
		lokstra_init.WithDbMigrations(true, "migrations"),
		lokstra_init.WithServerInitFunc(func() error {
			registerRouters()
			return nil
		}),
	)
}
