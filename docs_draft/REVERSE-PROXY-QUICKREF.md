# Reverse Proxy Quick Reference

## YAML Configuration

### Minimal Example
```yaml
servers:
  - name: gateway
    apps:
      - addr: ":8080"
        reverse-proxies:
          - prefix: /api
            target: http://backend:8080
```

### With Strip Prefix
```yaml
reverse-proxies:
  - prefix: /api
    strip-prefix: true  # /api/users → backend/users
    target: http://backend:8080
```

### Multiple Backends
```yaml
reverse-proxies:
  - prefix: /users
    strip-prefix: true
    target: http://user-service:8001
  
  - prefix: /orders
    strip-prefix: true
    target: http://order-service:8002
```

### Hybrid (Proxy + Router)
```yaml
apps:
  - addr: ":8080"
    reverse-proxies:
      - prefix: /external
        target: http://external-api:8080
    routers:
      - my-router
```

## Code Equivalent

### Direct Code
```go
r := lokstra.NewRouter("gateway")
r.ANYPrefix("/api", lokstra_handler.MountReverseProxy("/api", "http://backend:8080"))
app := lokstra.NewApp("gateway", ":8080", r)
```

### Using App Method
```go
app := lokstra.NewApp("gateway", ":8080")
app.AddReverseProxies([]*app.ReverseProxyConfig{
    {
        Prefix:      "/api",
        StripPrefix: true,
        Target:      "http://backend:8080",
    },
})
```

## Strip Prefix Behavior

| strip-prefix | Request Path | Forwarded Path |
|--------------|--------------|----------------|
| `true` | `/api/users` | `backend/users` |
| `false` | `/api/users` | `backend/api/users` |
| (omitted) | `/api/users` | `backend/api/users` |

## Router Priority

```
1. Reverse Proxies ← FIRST
2. Regular Routers
3. 404 Not Found
```

## Use Cases

### API Gateway
```yaml
# Route all API traffic
reverse-proxies:
  - prefix: /
    target: http://api-server:8080
```

### Microservices
```yaml
# Route by service
reverse-proxies:
  - prefix: /users
    target: http://users:8001
  - prefix: /posts
    target: http://posts:8002
```

### API Versioning
```yaml
# Route by version
reverse-proxies:
  - prefix: /v1
    target: http://api-v1:8080
  - prefix: /v2
    target: http://api-v2:8081
```

### Development
```yaml
# Forward to local backend
reverse-proxies:
  - prefix: /api
    strip-prefix: true
    target: http://localhost:3000
```

## Testing

```bash
# Build
go build ./core/app

# Test
go test ./core/app -run TestAddReverseProxies

# Run example
go run ./cmd_draft/examples/reverse-proxy-gateway
```

## Key Points

✅ Config-driven (no code changes)  
✅ Prepended before regular routers  
✅ Multiple backends supported  
✅ Hybrid mode available  
✅ Production-ready (`httputil.ReverseProxy`)

## Files

- **Config:** `core/config/config.go`
- **Implementation:** `core/app/app.go`
- **Loader:** `lokstra_registry/config.go`
- **Tests:** `core/app/app_reverse_proxy_test.go`
- **Examples:** `cmd_draft/examples/reverse-proxy-gateway/`
- **Docs:** `docs_draft/reverse-proxy-config-implementation.md`
