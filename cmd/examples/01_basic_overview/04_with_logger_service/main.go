package main

import (
	"lokstra"
	"lokstra/serviceapi/logger_api"
	"lokstra/services/logger"
)

func main() {
	// Register logger service factory with type name "logger"
	// This allows Lokstra to create services of type "main.logger"
	// If domain is omitted, default domain "main" is used
	lokstra.RegisterServiceFactory("logger", logger.ServiceFactory)

	// === Create logger via NewService (Named Service approach) ===
	logger1, _ := lokstra.NewService[logger_api.Logger]("logger", "default")
	logger1.Info("Log created using NewService")

	// === Get logger from service registry (must be created first) ===
	// logger2 is same as logger1, but retrieved using GetService
	logger2 := lokstra.GetService[logger_api.Logger]("logger:default")
	logger2.Info("Log retrieved using GetService")

	if logger1 == logger2 {
		logger1.Info("logger1 and logger2 are the same instance") // This should be true
	} else {
		logger1.Error("logger1 and logger2 are different instances") // This should not happen
	}

	// === Create logger directly from logger package (Direct approach) ===
	logger3 := logger.NewService(logger_api.LogLevelInfo)
	logger3.Info("Log created directly using logger.NewService")
}
