# Migration Guide - 03-crud-api

## 🎯 Two Approaches in One Example!

This example demonstrates **BOTH** approaches so you can see and compare:

### Run Mode 1: By Code (Manual - Old Paradigm)
```bash
go run main.go --mode=code
# or just: go run main.go
```

### Run Mode 2: By Config (YAML + Lazy DI - New Paradigm)
```bash
go run main.go --mode=config
```

---

## 📊 Side-by-Side Comparison

### APPROACH 1: Run by Code (Manual)

```go
func runWithCode() {
    log.Println("📝 APPROACH 1: Manual instantiation (run by code)")
    
    // 1. Create services manually
    db := NewDatabase()
    userSvc := &UserService{
        DB: service.Value(db), // Wrap in Cached for consistency
    }
    
    // 2. Cache service for handlers
    userService = service.Value(userSvc)
    
    // 3. Setup router and run
    setupRouterAndRun()
}
```

**Characteristics:**
- ✅ **Simple & Direct** - Easy to understand
- ✅ **Full Control** - You see everything
- ✅ **Good for small apps** - No overhead
- ❌ **Manual wiring** - You connect dependencies manually
- ❌ **Hardcoded order** - Must create DB before UserService
- ❌ **No config file** - Everything in code
- ❌ **Single environment** - No easy dev/staging/prod switching

**Best for:**
- Learning and prototyping
- Small applications (< 5 services)
- Single environment
- When you want explicit control

---

### APPROACH 2: Run by Config (YAML + Lazy DI)

**config.yaml:**
```yaml
services:
  database:
    type: database-factory
    config:
      seed_data: true
  
  user-service:
    type: user-service-factory
    depends-on: [database]  # Auto lazy-loaded!

deployments:
  development:
    servers:
      api:
        apps:
          - port: 3002
            services: [database, user-service]
```

**main.go:**
```go
func runWithConfig() {
    log.Println("⚙️ APPROACH 2: YAML Configuration + Lazy DI (run by config)")
    
    // 1. Get global registry
    reg := deploy.Global()
    
    // 2. Register service factories
    reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
    reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)
    
    // 3. Load and build deployment from YAML
    dep, _ := loader.LoadAndBuild(
        []string{"config.yaml"},
        "development",
        reg,
    )
    
    // 4. Get server and app
    server, _ := dep.GetServer("api")
    deployApp := server.Apps()[0]
    
    // 5. Get user service instance (lazy loaded)
    userServiceRaw, _ := deployApp.GetService("user-service")
    
    // 6. Cache service reference for handlers
    userService = service.Value(userServiceRaw.(*UserService))
    
    // 7. Setup router and run
    setupRouterAndRun()
}
```

**Characteristics:**
- ✅ **Declarative config** - Services defined in YAML
- ✅ **Lazy loading** - DB only created when UserService needs it
- ✅ **Type-safe DI** - `service.Cast[T]()` for type safety
- ✅ **Multi-environment** - Easy dev/staging/prod configs
- ✅ **Validation** - JSON Schema validates config
- ✅ **No initialization order issues** - Lazy DI handles it
- ❌ **More setup** - Factory functions needed
- ❌ **Learning curve** - Need to understand factory pattern

**Best for:**
- Production applications
- Multiple environments
- Complex service dependencies (5+ services)
- Team environments (config easier to review)
- When you need validation

---

## 🔍 Key Differences Explained

### 1. Service Creation

**By Code:**
```go
db := NewDatabase()                    // Created immediately
userSvc := &UserService{DB: service.Value(db)}  // DB must exist first
```

**By Config:**
```yaml
services:
  database:
    type: database-factory            # Not created yet!
  
  user-service:
    type: user-service-factory
    depends-on: [database]            # Will lazy-load when needed
```

```go
// Factory receives lazy-loaded deps
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB: service.Cast[*Database](deps["database"]), // Lazy Cached[*Database]
    }
}

// Database only created when userService.DB.Get() is called!
```

---

### 2. Dependency Injection

**By Code:**
```go
// Manual - YOU control the order
db := NewDatabase()           // Step 1: Create DB first
userSvc := &UserService{      // Step 2: Pass DB to UserService
    DB: service.Value(db),
}
// If you switch the order → compilation error or panic!
```

**By Config:**
```yaml
services:
  user-service:
    depends-on: [database]    # Framework resolves automatically
```

```go
// In UserService methods:
func (s *UserService) GetAll() ([]*User, error) {
    return s.DB.MustGet().GetAll()    // DB loaded on first call
}
// No initialization order issues!
// No circular dependency risks!
```

---

### 3. Configuration

**By Code:**
```go
// Hardcoded in code
port := ":3002"
timeout := 30 * time.Second
```

**By Config:**
```yaml
# config.yaml
configs:
  PORT: 3002
  TIMEOUT: 30

deployments:
  development:
    config-overrides:
      PORT: 3002
  
  production:
    config-overrides:
      PORT: ${PORT}    # From environment variable
```

---

### 4. Multi-Environment

**By Code:**
```go
// Need manual flags or env vars
var env string
flag.StringVar(&env, "env", "dev", "Environment")

if env == "prod" {
    dbHost = os.Getenv("DB_HOST")
} else {
    dbHost = "localhost"
}
```

**By Config:**
```yaml
deployments:
  development:
    config-overrides:
      DB_HOST: localhost
  
  production:
    config-overrides:
      DB_HOST: ${DB_HOST}    # From env
```

```go
// One line change!
dep, _ := loader.LoadAndBuild(
    []string{"config.yaml"},
    "production",  // Just change this!
    reg,
)
```

---

## 📈 When to Use Which?

### Use "Run by Code" When:
- ✅ Learning Lokstra framework
- ✅ Prototyping quickly
- ✅ Simple apps (1-3 services)
- ✅ Single environment only
- ✅ Want full explicit control
- ✅ Don't need config validation

### Use "Run by Config" When:
- ✅ Production applications
- ✅ Multiple environments (dev/staging/prod)
- ✅ Complex dependencies (5+ services)
- ✅ Team development (easier code review)
- ✅ Need config validation
- ✅ Want lazy loading benefits

---

## 🎓 Learning Path

1. **Start with "code" mode** - Understand the basics
2. **Run both modes** - See they produce same result
3. **Compare the code** - See what's different
4. **Try config mode** - Experience declarative approach
5. **Modify config.yaml** - See how easy it is to change
6. **Add new service** - Practice factory pattern

---

## 🧪 Try It Yourself!

### Test Both Modes:

```bash
# Mode 1: By Code
go run main.go --mode=code
curl http://localhost:3002/api/v1/users

# Mode 2: By Config
go run main.go --mode=config
curl http://localhost:3002/api/v1/users

# They produce identical results!
```

### Experiment:

1. **Add a new service** - Try adding a Logger service
2. **Change config** - Edit `config.yaml` to add config values
3. **Add environment** - Add a "staging" deployment
4. **Break dependency** - Remove `depends-on` and see what happens

---

## 📝 Summary

| Aspect | By Code | By Config |
|--------|---------|-----------|
| **Simplicity** | ✅ Very simple | ⚠️ More setup |
| **Control** | ✅ Full control | ✅ Controlled by config |
| **Initialization** | ❌ Manual order | ✅ Auto lazy-load |
| **Multi-env** | ❌ Manual flags | ✅ Built-in |
| **Validation** | ❌ None | ✅ JSON Schema |
| **Scalability** | ⚠️ 1-3 services | ✅ 5+ services |
| **Team Work** | ⚠️ Code review | ✅ Easy review |
| **Best for** | Learning, small apps | Production, teams |

---

**Both approaches are valid!** Choose based on your needs:
- Small/Learning → Use **Code**
- Production/Complex → Use **Config**

---

*This example shows both so you can learn when and why to use each approach.*
