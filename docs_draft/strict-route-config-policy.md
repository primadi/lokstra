# Strict Route Configuration Policy

## Philosophy: Code-First, Config for Deployment Customization

Route configuration in YAML follows a **strict middleware-only policy**:

### ✅ What Routes Can Do in Config:
1. **Add middleware** to existing routes (`use`)
2. **Disable routes** for specific environments (`enable: false`)  
3. **Override parent middleware inheritance** (`override-parent-mw: true`)

### ❌ What Routes CANNOT Do in Config:
1. **Change path** - paths must be defined in code
2. **Change method** - HTTP methods must be defined in code
3. **Change handler** - handlers must be defined in code
4. **Create new routes** - routes must be registered in code first

## Configuration Structure

```yaml
routers:
  - name: router-name              # Must match code registration
    use: [middleware1, middleware2] # Router-level middleware (inherited by all routes)
    routes:
      - name: route-name           # Must match route name in code
        use: [middleware3]         # Additional middleware for this route only
        enable: true               # Default: true, set false to disable route
        override-parent-mw: false  # Default: false, set true to not inherit router middleware
```

## Middleware Inheritance Behavior

### Normal Inheritance (default):
```yaml
routers:
  - name: api
    use: [logger, cors]    # Router middleware
    routes:
      - name: users
        use: [auth, cache] # Route middleware
```
**Result**: Route gets `[logger, cors, auth, cache]` (router + route middleware)

### Override Parent Middleware:
```yaml
routers:
  - name: api  
    use: [logger, cors]    # Router middleware
    routes:
      - name: public-info
        use: [cache]       # Route middleware only
        override-parent-mw: true
```
**Result**: Route gets `[cache]` only (no router middleware inherited)

## Example: Complete Configuration

```yaml
middlewares:
  - name: logger
    type: logger
  - name: cors
    type: cors
  - name: auth
    type: auth
  - name: rate-limit
    type: rate-limit

routers:
  - name: api
    use: [logger, cors]              # All API routes get logger + cors
    routes:
      - name: health
        use: [rate-limit]            # health gets: logger, cors, rate-limit
        
      - name: public-data
        use: [cache]                 # public-data gets: logger, cors, cache
        
      - name: internal-debug
        use: [auth]                  # internal-debug gets: auth only (no logger, cors)
        override-parent-mw: true
        
      - name: maintenance-endpoint
        enable: false                # Route disabled in this environment

  - name: admin
    use: [logger, auth]              # All admin routes need auth
    routes:
      - name: dashboard
        use: [audit]                 # dashboard gets: logger, auth, audit
        
      - name: public-status
        use: [cache]                 # public-status gets: cache only
        override-parent-mw: true     # No auth required for public status
```

## Benefits of This Approach

### 1. **Clear Separation of Concerns**
- **Code**: Business logic, paths, methods, handlers  
- **Config**: Deployment-specific middleware, environment toggles

### 2. **Type Safety**
- Routes, paths, methods validated at compile time
- Configuration errors caught at startup, not runtime

### 3. **Deployment Flexibility**
- Same code, different middleware per environment
- Can disable features without code changes
- Can add security/monitoring without touching business logic

### 4. **No Configuration Needed for Basic Cases**
If you're happy with middleware defined in code, **no router configuration needed at all**:

```yaml
# Minimal config - no router configuration
services:
  - name: database
    type: postgres
    config: {...}

servers:
  - name: web-server
    services: [database]
```

## Migration from Current Examples

### Before (with path/method):
```yaml
routes:
  - name: users
    path: /api/users    # ❌ Removed
    method: GET         # ❌ Removed
    use: [auth]
```

### After (middleware-only):
```yaml  
routes:
  - name: users         # ✅ Must match code registration
    use: [auth]         # ✅ Add middleware only
```

This makes configuration **simpler**, **safer**, and **more focused** on deployment concerns rather than application structure.