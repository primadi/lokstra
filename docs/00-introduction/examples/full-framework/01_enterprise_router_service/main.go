package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	lokstra.Bootstrap()

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
	fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println("")

	deploy.SetLogLevelFromEnv()

	lokstra_registry.LoadConfigFromFolder("config")

	dsn := lokstra_registry.GetConfig("global-db.dsn", "")
	schema := lokstra_registry.GetConfig("global-db.schema", "public")

	// Just to show that we can access nested config values
	fmt.Printf("Using Global DB DSN: %s, Schema: %s\n", dsn, schema)

	type dbConfig struct {
		DSN    string `json:"dsn"`
		Schema string `json:"schema"`
	}
	fullDBConfig := lokstra_registry.GetConfig("global-db", dbConfig{})

	// Print full nested config struct
	fmt.Printf("Full Global DB Config: %+v\n", fullDBConfig)

	// 1. Register service types from all modules
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config folder
	if err := lokstra_registry.InitAndRunServer(); err != nil {
		panic(err)
	}
}
