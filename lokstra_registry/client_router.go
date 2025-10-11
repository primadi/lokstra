package lokstra_registry

import (
	"sync"
	"time"

	"github.com/primadi/lokstra/api_client"
)

// clientRouterRegistry now uses composite key: routerName@serverName
var clientRouterRegistry sync.Map

// runningClientRouterRegistry maps routerName to selected *ClientRouter for runtime use
// Built by buildRunningClientRouterRegistry() before Start/Run
var runningClientRouterRegistry sync.Map

var currentServerName = ""

// sets the name of the currently running server
func SetCurrentServerName(serverName string) {
	currentServerName = serverName
}

// gets the name of the currently running server
func GetCurrentServerName() string {
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
	runningClientRouterRegistry = sync.Map{}

	currentSrv := currentServerName

	// First pass: Add routers from currentServerName (priority)
	clientRouterRegistry.Range(func(key, value any) bool {
		cr := value.(*api_client.ClientRouter)
		if cr.ServerName == currentSrv {
			// If server is registered, check deployment-id
			// If not registered yet (early call), add anyway since RegisterClientRouter
			// only called for current deployment
			srv := GetServer(cr.ServerName)
			if srv == nil || srv.DeploymentID == currentDeploymentID || srv.DeploymentID == "" {
				runningClientRouterRegistry.Store(cr.RouterName, cr)
			}
		}
		return true
	})

	// Second pass: Add routers from other servers with same deployment-id
	clientRouterRegistry.Range(func(key, value any) bool {
		cr := value.(*api_client.ClientRouter)

		// Check if already added in first pass
		if _, exists := runningClientRouterRegistry.Load(cr.RouterName); exists {
			return true
		}

		// Check if server has same deployment-id
		// If server not registered yet, add anyway (will be from same deployment)
		srv := GetServer(cr.ServerName)
		if srv == nil || srv.DeploymentID == currentDeploymentID || srv.DeploymentID == "" {
			runningClientRouterRegistry.Store(cr.RouterName, cr)
		}
		return true
	})
}

// registers where a router can be accessed
// Uses composite key: routerName@serverName to allow multiple servers to have same router name
func RegisterClientRouter(routerName, serverName, baseURL, addr string, timeout time.Duration) {
	isLocal := (serverName == currentServerName)

	cr := &api_client.ClientRouter{
		RouterName: routerName,
		ServerName: serverName,
		FullURL:    baseURL + addr,
		IsLocal:    isLocal,
		Timeout:    timeout,
	}

	// If it's local, try to get the actual router instance
	if isLocal {
		if localRouter := GetRouter(routerName); localRouter != nil {
			cr.Router = localRouter
		}
	}

	// Use composite key: routerName@serverName
	key := routerName + "@" + serverName

	clientRouterRegistry.Store(key, cr)
}

// gets a client to communicate with a router (local or remote)
// Uses runningClientRouterRegistry which is pre-built with deployment-id filtering
// and priority selection (currentServer first, then other servers with same deployment-id)
// Note: buildRunningClientRouterRegistry() should be called before Start/Run
func GetClientRouter(routerName string) *api_client.ClientRouter {
	// Direct lookup from running registry (O(1))
	crAny, exists := runningClientRouterRegistry.Load(routerName)

	// DEBUG: List available routers if not found
	if !exists {
		// runningClientRouterMutex.RLock()
		// availableRouters := make([]string, 0, len(runningClientRouterRegistry))
		// for name := range runningClientRouterRegistry {
		// 	availableRouters = append(availableRouters, name)
		// }
		// runningClientRouterMutex.RUnlock()
		// fmt.Printf("[DEBUG GetClientRouter] Router '%s' not found. Available: %v\n", routerName, availableRouters)
		return nil
	}

	return crAny.(*api_client.ClientRouter)
}

// GetClientRouterCached gets a client with optional caching (legacy compatibility)
// Deprecated: Use GetClientRouter instead, caching is not needed for one-time service registration
func GetClientRouterCached(routerName string, current *api_client.ClientRouter) *api_client.ClientRouter {
	// If current is provided and matches, reuse it (cache hit)
	if current != nil && current.RouterName == routerName {
		return current
	}

	return GetClientRouter(routerName)
}

// GetClientRouterOnServer gets a ClientRouter on a specific server
// This allows explicit server targeting
// Only returns router if it's in the runningClientRouterRegistry (same deployment-id)
func GetClientRouterOnServer(routerName, serverName string) *api_client.ClientRouter {
	// Check if router exists in running registry first (deployment-id validation)
	runningCrAny, exists := runningClientRouterRegistry.Load(routerName)

	if !exists {
		return nil
	}

	runningCr := runningCrAny.(*api_client.ClientRouter)

	// If the running router is already on the target server, return it
	if runningCr.ServerName == serverName {
		return runningCr
	}

	// Look for specific server in full registry
	key := routerName + "@" + serverName

	crAny, exists := clientRouterRegistry.Load(key)

	if !exists {
		return nil
	}

	cr := crAny.(*api_client.ClientRouter)

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
