# Router 404 Handling Issue

## Problem

When a request is made to a path that doesn't exist in the router, instead of returning a proper 404 Not Found error, the router appears to fallback to the root `/` endpoint.

**Example:**
```http
GET http://localhost:3004/users/999/orders
```

**Current Behavior:** Returns the root `/` response:
```json
{
  "status": "success",
  "data": {
    "server": "user-server",
    "endpoints": ["GET /users", "GET /users/{id}"]
  }
}
```

**Expected Behavior:** Should return 404 Not Found:
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "Route not found"
  }
}
```

## Current Router Setup

user-api router only has these routes:
- `GET /` (root info endpoint)
- `GET /users` (list users)
- `GET /users/{id}` (get user by ID)

The path `/users/999/orders` should NOT match any of these routes.

## Investigation Needed

1. **Check Router Engine Behavior**
   - Default engine is "chi" router
   - Verify chi router's 404 handling
   - Check if there's a wildcard catch-all route

2. **Check Path Matching Logic**
   - Is `/users/999/orders` somehow matching `/users/{id}`?
   - Is there a fallback mechanism to `/`?

3. **Verify RouterEngine Implementation**
   - Look at `core/router/engine/chi_router.go`
   - Check if NotFound handler is registered
   - Verify pattern matching is strict

## Possible Solutions

### Option 1: Add Custom 404 Handler
```go
router := lokstra.NewRouter("user-api")

// Register custom 404 handler
router.NotFound(func(ctx *request.Context) error {
    return ctx.Api.Error(404, "NOT_FOUND", "Route not found")
})
```

### Option 2: Fix Router Engine
Update `core/router/engine/chi_router.go` to ensure proper 404 handling:
```go
func (c *ChiRouter) Build() {
    // ... existing build logic
    
    // Set custom 404 handler
    c.mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
        // Return JSON 404 response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]any{
            "status": "error",
            "error": map[string]string{
                "code": "NOT_FOUND",
                "message": "Route not found",
            },
        })
    })
}
```

### Option 3: Strict Path Matching
Ensure path parameters don't match paths with additional segments:
- `/users/{id}` should NOT match `/users/999/orders`
- Pattern should be strict: exact segment count

## Priority

**Medium** - This affects API clarity and proper error handling, but doesn't break core functionality.

## Workaround

For now, document which endpoints exist on which servers:
- **user-service**: Only `/users` and `/users/{id}`
- **order-service**: Includes `/users/{user_id}/orders` and `/orders/{id}`
- **monolith**: All endpoints

Clients should know which server to call for which endpoint.

## Next Steps

1. Investigate chi router's NotFound behavior
2. Add tests for 404 scenarios
3. Implement proper 404 handler
4. Update all routers to use consistent error responses
