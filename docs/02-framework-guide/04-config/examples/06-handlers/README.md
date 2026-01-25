# Example 06 - Handler Configurations

**Complete runnable example** demonstrating handler configurations: SPA mounting and static file serving.

> **Note:** Reverse proxy examples require external services and are documented separately in the full README. This runnable example focuses on SPA and static file serving.

## What's Demonstrated

- ✅ SPA mounting (Single Page Applications)
- ✅ Static file serving
- ✅ Multiple SPAs on different paths
- ✅ Handler mount order
- ✅ Complete runnable example with actual HTML files

## File Structure

```
06-handlers/
├── main.go                 # Application entry
├── user_service.go         # Business service
├── user_repository.go      # Repository
├── config.yaml             # Handler configurations
├── go.mod
├── test.http
├── dist/                   # SPA builds
│   ├── admin/
│   │   └── index.html     # Admin dashboard SPA
│   └── landing/
│       └── index.html     # Landing page SPA
├── public/                 # Static assets
│   └── assets/
│       ├── logo.svg       # Logo image
│       └── style.css      # Stylesheet
└── README.md
```

## Configuration Highlights

### SPA Mounts

```yaml
mount-spa:
  # Admin dashboard at /admin
  - prefix: "/admin"
    dir: "./dist/admin"
  
  # Landing page at root
  - prefix: "/"
    dir: "./dist/landing"
```

**Behavior:**
- Routes without file extension → serve `index.html`
- Static files (`.js`, `.css`, `.png`) → serve directly
- 404 for missing files

**Examples:**
- `GET /admin` → `dist/admin/index.html`
- `GET /admin/users` → `dist/admin/index.html` (SPA client routing)
- `GET /admin/logo.png` → `dist/admin/logo.png` (if exists)
- `GET /` → `dist/landing/index.html`

### Static File Mounts

```yaml
mount-static:
  # Public assets
  - prefix: "/assets"
    dir: "./public/assets"
```

**Behavior:**
- Files served directly
- Directories → append `/index.html`
- 404 for missing files

**Examples:**
- `GET /assets/logo.svg` → `public/assets/logo.svg`
- `GET /assets/style.css` → `public/assets/style.css`

## Handler Mount Order

Handlers are mounted in this priority:

1. **Business Routers** - `/api/users` (from `published-services`)
2. **SPA Mounts** - `/admin/*`, `/*`
3. **Static Mounts** - `/assets/*`

**Important:** Most specific routes first!

```yaml
apps:
  - addr: ":8080"
    published-services: [user-service]  # 1st: /api/users
    mount-spa:
      - prefix: "/admin"                 # 2nd: /admin/*
      - prefix: "/"                      # 3rd: /* (catch-all)
    mount-static:
      - prefix: "/assets"                # 4th: /assets/*
```

## Run

```bash
# First time: generate code
go run .

# The application will start on :8080
```

## Test

### 1. Test Business API
```bash
curl http://localhost:8080/api/users
```

### 2. Test Admin SPA
```bash
# Root
curl http://localhost:8080/admin

# Client-side route (still serves index.html)
curl http://localhost:8080/admin/users

# Or open in browser
open http://localhost:8080/admin
```

### 3. Test Landing Page
```bash
curl http://localhost:8080/

# Or open in browser
open http://localhost:8080/
```

### 4. Test Static Assets
```bash
# Logo
curl http://localhost:8080/assets/logo.svg

# CSS
curl http://localhost:8080/assets/style.css
```

Or use **test.http** in VS Code with REST Client extension.

## How It Works

### 1. SPA Mounting

When you access `/admin/users`:

1. Request arrives at router
2. No exact match for `/admin/users`
3. Falls through to SPA handler at `/admin`
4. Handler checks if `/users` is a file → NO
5. Serves `dist/admin/index.html`
6. Client-side router (React/Vue/Angular) handles `/users` route

### 2. Static File Serving

When you access `/assets/logo.svg`:

1. Request arrives at router
2. Falls through to static handler at `/assets`
3. Handler serves `public/assets/logo.svg` directly

### 3. Routing Priority

```
Request: GET /api/users
  ✓ Matches business router → Handle

Request: GET /admin
  ✗ No match in business router
  ✓ Matches SPA at /admin → Serve index.html

Request: GET /assets/logo.svg
  ✗ No match in business router
  ✗ No match in SPA
  ✓ Matches static at /assets → Serve file

Request: GET /unknown
  ✗ No match in business router
  ✗ No match in specific SPAs
  ✓ Matches root SPA at / → Serve index.html (404 page)
```

## Architecture Patterns

### Pattern 1: API + Admin SPA
```yaml
apps:
  - addr: ":8080"
    published-services: [api-service]
    mount-spa:
      - prefix: "/admin"
        dir: "./dist/admin"
```

**Use case:** Admin dashboard alongside API

### Pattern 2: API + Multiple SPAs
```yaml
apps:
  - addr: ":8080"
    published-services: [api-service]
    mount-spa:
      - prefix: "/admin"
        dir: "./dist/admin"
      - prefix: "/app"
        dir: "./dist/app"
      - prefix: "/"
        dir: "./dist/landing"
```

**Use case:** Multi-tenant or multi-app platform

### Pattern 3: SPA + Static Assets
```yaml
apps:
  - addr: ":8080"
    mount-spa:
      - prefix: "/"
        dir: "./dist/spa"
    mount-static:
      - prefix: "/assets"
        dir: "./public/assets"
```

**Use case:** Pure frontend app with static assets

## Production Considerations

### 1. SPA Caching

**Problem:** index.html should never be cached
**Solution:** Add middleware

```go
// Add no-cache headers for HTML
func noCacheMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, ".html") {
            w.Header().Set("Cache-Control", "no-cache, no-repository, must-revalidate")
        }
        next.ServeHTTP(w, r)
    })
}
```

### 2. Static Asset Caching

**Problem:** Assets should be cached for performance
**Solution:** Use versioned filenames

```html
<!-- Good: versioned assets -->
<script src="/assets/app.12345.js"></script>

<!-- Bad: no versioning -->
<script src="/assets/app.js"></script>
```

### 3. Compression

**Add gzip middleware** for better performance:

```yaml
middleware-definitions:
  gzip:
    type: gzip-compression
```

### 4. CDN for Production

For production, consider using CDN for static assets:
- AWS CloudFront
- Cloudflare
- Azure CDN

## Summary

This example demonstrates how to:
- ✅ Mount multiple SPAs on different paths
- ✅ Serve static assets efficiently
- ✅ Configure handler mount order
- ✅ Build a complete runnable application

All configured **declaratively in YAML** without writing handler code! 🎉

## Next Steps

- See full [README](README_FULL.md) for reverse proxy examples
- See [01-basic-config](../01-basic-config/) for service basics
- See [02-multi-file](../02-multi-file/) for environment configs

## What's Demonstrated

- ✅ Reverse proxy configuration
- ✅ Path stripping and rewriting
- ✅ SPA mounting (Single Page Applications)
- ✅ Static file serving
- ✅ Multiple apps with different handlers
- ✅ CDN-like static server

## Configuration Highlights

### 1. Reverse Proxy

**Simple proxy with prefix stripping:**
```yaml
reverse-proxies:
  - prefix: "/api/v2"
    strip-prefix: true
    target: "http://backend:9000"
```

**Request flow:**
```
Client: GET /api/v2/users
  ↓ (strip-prefix: true)
Backend: GET /users
```

**With path rewriting:**
```yaml
reverse-proxies:
  - prefix: "/graphql"
    target: "http://graphql-server:4000"
    rewrite:
      from: "^/graphql"
      to: "/api/graphql"
```

**Request flow:**
```
Client: POST /graphql
  ↓ (rewrite)
Backend: POST /api/graphql
```

### 2. SPA Mounting

```yaml
mount-spa:
  - prefix: "/admin"
    dir: "./dist/admin-spa"
```

**Behavior:**
- `/admin` → serves `dist/admin-spa/index.html`
- `/admin/users` → serves `dist/admin-spa/index.html`
- `/admin/logo.png` → serves `dist/admin-spa/logo.png`
- Routes without extensions → fallback to `index.html`

**Use cases:**
- React apps
- Vue apps
- Angular apps
- Any client-side routed SPA

### 3. Static File Serving

```yaml
mount-static:
  - prefix: "/assets"
    dir: "./public/assets"
```

**Behavior:**
- `/assets/logo.png` → serves `public/assets/logo.png`
- `/assets/css` → serves `public/assets/css/index.html`
- `/assets/css/` → serves `public/assets/css/index.html`
- Paths without extensions → append `/index.html`

**Use cases:**
- Static assets (images, CSS, JS)
- Download files
- Documentation sites
- CDN-like serving

## Architecture Patterns

### API Gateway Pattern
```yaml
apps:
  - addr: ":8080"
    reverse-proxies:
      - prefix: "/api/users"
        target: "http://user-service:9001"
      - prefix: "/api/orders"
        target: "http://order-service:9002"
      - prefix: "/api/payments"
        target: "http://payment-service:9003"
```

### Backend-for-Frontend (BFF)
```yaml
apps:
  - addr: ":8080"
    published-services: [bff-service]
    reverse-proxies:
      - prefix: "/internal/users"
        target: "http://user-service:9001"
      - prefix: "/internal/orders"
        target: "http://order-service:9002"
    mount-spa:
      - prefix: "/"
        dir: "./dist/web-app"
```

### Monolith + SPA
```yaml
apps:
  - addr: ":8080"
    published-services: [api-service]
    mount-spa:
      - prefix: "/"
        dir: "./dist/spa"
```

### CDN Server
```yaml
apps:
  - addr: ":8081"
    mount-static:
      - prefix: "/images"
        dir: "./cdn/images"
      - prefix: "/videos"
        dir: "./cdn/videos"
```

## Request Routing Order

Handlers are mounted in this order:

1. **Reverse Proxies** (prepended first)
2. **Business Routers** (from `routers` and `published-services`)
3. **SPA Mounts** (added after business routers)
4. **Static Mounts** (added after SPA mounts)

**Example:**
```yaml
apps:
  - addr: ":8080"
    reverse-proxies: [...]      # 1st priority
    routers: [user-router]       # 2nd priority
    published-services: [...]    # 2nd priority (auto-generated routers)
    mount-spa: [...]             # 3rd priority
    mount-static: [...]          # 4th priority
```

## Run

```bash
go run main.go
```

## Test Reverse Proxy

```bash
# Test API v2 proxy (strips /api/v2 prefix)
curl http://localhost:8080/api/v2/users

# Test GraphQL proxy (rewrites path)
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ users { id name } }"}'
```

## Test SPA Mount

```bash
# All these serve index.html
curl http://localhost:8080/admin
curl http://localhost:8080/admin/users
curl http://localhost:8080/admin/settings

# Static files served directly
curl http://localhost:8080/admin/logo.png
curl http://localhost:8080/admin/app.js
```

## Test Static Files

```bash
# Serve file directly
curl http://localhost:8080/assets/logo.png

# Directory → index.html
curl http://localhost:8080/assets/css/
```

## Production Considerations

### 1. Reverse Proxy

**Pros:**
- ✅ Simple routing to microservices
- ✅ Path rewriting for API versioning
- ✅ No code changes needed

**Cons:**
- ⚠️ No circuit breakers (add middleware)
- ⚠️ No retry logic (add middleware)
- ⚠️ No load balancing across multiple targets

**Best for:**
- Internal service-to-service routing
- API gateway patterns
- Legacy system integration

### 2. SPA Mounting

**Pros:**
- ✅ Client-side routing works out of the box
- ✅ No separate web server needed
- ✅ Single deployment artifact

**Cons:**
- ⚠️ No cache headers by default
- ⚠️ No gzip compression by default

**Best practices:**
```yaml
apps:
  - addr: ":8080"
    # Add compression middleware
    routers: [compression-router]
    mount-spa:
      - prefix: "/"
        dir: "./dist/spa"
```

### 3. Static Files

**Pros:**
- ✅ Simple file serving
- ✅ Index.html fallback

**Cons:**
- ⚠️ Not optimized for large files
- ⚠️ No range request support

**Best for:**
- Small to medium static assets
- Documentation sites
- Download files

**For production CDN:**
Consider using dedicated services like Cloudflare, AWS CloudFront, or nginx.

## Summary

This example demonstrates how to configure various handlers at the app level without writing any Go code. All routing, proxying, and file serving is configured declaratively in YAML! 🎉
