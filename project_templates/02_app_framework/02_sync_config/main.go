package main

import (
	"github.com/primadi/lokstra/lokstra_init"
)

func main() {
	lokstra_init.BootstrapAndRun(
		// lokstra_init.WithAnnotations(true), // default is true
		// lokstra_init.WithYAMLConfigPath(true, "config"), // default is true, path is empty (config folder)
		lokstra_init.WithPgSyncMap(true, "db_main"),
		lokstra_init.WithDbPoolAutoSync(true),
		lokstra_init.WithDbMigrations(true, "migrations"),
		lokstra_init.WithServerInitFunc(func() error {
			registerRouters()
			return nil
		}),
	)
}
