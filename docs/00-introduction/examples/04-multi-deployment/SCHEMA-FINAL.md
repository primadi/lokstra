# Example 04 Schema - Final Design

## 🎯 **Overview**

Example 04 demonstrates **full deployment pattern** with:
- ✅ Hybrid service scope (server-level + app-level)
- ✅ Router registry and discovery
- ✅ Microservices communication
- ✅ Remote service definitions

---

## 📋 **Complete Schema**

### **Root Structure**

```yaml
# Service definitions (shared across deployments)
service-definitions:
  service-name:
    type: factory-type
    depends-on: [dependency1, dependency2]
    config: {...}

# Remote service definitions  
remote-service-definitions:
  remote-name:
    url: "http://..."
    resource: "singular"
    resource-plural: "plural"

# Deployments
deployments:
  deployment-name:
    config-overrides: {...}
    servers:
      server-name:
        base-url: "http://..."
        required-services: [...]        # ← Server-level (SHARED)
        required-remote-services: [...] # ← Server-level (SHARED)
        apps:
          - addr: ":8080"
            required-services: [...]        # ← App-level (ISOLATED)
            routers: [...]                  # ← Auto-published
            required-remote-services: [...] # ← App-level (ISOLATED)
```

---

## 🏗️ **Service Hierarchy**

### **Two-Level Service Scope**

```
┌─────────────────────────────────────────────────────┐
│ Server: api-server                                  │
│                                                     │
│ Server-Level Services (SHARED):                    │
│   • database         ← Created ONCE               │
│   • redis-cache      ← Shared by all apps         │
│   • logger           ← Shared logging             │
│                                                     │
│ Server-Level Remote Services (SHARED):             │
│   • payment-api      ← Shared HTTP client         │
│                                                     │
│ ┌─────────────────┐ ┌─────────────────┐           │
│ │ App :8080       │ │ App :8081       │           │
│ │                 │ │                 │           │
│ │ App Services:   │ │ App Services:   │           │
│ │ • user-service  │ │ • order-service │           │
│ │                 │ │                 │           │
│ │ Uses Shared:    │ │ Uses Shared:    │           │
│ │ • database ─────┼─┼──→ [database]   │           │
│ │ • redis-cache ──┼─┼──→ [cache]      │           │
│ │ • payment-api ──┼─┼──→ [payment]    │           │
│ └─────────────────┘ └─────────────────┘           │
└─────────────────────────────────────────────────────┘
```

### **Service Resolution Order**

```
1. App-level local services         (highest priority)
2. Server-level local services       (shared)
3. App-level remote services
4. Server-level remote services      (shared HTTP clients)
```

---

## 📊 **Field Reference**

### **Server-Level Fields**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `base-url` | string | ✅ Yes | Server base URL (e.g., "http://localhost") |
| `required-services` | array | ❌ No | **Shared** local services (created once) |
| `required-remote-services` | array | ❌ No | **Shared** remote services (HTTP clients) |
| `apps` | array | ✅ Yes | Applications running on this server |

### **App-Level Fields**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `addr` | string | ✅ Yes | Listen address (e.g., ":8080", "unix:/tmp/app.sock") |
| `required-services` | array | ❌ No | **App-specific** local services |
| `routers` | array | ❌ No | Routers to run (auto-published for discovery) |
| `required-remote-services` | array | ❌ No | **App-specific** remote services |

---

## 💡 **Best Practices**

### **What Goes in Server-Level?**

✅ **Infrastructure Services** (shared):
- Database connection pools
- Cache clients (Redis, Memcached)
- Loggers
- Configuration loaders
- Message queue clients

✅ **Common Remote Services** (shared):
- External APIs used by all apps
- Shared microservices

### **What Goes in App-Level?**

✅ **Business Logic Services** (isolated):
- Domain-specific services
- Feature-specific services
- Services that differ per app

✅ **App-Specific Remote Services** (isolated):
- Remote services only used by one app
- Optional external integrations

---

## 📝 **Example Configurations**

### **Example 1: Monolith**

```yaml
deployments:
  monolith:
    servers:
      api:
        base-url: "http://localhost"
        # All services at server-level (shared by all apps/ports)
        required-services:
          - database
          - cache
          - user-service
          - order-service
          - product-service
        apps:
          - addr: ":8080"
            routers: [api-router]
```

**Pattern**: Everything shared (single process)

---

### **Example 2: Microservices**

```yaml
deployments:
  # User Service
  user-service:
    servers:
      user-api:
        base-url: "http://localhost"
        required-services:
          - database      # Shared infra
          - user-service  # Business logic
        apps:
          - addr: ":3004"
            routers: [user-api]

  # Order Service (needs User Service)
  order-service:
    servers:
      order-api:
        base-url: "http://localhost"
        required-services:
          - database       # Shared infra
          - order-service  # Business logic
        required-remote-services:
          - user-service-remote  # Remote dependency
        apps:
          - addr: ":3005"
            routers: [order-api]
```

**Pattern**: Separate services with remote dependencies

---

### **Example 3: Multi-Port API**

```yaml
deployments:
  production:
    servers:
      api-server:
        base-url: "http://localhost"
        # Shared infrastructure
        required-services:
          - database
          - redis-cache
          - logger
        # Shared remote services
        required-remote-services:
          - payment-gateway
          - email-service
        
        apps:
          # Public API
          - addr: ":8080"
            required-services:
              - user-service
              - product-service
            routers: [public-api]
          
          # Admin API
          - addr: ":8081"
            required-services:
              - admin-service
              - audit-service
            routers: [admin-api]
          
          # Partner API
          - addr: ":8082"
            required-services:
              - partner-service
            required-remote-services:
              - partner-analytics  # Only this app needs it
            routers: [partner-api]
```

**Pattern**: Different apps, shared infrastructure

---

## 🎯 **Key Design Decisions**

### **1. ✅ Hybrid Service Scope**
- **Server-level**: Shared infrastructure
- **App-level**: Isolated business logic
- **Why**: Resource efficiency + flexibility

### **2. ✅ Auto-Published Routers**
- All routers in `routers` field are auto-published
- No separate `published-routers` field
- **Why**: Simplicity (99% use case)

### **3. ✅ Both Local and Remote at Both Levels**
- Server can have `required-services` + `required-remote-services`
- App can have `required-services` + `required-remote-services`
- **Why**: Maximum flexibility

### **4. ✅ App-Level Addr Required**
- Every app must specify `addr`
- **Why**: Explicit port/socket binding

---

## ✨ **Final Schema (Go Structs)**

```go
type DeployConfig struct {
    ServiceDefinitions       map[string]*ServiceDef
    RemoteServiceDefinitions map[string]*RemoteServiceSimple
    Deployments              map[string]*DeploymentDefMap
}

type DeploymentDefMap struct {
    ConfigOverrides map[string]any
    Servers         map[string]*ServerDefMap
}

type ServerDefMap struct {
    BaseURL        string
    Services       []string  // Server-level (shared)
    RemoteServices []string  // Server-level (shared)
    Apps           []*AppDefMap
}

type AppDefMap struct {
    Addr           string
    Services       []string  // App-level (isolated)
    Routers        []string  // Auto-published
    RemoteServices []string  // App-level (isolated)
}
```

---

## 🎓 **Summary**

**Example 04 Schema Features**:
1. ✅ Hybrid service scope (server + app)
2. ✅ Auto-published routers (simple)
3. ✅ Remote service support (both levels)
4. ✅ Flexible deployment patterns
5. ✅ Clean, intuitive configuration

**Result**: Production-ready deployment configuration! 🚀
