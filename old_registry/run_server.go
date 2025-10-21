package old_registry

import (
	"fmt"
	"time"
)

// starts the server based on the provided configuration and server name
func StartServer() error {
	serverName := GetCurrentServerName()
	if serverName == "" {
		return fmt.Errorf("current server name is not set")
	}
	srv := GetServer(serverName)
	if srv == nil {
		return fmt.Errorf("server %s not found", serverName)
	}
	// Build running client router registry before starting
	buildRunningClientRouterRegistry()
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// prints server start information for the specified server
func PrintServerStartInfo() {
	serverName := GetCurrentServerName()
	if serverName == "" {
		panic("current server name is not set")
	}
	srv := GetServer(serverName)
	if srv == nil {
		panic("server " + serverName + " not found")
	}
	srv.PrintStartInfo()
}

// runs the server with a specified timeout for graceful shutdown
func RunServer(timeout time.Duration) error {
	serverName := GetCurrentServerName()
	if serverName == "" {
		return fmt.Errorf("current server name is not set")
	}
	srv := GetServer(serverName)
	if srv == nil {
		return fmt.Errorf("server %s not found", serverName)
	}
	// Build running client router registry before running
	buildRunningClientRouterRegistry()
	return srv.Run(timeout)
}
