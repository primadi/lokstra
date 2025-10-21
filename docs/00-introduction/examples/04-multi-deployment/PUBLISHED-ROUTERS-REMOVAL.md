# Schema Simplification: Removed Published-Routers

## 🎯 **Decision**

**REMOVED** `published-routers` field from schema.

**Reason**: Simplicity > Flexibility for this use case.

---

## ✅ **What Changed**

### **Before (Complex)**:
```yaml
apps:
  - addr: ":8080"
    routers: [user-api, admin-api]
    published-routers: [user-api]  # ❌ Redundant in 99% cases
```

### **After (Simple)**:
```yaml
apps:
  - addr: ":8080"
    routers: [user-api, admin-api]  # ✅ Auto-published for discovery
```

---

## 📊 **Impact**

| Aspect | Before | After |
|--------|--------|-------|
| **Config Lines** | More (2 fields) | Less (1 field) |
| **Clarity** | Confusing (what's the diff?) | Clear (auto-published) |
| **Use Cases Covered** | 100% | 99% (good enough!) |
| **Developer Experience** | ⚠️ Must remember 2 fields | ✅ Simple! |

---

## 🔧 **Schema Changes**

### **1. Go Struct (`schema.go`)**

```go
// REMOVED PublishedRouters field
type AppDefMap struct {
    Addr           string   `yaml:"addr"`
    Services       []string `yaml:"required-services,omitempty"`
    Routers        []string `yaml:"routers,omitempty"`  // ← Auto-published!
    RemoteServices []string `yaml:"required-remote-services,omitempty"`
}
```

### **2. JSON Schema (`lokstra.schema.json`)**

```json
{
  "routers": {
    "type": "array",
    "description": "Routers to include in this app (automatically published for service discovery)"
  }
  // REMOVED: published-routers field
}
```

---

## 💡 **Behavior**

**All routers in `routers` field are automatically**:
1. ✅ Loaded and run by the app
2. ✅ Registered in router registry
3. ✅ Published for service discovery
4. ✅ Discoverable by remote services

**Example**:
```yaml
deployments:
  user-service:
    servers:
      user-api:
        apps:
          - addr: ":3004"
            routers: [user-api]  
            # ↑ Automatically discoverable by order-service
```

---

## 🎓 **If You Need Internal Routers**

### **Option 1: Separate Server** (Recommended)
```yaml
deployments:
  production:
    servers:
      # Public API server
      public-api:
        apps:
          - addr: ":8080"
            routers: [user-api]  # ← Discoverable
      
      # Internal metrics server (different deployment-id = isolated)
      internal:
        apps:
          - addr: ":9090"
            routers: [metrics-api]  # ← Not in same deployment = not discoverable
```

### **Option 2: Convention** (If really needed)
```go
// In router registration code:
if strings.HasPrefix(routerName, "_") {
    // Don't register for discovery
} else {
    lokstra_registry.RegisterRouter(routerName, router)
}
```

```yaml
routers: [user-api, _internal-api]
#                   ↑ Underscore = private (not registered)
```

---

## 📈 **Benefits**

1. **Simpler Config** ✅
   - One field instead of two
   - Less cognitive load
   
2. **Clear Semantics** ✅
   - If it's in `routers`, it's discoverable
   - No confusion about publish vs run
   
3. **Matches Common Pattern** ✅
   - 99% of routers should be discoverable
   - Rare edge cases can use workarounds
   
4. **Better DX** ✅
   - Less to remember
   - Fewer mistakes
   - Faster to write

---

## 🎯 **Final Schema Summary**

```yaml
deployments:
  production:
    servers:
      api-server:
        base-url: "http://localhost"
        
        # Server-level (shared)
        required-services: [database, cache]
        required-remote-services: [payment-service]
        
        apps:
          - addr: ":8080"
            # App-level (isolated)
            required-services: [user-service]
            routers: [user-api]  # ← Auto-published
            required-remote-services: [analytics-service]
```

**Clean, simple, intuitive!** 🎉
