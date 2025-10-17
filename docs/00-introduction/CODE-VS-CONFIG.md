# Code vs Config: Deployment Modes

Lokstra supports **two equivalent ways** to define deployments:

1. **Code Mode**: Define everything programmatically
2. **Config Mode**: Define everything in YAML files

Both modes use the **same underlying structure** and produce **identical results**.

---

## Parallel Structure

| Step | Code Mode | Config Mode |
|------|-----------|-------------|
| **1. Registry** | `reg := deploy.Global()` | `reg := deploy.Global()` |
| **2. Register Types** | `reg.RegisterServiceType(...)` | `reg.RegisterServiceType(...)` |
| **3. Define Services** | `reg.DefineService(&schema.ServiceDef{...})` | `services:` in YAML |
| **4. Build Deployment** | `dep := deploy.New("development")` | `loader.LoadAndBuild(...)` |
| **5. Create Structure** | `server := dep.NewServer(...)`<br>`app := server.NewApp(...)`<br>`app.AddService(...)` | Done automatically by loader |
| **6. Lazy Load** | `service.LazyLoadFrom[T](app, "name")` | `service.LazyLoadFrom[T](app, "name")` |

**Key Insight**: Steps 3-5 in **code mode** are equivalent to the **YAML definition** in config mode.

---

## Example: User Service

### Code Mode

```go
// 1. Get registry
reg := deploy.Global()

// 2. Register service factories
reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

// 3. Define services (like YAML structure)
reg.DefineService(&schema.ServiceDef{
    Name: "database",
    Type: "database-factory",
})
reg.DefineService(&schema.ServiceDef{
    Name:      "user-service",
    Type:      "user-service-factory",
    DependsOn: []string{"database"},
})

// 4. Build deployment structure
dep := deploy.New("development")
server := dep.NewServer("api", ":3002")
app := server.NewApp(3002)

// 5. Add services to app
app.AddService("database")
app.AddService("user-service")

// 6. Lazy load service
userService := service.LazyLoadFrom[*UserService](app, "user-service")
```

### Config Mode

```yaml
# config.yaml - Steps 3-5 in declarative form
services:
  - name: database
    type: database-factory

  - name: user-service
    type: user-service-factory
    depends-on:
      - database

deployments:
  development:
    servers:
      - name: api
        base-url: ":3002"
        apps:
          - port: 3002
            services:
              - database
              - user-service
```

```go
// 1. Get registry
reg := deploy.Global()

// 2. Register service factories
reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

// 3-5. Load from YAML (does Define + Build + Add automatically)
dep, err := loader.LoadAndBuild(
    []string{"config.yaml"},
    "development",
    reg,
)

// 6. Lazy load service (SAME as code mode)
server, _ := dep.GetServer("api")
app := server.Apps()[0]
userService := service.LazyLoadFrom[*UserService](app, "user-service")
```

---

## When to Use Each Mode

### Use Code Mode When:
- ✅ Prototyping or learning the framework
- ✅ Simple single-deployment apps
- ✅ You prefer type-safety and IDE autocomplete
- ✅ Configuration is dynamic (based on runtime conditions)

### Use Config Mode When:
- ✅ **Production deployments** (recommended)
- ✅ Multiple environments (dev, staging, prod)
- ✅ Multiple deployments in one project
- ✅ Non-developers need to modify configuration
- ✅ Configuration management and versioning

---

## Mixing Both Modes

You can combine both approaches:

```go
// Load base config from YAML
dep, _ := loader.LoadAndBuild(files, "production", reg)

// Add additional services in code
server, _ := dep.GetServer("api")
app := server.Apps()[0]
app.AddService("monitoring-service")  // Added programmatically
```

---

## Best Practices

### 1. **Always Use Global Registry**
Both modes should use `deploy.Global()` for consistency.

### 2. **Register Types Before Defining Services**
```go
// ✅ Correct order
reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
reg.DefineService(&schema.ServiceDef{Name: "db", Type: "database-factory"})

// ❌ Wrong - will panic (service type not found)
reg.DefineService(&schema.ServiceDef{Name: "db", Type: "database-factory"})
reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
```

### 3. **Use LazyLoadFrom() for Services**
Both modes should use the same loading pattern:
```go
// ✅ Type-safe and lazy
userService := service.LazyLoadFrom[*UserService](app, "user-service")

// ❌ Don't use manual GetService + cast
rawService, _ := app.GetService("user-service")
userService := rawService.(*UserService)
```

### 4. **Keep Service Factories Separate**
Define factories once, use in both modes:
```go
// services.go
func UserServiceFactory(cfg service.Config) (any, error) {
    db := service.MustCast[*Database](cfg.GetDependency("database"))
    return &UserService{DB: db}, nil
}
```

---

## Migration Path

If you start with **code mode** and want to move to **config mode**:

1. Extract your `DefineService` calls → `services:` section in YAML
2. Extract your deployment structure → `deployments:` section in YAML
3. Keep service factories and business logic in code
4. Everything after `LazyLoadFrom()` stays the same!

**Example**: See `docs/00-introduction/examples/03-crud-api/` for a working dual-mode example.

---

## Summary

- Both modes use **identical runtime behavior**
- **Code mode** = explicit, programmatic, type-safe
- **Config mode** = declarative, maintainable, environment-friendly
- Choose based on your **deployment complexity** and **team needs**
- Production apps typically use **config mode** with YAML files
