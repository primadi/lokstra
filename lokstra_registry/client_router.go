package lokstra_registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/primadi/lokstra/core/router"
)

// ClientRouter stores information about where a router can be accessed
type ClientRouter struct {
	routerName string
	serverName string
	fullURL    string
	isLocal    bool
	router     router.Router
}

// clientRouterRegistry now uses composite key: routerName@serverName
var clientRouterRegistry = make(map[string]*ClientRouter)

// runningClientRouterRegistry maps routerName to selected *ClientRouter for runtime use
// Built by buildRunningClientRouterRegistry() before Start/Run
var runningClientRouterRegistry = make(map[string]*ClientRouter)
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
	runningClientRouterRegistry = make(map[string]*ClientRouter)

	// First pass: Add routers from currentServerName (priority)
	for _, cr := range clientRouterRegistry {
		if cr.serverName == currentServerName {
			srv := GetServer(cr.serverName)
			if srv != nil && srv.DeploymentID == currentDeploymentID {
				runningClientRouterRegistry[cr.routerName] = cr
			}
		}
	}

	// Second pass: Add routers from other servers with same deployment-id
	for _, cr := range clientRouterRegistry {
		// Skip if already added in first pass
		if _, exists := runningClientRouterRegistry[cr.routerName]; exists {
			continue
		}

		// Check if server has same deployment-id
		srv := GetServer(cr.serverName)
		if srv != nil && srv.DeploymentID == currentDeploymentID {
			runningClientRouterRegistry[cr.routerName] = cr
		}
	}
}

// registers where a router can be accessed
// Uses composite key: routerName@serverName to allow multiple servers to have same router name
func RegisterClientRouter(routerName, serverName, baseURL, addr string) {
	isLocal := (serverName == currentServerName)

	cr := &ClientRouter{
		routerName: routerName,
		serverName: serverName,
		fullURL:    baseURL + addr,
		isLocal:    isLocal,
	}

	// If it's local, try to get the actual router instance
	if isLocal {
		if localRouter, exists := routerRegistry[routerName]; exists {
			cr.router = localRouter
		}
	}

	// Use composite key: routerName@serverName
	key := routerName + "@" + serverName
	clientRouterRegistry[key] = cr
}

// gets a client to communicate with a router (local or remote)
// Uses runningClientRouterRegistry which is pre-built with deployment-id filtering
// and priority selection (currentServer first, then other servers with same deployment-id)
// Results are cached in current parameter
// Note: buildRunningClientRouterRegistry() should be called before Start/Run
func GetClientRouter(routerName string, current *ClientRouter) *ClientRouter {
	// If current is provided and matches, reuse it (cache hit)
	if current != nil && current.routerName == routerName {
		return current
	}

	// Direct lookup from running registry (O(1))
	cr, exists := runningClientRouterRegistry[routerName]
	if !exists {
		return nil
	}

	return cr
}

// GetClientRouterOnServer gets a ClientRouter on a specific server
// This allows explicit server targeting
// Only returns router if it's in the runningClientRouterRegistry (same deployment-id)
func GetClientRouterOnServer(routerName, serverName string, current *ClientRouter) *ClientRouter {
	// If current is provided and matches both router and server, reuse it (cache hit)
	if current != nil && current.routerName == routerName && current.serverName == serverName {
		return current
	}

	// Check if router exists in running registry first (deployment-id validation)
	runningCr, exists := runningClientRouterRegistry[routerName]
	if !exists {
		return nil
	}

	// If the running router is already on the target server, return it
	if runningCr.serverName == serverName {
		return runningCr
	}

	// Look for specific server in full registry
	key := routerName + "@" + serverName
	cr, exists := clientRouterRegistry[key]
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

// performs a GET request to the router
func (c *ClientRouter) GET(path string) (*http.Response, error) {
	return c.makeRequest("GET", path, nil)
}

// performs a POST request to the router
func (c *ClientRouter) POST(path string, body any) (*http.Response, error) {
	return c.makeRequest("POST", path, body)
}

// makeRequest handles both local (router.ServeHTTP) and remote (HTTP) calls
func (c *ClientRouter) makeRequest(method, path string, body any) (*http.Response, error) {
	if c.isLocal && c.router != nil {
		// Use router.ServeHTTP for same-server communication (faster than httptest)
		return c.makeLocalRequest(method, path, body)
	}
	// Use HTTP for remote communication
	return c.makeRemoteRequest(method, path, body)
}

// makeLocalRequest uses router.ServeHTTP for zero-overhead local calls
func (c *ClientRouter) makeLocalRequest(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Use router.ServeHTTP directly (faster than httptest roundtrip)
	c.router.ServeHTTP(w, req)

	return w.Result(), nil
}

// makeRemoteRequest uses standard HTTP client for remote calls
func (c *ClientRouter) makeRemoteRequest(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	url := c.fullURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make HTTP call with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client.Do(req)
}

// parses HTTP response body to target struct
func ParseJSONResponse[T any](resp *http.Response, target *T) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, target)
}
