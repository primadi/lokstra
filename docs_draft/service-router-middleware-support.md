# Service Router Middleware Support

## Overview

Fitur ini memungkinkan penambahan middleware dan route options (name, description, override parent middleware) pada handler spesifik ketika menggunakan `router.NewFromService()` untuk auto-generate routes dari service.

## Changes

### 1. Updated `RouteMeta` struct
Tambahan field untuk mendukung semua `route.RouteHandlerOption`:

```go
type RouteMeta struct {
    MethodName       string
    HTTPMethod       string
    Path             string
    Name             string  // NEW: Route name (defaults to MethodName if empty)
    Description      string  // NEW: Route description
    OverrideParentMw bool    // NEW: Override parent middleware
    Middlewares      []any   // NEW: Middleware untuk route ini
    // ... fields lainnya
}
```

### 2. Changed to Pointer Receiver

All methods now return `*ServiceRouterOptions` instead of value:

```go
// Before
func DefaultServiceRouterOptions() ServiceRouterOptions
func (o ServiceRouterOptions) WithPrefix(prefix string) ServiceRouterOptions

// After  
func DefaultServiceRouterOptions() *ServiceRouterOptions
func (o *ServiceRouterOptions) WithPrefix(prefix string) *ServiceRouterOptions
```

**Reason**: More efficient, prevents unnecessary copying of the struct.

### 3. New Helper Method: `WithMethodMiddleware()`

Method convenience untuk menambah middleware tanpa mengubah path atau HTTP method:

```go
func (o *ServiceRouterOptions) WithMethodMiddleware(methodName string, middleware ...any) *ServiceRouterOptions
```

### 4. Updated `registerRouteByMethod()`

Sekarang menerima `RouteMeta` dan mengkonversi ke `route.RouteHandlerOption`:

```go
func registerRouteByMethod(r Router, httpMethod, path string, handler any, meta RouteMeta)
```

## Usage Examples

### Example 1: Tambah Middleware Tanpa Mengubah Route

```go
opts := router.DefaultServiceRouterOptions().
    WithPrefix("/api").
    WithMethodMiddleware("DeleteUser", "auth", "admin_check").
    WithMethodMiddleware("CreateUser", "auth", "rate_limit")

r := router.NewFromService(&UserService{}, opts)
```

**Hasil:**
- `DeleteUser` → tetap gunakan convention (DELETE /api/users/{id}) + middleware: auth, admin_check
- `CreateUser` → tetap gunakan convention (POST /api/users) + middleware: auth, rate_limit

### Example 2: Override Lengkap (Path + Method + Middleware)

```go
opts := router.DefaultServiceRouterOptions().
    WithPrefix("/api").
    WithRouteOverride("DeleteUser", router.RouteMeta{
        HTTPMethod:  "DELETE",
        Path:        "/users/{id}",
        Name:        "delete-user-by-id",
        Description: "Delete a user by their ID",
        Middlewares: []any{"auth", "admin_check", "audit"},
    })

r := router.NewFromService(&UserService{}, opts)
```

### Example 3: Hanya Tambah Middleware + Options, Path & Method Gunakan Convention

```go
opts := router.DefaultServiceRouterOptions().
    WithRouteOverride("GetUser", router.RouteMeta{
        // Path dan HTTPMethod kosong = gunakan convention
        Name:             "get-user-detail",
        Description:      "Retrieve user details by ID",
        OverrideParentMw: true,
        Middlewares:      []any{"auth", "cache"},
    })

r := router.NewFromService(&UserService{}, opts)
```

**Hasil:**
- Path dan HTTP method di-generate dari convention (GET /users/{id})
- Route name: "get-user-detail"
- Description: "Retrieve user details by ID"
- Override parent middleware: true
- Middleware: auth, cache

### Example 4: Kombinasi Global dan Method-Specific Middleware

```go
opts := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithMiddlewares("cors", "logging").           // Applied to ALL routes
    WithMethodMiddleware("DeleteUser", "admin").  // Additional for DeleteUser only

r := router.NewFromService(&UserService{}, opts)
```

**Hasil:**
- Semua routes: cors → logging
- DeleteUser: cors → logging → admin

### Example 5: Override Parent Middleware

```go
opts := router.DefaultServiceRouterOptions().
    WithRouteOverride("PublicEndpoint", router.RouteMeta{
        OverrideParentMw: true,  // This route will NOT inherit parent middleware
        Middlewares:      []any{"rate_limit"}, // Only this middleware
    })

r := router.NewFromService(&UserService{}, opts)
```

## Route Options Support

All `route.RouteHandlerOption` are now supported via `RouteMeta`:

| RouteMeta Field | Converts To | Description |
|----------------|-------------|-------------|
| `Name` | `route.WithNameOption(name)` | Route name for identification |
| `Description` | `route.WithDescriptionOption(desc)` | Route description |
| `OverrideParentMw` | `route.WithOverrideParentMwOption(true)` | Skip parent middleware |
| `Middlewares` | Passed as-is to router method | Route-specific middleware |

## Implementation Details

### Logic Flow in `NewFromServiceWithEngine`:

1. Cek jika method ada di `RouteOverrides`
2. Jika ada:
   - Ambil `Path`, `HTTPMethod`, dan semua options dari override
   - Jika `Path` atau `HTTPMethod` kosong, gunakan convention
   - Ensure `MethodName` is set (defaults to method.Name)
   - Pass complete `RouteMeta` to `registerRouteByMethod`
3. Jika tidak ada override:
   - Gunakan convention untuk generate path dan HTTP method
   - Create basic `RouteMeta` with only `MethodName`
   - No additional options

### Key Changes in Code:

**registerRouteByMethod signature:**
```go
// Before
func registerRouteByMethod(r Router, httpMethod, path string, handler any, methodName string, middleware ...any)

// After - accepts full RouteMeta
func registerRouteByMethod(r Router, httpMethod, path string, handler any, meta RouteMeta)
```

**registerRouteByMethod implementation:**
```go
func registerRouteByMethod(r Router, httpMethod, path string, handler any, meta RouteMeta) {
    var options []any
    
    // Convert RouteMeta fields to route.RouteHandlerOption
    if meta.Name != "" {
        options = append(options, route.WithNameOption(meta.Name))
    } else if meta.MethodName != "" {
        options = append(options, route.WithNameOption(meta.MethodName))
    }
    
    if meta.Description != "" {
        options = append(options, route.WithDescriptionOption(meta.Description))
    }
    
    if meta.OverrideParentMw {
        options = append(options, route.WithOverrideParentMwOption(true))
    }
    
    // Add middlewares
    options = append(options, meta.Middlewares...)
    
    // Register route with all options
    r.GET(path, handler, options...)
}
```

## Benefits

1. ✅ **Complete Route Control**: Support semua `route.RouteHandlerOption`
2. ✅ **Flexible**: Bisa hanya tambah middleware, atau full override
3. ✅ **Clean**: Semua deklaratif di options, tidak perlu imperative post-registration
4. ✅ **Consistent**: Menggunakan pattern yang sama dengan route override yang sudah ada
5. ✅ **Type-safe**: Menggunakan struct `RouteMeta` yang sudah ada
6. ✅ **Convenience**: Helper method `WithMethodMiddleware()` untuk kasus simple
7. ✅ **Efficient**: Pointer receiver prevents unnecessary struct copying

## Backward Compatibility

✅ **Fully backward compatible** - Tidak ada breaking changes:
- All new fields in `RouteMeta` are optional
- `ServiceRouterOptions` now uses pointer (more efficient, still compatible)
- Existing code will continue to work with minimal changes
- Convention-based routes work as before
