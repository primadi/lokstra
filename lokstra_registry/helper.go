package lokstra_registry

import (
	"time"
)

// RunConfiguredServer initializes and runs the server based on loaded config.
// Must be called after LoadConfig() and service/middleware registration.
//
// This function:
//  1. Reads server selection from config (or uses first server if not specified)
//  2. Reads shutdown timeout from config (default: 30s)
//  3. Runs the selected server
//
// Config keys used:
//   - server: Server composite key "deployment.server" (optional, uses first if not specified)
//   - shutdown_timeout: Graceful shutdown timeout duration (optional, default: 30s)
//
// Example:
//
//	if err := lokstra_registry.RunConfiguredServer(); err != nil {
//	    logger.LogPanic(err)
//	}
func RunConfiguredServer() error {
	server := GetConfig("server", "")

	var timeout time.Duration
	timeoutStr := GetConfig("shutdown_timeout", "30s")
	if dur, err := time.ParseDuration(timeoutStr); err == nil {
		timeout = dur
	} else {
		timeout = 30 * time.Second
	}

	// Run server
	return RunServer(server, timeout)
}

// return runtime mode: dev, debug, or prod
func GetRuntimeMode() string {
	return GetConfig("runtime.mode", "prod")
}
