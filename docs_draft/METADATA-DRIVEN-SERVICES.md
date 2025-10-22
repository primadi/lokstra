# Metadata-Driven Service Registration

## Overview

This document describes the **metadata-driven service registration** pattern in Lokstra framework. This pattern provides a **declarative, type-safe** way to register services with automatic router generation and smart remote service proxy creation.

## Problem Statement

### Before (Manual Configuration):

**Remote Factory (Complex):**
```go
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    baseURL := config["base-url"].(string)
    
    // Manual proxy.Service creation
    service := proxy.NewService(
        baseURL,
        autogen.ConversionRule{
            Convention:     "rest",
            Resource:       "user",
            ResourcePlural: "users",
        },
        autogen.RouteOverride{},
    )
    
    return &UserServiceRemote{service: service}
}
```

**Issues:**
- ❌ Resource metadata duplicated in every remote factory
- ❌ Convention hardcoded in multiple places
- ❌ Inconsistent with local factory pattern
- ❌ No central source of truth for service metadata
- ❌ Router definitions must be written manually

## Solution: Metadata Registration Pattern

### 1. Register Service with Metadata

```go
lokstra_registry.RegisterServiceType("user-service-factory", 
    UserServiceFactory,      // Local factory
    UserServiceRemoteFactory, // Remote factory
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)
```

**Metadata Options:**
- `WithResource(singular, plural)` - Resource name for REST endpoints
- `WithConvention(type)` - Convention type: "rest", "rpc", "graphql"
- `WithRouteOverride(method, path)` - Custom route paths
- `WithHiddenMethods(...methods)` - Hide methods from auto-router
- `WithPathPrefix(prefix)` - Path prefix for all routes
- `WithMiddlewares(...names)` - Middleware names to apply

### 2. Simplified Remote Factory

```go
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceRemote{
        service: service.CastProxyService(config["remote"]),
    }
}
```

**What changed:**
- ✅ Framework pre-instantiates `proxy.Service` with metadata
- ✅ Factory just casts and injects (like local factory pattern)
- ✅ Consistent pattern across local and remote

### 3. Simplified Remote Constructor

```go
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{
        service: proxyService,
    }
}
```

**What changed:**
- ✅ No config parsing
- ✅ No proxy.Service construction
- ✅ Simple dependency injection

### 4. Auto-Generated Routers

**YAML Configuration:**
```yaml
apps:
  - addr: ":3004"
    published-services:
      - user-service  # Auto-generates router with metadata
```

**What happens:**
1. Loader scans `published-services`
2. Gets metadata from `RegisterServiceType`
3. Creates router definition automatically
4. Router instantiated at runtime with proper resource/convention

## Architecture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. SERVICE REGISTRATION (main.go)                               │
├─────────────────────────────────────────────────────────────────┤
│ RegisterServiceType("user-service-factory",                     │
│     UserServiceFactory,                                         │
│     UserServiceRemoteFactory,                                   │
│     WithResource("user", "users"),                              │
│     WithConvention("rest"))                                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. METADATA STORAGE (GlobalRegistry)                            │
├─────────────────────────────────────────────────────────────────┤
│ ServiceFactoryEntry {                                           │
│     Local: UserServiceFactory                                   │
│     Remote: UserServiceRemoteFactory                            │
│     Metadata: {                                                 │
│         Resource: "user"                                        │
│         ResourcePlural: "users"                                 │
│         Convention: "rest"                                      │
│     }                                                           │
│ }                                                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. AUTO-ROUTER GENERATION (Loader)                              │
├─────────────────────────────────────────────────────────────────┤
│ For each published-service:                                     │
│   1. Get service definition                                     │
│   2. Get metadata from factory registration                     │
│   3. DefineRouter("user-service-router", RouterDef{             │
│        Service: "user-service",                                 │
│        Convention: "rest",                                      │
│        Resource: "user",                                        │
│        ResourcePlural: "users"                                  │
│      })                                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. REMOTE SERVICE INJECTION (AddRemoteServiceByName)             │
├─────────────────────────────────────────────────────────────────┤
│ 1. Get metadata from factory registration                       │
│ 2. Create proxy.Service with metadata:                          │
│    proxy.NewService(baseURL, ConversionRule{                    │
│        Convention: metadata.Convention,                         │
│        Resource: metadata.Resource,                             │
│        ResourcePlural: metadata.ResourcePlural                  │
│    })                                                           │
│ 3. Inject as config["remote"]                                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. REMOTE FACTORY EXECUTION                                      │
├─────────────────────────────────────────────────────────────────┤
│ UserServiceRemoteFactory(deps, config) {                        │
│     return &UserServiceRemote{                                  │
│         service: service.CastProxyService(config["remote"])     │
│     }                                                           │
│ }                                                               │
└─────────────────────────────────────────────────────────────────┘
```

## Benefits

### ✅ Single Source of Truth
Metadata defined once at registration, used everywhere:
- Auto-router generation
- Remote service proxy creation
- Convention-based routing
- Resource name mapping

### ✅ Consistent Factory Pattern
```go
// LOCAL FACTORY
func UserServiceFactory(deps, config) any {
    return &UserServiceImpl{
        DB: service.Cast[*Database](deps["database"]),
    }
}

// REMOTE FACTORY (now consistent!)
func UserServiceRemoteFactory(deps, config) any {
    return &UserServiceRemote{
        service: service.CastProxyService(config["remote"]),
    }
}
```

### ✅ Type-Safe Options
```go
deploy.WithResource("user", "users")      // Compile-time checked
deploy.WithConvention("rest")             // Type-safe
deploy.WithHiddenMethods("Internal...")   // Autocompletion
```

### ✅ YAML Override Support
Metadata from registration can be overridden in YAML config if needed:
```yaml
remote-service-definitions:
  user-service-remote:
    url: "http://localhost:3004"
    resource: "person"  # Override default "user"
    resource-plural: "people"
```

### ✅ Auto-Router Generation
No manual router definitions needed:
```yaml
# OLD WAY
routers:
  user-api:
    service: user-service
    convention: rest
    resource: user
    resource-plural: users

# NEW WAY
apps:
  - addr: ":3004"
    published-services:
      - user-service  # Auto-generated!
```

### ✅ Clean Constructors
```go
// OLD WAY - Config parsing
func NewUserServiceRemote(baseURL string, config ...map[string]any) *UserServiceRemote {
    convention := ""
    resource := "user"
    if len(config) > 0 {
        // Parse config...
    }
    service := proxy.NewService(baseURL, ...)
    return &UserServiceRemote{service: service}
}

// NEW WAY - Simple injection
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{service: proxyService}
}
```

## Code Comparison

### Remote Service Constructor

**Before:**
```go
func NewUserServiceRemote(baseURL string, config ...map[string]any) *UserServiceRemote {
    // 15+ lines of config parsing and proxy.Service construction
    convention := ""
    resource := "user"
    resourcePlural := "users"
    
    if len(config) > 0 {
        cfg := config[0]
        if conv, ok := cfg["convention"].(string); ok {
            convention = conv
        }
        if res, ok := cfg["resource"].(string); ok {
            resource = res
        }
        if plural, ok := cfg["resource-plural"].(string); ok {
            resourcePlural = plural
        }
    }
    
    service := proxy.NewService(baseURL, autogen.ConversionRule{...}, ...)
    return &UserServiceRemote{service: service}
}
```

**After:**
```go
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{service: proxyService}
}
```

**Reduction:** 15 lines → 3 lines (80% less code)

### Remote Factory

**Before:**
```go
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    baseURL := config["base-url"].(string)
    return appservice.NewUserServiceRemote(baseURL, config)
}
```

**After:**
```go
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return &appservice.UserServiceRemote{
        service: service.CastProxyService(config["remote"]),
    }
}
```

**Pattern:** Now identical to local factory (dependency casting)

## Testing

### User Service (Standalone)
```bash
go run . -server="microservice.user-server"
```

**Output:**
```
✨ Auto-generated router 'user-service-router' from service 'user-service'
Starting [user-server] with 1 router(s) on address :3004
[user-auto] GET /users/{id} -> user-auto.GET[users_{id}]
[user-auto] GET /users -> user-auto.GET[users]
🟢 Starting server 'user-server' on :3004
```

### Order Service (with Remote User Service)
```bash
go run . -server="microservice.order-server"
```

**Output:**
```
✨ Auto-generated router 'order-service-router' from service 'order-service'
Starting [order-server] with 1 router(s) on address :3005
[order-auto] GET /orders/{id} -> order-auto.GET[orders_{id}]
🟢 Starting server 'order-server' on :3005
```

**Verification:**
- ✅ Auto-router generated from metadata
- ✅ Remote user-service proxy created with metadata
- ✅ Order service uses remote user-service seamlessly
- ✅ All HTTP calls work correctly

## Migration Guide

### For Existing Services

1. **Update Registration** - Add metadata options:
```go
lokstra_registry.RegisterServiceType("user-service-factory", 
    UserServiceFactory, 
    UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),  // Add this
    deploy.WithConvention("rest"),          // Add this
)
```

2. **Simplify Remote Constructor** - Remove config parsing:
```go
// OLD
func NewUserServiceRemote(baseURL string, config ...map[string]any) *UserServiceRemote

// NEW
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote
```

3. **Update Remote Factory** - Use CastProxyService:
```go
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceRemote{
        service: service.CastProxyService(config["remote"]),
    }
}
```

4. **Use published-services** in YAML:
```yaml
apps:
  - addr: ":3004"
    published-services:
      - user-service
```

## Summary

The metadata-driven pattern provides:

1. **Declarative Registration** - Metadata defined once, used everywhere
2. **Consistent Factories** - Local and remote use same pattern
3. **Auto-Router Generation** - No manual router definitions
4. **Smart Proxy Creation** - Framework handles complexity
5. **Type Safety** - Compile-time checking with options
6. **Clean Code** - Less boilerplate, more readability

**Result:** Services are easier to define, routers are auto-generated, and remote services work seamlessly with minimal configuration! 🚀
