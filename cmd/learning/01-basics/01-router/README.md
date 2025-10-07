# 01-router: Basic Router

## What You'll Learn
- Create a Lokstra router
- Define HTTP routes (GET, POST, PUT, DELETE, PATCH, **ANY**)
- **Exact match vs Prefix match** routes
- Handle path parameters with **two syntax styles**
- Use **ANY method** to accept all HTTP methods
- Use the router with standard library's `http.ListenAndServe`

## Key Concepts

### Router
A **Router** is the core building block of Lokstra. It:
- Maps HTTP requests to handler functions
- Supports all standard HTTP methods
- Can handle path parameters (`:name` or `{name}`)
- Implements `http.Handler` interface
- Supports **exact match** and **prefix match** routing

### Handler Function
```go
func(c *lokstra.RequestContext) error
```
- Takes a `RequestContext` which contains request and response helpers
- Returns an error (nil if successful)
- Use `c.Api` methods for standard responses

### Path Parameter Syntax

Lokstra supports **two styles** for path parameters:

```go
// Style 1: Colon syntax
router.GET("/hello/:name", handler)

// Style 2: Curly braces syntax
router.GET("/user/{id}", handler)
```

Both styles work identically. Access parameters with:
```go
name := c.Req.PathParam("name", "default")
```

### Exact Match vs Prefix Match

#### Exact Match (GET, POST, PUT, DELETE, PATCH)
Routes match **only the exact path**:

```go
router.GET("/api/users", handler)
// ✅ Matches: /api/users
// ❌ Does NOT match: /api/users/123
```

#### Prefix Match (GETPrefix, POSTPrefix, PUTPrefix, DELETEPrefix, PATCHPrefix)
Routes match **the path and all sub-paths**:

```go
router.GETPrefix("/api/products", handler)
// ✅ Matches: /api/products/
// ✅ Matches: /api/products/123
// ✅ Matches: /api/products/search
// ✅ Matches: /api/products/category/electronics
```

**Use Cases:**
- **Exact**: Specific API endpoints with known paths
- **Prefix**: Catch-all handlers, file servers, reverse proxies, SPA serving

### ANY Method - Accepts All HTTP Methods

The `ANY` and `ANYPrefix` methods match **all HTTP methods** (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, etc.):

```go
// ANY: Exact match for all methods
router.ANY("/api/flexible", handler)
// ✅ Matches: GET /api/flexible
// ✅ Matches: POST /api/flexible
// ✅ Matches: PUT /api/flexible
// ❌ Does NOT match: GET /api/flexible/123

// ANYPrefix: Match all methods and sub-paths
router.ANYPrefix("/api/wildcard", handler)
// ✅ Matches: GET /api/wildcard/
// ✅ Matches: POST /api/wildcard/create
// ✅ Matches: PUT /api/wildcard/update/123
// ✅ Matches: DELETE /api/wildcard/remove/456
```

**Use Cases:**
- **Generic handlers**: Webhooks that accept any method
- **Debugging endpoints**: Log all requests regardless of method
- **Fallback handlers**: Catch-all for unmatched routes
- **Proxy endpoints**: Forward any request type

## Running the Example

```bash
cd cmd/learning/01-basics/01-router
go run main.go
```

The server will show all registered routes and example URLs.

## Testing

### Basic Routes
```bash
# Simple GET
curl http://localhost:8080/hello
```

### Path Parameters (Two Syntax Styles)
```bash
# Colon syntax (:name)
curl http://localhost:8080/hello/John

# Curly braces syntax ({id})
curl http://localhost:8080/user/123

# Multiple parameters
curl http://localhost:8080/user/10/post/20
```

### Exact Match vs Prefix Match
```bash
# Exact match - only matches /api/users
curl http://localhost:8080/api/users

# Exact match - fails on sub-path
curl http://localhost:8080/api/users/123  # 404

# Prefix match - matches all sub-paths
curl http://localhost:8080/api/products/
curl http://localhost:8080/api/products/123
curl http://localhost:8080/api/products/search
```

### Other HTTP Methods
```bash
# POST
curl -X POST http://localhost:8080/greet
curl -X POST http://localhost:8080/api/submit/order

# PUT
curl -X PUT http://localhost:8080/update/123
curl -X PUT http://localhost:8080/api/update/product/123

# DELETE
curl -X DELETE http://localhost:8080/delete/123
curl -X DELETE http://localhost:8080/api/remove/item/456

# PATCH
curl -X PATCH http://localhost:8080/patch/123
curl -X PATCH http://localhost:8080/api/patch/resource/789

# ANY - accepts all HTTP methods
curl http://localhost:8080/api/flexible              # GET
curl -X POST http://localhost:8080/api/flexible      # POST
curl -X PUT http://localhost:8080/api/flexible       # PUT
curl -X DELETE http://localhost:8080/api/flexible    # DELETE

# ANYPrefix - accepts all methods on all sub-paths
curl http://localhost:8080/api/wildcard/             # GET
curl -X POST http://localhost:8080/api/wildcard/create
curl -X PUT http://localhost:8080/api/wildcard/update/123
curl -X DELETE http://localhost:8080/api/wildcard/remove/456
```

Or use the provided `test.http` file with your HTTP client (VS Code REST Client, IntelliJ HTTP Client, etc.).

## Code Highlights

### Defining Routes
```go
// Exact match routes
router.GET("/api/users", handlerFunc)
router.POST("/api/users", handlerFunc)

// Prefix match routes
router.GETPrefix("/api/products", handlerFunc)
router.POSTPrefix("/api/submit", handlerFunc)

// ANY method - accepts all HTTP methods
router.ANY("/api/flexible", handlerFunc)
router.ANYPrefix("/api/wildcard", handlerFunc)

// Path parameters (two styles)
router.GET("/hello/:name", handlerFunc)      // Colon style
router.GET("/user/{id}", handlerFunc)        // Curly braces style
```

### Handler Implementation
```go
router.GET("/api/products", func(c *lokstra.RequestContext) error {
    // Access the matched path
    path := c.R.URL.Path
    
    // Return standardized response
    return c.Api.Ok(fmt.Sprintf("Matched: %s", path))
})
```

## What's Next?
- **02-app**: Learn how to combine multiple routers into an App
- **03-server**: Learn how to create servers with multiple apps
- **04-handlers**: Learn advanced handler patterns (smart bind, manual bind)
- **05-config**: Learn YAML configuration for scalable deployment
