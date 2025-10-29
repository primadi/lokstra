# CORS Middleware Example

Demonstrates Cross-Origin Resource Sharing (CORS) configuration in Lokstra.

## What You'll Learn

- Configure CORS middleware
- Understand CORS headers
- Handle preflight OPTIONS requests
- Set allowed origins, methods, and headers
- Enable credentials support

## Running

```bash
cd docs/01-essentials/03-middleware/examples/03-cors-middleware
go run main.go
```

## Testing

### With curl (simulate browser):

**Allowed origin:**
```bash
curl -H "Origin: http://localhost:3001" http://localhost:3000/users
```

Expected response headers:
```
Access-Control-Allow-Origin: http://localhost:3001
Access-Control-Allow-Credentials: true
```

**Disallowed origin:**
```bash
curl -H "Origin: http://evil.com" http://localhost:3000/users
```

Expected: NO CORS headers in response

**Preflight (OPTIONS) request:**
```bash
curl -X OPTIONS \
  -H "Origin: http://localhost:3001" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:3000/users
```

Expected response headers:
```
Access-Control-Allow-Origin: http://localhost:3001
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key
Access-Control-Max-Age: 3600
```

### With browser:

1. Open http://localhost:3001 in browser
2. Open browser console
3. Run:
```javascript
fetch('http://localhost:3000/users')
  .then(r => r.json())
  .then(console.log)
  .catch(console.error)
```

Should work because localhost:3001 is in allowed origins!

## Configuration

```go
corsConfig := map[string]any{
    "allowed_origins": []string{
        "http://localhost:3001",
        "http://localhost:8080",
        "https://myapp.com",
    },
    "allowed_methods": []string{
        "GET", "POST", "PUT", "DELETE", "OPTIONS",
    },
    "allowed_headers": []string{
        "Content-Type",
        "Authorization",
        "X-API-Key",
    },
    "allow_credentials": true,
    "max_age": 3600, // Cache preflight for 1 hour
}

router.Use(cors.Middleware(corsConfig))
```

## CORS Headers Explained

### Response Headers:

| Header | Purpose | Example |
|--------|---------|---------|
| `Access-Control-Allow-Origin` | Which origin is allowed | `http://localhost:3001` |
| `Access-Control-Allow-Methods` | Which HTTP methods allowed | `GET, POST, PUT, DELETE` |
| `Access-Control-Allow-Headers` | Which request headers allowed | `Content-Type, Authorization` |
| `Access-Control-Allow-Credentials` | Can send cookies/auth | `true` |
| `Access-Control-Max-Age` | Cache preflight (seconds) | `3600` |

### Request Headers (from browser):

| Header | Purpose | Example |
|--------|---------|---------|
| `Origin` | Where request came from | `http://localhost:3001` |
| `Access-Control-Request-Method` | Preflight: what method will be used | `POST` |
| `Access-Control-Request-Headers` | Preflight: what headers will be sent | `Content-Type` |

## What is Preflight?

Before making certain requests (POST, PUT, DELETE, custom headers), browsers send an **OPTIONS** request first to check if it's allowed.

**Simple request (no preflight):**
- GET
- POST with simple Content-Type (form-urlencoded, multipart/form-data)
- No custom headers

**Complex request (requires preflight):**
- POST/PUT/DELETE with JSON
- Custom headers (Authorization, X-API-Key)
- Any method other than GET/POST

## Common Issues

### Issue 1: CORS Error in Browser
```
Access to fetch at 'http://localhost:3000/users' from origin 
'http://localhost:3001' has been blocked by CORS policy
```

**Solution**: Add origin to `allowed_origins`:
```go
"allowed_origins": []string{
    "http://localhost:3001",  // Add this
}
```

### Issue 2: Credentials Not Working
```
Credentials flag is true, but the 'Access-Control-Allow-Credentials' 
header is ''
```

**Solution**: Enable credentials:
```go
"allow_credentials": true,
```

### Issue 3: Custom Header Blocked
```
Request header X-API-Key is not allowed by 
Access-Control-Allow-Headers in preflight response
```

**Solution**: Add header to allowed list:
```go
"allowed_headers": []string{
    "Content-Type",
    "X-API-Key",  // Add this
}
```

## Production Tips

### For development:
```go
corsConfig := map[string]any{
    "allowed_origins": []string{"*"},  // Allow all
    "allowed_methods": []string{"*"},
    "allowed_headers": []string{"*"},
}
```

### For production:
```go
corsConfig := map[string]any{
    // Specific origins only!
    "allowed_origins": []string{
        "https://myapp.com",
        "https://www.myapp.com",
    },
    // Only needed methods
    "allowed_methods": []string{"GET", "POST", "PUT", "DELETE"},
    // Only needed headers
    "allowed_headers": []string{"Content-Type", "Authorization"},
    "allow_credentials": true,
    "max_age": 86400,  // 24 hours
}
```

## Key Takeaways

- ✅ CORS is a browser security feature, not a server security feature
- ✅ Use `allowed_origins` to whitelist specific domains
- ✅ Preflight requests (OPTIONS) happen automatically for complex requests
- ✅ `allow_credentials: true` needed for cookies/auth
- ✅ Use wildcard (`*`) only in development, never production!
- ✅ CORS doesn't protect against all attacks - still need server-side validation
