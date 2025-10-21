# Example 04: Multi-Deployment - Design Proposal

## 🎯 Goals

1. Show **full deployment pattern** with router registry
2. Demonstrate **service discovery** across deployments
3. Introduce **published routers** concept
4. Show **required-services** pattern

---

## 🤔 Design Questions

### **Q1: Router Registration - Manual vs Auto?**

**Options:**
- **A. Manual in code** (Example 04 simple)
- **B. Auto from config** (Example 05 auto-router)

**Decision**: **A - Manual** untuk Example 04
- Focus on deployment concepts, not auto-generation
- Auto-router sudah di Example 05

### **Q2: Published Routers - Where to declare?**

**Current Problem**: 
```yaml
# ❌ No way to publish routers for discovery
user-service:
  servers:
    user-api:
      apps:
        - addr: ":3004"
          required-services: [user-service]
          # Missing: published-routers ???
```

**Proposal**:
```yaml
user-service:
  servers:
    user-api:
      base-url: "http://localhost"
      apps:
        - addr: ":3004"
          required-services: [user-service]
          routers: [user-api]  # ← Routers hosted by this app
          published-routers: [user-api]  # ← Routers published for discovery
```

**Meaning**:
- `routers`: Which routers this app runs
- `published-routers`: Which routers are discoverable by other services

### **Q3: Required-Services - App Level vs Server Level?**

**Current**: App level
```yaml
servers:
  user-api:
    apps:
      - addr: ":3004"
        required-services: [database, user-service]  # ← Per app
```

**Question**: Should it be server level?
```yaml
servers:
  user-api:
    required-services: [database, user-service]  # ← Per server?
    apps:
      - addr: ":3004"
```

**Analysis**:

| Aspect | App Level | Server Level |
|--------|-----------|--------------|
| **Lazy Loading** | ✅ Per app | ✅ Per server (shared) |
| **Service Scope** | App-specific | Server-wide |
| **Flexibility** | High (different services per app) | Lower (all apps share) |
| **Typical Use Case** | Different apps, different services | Same services, multiple ports |

**Recommendation**: **App Level** (current design is correct!)

**Why?**
- Services are **lazy-loaded** - created on first request
- Each app might need **different services**
- More flexible for complex deployments

**Example**:
```yaml
servers:
  api-server:
    apps:
      - addr: ":8080"  # Public API
        required-services: [user-service, order-service]
      - addr: ":8081"  # Admin API
        required-services: [admin-service, audit-service]
```

Different apps = different responsibilities = different services ✅

---

## ✅ Proposed Schema Changes

### Add `published-routers` to AppDefMap:

```go
// AppDefMap is an app using map structure
type AppDefMap struct {
    Addr              string   `yaml:"addr" json:"addr"`
    Services          []string `yaml:"required-services,omitempty"`
    Routers           []string `yaml:"routers,omitempty"`
    PublishedRouters  []string `yaml:"published-routers,omitempty"`  // ← NEW!
    RemoteServices    []string `yaml:"required-remote-services,omitempty"`
}
```

### Example 04 Config (Proposed):

```yaml
# User Service Deployment
user-service:
  servers:
    user-api:
      base-url: "http://localhost"
      apps:
        - addr: ":3004"
          required-services:
            - database
            - user-service
          routers:
            - user-api  # Router defined in code
          published-routers:
            - user-api  # Make it discoverable

# Order Service Deployment
order-service:
  servers:
    order-api:
      base-url: "http://localhost"
      apps:
        - addr: ":3005"
          required-services:
            - database
            - order-service
          routers:
            - order-api  # Router defined in code
          published-routers:
            - order-api  # Make it discoverable
          required-remote-services:
            - user-service-remote  # Proxy to user-service
```

### Code Pattern:

```go
func runUserService(dep *deploy.Deployment) {
    // 1. Get app
    server, _ := dep.GetServer("user-api")
    app := server.Apps()[0]
    
    // 2. Lazy load services
    userService := service.LazyLoadFrom[*UserService](app, "user-service")
    
    // 3. Create and register router
    r := lokstra.NewRouter("user-api")
    r.GET("/users", handler.list)
    r.GET("/users/{id}", handler.get)
    
    // 4. Register router for discovery
    lokstra_registry.RegisterRouter("user-api", r)
    
    // 5. Set current deployment and server
    lokstra_registry.SetCurrentDeployment(dep)
    lokstra_registry.SetCurrentServer("user-api")
    
    // 6. Run current server (framework builds and runs)
    lokstra_registry.RunCurrentServer(30 * time.Second)
}
```

---

## 📊 Summary

| Question | Answer |
|----------|--------|
| Router Registration? | ✅ **YES** - Manual in code, registered to registry |
| Published Routers? | ✅ **NEW** - Add `published-routers` to schema |
| Required-Services Level? | ✅ **App Level** - Correct as-is (flexibility) |
| Example 04 Pattern? | ✅ **Full Deployment** - Use current server pattern |

---

## 🔧 Implementation Steps

1. ✅ Add `published-routers` to schema
2. ✅ Add `PublishedRouters []string` to `AppDefMap`
3. ✅ Update Example 04 config.yaml
4. ✅ Update Example 04 main.go to use current server pattern
5. ✅ Implement `lokstra_registry` deployment functions
6. ✅ Update README with clear explanation

---

**Next**: Implement these changes? 🚀
