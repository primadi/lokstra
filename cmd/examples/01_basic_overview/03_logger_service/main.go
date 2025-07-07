package main

import (
	"lokstra"
	"lokstra/serviceapi/logger_api"
	"lokstra/services/logger"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	// Register logger service factory with type name "logger"
	// This allows Lokstra to create services of type "main.logger"
	// If domain is omitted, default domain "main" is used
	ctx.RegisterServiceFactory("logger", logger.ServiceFactory)

	// === Create logger via NewService (Named Service approach) ===
	logger1, _ := logger_api.NewLogger(ctx, "logger", "default", logger_api.LogLevelInfo)
	logger1.Info("Log created using NewService")

	// === Get logger from service registry (must be created first) ===
	// logger2 is same as logger1, but retrieved using GetService
	logger2, err := logger_api.GetLogger(ctx, "logger:default")
	if err != nil {
		logger1.Error("Failed to retrieve logger using GetService", logger_api.Field{
			Key:   "error",
			Value: err})
		return
	}

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
