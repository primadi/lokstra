# Route Priorities Deep Dive

> **Understand route matching order and prevent conflicts**

This example demonstrates how Lokstra prioritizes routes when multiple patterns could match the same URL.

## Route Priority Rules

Lokstra follows these priority rules (highest to lowest):

1. **Exact matches** - Routes without parameters
2. **Parameterized routes** - Routes with `:param`
3. **Wildcard routes** - Routes with `*wildcard`

### Priority Order

```
/users/active          (1. Exact - highest priority)
/users/:id             (2. Parameter - medium priority)
/users/*path           (3. Wildcard - lowest priority)
```

---

## Examples

### 1. Exact vs Parameter

```go
// Register order doesn't matter - exact always wins
r.Get("/users/me", GetCurrentUser)      // Matches: /users/me
r.Get("/users/:id", GetUserByID)        // Matches: /users/123
```

**Test**:
- `GET /users/me` → Calls `GetCurrentUser` ✅
- `GET /users/123` → Calls `GetUserByID` ✅

---

### 2. Parameter vs Wildcard

```go
r.Get("/files/:id", GetFileByID)        // Matches: /files/123
r.Get("/files/*path", GetFilePath)      // Matches: /files/docs/readme.txt
```

**Test**:
- `GET /files/123` → Calls `GetFileByID` ✅
- `GET /files/docs/readme.txt` → Calls `GetFilePath` ✅

---

### 3. Multiple Exact Routes

```go
r.Get("/users/active", GetActiveUsers)
r.Get("/users/inactive", GetInactiveUsers)
r.Get("/users/pending", GetPendingUsers)
r.Get("/users/:id", GetUserByID)
```

All exact routes have equal priority. Parameter route only matches when no exact route matches.

---

### 4. Nested Parameters

```go
r.Get("/users/:userId/posts/:postId", GetUserPost)
r.Get("/users/:userId/posts/latest", GetLatestPost)
```

**Priority**:
- `/users/1/posts/latest` → Exact segment wins → `GetLatestPost` ✅
- `/users/1/posts/123` → Falls through to → `GetUserPost` ✅

---

### 5. Wildcard Captures Everything

```go
r.Get("/api/v1/*path", HandleAPIv1)
r.Get("/api/v2/*path", HandleAPIv2)
```

Wildcard captures all remaining path segments:
- `/api/v1/users/123` → `path = "users/123"`
- `/api/v2/posts/456/comments` → `path = "posts/456/comments"`

---

## Common Pitfalls

### ❌ Ambiguous Routes

```go
// DON'T: Both use parameters at same position
r.Get("/posts/:id", GetPost)
r.Get("/posts/:slug", GetPostBySlug)  // ❌ Conflict!
```

**Problem**: Router can't distinguish between ID and slug  
**Solution**: Use different paths or exact routes

```go
// ✅ FIX 1: Different paths
r.Get("/posts/id/:id", GetPost)
r.Get("/posts/slug/:slug", GetPostBySlug)

// ✅ FIX 2: Exact routes for known values
r.Get("/posts/latest", GetLatestPost)
r.Get("/posts/:id", GetPost)
```

---

### ❌ Wildcard Blocking Routes

```go
// DON'T: Register wildcard before specific routes
r.Get("/files/*path", ServeFiles)
r.Get("/files/upload", HandleUpload)  // ❌ Never reached!
```

**Problem**: Wildcard matches everything, including `/files/upload`  
**Solution**: Register specific routes first (but Lokstra auto-prioritizes exact routes anyway)

```go
// ✅ This works correctly (exact route has priority)
r.Get("/files/upload", HandleUpload)  // Exact - always matches first
r.Get("/files/*path", ServeFiles)     // Wildcard - matches rest
```

---

## Best Practices

### ✅ Do

```go
// Use exact routes for known endpoints
r.Get("/users/me", GetCurrentUser)
r.Get("/users/active", GetActiveUsers)
r.Get("/users/search", SearchUsers)
r.Get("/users/:id", GetUserByID)

// Use descriptive parameter names
r.Get("/posts/:postId/comments/:commentId", GetComment)

// Use wildcards for file serving
r.Get("/static/*filepath", ServeStaticFiles)
```

### ❌ Don't

```go
// Don't create ambiguous patterns
r.Get("/items/:id", GetItemByID)
r.Get("/items/:name", GetItemByName)  // ❌ Ambiguous

// Don't rely on registration order
// (exact routes always have priority regardless of order)

// Don't use wildcard for structured APIs
r.Get("/api/*endpoint", HandleAPI)  // ❌ Too broad
```

---

## Route Testing Checklist

When adding new routes, test:

1. ✅ Exact match works
2. ✅ Parameter extraction works
3. ✅ No conflicts with existing routes
4. ✅ Wildcard doesn't block specific routes
5. ✅ 404 for invalid paths

---

## Running

```bash
go run main.go

# Test with test.http file
```

---

## Key Takeaways

✅ **Priority**: Exact > Parameter > Wildcard  
✅ **Registration order doesn't affect priority**  
✅ **Exact routes never conflict** (different paths)  
✅ **Avoid ambiguous parameter routes** at same position  
✅ **Wildcards match everything** after the prefix  
✅ **Test route matching** before deploying

---

## Related Examples

- [01-all-handler-forms](../01-all-handler-forms/) - Handler patterns
- [02-parameter-binding](../02-parameter-binding/) - Parameter extraction
- [05-error-handling](../05-error-handling/) - 404 handling
