# Reverse Proxy Configuration - Implementation Summary

## Overview

Fitur **Reverse Proxy Configuration** memungkinkan Lokstra digunakan sebagai API Gateway yang di-konfigurasi melalui YAML, tanpa perlu menulis kode Go.

## Struktur Konfigurasi

### Config Struct (`core/config/config.go`)

```go
// ReverseProxyConfig represents reverse proxy configuration for an app
type ReverseProxyConfig struct {
    Prefix      string `yaml:"prefix"`       // URL prefix (/api)
    StripPrefix bool   `yaml:"strip-prefix"` // Strip prefix sebelum forward
    Target      string `yaml:"target"`       // Target backend URL
}

// App configuration
type App struct {
    Name           string
    Addr           string
    ReverseProxies []*ReverseProxyConfig `yaml:"reverse-proxies"`
    Routers        []string
}
```

### App Method (`core/app/app.go`)

```go
// AddReverseProxies membuat router untuk reverse proxy dan mount-nya
func (a *App) AddReverseProxies(proxies []*ReverseProxyConfig) {
    if len(proxies) == 0 {
        return
    }
    
    // Create dedicated router
    proxyRouter := router.New(a.name + "-reverse-proxy")
    
    for _, proxy := range proxies {
        stripPrefix := ""
        if proxy.StripPrefix {
            stripPrefix = proxy.Prefix
        }
        
        handler := lokstra_handler.MountReverseProxy(stripPrefix, proxy.Target)
        proxyRouter.ANYPrefix(proxy.Prefix, handler)
    }
    
    // Add proxy router to app (prepended)
    a.AddRouter(proxyRouter)
}
```

### Config Loader (`lokstra_registry/config.go`)

```go
// After creating app, add reverse proxies
a := app.New(appConfig.GetName(i), appConfig.Addr, routers...)

if len(appConfig.ReverseProxies) > 0 {
    proxies := make([]*app.ReverseProxyConfig, len(appConfig.ReverseProxies))
    for j, rp := range appConfig.ReverseProxies {
        proxies[j] = &app.ReverseProxyConfig{
            Prefix:      rp.Prefix,
            StripPrefix: rp.StripPrefix,
            Target:      rp.Target,
        }
    }
    a.AddReverseProxies(proxies)
}
```

## Usage Examples

### 1. Pure Reverse Proxy Gateway

```yaml
servers:
  - name: api-gateway
    base-url: http://localhost
    apps:
      - name: gateway
        addr: ":8080"
        reverse-proxies:
          - prefix: /api
            strip-prefix: true
            target: http://api-server:8080
```

**Effect:**
- `GET /api/users` → `GET http://api-server:8080/users`
- `POST /api/products` → `POST http://api-server:8080/products`

### 2. Multiple Backend Services

```yaml
reverse-proxies:
  - prefix: /users
    strip-prefix: true
    target: http://user-service:8001
  
  - prefix: /orders
    strip-prefix: true
    target: http://order-service:8002
  
  - prefix: /payments
    strip-prefix: false  # Keep prefix
    target: http://payment-service:8003
```

### 3. Hybrid Mode (Proxy + Routers)

```yaml
apps:
  - name: hybrid
    addr: ":8090"
    reverse-proxies:
      - prefix: /external
        target: http://external-api:8080
    routers:
      - internal-router  # Regular router
```

**Execution order:**
1. Reverse proxies mounted first
2. Then regular routers

## Code Equivalent

Config-based approach:
```yaml
reverse-proxies:
  - prefix: /api
    strip-prefix: true
    target: http://backend:8080
```

Equivalent code:
```go
r := lokstra.NewRouter("gateway")
r.ANYPrefix("/api", lokstra_handler.MountReverseProxy("/api", "http://backend:8080"))
app := lokstra.NewApp("gateway", ":8080", r)
```

## Benefits

✅ **Zero Code** - Purely configuration-driven  
✅ **Dynamic** - Edit YAML without rebuild  
✅ **Simple** - Easy to understand and maintain  
✅ **Flexible** - Combine with routers if needed  
✅ **Production-Ready** - Uses standard `httputil.ReverseProxy`

## Implementation Details

### Flow

1. **Config Load** → `lokstra_registry.LoadConfig()`
2. **Parse YAML** → Extract `reverse-proxies` from app config
3. **Create App** → `app.New(...)`
4. **Mount Proxies** → `app.AddReverseProxies(proxies)`
5. **Auto-Router** → Creates `<app-name>-reverse-proxy` router
6. **Register Routes** → Each proxy becomes `ANYPrefix` route

### Strip Prefix Behavior

| Config | Request | Forwarded To |
|--------|---------|--------------|
| `strip-prefix: true` | `/api/users` | `target/users` |
| `strip-prefix: false` | `/api/users` | `target/api/users` |

### Router Priority

1. **Reverse Proxies** (mounted first)
2. **Regular Routers** (mounted after)

This ensures proxies can act as catch-all before specific routers handle requests.

## Testing

See `cmd_draft/examples/reverse-proxy-gateway/` for:
- ✅ Config examples (`config.yaml`)
- ✅ Code examples (`main.go`)
- ✅ Documentation (`README.md`)

## Future Enhancements

Potential improvements:
- [ ] `strip-prefix` as string (for custom prefix stripping)
- [ ] Middleware support for reverse proxies
- [ ] Load balancing (multiple targets)
- [ ] Circuit breaker integration
- [ ] Request/response transformation
- [ ] Authentication/authorization hooks

## Files Changed

1. ✅ `core/config/config.go` - Added `ReverseProxyConfig` struct
2. ✅ `core/app/app.go` - Added `AddReverseProxies()` method
3. ✅ `lokstra_registry/config.go` - Auto-mount proxies from config
4. ✅ `cmd_draft/examples/reverse-proxy-gateway/` - Examples & docs

---

**Date:** October 14, 2025  
**Status:** ✅ Implemented & Tested  
**Breaking Changes:** None (backward compatible)
