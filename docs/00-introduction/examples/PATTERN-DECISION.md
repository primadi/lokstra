# Examples Pattern Decision

## 📋 Overview

Examples dalam `docs/00-introduction` dirancang dengan **progressive learning approach** - dari simple ke complex.

## 🎯 Pattern per Example

### **Example 01-02: Foundation**
**Pattern**: Manual router + direct `lokstra.NewApp()`

**Focus**: 
- Basic routing
- Handler patterns
- Request/Response handling

**No Registry**: Router tidak perlu didaftarkan ke registry

```go
r := lokstra.NewRouter("api")
r.GET("/hello", handler)

app := lokstra.NewApp("my-app", ":8080", r)
app.Run(30 * time.Second)
```

---

### **Example 03: Services + DI**
**Pattern**: Manual deployment structure + Services + Lazy DI

**Focus**:
- Service definition and registration
- Dependency injection
- Lazy loading pattern
- Code vs Config comparison

**No Registry**: Router tidak perlu didaftarkan (masih simple pattern)

**Two Approaches**:

#### Code Mode:
```go
// 1. Register service factories
lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

// 2. Define services
lokstra_registry.DefineService(&schema.ServiceDef{
    Name: "database",
    Type: "database-factory",
})

// 3. Build deployment (for service container)
dep := deploy.New("development")
server := dep.NewServer("api", "http://localhost")
app := server.NewApp(":3002")
app.AddService("database")

// 4. Lazy load service
userService := service.LazyLoadFrom[*UserService](app, "user-service")

// 5. Create router manually
r := lokstra.NewRouter("api")
// ... setup routes

// 6. Run with lokstra.NewApp()
lokstraApp := lokstra.NewApp("crud-api", ":3002", r)
lokstraApp.Run(30 * time.Second)
```

#### Config Mode:
```go
// 1. Register service factories
lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

// 2. Load deployment from YAML
dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "development")

// 3. Get app and lazy load service
server, _ := dep.GetServer("api")
app := server.Apps()[0]
userService := service.LazyLoadFrom[*UserService](app, "user-service")

// 4. Create router manually
r := lokstra.NewRouter("api")
// ... setup routes

// 5. Run with lokstra.NewApp()
lokstraApp := lokstra.NewApp("crud-api", ":3002", r)
lokstraApp.Run(30 * time.Second)
```

**Key Point**: Deployment structure hanya untuk **service container**, bukan untuk router management.

---

### **Example 04: Multi-Deployment** (TODO)
**Pattern**: Full deployment pattern + Current server

**Focus**:
- Multiple servers in one deployment
- Current server selection
- Router registry for cross-service communication
- Deployment-based routing

**With Registry**: Router didaftarkan untuk deployment pattern

```go
// Register service factories
lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

// Load deployment
dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "production")

// Set current server and run
lokstra_registry.SetCurrentDeployment(dep)
lokstra_registry.SetCurrentServer("user-service")
lokstra_registry.RunCurrentServer(30 * time.Second)
```

**Benefit**: Framework automatically builds routers from config and handles server lifecycle.

---

### **Example 05: Auto-Router + Proxy**
**Pattern**: Convention-based routing + Remote service calls

**Focus**:
- Auto-router generation from services
- REST/RPC conventions
- Proxy service pattern
- Convention over configuration

```go
// Server mode: Auto-generate router from service
router := autogen.NewFromService(orderService, autogen.REST())

// Client mode: Use proxy for remote calls
proxy.Call(proxyService, "GetUserByID", userID)
```

---

## 🎓 Learning Progression

```
01-02: Basic Routing
   ↓
03: Services + DI (manual deployment for DI only)
   ↓
04: Full Deployment Pattern (current server + router registry)
   ↓
05: Auto-Router + Conventions (eliminate manual routing)
```

## ✅ Key Decisions

1. **Router Registry Usage**:
   - Examples 01-03: **NO** registry (simple pattern)
   - Example 04+: **YES** registry (full deployment pattern)

2. **Deployment Structure**:
   - Example 01-02: **None** (direct app creation)
   - Example 03: **Partial** (only for service container)
   - Example 04+: **Full** (complete deployment lifecycle)

3. **Running Pattern**:
   - Example 01-03: `lokstra.NewApp()` + `app.Run()`
   - Example 04+: `lokstra_registry.RunCurrentServer()`

4. **LoadAndBuild Signature**:
   - Simplified: `LoadAndBuild(paths, deploymentName)` 
   - Always uses `deploy.Global()` internally
   - No need to pass registry parameter

## 🔧 Implementation Notes

### lokstra_registry Package-Level Functions

Untuk consistency dan kemudahan, semua registry operations menggunakan **package-level functions**:

```go
// ❌ OLD (verbose)
reg := lokstra_registry.Global()
reg.RegisterServiceType("my-service", factory, nil)

// ✅ NEW (clean)
lokstra_registry.RegisterServiceType("my-service", factory, nil)
```

### Deployment Functions (Example 04+)

Functions untuk full deployment pattern:

```go
// Set current deployment and server
lokstra_registry.SetCurrentDeployment(dep)
lokstra_registry.SetCurrentServer("server-name")

// Print and run
lokstra_registry.PrintCurrentServerInfo()
lokstra_registry.RunCurrentServer(30 * time.Second)
```

---

## 📚 Summary

| Example | Pattern | Registry | Running |
|---------|---------|----------|---------|
| 01-02 | Manual | ❌ No | `lokstra.NewApp()` |
| 03 | Services+DI | ❌ No | `lokstra.NewApp()` |
| 04 | Full Deployment | ✅ Yes | `RunCurrentServer()` |
| 05 | Auto-Router | ✅ Yes | `RunCurrentServer()` |

**Principle**: Start simple, add complexity progressively as needed.
