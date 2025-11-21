# Lokstra Bootstrap Flows

This document explains the two approaches for bootstrapping a Lokstra application.

## Table of Contents

1. [Flow Comparison](#flow-comparison)
2. [Old Flow (Current)](#old-flow-current)
3. [New Flow (Recommended)](#new-flow-recommended)
4. [Migration Guide](#migration-guide)
5. [Examples](#examples)

---

## Flow Comparison

### Old Flow (Current)

```
1. RegisterServiceTypes()       // Using code only
2. RegisterMiddlewareTypes()    // Using code only
3. RunServerFromConfig()        // Load config + start server
   └─ LoadAndBuild()           // Register lazy services from YAML
   └─ RunServer()              // Start server
```

**Problems:**
- ❌ Config not available during service/middleware registration
- ❌ Services that need config must use lazy loading
- ❌ Config validation happens late (at server start)

### New Flow (Recommended)

```
1. LoadConfig()                // Load YAML config
   └─ LoadAndBuild()          // Register lazy services from YAML
2. RegisterServiceTypes()      // Config is available here!
3. RegisterMiddlewareTypes()   // Config is available here!
4. InitAndRunServer()          // Start server only
```

**Benefits:**
- ✅ Config available during service/middleware registration
- ✅ Early config validation
- ✅ Services can read config in factories without lazy loading
- ✅ More intuitive and easier to debug

---

## Old Flow (Current)

### Code Pattern

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra.Bootstrap()
    deploy.SetLogLevelFromEnv()
    
    // 1. Register service types (no config available)
    registerServiceTypes()
    
    // 2. Register middleware types (no config available)
    registerMiddlewareTypes()
    
    // 3. Load config + run server
    lokstra_registry.RunServerFromConfigFolder("config")
}
```

### Service Factory (Old Flow)

```go
// Problem: Can't access config during registration
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    // config parameter only contains service-level config from YAML
    // Cannot access global config like database.dsn
    
    return &UserServiceImpl{
        UserRepo: service.Cast[UserRepository](deps["user-repository"]),
    }
}
```

---

## New Flow (Recommended)

### Code Pattern

```go
package main

import (
    "log"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra.Bootstrap()
    deploy.SetLogLevelFromEnv()
    
    // 1. Load config FIRST
    if err := lokstra_registry.LoadConfigFromFolder("config"); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // 2. Register service types (config IS available now!)
    registerServiceTypes()
    
    // 3. Register middleware types (config IS available now!)
    registerMiddlewareTypes()
    
    // 4. Initialize and run server
    if err := lokstra_registry.InitAndRunServer(); err != nil {
        log.Fatal("Failed to run server:", err)
    }
}
```

### Service Factory (New Flow)

```go
import "github.com/primadi/lokstra/lokstra_registry"

// Solution: Access global config during registration
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    // Now you can access global config!
    dbDSN := lokstra_registry.GetConfig("database.dsn", "postgres://localhost/mydb")
    cacheEnabled := lokstra_registry.GetConfig("cache.enabled", true)
    
    log.Printf("Creating UserService with DSN: %s, Cache: %v", dbDSN, cacheEnabled)
    
    return &UserServiceImpl{
        UserRepo: service.Cast[UserRepository](deps["user-repository"]),
        CacheEnabled: cacheEnabled,
    }
}
```

---

## Migration Guide

### Step 1: Update main.go

**Before (Old Flow):**
```go
func main() {
    lokstra.Bootstrap()
    deploy.SetLogLevelFromEnv()
    
    registerServiceTypes()
    registerMiddlewareTypes()
    
    lokstra_registry.RunServerFromConfigFolder("config")
}
```

**After (New Flow):**
```go
func main() {
    lokstra.Bootstrap()
    deploy.SetLogLevelFromEnv()
    
    // Load config first
    if err := lokstra_registry.LoadConfigFromFolder("config"); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Register services (config available)
    registerServiceTypes()
    registerMiddlewareTypes()
    
    // Start server
    if err := lokstra_registry.InitAndRunServer(); err != nil {
        log.Fatal("Failed to run server:", err)
    }
}
```

### Step 2: Update Service Factories (Optional)

You can now access global config in your factories:

```go
func MyServiceFactory(deps map[string]any, config map[string]any) any {
    // Access global config from YAML
    apiKey := lokstra_registry.GetConfig("external.api_key", "")
    timeout := lokstra_registry.GetConfig("external.timeout", 30)
    
    return &MyServiceImpl{
        APIKey: apiKey,
        Timeout: time.Duration(timeout) * time.Second,
    }
}
```

### Step 3: Add Global Config to YAML (Optional)

You can now use a global configs section:

```yaml
# config/app.yaml
configs:
  database:
    dsn: "postgres://localhost:5432/mydb"
  cache:
    enabled: true
    ttl: 300
  external:
    api_key: "${API_KEY}"
    timeout: 30

service-definitions:
  user-service:
    type: user-service-factory
    # Service-level config still works
    config:
      some_service_specific_setting: "value"
```

---

## Examples

### Example 1: Simple App

```go
package main

import (
    "log"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra.Bootstrap()
    
    // Load single config file
    if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
        log.Fatal(err)
    }
    
    registerServiceTypes()
    
    if err := lokstra_registry.InitAndRunServer(); err != nil {
        log.Fatal(err)
    }
}
```

### Example 2: Multiple Config Files

```go
package main

import (
    "log"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra.Bootstrap()
    
    // Load multiple config files
    if err := lokstra_registry.LoadConfig(
        "config/base.yaml",
        "config/services.yaml",
        "config/deployments.yaml",
    ); err != nil {
        log.Fatal(err)
    }
    
    registerServiceTypes()
    registerMiddlewareTypes()
    
    if err := lokstra_registry.InitAndRunServer(); err != nil {
        log.Fatal(err)
    }
}
```

### Example 3: Config Folder

```go
package main

import (
    "log"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra.Bootstrap()
    
    // Load all YAML files from folder
    if err := lokstra_registry.LoadConfigFromFolder("config"); err != nil {
        log.Fatal(err)
    }
    
    registerServiceTypes()
    registerMiddlewareTypes()
    
    if err := lokstra_registry.InitAndRunServer(); err != nil {
        log.Fatal(err)
    }
}
```

### Example 4: Accessing Config in Service Factory

```go
import "github.com/primadi/lokstra/lokstra_registry"

func DatabaseServiceFactory(deps map[string]any, config map[string]any) any {
    // Get from global config
    dsn := lokstra_registry.GetConfig("database.dsn", "postgres://localhost/mydb")
    maxConns := lokstra_registry.GetConfig("database.max_connections", 25)
    
    // Get from service-level config (still works!)
    poolSize := 10
    if ps, ok := config["pool_size"].(int); ok {
        poolSize = ps
    }
    
    log.Printf("Creating database service: DSN=%s, MaxConns=%d, PoolSize=%d", 
        dsn, maxConns, poolSize)
    
    return &DatabaseService{
        DSN: dsn,
        MaxConnections: maxConns,
        PoolSize: poolSize,
    }
}
```

---

## API Reference

### New Functions

#### `LoadConfig(configPaths ...string) error`

Loads YAML configuration file(s) and registers lazy load services.

```go
// Single file
err := lokstra_registry.LoadConfig("config.yaml")

// Multiple files
err := lokstra_registry.LoadConfig(
    "config/base.yaml",
    "config/services.yaml",
)
```

#### `LoadConfigFromFolder(configFolder string) error`

Loads all YAML files from the specified folder.

```go
err := lokstra_registry.LoadConfigFromFolder("config")
```

#### `InitAndRunServer() error`

Initializes and runs the server based on loaded config.

```go
// Reads these config keys:
// - server: Server selection (optional, uses first if not specified)
// - shutdown_timeout: Graceful shutdown timeout (optional, default: 30s)

err := lokstra_registry.InitAndRunServer()
```

#### `GetConfig[T any](key string, defaultValue T) T`

Retrieves a configuration value with type safety.

```go
// String
dsn := lokstra_registry.GetConfig("database.dsn", "postgres://localhost/mydb")

// Int
maxConns := lokstra_registry.GetConfig("database.max_connections", 25)

// Bool
cacheEnabled := lokstra_registry.GetConfig("cache.enabled", true)

// Duration (as int seconds)
timeoutSec := lokstra_registry.GetConfig("timeout", 30)
timeout := time.Duration(timeoutSec) * time.Second
```

---

## FAQ

### Q: Do I need to migrate to the new flow?

**A:** No, the old flow still works. However, the new flow is recommended for new projects and provides better developer experience.

### Q: Can I use both flows in the same project?

**A:** No, choose one flow per project. They are mutually exclusive.

### Q: What if I don't need to access config in my services?

**A:** The new flow is still recommended for better organization and early config validation.

### Q: Does this change how YAML config works?

**A:** No, YAML structure remains the same. You can now optionally add a `configs` section for global values.

### Q: What about environment variables?

**A:** Environment variable substitution (e.g., `${DATABASE_URL}`) still works in both flows.

---

## Conclusion

The **new flow** provides better separation of concerns and makes config available during service registration. It's the recommended approach for all new projects.

**Key Takeaway:**

```
Old: Register → Load Config → Run
New: Load Config → Register → Run ✅
```
