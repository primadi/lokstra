# Lokstra Examples - New Paradigm

> 🎯 **Same examples, modern approach with YAML + Lazy DI**

These are the same examples as before, now updated to use Lokstra's new paradigm where beneficial.

---

## 📂 Examples

### [01-hello-world](./01-hello-world/) - **NO CHANGES**
**Simplest API with 3 endpoints**

- Simple router with GET handlers
- Auto JSON responses
- Basic string and map returns

**Status:** ❌ **No changes needed** - Simple examples don't need config!

```bash
cd 01-hello-world && go run main.go
curl http://localhost:3000/
```

---

### [02-handler-forms](./02-handler-forms/) - **NO CHANGES**
**All 29 handler variations**

- Request binding (JSON, path, query, header)
- Response forms (string, map, struct, error handling)
- Context access patterns

**Status:** ❌ **No changes needed** - Handler forms are framework features!

```bash
cd 02-handler-forms && go run main.go
```

---

### [03-crud-api](./03-crud-api/) - **✅ UPDATED**
**Full CRUD API with database service**

**OLD Approach:**
- Manual service instantiation
- Manual registry registration
- Hardcoded dependencies

**NEW Approach:**
- Service factories with YAML config
- Lazy dependency injection
- Type-safe service loading

```bash
cd 03-crud-api && go run main.go
curl http://localhost:3000/users
```

**What Changed:**
- ✅ Added `config.yaml` for service configuration
- ✅ Database and UserService now use factories
- ✅ Lazy DI for Database → UserService
- ❌ Router and handlers unchanged

---

### [04-multi-deployment](./04-multi-deployment/) - **✅ UPDATED**
**Monolith vs Microservices**

**OLD Approach:**
```go
flag.StringVar(&mode, "mode", "monolith", "...")
if mode == "monolith" {
    runMonolithServer()
} else if mode == "user-service" {
    runUserServiceServer()
}
```

**NEW Approach:**
```yaml
deployments:
  monolith: {...}
  microservices-user: {...}
  microservices-order: {...}
```

```bash
# Choose deployment declaratively!
go run main.go --deployment=monolith
go run main.go --deployment=microservices-user
go run main.go --deployment=microservices-order
```

**What Changed:**
- ✅ One `config.yaml` instead of multiple flag checks
- ✅ Declarative deployment configurations
- ✅ Shared service definitions, different deployments
- ✅ Better separation of concerns

---

## 🔄 When to Use New Paradigm?

### ❌ DON'T Update (Keep Simple):
- **01-hello-world** - Too simple, no benefit
- **02-handler-forms** - Framework features, not config
- Simple scripts and prototypes
- Learning examples

### ✅ DO Update (Clear Benefits):
- **03-crud-api** - Service dependencies benefit from lazy DI
- **04-multi-deployment** - Multi-environment is perfect fit
- Production applications
- Complex service graphs
- Multiple environments

---

## 📊 Comparison Table

| Example | Old Paradigm | New Paradigm | Status |
|---------|-------------|--------------|--------|
| **01-hello-world** | Manual router | Manual router | ❌ No change needed |
| **02-handler-forms** | Manual router | Manual router | ❌ No change needed |
| **03-crud-api** | Manual services | YAML + Factories | ✅ Updated |
| **04-multi-deployment** | Flags + functions | YAML deployments | ✅ Updated |

---

## 🎯 Key Improvements in Updated Examples

### Example 03 (CRUD API)

**Before:**
```go
// main.go
db := NewDatabase()
lokstra_registry.Register("database", db)

userSvc := NewUserService(db)
lokstra_registry.Register("userService", userSvc)
```

**After:**
```yaml
# config.yaml
services:
  database:
    type: database-factory
  
  user-service:
    type: user-service-factory
    depends-on: [database]  # Auto lazy-loaded!
```

```go
// main.go
reg := deploy.Global()
reg.RegisterServiceType("database-factory", dbFactory, nil)
reg.RegisterServiceType("user-service-factory", userServiceFactory, nil)

dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "development", reg)
```

**Benefits:**
- ✅ No manual wiring
- ✅ Lazy loading (DB created on first use)
- ✅ Type-safe with `service.Cached[T]`
- ✅ Easy to test (mock config)

### Example 04 (Multi-Deployment)

**Before:**
```go
// main.go - 160 lines with flag parsing
flag.StringVar(&mode, "mode", "monolith", "...")
if mode == "monolith" {
    registerMonolithServices()
    // ... setup routes
    runMonolithServer()
} else if mode == "user-service" {
    registerUserServices()
    // ... setup routes
    runUserServiceServer()
} // ... more conditions
```

**After:**
```yaml
# config.yaml - Declarative!
deployments:
  monolith:
    servers:
      main:
        apps:
          - port: 3003
            services: [user-service, order-service]
  
  microservices-user:
    servers:
      user-api:
        apps:
          - port: 3004
            services: [user-service]
  
  microservices-order:
    servers:
      order-api:
        apps:
          - port: 3005
            services: [order-service]
```

```go
// main.go - Clean and simple!
deployment := flag.String("deployment", "monolith", "Deployment to run")
flag.Parse()

dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, *deployment, reg)
server, _ := dep.GetServer("main") // or "user-api", "order-api"
// ... start server
```

**Benefits:**
- ✅ One config file for all deployments
- ✅ No conditional logic in code
- ✅ Easy to add new deployments
- ✅ Services defined once, reused everywhere

---

## 📚 Migration Notes

Each updated example includes:
1. **MIGRATION.md** - Detailed before/after comparison
2. **config.yaml** - New configuration file
3. **Original code (commented)** - See what changed
4. **New code** - Using YAML + lazy DI

### Files Added:
- `config.yaml` - Service and deployment configuration
- `MIGRATION.md` - Explains the changes
- Service factories in `main.go`

### Files Unchanged:
- `README.md` - Updated to explain new approach
- `test.http` - Testing commands same
- Handler logic - Business logic unchanged
- Router setup - Routing unchanged

---

## 🚀 Running Examples

All examples work the same way:

```bash
# Navigate to any example
cd 01-hello-world  # or 02, 03, 04

# Run it
go run main.go

# Test it (use test.http or curl from README)
curl http://localhost:3000/
```

### For 04-multi-deployment:
```bash
cd 04-multi-deployment

# Run different deployments
go run main.go --deployment=monolith
go run main.go --deployment=microservices-user
go run main.go --deployment=microservices-order
```

---

## 🔗 Documentation

- **[Old Examples](../examples_old/)** - Original approach (still valid!)
- **[YAML Quick Ref](../../core/deploy/YAML-QUICK-REF.md)** - Config syntax
- **[Integration Guide](../../core/deploy/INTEGRASI-SISTEM-LAMA.md)** - Old vs new
- **[Complete Journey](../../core/deploy/COMPLETE-JOURNEY.md)** - Full implementation

---

## ⚠️ Important Notes

1. **Old paradigm still works 100%** - Not deprecated!
2. **Simple examples don't need YAML** - Keep them simple!
3. **New paradigm shines with complexity** - 5+ services, multi-env
4. **Choose what fits your needs** - Both approaches are valid

---

*Updated examples show when and how to use the new paradigm effectively.*
