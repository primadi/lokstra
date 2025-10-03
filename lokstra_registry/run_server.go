package lokstra_registry

import "time"

// starts the server based on the provided configuration and server name
func StartServer() {
	serverName := GetCurrentServerName()
	if serverName == "" {
		panic("current server name is not set")
	}
	srv := GetServer(serverName)
	if srv == nil {
		panic("server " + serverName + " not found")
	}
	// Build running client router registry before starting
	buildRunningClientRouterRegistry()
	srv.Start()
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
func RunServer(timeout time.Duration) {
	serverName := GetCurrentServerName()
	if serverName == "" {
		panic("current server name is not set")
	}
	srv := GetServer(serverName)
	if srv == nil {
		panic("server " + serverName + " not found")
	}
	// Build running client router registry before running
	buildRunningClientRouterRegistry()
	srv.Run(timeout)
}
