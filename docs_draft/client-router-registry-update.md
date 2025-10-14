# Client Router Registry Strategy Update

## Overview
The `clientRouterRegistry` has been updated to support multi-server deployments with proper deployment-id isolation.

## Key Changes

### 1. Composite Key Structure
The registry now uses composite keys in format: `routerName@serverName`
- Allows multiple servers to have routers with the same name
- Enables precise router targeting across different servers

### 2. Updated `RegisterClientRouter`
```go
func RegisterClientRouter(routerName, serverName, baseURL, addr string)
```
- Now stores routers using composite key: `routerName@serverName`
- Multiple servers can register routers with same name without conflicts

### 3. Enhanced `GetClientRouter` with Smart Search
```go
func GetClientRouter(routerName string, current *ClientRouter) *ClientRouter
```

**Search Strategy (in order):**

1. **Same Server First**: Searches for router in `currentServerName`
   - Fastest path for local communication
   - Uses direct key lookup: `routerName@currentServerName`

2. **Same Deployment-ID**: If not found locally, searches across all servers with matching `deployment-id`
   - Maintains deployment isolation
   - Returns first match found
   - Enables cross-server communication within same deployment

**Caching**: Results are cached in the `current` parameter for performance

### 4. New `GetClientRouterOnServer` Function
```go
func GetClientRouterOnServer(routerName, serverName string, current *ClientRouter) *ClientRouter
```

**Purpose**: Get a specific router on a named server with deployment-id validation

**Features**:
- Direct server targeting via `routerName@serverName` lookup
- Validates that target server shares same `deployment-id` as current server
- Returns `nil` if server not found or deployment-id mismatch
- Supports caching via `current` parameter

**Use Cases**:
- Explicit routing to specific server instances
- Load balancing across known servers
- Service-to-service communication with server affinity

## Deployment-ID Management

### Getting Current Deployment-ID
```go
func GetCurrentDeploymentId() string
```
- Returns deployment-id of current server
- Used for deployment isolation checks

### How It Works
1. Each `Server` has a `DeploymentID` field
2. Servers with same `DeploymentID` can communicate
3. Servers with different `DeploymentID` are isolated
4. Empty `DeploymentID` allows unrestricted access

## Example Usage

### Basic Router Access
```go
// Get router (searches current server first, then same deployment)
var productClient *ClientRouter
productClient = GetClientRouter("product-api", productClient)
if productClient != nil {
    resp, err := productClient.GET("/products")
}
```

### Targeted Server Access
```go
// Get router from specific server (with deployment-id check)
var orderClient *ClientRouter
orderClient = GetClientRouterOnServer("order-api", "order-server-1", orderClient)
if orderClient != nil {
    resp, err := orderClient.POST("/orders", orderData)
}
```

### Multi-Server Scenario
```go
// Deployment A
RegisterClientRouter("api", "server1", "http://localhost", ":8081")
RegisterClientRouter("api", "server2", "http://localhost", ":8082")

// Deployment B
RegisterClientRouter("api", "server3", "http://localhost", ":8083")

// From server1 (Deployment A):
SetCurrentServerName("server1")
client := GetClientRouter("api", nil)
// Returns: api@server1 (same server)

// If api@server1 doesn't exist, searches server2 (same deployment-id)
// Will NOT return server3 (different deployment-id)
```

## Benefits

1. **Deployment Isolation**: Prevents cross-deployment communication
2. **Smart Routing**: Automatically finds routers using intelligent search
3. **Performance**: Local routers use direct `ServeHTTP` (no network overhead)
4. **Flexibility**: Supports both auto-discovery and explicit targeting
5. **Caching**: Reuses router instances for better performance
6. **Multi-Tenancy**: Same router names can exist across different servers

## Migration Notes

### Breaking Changes
- `clientRouterRegistry` key format changed from `routerName` to `routerName@serverName`
- Old code directly accessing registry will need updates

### Backward Compatibility
- API signatures remain the same
- Existing calls to `RegisterClientRouter` and `GetClientRouter` work as before
- Just need to ensure `DeploymentID` is set on servers for isolation

## Configuration Example

```go
// Setup deployment A
server1 := server.New("server1")
server1.DeploymentID = "deployment-a"
server1.BaseUrl = "http://localhost:8081"

server2 := server.New("server2")
server2.DeploymentID = "deployment-a"
server2.BaseUrl = "http://localhost:8082"

// Setup deployment B
server3 := server.New("server3")
server3.DeploymentID = "deployment-b"
server3.BaseUrl = "http://localhost:8083"

// Register routers
RegisterClientRouter("product-api", "server1", server1.BaseUrl, "/api/v1")
RegisterClientRouter("order-api", "server2", server2.BaseUrl, "/api/v1")
RegisterClientRouter("product-api", "server3", server3.BaseUrl, "/api/v1")

// From server1: can access product-api@server1 and order-api@server2
// From server1: cannot access product-api@server3 (different deployment)
```
