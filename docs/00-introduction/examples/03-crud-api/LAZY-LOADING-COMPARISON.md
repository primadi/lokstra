# Lazy Loading: Code vs Config Comparison

## 🎯 Both Modes Use LAZY LOADING

Kedua mode (`--mode=code` dan `--mode=config`) sama-sama menggunakan **lazy loading** - service **TIDAK** dibuat saat startup, tapi baru dibuat saat **FIRST HTTP REQUEST**.

---

## 📊 Side-by-Side Comparison

### Mode 1: `--mode=code` (Manual Lazy Loading)

```go
func runWithCode() {
    // Define lazy-loaded services
    db := service.LazyLoadWith(func() *Database {
        log.Println("🔧 Creating Database instance (lazy)...")
        return NewDatabase()
    })

    userService = service.LazyLoadWith(func() *UserService {
        log.Println("🔧 Creating UserService instance (lazy)...")
        return &UserService{
            DB: db,  // Pass lazy DB reference
        }
    })

    log.Println("✅ Services configured (lazy - will be created on first HTTP request)")
    
    setupRouterAndRun()
}
```

**Key Points:**
- ✅ Manual lazy load using `service.LazyLoadWith()`
- ✅ Dependencies passed as `service.Cached[T]` references
- ✅ No YAML, no registry - pure code
- ✅ Type-safe with generics
- ✅ Services created on FIRST access (when first HTTP request arrives)

---

### Mode 2: `--mode=config` (YAML + Lazy Loading)

```yaml
# config.yaml
services:
  - name: database
    type: database-factory
    config: {}

  - name: user-service
    type: user-service-factory
    depends_on:
      - database
```

```go
func runWithConfig() {
    // 1. Register factories
    reg := deploy.Global()
    reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
    reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

    // 2. Load deployment from YAML (lazy definitions)
    dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "development", reg)
    
    // 3. Get service (LAZY - not created yet!)
    server, _ := dep.GetServer("api")
    deployApp := server.Apps()[0]
    userServiceRaw, _ := deployApp.GetService("user-service")
    
    // 4. Wrap in Cached for handlers
    userService = service.Value(userServiceRaw.(*UserService))
    
    setupRouterAndRun()
}
```

**Key Points:**
- ✅ YAML-driven configuration
- ✅ Factory pattern with dependency injection
- ✅ Dependencies auto-resolved from YAML `depends_on`
- ✅ Services created lazily via `deployApp.GetService()`
- ✅ Type-safe after casting

---

## 🔍 When Are Services Created?

### Startup Phase (Both Modes)

```
🚀 Starting CRUD API in 'code' mode...
📝 APPROACH 1: Manual registration + Lazy loading (run by code)
✅ Services configured (lazy - will be created on first HTTP request)
🌐 Starting server: crud-api
   📍 Address: :3002
   ⏱️  Timeout: 30s
✅ Server is running...
```

**Services NOT created yet!** ⏳

---

### First HTTP Request

```
GET http://localhost:3002/api/v1/users
```

**NOW services are created:**

```
   🔧 Creating Database instance (lazy)...
   🔧 Creating UserService instance (lazy)...
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
  ]
}
```

---

### Subsequent Requests

Services are **already created** and **cached**:

```
GET http://localhost:3002/api/v1/users/1
```

**No creation logs** - uses cached instances! ✅

---

## 📐 Architecture Diagram

### Mode 1: Code (Manual Lazy)

```
main()
  ↓
runWithCode()
  ↓
service.LazyLoadWith(() => NewDatabase())         ← NOT created yet
  ↓
service.LazyLoadWith(() => &UserService{DB: db})  ← NOT created yet
  ↓
setupRouterAndRun()
  ↓
app.Run() ← Server starts
  ↓
[FIRST HTTP REQUEST arrives]
  ↓
listUsersHandler()
  ↓
userService.MustGet() ← TRIGGERS LAZY CREATION
  ↓
  ├─> Create UserService
  │     ↓
  │   db.MustGet() ← TRIGGERS DB CREATION
  │     ↓
  │   Create Database ✅
  │
  └─> Return UserService ✅
```

---

### Mode 2: Config (YAML + Lazy DI)

```
main()
  ↓
runWithConfig()
  ↓
deploy.Global().RegisterServiceType(...)  ← Register factories
  ↓
loader.LoadAndBuild(...)  ← Parse YAML, build deployment
  ↓
deployApp.GetService("user-service")  ← Get lazy reference (NOT created yet)
  ↓
service.Value(userServiceRaw)  ← Wrap in Cached
  ↓
setupRouterAndRun()
  ↓
app.Run() ← Server starts
  ↓
[FIRST HTTP REQUEST arrives]
  ↓
listUsersHandler()
  ↓
userService.MustGet() ← TRIGGERS LAZY CREATION via deployment
  ↓
  ├─> Call UserServiceFactory(deps, config)
  │     ↓
  │   deps["database"].(*service.Cached[any]).Cast[*Database]()
  │     ↓
  │   TRIGGERS DatabaseFactory() ← Create Database ✅
  │     ↓
  │   Return UserService ✅
  │
  └─> Cache result ✅
```

---

## 🎯 Key Differences

| Aspect | Mode 1: Code | Mode 2: Config |
|--------|-------------|---------------|
| **Definition** | `service.LazyLoadWith()` | YAML + Factory |
| **Dependencies** | Manual `DB: db` | Auto from `depends_on` |
| **Creation Logic** | Inline lambda | Factory function |
| **Configuration** | Hardcoded in code | External YAML file |
| **Type Safety** | Generics `Cached[T]` | After type assertion |
| **Registry** | None | `deploy.GlobalRegistry` |
| **Best For** | Simple apps, learning | Production, multi-env |

---

## ✅ Common Ground

**Both modes share:**

1. ✅ **Lazy initialization** - services created on first use
2. ✅ **Thread-safe** - `sync.Once` ensures single creation
3. ✅ **Cached** - subsequent calls reuse same instance
4. ✅ **Type-safe** - compile-time type checking
5. ✅ **Same handlers** - both use package-level `userService` variable
6. ✅ **Same router** - identical HTTP endpoints
7. ✅ **No paradigm mixing** - both use NEW paradigm (no lokstra_registry)

---

## 🚀 Testing Lazy Loading

### Test Mode 1 (Code)

```bash
# Terminal 1: Start server
go run . -mode code

# Terminal 2: Make request
curl http://localhost:3002/api/v1/users
```

**Expected logs:**
```
🚀 Starting CRUD API in 'code' mode...
✅ Services configured (lazy - will be created on first HTTP request)
✅ Server is running...
   🔧 Creating Database instance (lazy)...      ← First request triggers this
   🔧 Creating UserService instance (lazy)...  ← Then this
```

---

### Test Mode 2 (Config)

```bash
# Terminal 1: Start server
go run . -mode config

# Terminal 2: Make request
curl http://localhost:3002/api/v1/users
```

**Expected logs:**
```
🚀 Starting CRUD API in 'config' mode...
✅ Services loaded from YAML config
✅ Server is running...
   [Service creation happens on first request via factories]
```

---

## 💡 Best Practices

### When to Use Mode 1 (Code)

- ✅ Small applications
- ✅ Learning/prototyping
- ✅ Single environment
- ✅ Services with no configuration
- ✅ Quick startup

### When to Use Mode 2 (Config)

- ✅ Production applications
- ✅ Multiple environments (dev/staging/prod)
- ✅ Services with complex configuration
- ✅ Need to change config without recompile
- ✅ Team collaboration (config in VCS)

---

## 🎓 Key Takeaways

1. **LAZY is KING** 👑
   - Both modes use lazy loading
   - Services created on FIRST use
   - Zero startup overhead

2. **Type Safety Matters** 🔒
   - `service.Cached[T]` provides compile-time safety
   - Generic type checking prevents runtime errors

3. **Choose Your Style** 🎨
   - Code: Explicit, inline, simple
   - Config: Declarative, flexible, scalable

4. **NO OLD PARADIGM** 🚫
   - No `lokstra_registry` dependency
   - Pure NEW paradigm (deploy + service)
   - Production-ready architecture

---

Perfect! Both modes demonstrate **lazy loading** correctly! 🎉
