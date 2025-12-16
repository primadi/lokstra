package main

import (
	"fmt"

	"github.com/primadi/lokstra/lokstra_init"
)

// NEW RECOMMENDED FLOW
// This flow separates config loading from service registration,
// allowing services to access config during registration.
func main() {
	if err := lokstra_init.BootstrapAndRun(
		// lokstra.WithLogLevel(logger.LogLevelDebug),
		lokstra_init.WithDbMigrations(false, "migrations"),
		lokstra_init.WithServerInitFunc(func() error {
			fmt.Println("")
			fmt.Println("╔═══════════════════════════════════════════════╗")
			fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
			fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
			fmt.Println("║   [Config First]                              ║")
			fmt.Println("╚═══════════════════════════════════════════════╝")
			fmt.Println("")

			registerRouters()
			registerMiddlewareTypes()

			return nil
		})); err != nil {
		panic("❌ Failed to initialize lokstra:" + err.Error())
	}
}
