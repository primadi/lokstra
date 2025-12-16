package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	lokstra_init.Bootstrap()

	if _, err := loader.LoadConfig(
		"config/base.yaml",
		"config/dev.yaml", // or production.yaml for prod
	); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}

}
