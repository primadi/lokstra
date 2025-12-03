package main

import (
	"github.com/primadi/lokstra"
)

func main() {
	// 1. Bootstrap Lokstra framework
	lokstra.Bootstrap()

	// 2. Load application config
	if err := lokstra.LoadConfigFromFolder("config"); err != nil {
		panic(err)
	}

	// 3. auto db migrations
	if err := lokstra.CheckDbMigrationsAuto("migrations"); err != nil {
		panic(err)
	}

	// 4. Register routers
	registerRouters()

	// 5. Run the server
	if err := lokstra.InitAndRunServer(); err != nil {
		panic(err)
	}
}
