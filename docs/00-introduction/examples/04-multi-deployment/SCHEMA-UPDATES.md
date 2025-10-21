# Schema Updates: Hybrid Service Pattern

## 🎯 **Problem Solved**

**Before**: Services defined only at app level caused duplication of infrastructure services.

**Example Problem**:
```yaml
# ❌ OLD: Database created TWICE (waste!)
servers:
  api-server:
    apps:
      - addr: ":8080"
        required-services: [database, user-service]  # database #1
      - addr: ":8081"  
        required-services: [database, order-service] # database #2 (duplicate!)
```

**After**: Hybrid pattern - server-level (shared) + app-level (isolated).

```yaml
# ✅ NEW: Database created ONCE (shared!)
servers:
  api-server:
    required-services: [database]  # ← Shared by all apps
    apps:
      - addr: ":8080"
        required-services: [user-service]   # ← App-specific
      - addr: ":8081"
        required-services: [order-service]  # ← App-specific
```

---

## 📋 **Schema Changes**

### **1. ServerDefMap - Added Fields**

```go
type ServerDefMap struct {
    BaseURL        string       `yaml:"base-url"`
    Services       []string     `yaml:"required-services,omitempty"`        // ← NEW: Shared
    RemoteServices []string     `yaml:"required-remote-services,omitempty"` // ← NEW: Shared
    Apps           []*AppDefMap `yaml:"apps"`
}
```

### **2. AppDefMap - Added Field**

```go
type AppDefMap struct {
    Addr             string   `yaml:"addr"`
    Services         []string `yaml:"required-services,omitempty"`        // App-specific
    Routers          []string `yaml:"routers,omitempty"`
    PublishedRouters []string `yaml:"published-routers,omitempty"`        // ← NEW: For discovery
    RemoteServices   []string `yaml:"required-remote-services,omitempty"` // App-specific
}
```

---

## 🏗️ **Architecture**

### **Service Scope Hierarchy**:

```
┌─────────────────────────────────────────────────────────┐
│ Server: api-server                                      │
│                                                         │
│ Server-Level (SHARED):                                 │
│   • database         ← Created once, reused everywhere │
│   • redis-cache      ← Shared connection pool          │
│   • logger           ← Shared logger instance          │
│                                                         │
│ ┌──────────────────┐ ┌──────────────────┐             │
│ │ App :8080        │ │ App :8081        │             │
│ │ (User API)       │ │ (Order API)      │             │
│ │                  │ │                  │             │
│ │ App-Level:       │ │ App-Level:       │             │
│ │ • user-service   │ │ • order-service  │             │
│ │                  │ │                  │             │
│ │ Uses shared:     │ │ Uses shared:     │             │
│ │ • database ───────┼─┼─→ [database]    │             │
│ │ • redis-cache ────┼─┼─→ [redis-cache] │             │
│ └──────────────────┘ └──────────────────┘             │
└─────────────────────────────────────────────────────────┘
```

### **Service Resolution Order**:

```
1. App-level local services     (highest priority)
2. Server-level local services   (shared)
3. App-level remote services
4. Server-level remote services  (shared HTTP clients)
```

---

## 📊 **Use Cases**

### **Use Case 1: Shared Infrastructure**

```yaml
servers:
  api-server:
    # Shared infrastructure services
    required-services:
      - database          # PostgreSQL pool (shared)
      - redis-cache       # Redis client (shared)
      - logger            # Logger instance (shared)
      - config-service    # Config loader (shared)
    
    apps:
      - addr: ":8080"
        required-services: [user-service]
      - addr: ":8081"
        required-services: [order-service]
      - addr: ":8082"
        required-services: [admin-service]
```

**Result**: 
- 1 database connection pool (not 3!)
- 1 Redis client (not 3!)
- Resource efficient ✅

---

### **Use Case 2: Shared Remote Services**

```yaml
servers:
  order-server:
    # All apps need these remote services
    required-remote-services:
      - user-service-remote     # HTTP client to user-service
      - payment-service-remote  # HTTP client to payment-service
    
    apps:
      - addr: ":8080"  # Public Order API
        required-services: [order-service]
      
      - addr: ":8081"  # Admin Order API
        required-services: [order-admin-service]
      
      - addr: ":8082"  # Reporting API
        required-services: [order-reporting-service]
        # Needs extra remote service (app-specific)
        required-remote-services:
          - analytics-service-remote
```

**Result**:
- Shared HTTP clients for user & payment services
- App-specific remote service for analytics
- Flexibility + efficiency ✅

---

### **Use Case 3: Different Services per App**

```yaml
servers:
  api-server:
    # Infrastructure shared by all
    required-services: [database, cache]
    
    apps:
      - addr: ":8080"  # Public API
        required-services: [user-service, product-service]
      
      - addr: ":8081"  # Admin API (different services!)
        required-services: [admin-service, audit-service]
      
      - addr: ":8082"  # Webhook API (different services!)
        required-services: [webhook-service, event-service]
```

**Result**: 
- Each app has its own business logic services
- All share same database & cache
- Isolation + sharing ✅

---

## ✅ **Benefits**

| Benefit | Before | After |
|---------|--------|-------|
| **Resource Usage** | ❌ Duplicated services | ✅ Shared infrastructure |
| **Flexibility** | ⚠️ Limited | ✅ App-specific + shared |
| **Typical Pattern** | ❌ Doesn't match real-world | ✅ Matches microservices pattern |
| **Code Clarity** | ⚠️ Unclear what's shared | ✅ Explicit hierarchy |

---

## 🚀 **Migration Guide**

### **Old Config (App-level only)**:
```yaml
servers:
  api:
    apps:
      - addr: ":8080"
        required-services:
          - database
          - cache
          - user-service
```

### **New Config (Hybrid)**:
```yaml
servers:
  api:
    # Move infrastructure to server-level
    required-services:
      - database
      - cache
    apps:
      - addr: ":8080"
        # Keep business logic at app-level
        required-services:
          - user-service
```

---

## 📝 **Best Practices**

### **Server-Level Services (Shared)**:
- ✅ Database connection pools
- ✅ Cache clients (Redis, Memcached)
- ✅ Loggers
- ✅ Configuration loaders
- ✅ Message queue clients
- ✅ Common remote services (used by all apps)

### **App-Level Services (Isolated)**:
- ✅ Business logic services
- ✅ Domain-specific services
- ✅ App-specific remote services
- ✅ Services that should be isolated per app

---

## 🎯 **Summary**

**Hybrid Pattern = Best of Both Worlds**:
- Server-level: Shared infrastructure (efficient)
- App-level: Isolated business logic (flexible)
- Remote services: Both levels supported
- Published routers: For service discovery

**Result**: Clean, efficient, flexible deployment configuration! 🎉
