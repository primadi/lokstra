# LoadAndBuild Analysis & Refactoring Plan

## Current State Analysis

### What LoadAndBuild Does:

1. **Load YAML config** (`LoadConfig`)
2. **Register Configs** → `DefineConfig`
3. **Resolve Configs** → `ResolveConfigs`
4. **Register Middlewares** → `DefineMiddleware`
5. **Register Services** → `RegisterLazyService` (✅ sudah diupdate Opsi 2)
6. **Auto-generate Routers** for published services
7. **Register Routers** → `DefineRouter`
8. **Build Deployment Topology** (server → apps → services/routers)
9. **Register Deployments** → `RegisterDeployment`

---

## Question 1: DefineMiddleware vs RegisterMiddlewareName

### Current Implementation:

**DefineMiddleware** (YAML loader):
```go
// schema.MiddlewareDef from YAML
type MiddlewareDef struct {
    Name   string
    Type   string         // Factory type
    Config map[string]any
}

// Stores in g.middlewares map[string]*schema.MiddlewareDef
registry.DefineMiddleware(&schema.MiddlewareDef{
    Name: "logger-debug",
    Type: "logger",
    Config: map[string]any{"level": "debug"},
})
```

**RegisterMiddlewareName** (code API):
```go
// Stores in g.middlewareEntries sync.Map → MiddlewareEntry
registry.RegisterMiddlewareName("logger-debug", "logger", 
    map[string]any{"level": "debug"})
```

### Problem: Dual Storage (Same as services before Opsi 2!)

- `g.middlewares map[string]*schema.MiddlewareDef` - YAML definitions
- `g.middlewareEntries sync.Map` - Runtime entries (from RegisterMiddlewareName)

### Answer: **YES, DefineMiddleware can be removed!**

Similar to Opsi 2 for services, we should:
1. Remove `g.middlewares` map
2. Remove `DefineMiddleware` method
3. Update YAML loader to use `RegisterMiddlewareName`
4. Benefits: Single API for YAML and code

---

## Question 2: DefineRouter - Still Needed?

### Current Router Registration:

**DefineRouter** (used by loader):
```go
registry.DefineRouter("user-service-router", &schema.RouterDef{
    ServiceName:    "user-service",
    Resource:       "User",
    ResourcePlural: "Users",
    Convention:     "rest",
    PathPrefix:     "/api",
    Middlewares:    []string{"auth"},
    Hidden:         []string{"Delete"},
    Custom:         []schema.RouteDef{...},
})
```

**No equivalent in code API** - There's no `RegisterRouter` method!

### Analysis:

Router definitions are special because:
1. They contain **routing metadata** (resource, convention, path prefix)
2. They're **tied to services** (ServiceName field)
3. They're **auto-generated** from service metadata
4. Manual router definitions are **overrides** of auto-generated ones

### Answer: **YES, DefineRouter is still needed**

But we should add a code equivalent:
```go
// Proposed new API
registry.RegisterRouter("user-service-router", &RouterConfig{
    ServiceName:    "user-service",
    Resource:       "User",
    ResourcePlural: "Users",
    Convention:     "rest",
    PathPrefix:     "/api",
    Middlewares:    []string{"auth"},
    Hidden:         []string{"Delete"},
    Custom:         map[string]RouteConfig{...},
})
```

This would:
- Store in `g.routers` (keep existing storage)
- Be used by both YAML loader and code
- Allow manual router configuration without YAML

---

## Question 3: Can LoadAndBuild be done entirely by code?

### Answer: **YES! But needs some new APIs**

Currently missing code APIs:
1. ✅ Service registration - **DONE** (Opsi 2)
2. ❌ Middleware registration from YAML - **Need to update**
3. ❌ Router registration - **Need new API**
4. ❌ Deployment topology - **Need new API**

### What programmer needs to do by code:

```go
// 1. Register configs (already available)
lokstra_registry.DefineConfig("db-host", "localhost")
lokstra_registry.DefineConfig("db-port", "5432")

// 2. Resolve configs (already available)
deploy.Global().ResolveConfigs()

// 3. Register middleware TYPES (already available)
lokstra_registry.RegisterMiddlewareType("logger", loggerFactory)

// 4. Register middleware NAMES (already available) ✅
lokstra_registry.RegisterMiddlewareName("logger-debug", "logger", 
    map[string]any{"level": "debug"})

// 5. Register service factory TYPES (already available)
lokstra_registry.RegisterServiceType("user-service-factory", 
    userServiceFactory, nil)

// 6. Register service INSTANCES (NEW - Opsi 2) ✅
lokstra_registry.RegisterLazyService("user-service", 
    "user-service-factory", 
    map[string]any{
        "depends-on": []string{"user-repository"},
        "max-users": 1000,
    })

// 7. Register ROUTERS (MISSING - Need new API) ❌
// Current: No API, must use YAML or call deploy.Global().DefineRouter()
// Proposed:
lokstra_registry.RegisterRouter("user-service-router", &RouterConfig{
    ServiceName: "user-service",
    Resource: "User",
    // ... etc
})

// 8. Build DEPLOYMENT TOPOLOGY (MISSING - Need new API) ❌
// Current: Must use YAML
// Proposed:
lokstra_registry.RegisterDeployment("local-dev", &DeploymentConfig{
    Servers: map[string]*ServerConfig{
        "main-server": {
            BaseURL: "http://localhost:8080",
            Apps: []*AppConfig{
                {
                    Addr: ":8080",
                    PublishedServices: []string{"user-service"},
                    Routers: []string{"user-service-router"},
                },
            },
        },
    },
})
```

---

## Proposed Refactoring Steps (Opsi 2.5)

### Phase 1: Middleware Unification (Similar to Opsi 2)

1. Remove `g.middlewares map[string]*schema.MiddlewareDef`
2. Remove `DefineMiddleware` method
3. Update YAML loader to use `RegisterMiddlewareName`
4. Add wrapper: `lokstra_registry.RegisterMiddlewareName`

### Phase 2: Add Code APIs

1. Add `RegisterRouter` method (wrapper for DefineRouter)
   ```go
   func (g *GlobalRegistry) RegisterRouter(name string, config *RouterConfig) {
       // Convert RouterConfig to schema.RouterDef
       // Call DefineRouter
   }
   ```

2. Add `RegisterDeployment` method
   ```go
   func (g *GlobalRegistry) RegisterDeployment(name string, config *DeploymentConfig) {
       // Build topology programmatically
       // Store in deployments
   }
   ```

3. Add wrapper functions in `lokstra_registry`:
   - `RegisterRouter`
   - `RegisterDeployment`

### Phase 3: Make LoadAndBuild Optional

After Phase 1-2, programmer can choose:
- **Option A**: Use YAML (`LoadAndBuild`)
- **Option B**: Pure code (call registration APIs manually)
- **Option C**: Hybrid (some YAML, some code)

---

## Benefits of Full Code Support

1. **Type Safety** - Go compiler checks
2. **IDE Support** - Autocomplete, refactoring
3. **Dynamic Configuration** - Can compute values at runtime
4. **Testing** - Easier to test without YAML files
5. **Deployment Flexibility** - Different envs without YAML changes

---

## Recommendation

**Step-by-step approach:**

1. ✅ **DONE**: Opsi 2 (Service unification)
2. **NEXT**: Middleware unification (similar to Opsi 2)
3. **THEN**: Add `RegisterRouter` API
4. **FINALLY**: Add `RegisterDeployment` API

This allows gradual migration and keeps backward compatibility.

---

## Example: Full Code-Based Setup

After all phases complete:

```go
package main

import "github.com/primadi/lokstra/lokstra_registry"

func main() {
    // Configs
    lokstra_registry.DefineConfig("env", "production")
    
    // Middleware types
    lokstra_registry.RegisterMiddlewareType("auth", authFactory)
    lokstra_registry.RegisterMiddlewareName("jwt-auth", "auth", 
        map[string]any{"secret": "xxx"})
    
    // Service types
    lokstra_registry.RegisterServiceType("user-service-factory", 
        userServiceFactory, nil)
    lokstra_registry.RegisterLazyService("user-service", 
        "user-service-factory", nil)
    
    // Routers (NEW)
    lokstra_registry.RegisterRouter("user-service-router", &RouterConfig{
        ServiceName: "user-service",
        Resource: "User",
    })
    
    // Deployment (NEW)
    lokstra_registry.RegisterDeployment("production", &DeploymentConfig{
        Servers: map[string]*ServerConfig{
            "api-server": {
                BaseURL: "https://api.example.com",
                Apps: []*AppConfig{{
                    Addr: ":443",
                    PublishedServices: []string{"user-service"},
                }},
            },
        },
    })
    
    // No LoadAndBuild needed!
    // Just start server...
}
```

This is the ultimate goal! 🎯
