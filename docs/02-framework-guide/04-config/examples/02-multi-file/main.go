package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/logger"
)

func main() {
	lokstra.Bootstrap()

	if err := lokstra.LoadConfig(
		"config/base.yaml",
		"config/dev.yaml", // or production.yaml for prod
	); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}

}
