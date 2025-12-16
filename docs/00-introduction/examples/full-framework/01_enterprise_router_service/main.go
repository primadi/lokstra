package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	lokstra_init.Bootstrap()

	logger.LogInfo("")
	logger.LogInfo("╔═══════════════════════════════════════════════╗")
	logger.LogInfo("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
	logger.LogInfo("║   Domain-Driven Design with Bounded Contexts  ║")
	logger.LogInfo("╚═══════════════════════════════════════════════╝")
	logger.LogInfo("")

	logger.SetLogLevelFromEnv()

	if _, err := loader.LoadConfig("config"); err != nil {
		panic(err)
	}

	dsn := lokstra_registry.GetConfig("db_main.dsn", "")
	schema := lokstra_registry.GetConfig("db_main.schema", "public")

	// Just to show that we can access nested config values
	logger.LogInfo("Using Global DB DSN: %s, Schema: %s", dsn, schema)

	type dbConfig struct {
		DSN    string `json:"dsn"`
		Schema string `json:"schema"`
	}
	fullDBConfig := lokstra_registry.GetConfig("db_main", dbConfig{})

	// Print full nested config struct
	logger.LogInfo("Full Global DB Config: %+v", fullDBConfig)

	// 1. Register service types from all modules
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config folder
	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		panic(err)
	}
}
