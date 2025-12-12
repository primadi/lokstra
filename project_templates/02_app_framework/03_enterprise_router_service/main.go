package main

import (
	"fmt"

	"github.com/primadi/lokstra"
)

// NEW RECOMMENDED FLOW
// This flow separates config loading from service registration,
// allowing services to access config during registration.
func main() {
	if err := lokstra.BootstrapAndRun(
		// lokstra.WithLogLevel(logger.LogLevelDebug),
		lokstra.WithoutDbMigrations(),
		lokstra.WithServerInitFunc(func() error {
			fmt.Println("")
			fmt.Println("╔═══════════════════════════════════════════════╗")
			fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
			fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
			fmt.Println("║   [Config First]                              ║")
			fmt.Println("╚═══════════════════════════════════════════════╝")
			fmt.Println("")

			registerServiceTypes()
			registerRouters()
			registerMiddlewareTypes()

			return nil
		})); err != nil {
		panic("❌ Failed to initialize lokstra:" + err.Error())
	}
}
