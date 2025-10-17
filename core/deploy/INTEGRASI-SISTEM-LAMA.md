# Integrasi dengan Sistem Lama (Router & Middleware)

## Status Sistem Lama vs Baru

### 🟢 Sistem yang Masih Digunakan (Tidak Berubah)

#### 1. **Router System** (`core/router/`)
Status: **MASIH DIGUNAKAN & TIDAK BERUBAH**

```go
// Router lama masih berfungsi 100%
router := router.New()
router.RegisterHandler("User", "GetAll", handler)
router.RegisterHandler("User", "GetById", handler)

// Auto-router dari service masih bekerja
router.RegisterService(userService, "User")
```

**Yang baru hanya cara KONFIGURASI-nya** (bisa dari YAML):
```yaml
# Sekarang bisa dikonfigurasi via YAML
routers:
  user-router:
    service: user-service
    overrides:
      GetAll:
        hide: false
        middleware: [auth, logging]
```

#### 2. **Middleware System** (`middleware/`)
Status: **MASIH DIGUNAKAN & TIDAK BERUBAH**

```go
// Middleware lama masih berfungsi
authMW := middleware.Auth(...)
logMW := middleware.Logging(...)

// Masih bisa digunakan seperti biasa
router.Use(authMW, logMW)
```

**Yang baru hanya cara REGISTRASI-nya**:
```yaml
# Bisa dikonfigurasi via YAML (opsional)
middlewares:
  auth:
    type: auth-middleware
    config:
      jwt-secret: ${JWT_SECRET}
```

#### 3. **Service System** (`lokstra_registry/`)
Status: **MASIH DIGUNAKAN & ENHANCED**

```go
// Cara lama masih bekerja
lokstra_registry.Register("userService", userSvcInstance)
svc := lokstra_registry.Get[*UserService]("userService")

// Sekarang ada TAMBAHAN: lazy loading
lazy := service.LazyLoad[*UserService]("userService")
svc := lazy.Get() // Loaded on demand
```

### 🆕 Sistem Baru (Tambahan, Bukan Pengganti)

#### 1. **Deployment Configuration** (`core/deploy/`)
**BARU**: Framework untuk organizing services, routers, middleware dalam YAML

```yaml
# Ini BARU - cara deklaratif untuk setup
deployments:
  production:
    servers:
      api-server:
        base-url: https://api.com
        apps:
          - port: 8080
            services: [user-service, order-service]
            routers: [user-router, order-router]
```

#### 2. **Lazy Dependency Injection** (`core/service/lazy_load.go`)
**BARU**: Type-safe lazy loading untuk dependency injection

```go
// BARU - typed lazy loading
type UserService struct {
    DB     *service.Cached[*DBPool]    // Lazy
    Logger *service.Cached[*Logger]    // Lazy
}

func (us *UserService) GetUser(id int) {
    db := us.DB.Get() // Resolved on demand
}
```

#### 3. **YAML Configuration** (`core/deploy/loader/`)
**BARU**: Load config dari YAML dengan validation

```go
// BARU - load config dari YAML
config, err := loader.LoadConfig("config.yaml")
dep, err := loader.BuildDeployment(config, "production", registry)
```

---

## 🔄 Cara Integrasi: Sistem Lama + Sistem Baru

### Scenario 1: Existing App - Tanpa Perubahan
**Sistem lama tetap jalan 100% tanpa perubahan apapun!**

```go
// main.go - TIDAK PERLU DIUBAH SAMA SEKALI
func main() {
    // Router lama masih bekerja
    router := router.New()
    
    // Service registry lama masih bekerja
    lokstra_registry.Register("userService", &UserService{})
    
    // Middleware lama masih bekerja
    router.Use(middleware.Auth(), middleware.Logging())
    
    // Start server seperti biasa
    http.ListenAndServe(":8080", router)
}
```

### Scenario 2: New App - Pakai Sistem Baru Penuh

```go
// main.go - NEW STYLE
func main() {
    // 1. Register factories
    reg := deploy.Global()
    reg.RegisterServiceType("user-service", userServiceFactory, nil)
    reg.RegisterServiceType("order-service", orderServiceFactory, nil)
    
    // 2. Load config dari YAML
    dep, err := loader.LoadAndBuildFromDir("config", "production", reg)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. Get app
    server, _ := dep.GetServer("api-server")
    app := server.Apps()[0]
    
    // 4. Services sudah auto-wired dengan lazy DI
    userSvc, _ := app.GetService("user-service")
    
    // 5. Router masih pakai sistem lama
    router := router.New()
    router.RegisterService(userSvc, "User")
    
    // 6. Start server
    http.ListenAndServe(":8080", router)
}
```

### Scenario 3: Hybrid - Mix Lama & Baru

```go
// main.go - HYBRID STYLE
func main() {
    // Services dari YAML (baru)
    reg := deploy.Global()
    reg.RegisterServiceType("user-service", userServiceFactory, nil)
    
    dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "production", reg)
    server, _ := dep.GetServer("api")
    app := server.Apps()[0]
    
    // Router pakai cara lama (masih works!)
    router := router.New()
    
    // Get service dari deployment baru
    userSvc, _ := app.GetService("user-service")
    
    // Register ke router lama
    router.RegisterService(userSvc, "User")
    
    // Middleware pakai cara lama (masih works!)
    router.Use(
        middleware.Auth(),
        middleware.Logging(),
    )
    
    http.ListenAndServe(":8080", router)
}
```

---

## 🎯 Kapan Pakai Yang Mana?

### Pakai Sistem Lama Saja (Router + Middleware Langsung):
✅ **Ketika:**
- App kecil (< 5 services)
- Tidak perlu multi-environment
- Konfigurasi sederhana
- Tidak perlu lazy loading
- Sudah ada app yang jalan (jangan ubah yang sudah jalan!)

```go
// Simple & direct
router := router.New()
userSvc := &UserService{DB: dbPool, Logger: logger}
router.RegisterService(userSvc, "User")
router.Use(middleware.Auth())
```

### Pakai Sistem Baru (Deployment + YAML):
✅ **Ketika:**
- App besar (> 5 services)
- Multi-environment (dev, staging, prod)
- Complex dependencies
- Butuh lazy loading (avoid circular deps)
- Butuh config validation
- Ingin deployment declarative

```yaml
# config.yaml - Declarative!
services:
  user-service:
    type: user-service-factory
    depends-on: [db, logger, cache]
    
deployments:
  production:
    config-overrides:
      DB_HOST: prod-db.com
    servers:
      api: {...}
```

---

## 📦 Komponen yang TIDAK Berubah

### 1. Router API
```go
// ✅ MASIH SAMA
type Router interface {
    RegisterHandler(resource, method string, handler HandlerFunc)
    RegisterService(service any, resourceName string)
    Use(middlewares ...Middleware)
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// ✅ MASIH BISA DIPAKAI
router := router.New()
router.RegisterHandler("User", "GetAll", handler)
```

### 2. Middleware API
```go
// ✅ MASIH SAMA
type Middleware func(next HandlerFunc) HandlerFunc

// ✅ MASIH BISA DIPAKAI
func AuthMiddleware() Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx *Context) error {
            // Auth logic
            return next(ctx)
        }
    }
}
```

### 3. Service Registry
```go
// ✅ MASIH SAMA
lokstra_registry.Register("userService", instance)
svc := lokstra_registry.Get[*UserService]("userService")

// 🆕 TAMBAHAN (opsional)
lazy := service.LazyLoad[*UserService]("userService")
svc := lazy.Get()
```

---

## 🔌 Integration Points

### Point 1: Service Instantiation
**Lama:**
```go
// Manual instantiation
dbPool := &DBPool{Host: "localhost"}
logger := &Logger{Level: "info"}
userSvc := &UserService{DB: dbPool, Logger: logger}

lokstra_registry.Register("userService", userSvc)
```

**Baru:**
```go
// Factory-based dengan lazy DI
func userServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB:     service.Cast[*DBPool](deps["db"]),
        Logger: service.Cast[*Logger](deps["logger"]),
    }
}

// Config di YAML
services:
  user-service:
    type: user-service-factory
    depends-on: [db, logger]
```

### Point 2: Configuration
**Lama:**
```go
// Hardcoded config
dbHost := "localhost"
dbPort := 5432
logLevel := "info"
```

**Baru:**
```yaml
# config.yaml
configs:
  DB_HOST: localhost
  DB_PORT: 5432
  LOG_LEVEL: info

# Per environment
deployments:
  production:
    config-overrides:
      DB_HOST: prod-db.com
      LOG_LEVEL: warn
```

### Point 3: Router Setup
**Lama (masih works!):**
```go
router := router.New()
router.RegisterService(userService, "User")
router.Use(authMW, logMW)
```

**Baru (opsional - belum fully implemented):**
```yaml
routers:
  user-router:
    service: user-service
    overrides:
      GetAll:
        middleware: [auth, logging]
```

```go
// TODO: Auto-setup router dari YAML
// Ini belum implemented - masih manual setup
router := router.New()
// ... setup from config
```

---

## 🚧 Yang Belum Implemented

### 1. Router Auto-Setup dari YAML
**Status: BELUM IMPLEMENTED**

```yaml
# Schema sudah ada, tapi builder belum
routers:
  user-router:
    service: user-service
```

**Perlu ditambahkan di `builder.go`:**
```go
// TODO: Implement
func (a *App) SetupRouters(routerDefs map[string]*schema.RouterDefSimple) {
    for name, rtrDef := range routerDefs {
        svc, _ := a.GetService(rtrDef.Service)
        router := router.New()
        router.RegisterService(svc, name)
        // Apply overrides...
        a.routers[name] = router
    }
}
```

### 2. Middleware Auto-Registration dari YAML
**Status: BELUM IMPLEMENTED**

```yaml
# Schema ada, tapi tidak diproses
middlewares:
  auth:
    type: jwt-auth
    config:
      secret: ${JWT_SECRET}
```

**Perlu ditambahkan:**
```go
// TODO: Implement middleware factory system
reg.RegisterMiddlewareType("jwt-auth", jwtAuthFactory)
```

### 3. Remote Service Integration
**Status: PARTIAL**

```yaml
# Schema ada
remote-services:
  payment-api:
    url: https://payment.com
    resource: payment
```

**Perlu ditambahkan:**
```go
// TODO: Auto-create API client from remote service config
```

---

## 📝 Rekomendasi Migrasi

### Untuk Existing Apps:
1. **JANGAN UBAH** - Sistem lama masih 100% berfungsi
2. **Jika ingin lazy DI** - Tambahkan `service.Cached[T]` secara bertahap
3. **Jika ingin YAML config** - Tambahkan loader untuk services baru
4. **Router tetap manual** - Belum perlu ubah router setup

### Untuk New Apps:
1. **Mulai dengan YAML config** - Lebih scalable
2. **Pakai lazy DI** - Hindari circular dependencies
3. **Router tetap manual** - Tunggu auto-setup implemented
4. **Middleware manual** - Tunggu factory system

### Migration Path:
```
Phase 1: Add lazy DI to new services
  ↓
Phase 2: Move configs to YAML
  ↓
Phase 3: Use deployment builder for service wiring
  ↓
Phase 4: Keep router/middleware manual (belum auto)
  ↓
Future: Auto router/middleware setup
```

---

## 🎯 Kesimpulan

| Component | Status | Pakai Sistem Lama? | Pakai Sistem Baru? |
|-----------|--------|-------------------|-------------------|
| **Router** | ✅ Unchanged | ✅ YA - Manual setup | ⏳ Auto-setup belum ready |
| **Middleware** | ✅ Unchanged | ✅ YA - Manual register | ⏳ Factory belum ready |
| **Service Registry** | ✅ Enhanced | ✅ YA - Masih works | ✅ YA - Plus lazy loading |
| **Configuration** | 🆕 New option | ✅ YA - Hardcode/env | ✅ YA - YAML preferred |
| **Dependency Injection** | 🆕 New | ❌ Manual wiring | ✅ YA - Auto lazy DI |
| **Deployment Setup** | 🆕 New | ❌ N/A | ✅ YA - YAML declarative |

**TL;DR:**
- ✅ **Sistem lama 100% masih berfungsi**
- ✅ **Tidak perlu migrasi paksa**
- ✅ **Sistem baru = tambahan fitur, bukan pengganti**
- ⚠️ **Router/middleware auto-setup belum implemented**
- 🎯 **Hybrid approach recommended: YAML config + manual router**
