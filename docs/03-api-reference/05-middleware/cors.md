# CORS Middleware

> Cross-Origin Resource Sharing (CORS) handling

## Overview

CORS middleware handles Cross-Origin Resource Sharing requests, allowing web applications from different origins to access your API. It automatically handles preflight requests and sets appropriate CORS headers.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/cors"
```

---

## Usage

### Basic Usage

```go
// Allow all origins
router.Use(cors.Middleware([]string{"*"}))
```

---

### Specific Origins

```go
// Allow specific origins only
router.Use(cors.Middleware([]string{
    "https://app.example.com",
    "https://admin.example.com",
}))
```

---

### Multiple Origins

```go
allowedOrigins := []string{
    "https://app.example.com",
    "https://mobile.example.com",
    "https://admin.example.com",
    "http://localhost:3000", // Development
}

router.Use(cors.Middleware(allowedOrigins))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: cors
    params:
      allow_origins: ["*"]

  # Or specific origins
  - type: cors
    params:
      allow_origins:
        - "https://app.example.com"
        - "https://admin.example.com"
```

---

## Features

### Automatic Preflight Handling

CORS middleware automatically handles `OPTIONS` requests (preflight):

**Client Request:**
```http
OPTIONS /api/users HTTP/1.1
Origin: https://app.example.com
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Content-Type, Authorization
```

**Server Response:**
```http
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

### Credentials Support

CORS middleware always sets `Access-Control-Allow-Credentials: true`, allowing cookies and authorization headers:

```javascript
// Client-side JavaScript
fetch('https://api.example.com/users', {
    credentials: 'include', // Include cookies
    headers: {
        'Authorization': 'Bearer token123'
    }
})
```

---

### Allowed Methods

Automatically allows these HTTP methods:
- GET
- POST
- PUT
- DELETE
- OPTIONS

---

### Dynamic Headers

Echoes the `Access-Control-Request-Headers` from preflight requests, allowing any headers the client requests.

---

## Examples

### Allow All Origins (Development)

```go
// During development, allow all origins
router.Use(cors.Middleware([]string{"*"}))
```

**⚠️ Warning:** Only use `"*"` in development. In production, specify exact origins.

---

### Production Configuration

```go
// Production: whitelist specific origins
func configureCORS() []string {
    if os.Getenv("ENV") == "production" {
        return []string{
            "https://app.example.com",
            "https://admin.example.com",
        }
    }
    // Development: allow all
    return []string{"*"}
}

router.Use(cors.Middleware(configureCORS()))
```

---

### Multiple Environments

```go
var allowedOrigins []string

switch os.Getenv("ENV") {
case "production":
    allowedOrigins = []string{
        "https://app.example.com",
        "https://admin.example.com",
    }
case "staging":
    allowedOrigins = []string{
        "https://staging.example.com",
        "https://app.example.com",
    }
default: // development
    allowedOrigins = []string{"*"}
}

router.Use(cors.Middleware(allowedOrigins))
```

---

### Per-Router Configuration

```go
// Public API - allow all origins
publicRouter := lokstra.NewRouter()
publicRouter.Use(cors.Middleware([]string{"*"}))

// Internal API - restrict origins
internalRouter := lokstra.NewRouter()
internalRouter.Use(cors.Middleware([]string{
    "https://internal.example.com",
}))
```

---

### With Environment Variables

```go
originsStr := os.Getenv("ALLOWED_ORIGINS")
var allowedOrigins []string

if originsStr == "" {
    allowedOrigins = []string{"*"}
} else {
    allowedOrigins = strings.Split(originsStr, ",")
}

router.Use(cors.Middleware(allowedOrigins))
```

**Environment:**
```bash
ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
```

---

### Dynamic Origin Validation

```go
func getAllowedOrigins() []string {
    // Fetch from database or config service
    origins, err := configService.GetAllowedOrigins()
    if err != nil {
        log.Printf("Failed to get origins: %v", err)
        return []string{} // Deny all if config fails
    }
    return origins
}

router.Use(cors.Middleware(getAllowedOrigins()))
```

---

## Behavior

### Origin Header Present

When `Origin` header is present in request:

1. **Check if origin is allowed**
   - If `"*"` in config: allow any origin
   - Otherwise: check if origin is in whitelist

2. **If origin is forbidden:**
   - Return `403 Forbidden`
   - No CORS headers set

3. **If origin is allowed:**
   - Set `Access-Control-Allow-Origin: <origin>`
   - Set `Access-Control-Allow-Credentials: true`

4. **If OPTIONS request (preflight):**
   - Set `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
   - Echo `Access-Control-Allow-Headers` from request
   - Return `204 No Content`

---

### No Origin Header

If no `Origin` header is present (same-origin request), CORS headers are **not set** and request proceeds normally.

---

## Best Practices

### 1. Use Specific Origins in Production

```go
// ✅ Good - whitelist specific origins
router.Use(cors.Middleware([]string{
    "https://app.example.com",
    "https://admin.example.com",
}))

// 🚫 Bad - allows any origin in production
router.Use(cors.Middleware([]string{"*"}))
```

---

### 2. Include All Subdomains if Needed

```go
// ✅ Good - explicit subdomains
router.Use(cors.Middleware([]string{
    "https://app.example.com",
    "https://api.example.com",
    "https://admin.example.com",
    "https://mobile.example.com",
}))

// 🚫 Bad - wildcard subdomains not supported
router.Use(cors.Middleware([]string{
    "https://*.example.com", // Won't work
}))
```

---

### 3. Include Development Origins Conditionally

```go
// ✅ Good - development origins only in dev
allowedOrigins := []string{
    "https://app.example.com",
}

if os.Getenv("ENV") == "development" {
    allowedOrigins = append(allowedOrigins, 
        "http://localhost:3000",
        "http://localhost:8080",
    )
}

router.Use(cors.Middleware(allowedOrigins))
```

---

### 4. Place CORS Early in Middleware Chain

```go
// ✅ Good - CORS before authentication
router.Use(
    recovery.Middleware(&recovery.Config{}),
    cors.Middleware(allowedOrigins), // Early
    jwtauth.Middleware(&jwtauth.Config{}),
)

// 🚫 Bad - auth blocks preflight requests
router.Use(
    recovery.Middleware(&recovery.Config{}),
    jwtauth.Middleware(&jwtauth.Config{}),
    cors.Middleware(allowedOrigins), // Too late
)
```

---

### 5. Test Preflight Requests

```bash
# Test preflight
curl -X OPTIONS http://localhost:8080/api/users \
  -H "Origin: https://app.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization" \
  -v

# Should return:
# Access-Control-Allow-Origin: https://app.example.com
# Access-Control-Allow-Credentials: true
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Common Issues

### Issue: Preflight Fails with 401

**Problem:** Authentication middleware blocks OPTIONS requests

**Solution:** Place CORS before authentication:

```go
// ✅ Correct order
router.Use(
    cors.Middleware(allowedOrigins),     // First
    jwtauth.Middleware(&jwtauth.Config{
        SkipPaths: []string{"/auth/**"}, // Skip auth paths
    }),
)
```

---

### Issue: "No 'Access-Control-Allow-Origin' header"

**Problem:** Origin not in whitelist or CORS middleware not applied

**Solution:** Check configuration:

```go
// Verify origin is in list
allowedOrigins := []string{
    "https://app.example.com", // Must match exactly
}

// Check middleware is applied
router.Use(cors.Middleware(allowedOrigins))
```

---

### Issue: Credentials Not Allowed

**Problem:** Browser blocks credentials with wildcard origin

**Solution:** Don't use `"*"` when sending credentials:

```go
// 🚫 Bad - wildcard doesn't work with credentials
router.Use(cors.Middleware([]string{"*"}))

// ✅ Good - specific origin
router.Use(cors.Middleware([]string{
    "https://app.example.com",
}))
```

---

### Issue: Wrong Origin Format

**Problem:** Origin must include protocol

**Solution:** Always include `https://` or `http://`:

```go
// 🚫 Bad - missing protocol
router.Use(cors.Middleware([]string{
    "app.example.com",
}))

// ✅ Good - full origin
router.Use(cors.Middleware([]string{
    "https://app.example.com",
}))
```

---

## Performance

**Overhead:** ~500ns per request

**Impact:** Minimal - only performs header checks and string comparisons

---

## Client-Side Examples

### JavaScript Fetch

```javascript
// Simple request
fetch('https://api.example.com/users', {
    method: 'GET',
    credentials: 'include', // Include cookies
})

// Request with headers (triggers preflight)
fetch('https://api.example.com/users', {
    method: 'POST',
    credentials: 'include',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer token123'
    },
    body: JSON.stringify({name: 'John'})
})
```

---

### Axios

```javascript
// Configure globally
axios.defaults.withCredentials = true

// Make request
axios.post('https://api.example.com/users', {
    name: 'John'
}, {
    headers: {
        'Authorization': 'Bearer token123'
    }
})
```

---

### jQuery

```javascript
$.ajax({
    url: 'https://api.example.com/users',
    type: 'POST',
    xhrFields: {
        withCredentials: true
    },
    headers: {
        'Authorization': 'Bearer token123'
    },
    data: JSON.stringify({name: 'John'}),
    contentType: 'application/json'
})
```

---

## Testing

### Test CORS in Go

```go
func TestCORS(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(cors.Middleware([]string{
        "https://app.example.com",
    }))
    
    router.GET("/test", func(c *request.Context) error {
        return c.Api.Ok("success")
    })
    
    // Test with allowed origin
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Origin", "https://app.example.com")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
    assert.Equal(t, "https://app.example.com", 
        rec.Header().Get("Access-Control-Allow-Origin"))
    assert.Equal(t, "true", 
        rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSForbidden(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(cors.Middleware([]string{
        "https://app.example.com",
    }))
    
    // Test with forbidden origin
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Origin", "https://evil.com")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 403, rec.Code)
}
```

---

## See Also

- **[JWT Auth](./jwt-auth)** - Authentication middleware
- **[Recovery](./recovery)** - Panic recovery
- **[Request Logger](./request-logger)** - Request logging

---

## Related Guides

- **[Security Best Practices](../../04-guides/security/)** - Security patterns
- **[API Design](../../04-guides/api-design/)** - API design principles
- **[Frontend Integration](../../04-guides/frontend/)** - Frontend setup
