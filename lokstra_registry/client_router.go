package lokstra_registry

import (
	"sync"
	"time"

	"github.com/primadi/lokstra/api_client"
)

// clientRouterRegistry now uses composite key: routerName@serverName
var clientRouterRegistry = make(map[string]*api_client.ClientRouter)
var clientRouterMutex sync.RWMutex

// runningClientRouterRegistry maps routerName to selected *ClientRouter for runtime use
// Built by buildRunningClientRouterRegistry() before Start/Run
var runningClientRouterRegistry = make(map[string]*api_client.ClientRouter)
var runningClientRouterMutex sync.RWMutex

var currentServerName = ""
var currentServerMutex sync.RWMutex

// sets the name of the currently running server
func SetCurrentServerName(serverName string) {
	currentServerMutex.Lock()
	defer currentServerMutex.Unlock()
	currentServerName = serverName
}

// gets the name of the currently running server
func GetCurrentServerName() string {
	currentServerMutex.RLock()
	defer currentServerMutex.RUnlock()
	return currentServerName
}

// gets the deployment ID of the currently running server
func GetCurrentDeploymentId() string {
	srv := GetServer(currentServerName)
	if srv == nil {
		return ""
	}
	return srv.DeploymentID
}

// buildRunningClientRouterRegistry builds the runtime router registry
// by selecting routers from same deployment-id with priority:
// 1. Router from currentServerName (highest priority)
// 2. Router from first server found with same deployment-id
// This should be called before Start() or Run()
func buildRunningClientRouterRegistry() {
	currentDeploymentID := GetCurrentDeploymentId()

	// Clear existing running registry
	runningClientRouterMutex.Lock()
	runningClientRouterRegistry = make(map[string]*api_client.ClientRouter)
	runningClientRouterMutex.Unlock()

	currentServerMutex.RLock()
	currentSrv := currentServerName
	currentServerMutex.RUnlock()

	// First pass: Add routers from currentServerName (priority)
	clientRouterMutex.RLock()
	for _, cr := range clientRouterRegistry {
		if cr.ServerName == currentSrv {
			srv := GetServer(cr.ServerName)
			if srv != nil && srv.DeploymentID == currentDeploymentID {
				runningClientRouterMutex.Lock()
				runningClientRouterRegistry[cr.RouterName] = cr
				runningClientRouterMutex.Unlock()
			}
		}
	}

	// Second pass: Add routers from other servers with same deployment-id
	for _, cr := range clientRouterRegistry {
		// Check if already added in first pass
		runningClientRouterMutex.RLock()
		_, exists := runningClientRouterRegistry[cr.RouterName]
		runningClientRouterMutex.RUnlock()

		if exists {
			continue
		}

		// Check if server has same deployment-id
		srv := GetServer(cr.ServerName)
		if srv != nil && srv.DeploymentID == currentDeploymentID {
			runningClientRouterMutex.Lock()
			runningClientRouterRegistry[cr.RouterName] = cr
			runningClientRouterMutex.Unlock()
		}
	}
	clientRouterMutex.RUnlock()
}

// registers where a router can be accessed
// Uses composite key: routerName@serverName to allow multiple servers to have same router name
func RegisterClientRouter(routerName, serverName, baseURL, addr string, timeout time.Duration) {
	currentServerMutex.RLock()
	isLocal := (serverName == currentServerName)
	currentServerMutex.RUnlock()

	cr := &api_client.ClientRouter{
		RouterName: routerName,
		ServerName: serverName,
		FullURL:    baseURL + addr,
		IsLocal:    isLocal,
		Timeout:    timeout,
	}

	// If it's local, try to get the actual router instance
	if isLocal {
		if localRouter, exists := routerRegistry[routerName]; exists {
			cr.Router = localRouter
		}
	}

	// Use composite key: routerName@serverName
	key := routerName + "@" + serverName

	clientRouterMutex.Lock()
	defer clientRouterMutex.Unlock()
	clientRouterRegistry[key] = cr
}

// gets a client to communicate with a router (local or remote)
// Uses runningClientRouterRegistry which is pre-built with deployment-id filtering
// and priority selection (currentServer first, then other servers with same deployment-id)
// Results are cached in current parameter
// Note: buildRunningClientRouterRegistry() should be called before Start/Run
func GetClientRouter(routerName string, current *api_client.ClientRouter) *api_client.ClientRouter {
	// If current is provided and matches, reuse it (cache hit)
	if current != nil && current.RouterName == routerName {
		return current
	}

	// Direct lookup from running registry (O(1))
	runningClientRouterMutex.RLock()
	cr, exists := runningClientRouterRegistry[routerName]
	runningClientRouterMutex.RUnlock()

	if !exists {
		return nil
	}

	return cr
}

// GetClientRouterOnServer gets a ClientRouter on a specific server
// This allows explicit server targeting
// Only returns router if it's in the runningClientRouterRegistry (same deployment-id)
func GetClientRouterOnServer(routerName, serverName string, current *api_client.ClientRouter) *api_client.ClientRouter {
	// If current is provided and matches both router and server, reuse it (cache hit)
	if current != nil && current.RouterName == routerName && current.ServerName == serverName {
		return current
	}

	// Check if router exists in running registry first (deployment-id validation)
	runningClientRouterMutex.RLock()
	runningCr, exists := runningClientRouterRegistry[routerName]
	runningClientRouterMutex.RUnlock()

	if !exists {
		return nil
	}

	// If the running router is already on the target server, return it
	if runningCr.ServerName == serverName {
		return runningCr
	}

	// Look for specific server in full registry
	key := routerName + "@" + serverName

	clientRouterMutex.RLock()
	cr, exists := clientRouterRegistry[key]
	clientRouterMutex.RUnlock()

	if !exists {
		return nil
	}

	// Verify this server has same deployment-id
	currentDeploymentID := GetCurrentDeploymentId()
	if currentDeploymentID != "" {
		srv := GetServer(serverName)
		if srv == nil || srv.DeploymentID != currentDeploymentID {
			return nil
		}
	}

	return cr
}
