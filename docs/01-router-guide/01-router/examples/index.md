# Router Examples

> **Hands-on examples to learn Lokstra routing**  
> **Total time**: ~45 minutes for all examples

---

## 📚 Examples Overview

Work through these examples **in order**. Each builds on the previous:

| Example | Topic | Time | Key Concepts |
|---------|-------|------|--------------|
| [01](01-basic-routes/) | **Basic Routes** | 5 min | GET, POST, auto JSON, validation |
| [02](02-route-parameters/) | **Route Parameters** | 7 min | Path params, query params, combined |
| [03](03-route-groups/) | **Route Groups** | 7 min | API versioning, nested groups |
| [04](04-handler-forms/) | **Handler Forms** | 10 min | 5 essential handler patterns |
| [05](05-response-patterns/) | **Response Patterns** ⭐ | 15 min | 3 response types, 2 response paths |

---

## 🚀 Quick Start

Each example is **self-contained** and **runnable**:

```bash
# Navigate to any example
cd 01-basic-routes

# Run it
go run main.go

# In another terminal, test it
curl http://localhost:3000/ping
```

---

## 📝 What You'll Learn

### Example 01: Basic Routes
**Concepts**: Creating routers, GET/POST routes, auto JSON conversion

```go
r.GET("/ping", func() string {
    return "pong"  // Auto-converted to JSON
})

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    // Request auto-bound and validated
})
```

**Learning outcomes**:
- Router creation
- HTTP methods
- Simple return values
- Request binding
- Validation tags

---

### Example 02: Route Parameters
**Concepts**: Path parameters, query parameters, type conversion

```go
// Path parameter
r.GET("/users/{id}", func(req *GetUserRequest) (*User, error) {
    // req.ID extracted from URL
})

// Query parameters
type SearchRequest struct {
    Category string  `query:"category"`
    MinPrice float64 `query:"min_price"`
    MaxPrice float64 `query:"max_price"`
}
```

**Learning outcomes**:
- Path params with `path:"id"`
- Query params with `query:"name"`
- Default values
- Automatic type conversion
- Combining parameters

---

### Example 03: Route Groups
**Concepts**: API versioning, route organization, nested groups

```go
// API v1
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsersV1)  // Simple list

// API v2
v2 := r.AddGroup("/v2")
v2.GET("/users", getUsersV2)  // Enhanced with metadata

// Nested groups
admin := r.AddGroup("/admin")
adminUsers := admin.AddGroup("/users")
adminUsers.POST("", createUser)  // POST /admin/users
```

**Learning outcomes**:
- Creating groups
- API versioning pattern
- Nested groups
- Route organization
- Using `PrintRoutes()` for debugging

---

### Example 04: Handler Forms
**Concepts**: Different handler signatures, when to use each

```go
// Form 1: Simple
func() string { return "pong" }

// Form 2: With error (most common)
func() ([]User, error) { return users, nil }

// Form 3: Request binding
func(req *CreateUserRequest) (*User, error) { ... }

// Form 4: Full control
func(ctx *request.Context, req *GetUserRequest) (*User, error) { ... }

// Form 5: Custom response
func(ctx *request.Context, req *GetUserRequest) (*response.Response, error) { ... }
```

**Learning outcomes**:
- 5 essential handler forms
- When to use each form
- Context access
- Custom responses
- Decision guide

---

### Example 05: Response Patterns ⭐
**Concepts**: 3 response types, 2 response paths, when to use each

**3 Response Types**:

```go
// 1. Manual (http.ResponseWriter) - Full control
func(ctx *request.Context) error {
    ctx.W.Write([]byte(`{"message":"hello"}`))
    return nil
}

// 2. Generic (response.Response) - JSON, HTML, text, etc
func() (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(data)  // or .Html(), .Text()
    return resp, nil
}

// 3. Opinionated (response.ApiHelper) - Structured JSON API
func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    api.Ok(data)  // Standard format
    return api, nil
}
```

**2 Response Paths**:

```go
// Path 1: Via Context
func(ctx *request.Context) error {
    ctx.Resp.Json(data)  // or ctx.Api.Ok(data)
    return nil
}

// Path 2: Via Return
func() (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(data)
    return resp, nil
}
```

**Learning outcomes**:
- Manual vs Generic vs Opinionated responses
- When to use each response type
- Context path vs Return path
- ApiHelper standard JSON format
- Decision guide for choosing approach
- Best practices for REST APIs

---

## 🧪 Testing Examples

Each example includes a `test.http` file for easy testing with **REST Client** extension in VS Code:

1. Install [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension
2. Open `test.http` file
3. Click "Send Request" above each request

Or use `curl`:

```bash
# Example 01
curl http://localhost:3000/ping
curl http://localhost:3000/users
curl -X POST http://localhost:3000/users -H "Content-Type: application/json" -d '{"name":"Charlie","email":"charlie@example.com"}'

# Example 02
curl http://localhost:3000/users/1
curl "http://localhost:3000/products?category=electronics"

# Example 03
curl http://localhost:3000/v1/users
curl http://localhost:3000/v2/users
curl http://localhost:3000/admin/stats

# Example 04
curl http://localhost:3000/ping
curl -i http://localhost:3000/users/1/custom
```

---

## 📊 Learning Progression

```
Example 01: Basic Routes
   ↓
   Learn: Router creation, GET/POST, auto JSON
   
Example 02: Route Parameters
   ↓
   Learn: Path/query params, type conversion
   
Example 03: Route Groups
   ↓
   Learn: API versioning, organization
   
Example 04: Handler Forms
   ↓
   Learn: Different handler patterns, when to use each
```

---

## 💡 Tips

### Running Examples

**Option 1**: Run each separately
```bash
cd 01-basic-routes && go run main.go
```

**Option 2**: Build and run
```bash
cd 01-basic-routes
go build -o server .
./server
```

### Understanding Code

1. **Read README first** - Understand what you'll learn
2. **Read main.go** - Study the code
3. **Run the server** - See it in action
4. **Test with test.http** - Try all endpoints
5. **Modify and experiment** - Change code, see what happens

### Common Patterns

Most real APIs use **Example 02 + 03 + 04** patterns:
- Path/query parameters (02)
- API versioning with groups (03)
- Form 2 or 3 handlers (04)

---

## 🎯 After Examples

You now know:
- ✅ Router basics
- ✅ All parameter types
- ✅ Route organization
- ✅ Handler patterns

**Next steps**:
1. **Combine patterns** - Build a real API using all concepts
2. **Add middleware** - Learn [03-middleware](../../03-middleware/)
3. **Add services** - Learn [02-service](../../02-service/) (coming soon)

---

## 🔗 Additional Resources

- **Main Router Guide**: [../index](../index)
- **All 29 Handler Forms**: [Deep Dive](../../../02-deep-dive/router/handler-forms) (coming soon)
- **Router Lifecycle**: [Deep Dive](../../../02-deep-dive/router/lifecycle) (coming soon)

---

**Happy learning!** 🚀

If you have questions or find issues, please [open an issue](https://github.com/primadi/lokstra/issues).
