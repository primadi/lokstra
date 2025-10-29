# Running Client Router Registry Strategy

## Overview
Strategi baru menggunakan `runningClientRouterRegistry` yang di-build saat startup untuk menyimpan router terpilih dari deployment-id yang sama, tanpa menghapus data dari registry asli.

## Strategi Baru vs Lama

### ❌ Strategi Lama (Cleanup)
```go
// Menghapus server dan router dengan deployment-id berbeda
func cleanupDifferentDeployments() {
    for serverName, srv := range serverRegistry {
        if srv.DeploymentID != currentDeploymentID {
            delete(serverRegistry, serverName)  // ← Destructive
        }
    }
}
```
**Masalah**: Data original hilang, tidak bisa recovery

### ✅ Strategi Baru (Build Running Registry)
```go
// Build registry terpisah untuk runtime tanpa menghapus data asli
func buildRunningClientRouterRegistry() {
    // First pass: Prioritas router dari currentServerName
    // Second pass: Router dari server lain dengan deployment-id sama
}
```
**Keuntungan**: 
- Data original tetap utuh di `clientRouterRegistry`
- Runtime menggunakan `runningClientRouterRegistry` yang sudah difilter
- Bisa rebuild jika diperlukan

## Implementasi

### 1. Dua Registry Map

**Lokasi**: `lokstra_registry/client_router.go`

```go
// Original registry - stores ALL registered routers (composite key: routerName@serverName)
var clientRouterRegistry = make(map[string]*ClientRouter)

// Running registry - stores SELECTED routers for runtime (simple key: routerName)
// Built by buildRunningClientRouterRegistry() before Start/Run
var runningClientRouterRegistry = make(map[string]*ClientRouter)
```

**Perbedaan**:
| Registry | Key Format | Purpose | Modified |
|----------|-----------|---------|----------|
| `clientRouterRegistry` | `routerName@serverName` | Store ALL routers | By RegisterClientRouter() |
| `runningClientRouterRegistry` | `routerName` | Runtime lookup | By buildRunningClientRouterRegistry() |

### 2. buildRunningClientRouterRegistry()

**Lokasi**: `lokstra_registry/client_router.go`

```go
func buildRunningClientRouterRegistry() {
    currentDeploymentID := GetCurrentDeploymentId()
    
    if currentDeploymentID == "" {
        // No deployment-id: use all routers
        for _, cr := range clientRouterRegistry {
            if _, exists := runningClientRouterRegistry[cr.routerName]; !exists {
                runningClientRouterRegistry[cr.routerName] = cr
            }
        }
        return
    }

    // Clear existing
    runningClientRouterRegistry = make(map[string]*ClientRouter)

    // Pass 1: Add routers from currentServerName (PRIORITY)
    for _, cr := range clientRouterRegistry {
        if cr.serverName == currentServerName {
            srv := GetServer(cr.serverName)
            if srv != nil && srv.DeploymentID == currentDeploymentID {
                runningClientRouterRegistry[cr.routerName] = cr
            }
        }
    }

    // Pass 2: Add routers from other servers (same deployment-id)
    for _, cr := range clientRouterRegistry {
        // Skip if already added
        if _, exists := runningClientRouterRegistry[cr.routerName]; exists {
            continue
        }

        srv := GetServer(cr.serverName)
        if srv != nil && srv.DeploymentID == currentDeploymentID {
            runningClientRouterRegistry[cr.routerName] = cr
        }
    }
}
```

**Priority Logic**:
1. **First**: Router dari `currentServerName` dengan `deployment-id` sama
2. **Second**: Router dari server lain (first match) dengan `deployment-id` sama
3. **Result**: Satu router per `routerName` dalam `runningClientRouterRegistry`

### 3. GetClientRouter() - Simplified

**Lokasi**: `lokstra_registry/client_router.go`

```go
func GetClientRouter(routerName string, current *ClientRouter) *ClientRouter {
    // Cache hit
    if current != nil && current.routerName == routerName {
        return current
    }

    // Direct O(1) lookup from running registry
    cr, exists := runningClientRouterRegistry[routerName]
    if !exists {
        return nil
    }

    return cr
}
```

**Keuntungan**:
- ✅ **O(1) lookup** langsung dari map
- ✅ Tidak ada iteration
- ✅ Tidak ada conditional checking
- ✅ Super fast

### 4. GetClientRouterOnServer() - Target Specific Server

**Lokasi**: `lokstra_registry/client_router.go`

```go
func GetClientRouterOnServer(routerName, serverName string, current *ClientRouter) *ClientRouter {
    // Cache hit
    if current != nil && current.routerName == routerName && current.serverName == serverName {
        return current
    }

    // Check if router exists in running registry (deployment-id check)
    runningCr, exists := runningClientRouterRegistry[routerName]
    if !exists {
        return nil  // Router not in same deployment
    }

    // If already on target server
    if runningCr.serverName == serverName {
        return runningCr
    }

    // Look for specific server in full registry
    key := routerName + "@" + serverName
    cr, exists := clientRouterRegistry[key]
    if !exists {
        return nil
    }

    // Verify deployment-id
    currentDeploymentID := GetCurrentDeploymentId()
    if currentDeploymentID != "" {
        srv := GetServer(serverName)
        if srv == nil || srv.DeploymentID != currentDeploymentID {
            return nil
        }
    }

    return cr
}
```

**Logic**:
1. Cek apakah router ada di `runningClientRouterRegistry` (validasi deployment)
2. Jika running router sudah di target server, return langsung
3. Cari di `clientRouterRegistry` dengan key spesifik
4. Validasi deployment-id server target

### 5. Integrasi dengan StartServer() dan RunServer()

**Lokasi**: `lokstra_registry/run_server.go`

```go
func StartServer() {
    serverName := GetCurrentServerName()
    srv := GetServer(serverName)
    
    // Build running registry before start
    buildRunningClientRouterRegistry()
    
    srv.Start()
}

func RunServer(timeout time.Duration) {
    serverName := GetCurrentServerName()
    srv := GetServer(serverName)
    
    // Build running registry before run
    buildRunningClientRouterRegistry()
    
    if err := srv.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

## Example Scenario

### Setup: Multi-Deployment Environment

```go
// Deployment A
server1 := server.New("server1")
server1.DeploymentID = "deployment-a"
server1.BaseUrl = "http://localhost:8081"
RegisterServer("server1", server1)

server2 := server.New("server2")
server2.DeploymentID = "deployment-a"
server2.BaseUrl = "http://localhost:8082"
RegisterServer("server2", server2)

// Deployment B
server3 := server.New("server3")
server3.DeploymentID = "deployment-b"
server3.BaseUrl = "http://localhost:8083"
RegisterServer("server3", server3)

// Register client routers
RegisterClientRouter("product-api", "server1", "http://localhost:8081", "/api/v1")
RegisterClientRouter("order-api", "server2", "http://localhost:8082", "/api/v1")
RegisterClientRouter("product-api", "server3", "http://localhost:8083", "/api/v1")

// Set current server
SetCurrentServerName("server1")
```

### Before buildRunningClientRouterRegistry()

```go
clientRouterRegistry = {
    "product-api@server1": {routerName: "product-api", serverName: "server1"},
    "order-api@server2":   {routerName: "order-api", serverName: "server2"},
    "product-api@server3": {routerName: "product-api", serverName: "server3"},
}

runningClientRouterRegistry = {} // Empty
```

### After buildRunningClientRouterRegistry()

```go
clientRouterRegistry = {
    // UNCHANGED - all data preserved
    "product-api@server1": {routerName: "product-api", serverName: "server1"},
    "order-api@server2":   {routerName: "order-api", serverName: "server2"},
    "product-api@server3": {routerName: "product-api", serverName: "server3"},
}

runningClientRouterRegistry = {
    // Only deployment-a routers
    "product-api": {routerName: "product-api", serverName: "server1"}, // ← Priority: current server
    "order-api":   {routerName: "order-api", serverName: "server2"},
    // product-api@server3 NOT included (different deployment)
}
```

### Runtime Lookup

```go
// From server1 (deployment-a)
client := GetClientRouter("product-api", nil)
// Returns: product-api@server1 (from runningClientRouterRegistry)
// Fast O(1) lookup, no iteration

client := GetClientRouter("order-api", nil)
// Returns: order-api@server2 (same deployment)

client := GetClientRouter("payment-api", nil)
// Returns: nil (not in runningClientRouterRegistry)
```

## Priority Selection Example

```go
// Multiple routers with same name
RegisterClientRouter("api", "server1", "http://localhost:8081", "/v1")  // current
RegisterClientRouter("api", "server2", "http://localhost:8082", "/v1")  // same deployment
RegisterClientRouter("api", "server3", "http://localhost:8083", "/v1")  // same deployment

SetCurrentServerName("server1")
buildRunningClientRouterRegistry()

// Result in runningClientRouterRegistry:
runningClientRouterRegistry["api"] = api@server1  // ← Current server wins
```

**Priority Rule**: Current server ALWAYS wins when multiple routers have same name

## Benefits

### 1. 🔒 Data Preservation
- ✅ Original `clientRouterRegistry` tetap utuh
- ✅ Bisa rebuild `runningClientRouterRegistry` kapan saja
- ✅ Tidak ada data loss

### 2. 🚀 Performance
- ✅ `GetClientRouter()` = O(1) direct map lookup
- ✅ Tidak ada iteration di runtime
- ✅ Tidak ada conditional checking di setiap request

### 3. 🎯 Clear Separation
- ✅ `clientRouterRegistry` = source of truth (composite key)
- ✅ `runningClientRouterRegistry` = runtime cache (simple key)
- ✅ Easy to understand dan maintain

### 4. 🔄 Flexible
- ✅ Bisa rebuild registry jika configuration berubah
- ✅ Support dynamic reconfiguration (future)
- ✅ Easy to test (just rebuild)

### 5. 🐛 Easier Debugging
- ✅ Bisa inspect kedua registry terpisah
- ✅ Lihat apa yang registered vs apa yang running
- ✅ Clear priority logic

## Performance Comparison

| Operation | Old Strategy | New Strategy |
|-----------|-------------|--------------|
| Startup | O(M) - delete servers | O(M) - build running registry |
| GetClientRouter | O(M) - iterate servers | O(1) - map lookup |
| Memory | Less (deleted data) | More (keep all data) |
| Data Loss | Yes (servers deleted) | No (preserved) |
| Rebuild | Cannot | Can rebuild anytime |

**M** = number of servers

## Migration Notes

### Non-Breaking Changes
- ✅ API signature tidak berubah
- ✅ `GetClientRouter()` dan `GetClientRouterOnServer()` tetap sama
- ✅ Existing code tetap work

### Internal Changes
- Registry internal structure berubah (2 maps instead of 1)
- Build process di startup instead of cleanup

### Testing
```go
func TestBuildRunningClientRouterRegistry(t *testing.T) {
    // Setup multi-deployment
    // Call buildRunningClientRouterRegistry()
    // Verify runningClientRouterRegistry only has same deployment
    // Verify clientRouterRegistry unchanged
}
```

## Best Practices

1. **Always call before Start/Run**: 
   ```go
   buildRunningClientRouterRegistry()
   srv.Start()
   ```

2. **Set current server first**:
   ```go
   SetCurrentServerName("server1")
   buildRunningClientRouterRegistry()  // Uses current server
   ```

3. **Register all routers before build**:
   ```go
   RegisterClientRouter("api1", ...)
   RegisterClientRouter("api2", ...)
   buildRunningClientRouterRegistry()  // Build after all registered
   ```

4. **Don't modify after build** (race condition):
   ```go
   buildRunningClientRouterRegistry()
   // Don't call RegisterClientRouter() after this
   srv.Start()
   ```

## Summary

Strategi baru menggunakan **two-registry approach**:
- ✅ `clientRouterRegistry`: Store ALL routers (preserved)
- ✅ `runningClientRouterRegistry`: Runtime cache (filtered by deployment-id)
- ✅ Priority: Current server > Other servers (same deployment)
- ✅ O(1) lookup di runtime
- ✅ No data loss
- ✅ Rebuildable

