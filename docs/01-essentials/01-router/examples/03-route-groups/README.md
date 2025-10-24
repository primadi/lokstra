# Example 03: Route Groups

> **Learn API versioning and nested groups**  
> **Time**: 7 minutes • **Concepts**: Route groups, API versioning, nested groups

---

## 🎯 What You'll Learn

- Creating route groups with `AddGroup()`
- API versioning (v1, v2)
- Different response formats per version
- Nested groups (`/admin/users`)
- Organizing routes logically

---

## 🚀 Run It

```bash
cd docs/01-essentials/01-router/examples/03-route-groups
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 🧪 Test It

### 1. API Version 1 - Simple Responses
```bash
curl http://localhost:3000/v1/users
```

**Response** (simple list):
```json
[
  {"id": 1, "name": "Alice", "email": "alice@example.com", "version": "v1"},
  {"id": 2, "name": "Bob", "email": "bob@example.com", "version": "v1"}
]
```

```bash
curl http://localhost:3000/v1/users/1
```

**Response**:
```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com",
  "version": "v1"
}
```

---

### 2. API Version 2 - Enhanced with Metadata
```bash
curl http://localhost:3000/v2/users
```

**Response** (with metadata):
```json
{
  "data": [
    {"id": 1, "name": "Alice", "email": "alice@example.com", "version": "v2"},
    {"id": 2, "name": "Bob", "email": "bob@example.com", "version": "v2"}
  ],
  "meta": {
    "count": 2,
    "version": "v2"
  }
}
```

```bash
curl http://localhost:3000/v2/users/1
```

**Response**:
```json
{
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "version": "v2"
  },
  "meta": {
    "version": "v2",
    "retrieved": "2025-10-22T10:30:00Z"
  }
}
```

---

### 3. Admin Routes - Nested Group
```bash
curl http://localhost:3000/admin/stats
```

**Response**:
```json
{
  "total_users": 2,
  "next_id": 3
}
```

```bash
curl http://localhost:3000/admin/users
```

**Response**:
```json
[
  {"id": 1, "name": "Alice", "email": "alice@example.com"},
  {"id": 2, "name": "Bob", "email": "bob@example.com"}
]
```

```bash
# Create user via admin
curl -X POST http://localhost:3000/admin/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

---

## 📝 Key Concepts

### 1. Creating Groups
```go
r := lokstra.NewRouter("api")

// Create a group with prefix
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsersV1)  // Becomes: GET /v1/users
```

---

### 2. API Versioning Pattern
```go
// Version 1 - Simple
v1 := r.AddGroup("/v1")
v1.GET("/users", func() ([]User, error) {
    return users, nil  // Simple list
})

// Version 2 - Enhanced
v2 := r.AddGroup("/v2")
v2.GET("/users", func() (map[string]any, error) {
    return map[string]any{
        "data": users,
        "meta": map[string]int{"count": len(users)},
    }, nil
})
```

**Benefits**:
- Backward compatibility (v1 still works!)
- Gradual migration
- Different implementations per version

---

### 3. Nested Groups
```go
// Parent group
admin := r.AddGroup("/admin")
admin.GET("/stats", getStats)  // GET /admin/stats

// Nested group
adminUsers := admin.AddGroup("/users")
adminUsers.GET("", listUsers)          // GET /admin/users
adminUsers.POST("", createUser)        // POST /admin/users
adminUsers.DELETE("/{id}", deleteUser) // DELETE /admin/users/{id}
```

---

### 4. Route Organization
```go
r := lokstra.NewRouter("api")

// Public routes
r.GET("/", home)
r.GET("/health", health)

// API v1
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsersV1)
v1.GET("/products", getProductsV1)

// API v2
v2 := r.AddGroup("/v2")
v2.GET("/users", getUsersV2)
v2.GET("/products", getProductsV2)

// Admin
admin := r.AddGroup("/admin")
admin.GET("/stats", getStats)
```

**Result** - Clean, organized structure:
```
GET  /
GET  /health
GET  /v1/users
GET  /v1/products
GET  /v2/users
GET  /v2/products
GET  /admin/stats
```

---

## 🎓 What You Learned

- ✅ Creating route groups with `AddGroup()`
- ✅ API versioning strategy
- ✅ Different response formats per version
- ✅ Nested groups for hierarchical routes
- ✅ Organizing routes logically
- ✅ Using `PrintRoutes()` for debugging

---

## 💡 Best Practices

### 1. Version from Day One
```go
// ✅ Good - versioned from start
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsers)

// 🚫 Bad - no version, hard to change later
r.GET("/users", getUsers)
```

---

### 2. Keep Versions Stable
```go
// ✅ Good - v1 stays unchanged
v1.GET("/users", getUsersV1Simple)

// Add v2 with breaking changes
v2.GET("/users", getUsersV2Enhanced)

// 🚫 Bad - changing v1 breaks clients!
v1.GET("/users", getUsersNew) // Breaking change!
```

---

### 3. Use Nested Groups for Resources
```go
// ✅ Good - clear hierarchy
api := r.AddGroup("/api")
users := api.AddGroup("/users")
users.GET("", listUsers)
users.POST("", createUser)
users.GET("/{id}", getUser)

// 🚫 Acceptable but less organized
r.GET("/api/users", listUsers)
r.POST("/api/users", createUser)
r.GET("/api/users/{id}", getUser)
```

---

## 🔍 Debugging Routes

Use `PrintRoutes()` to see all registered routes:

```go
r.PrintRoutes()
```

**Output**:
```
[api] GET / -> home
[api] GET /v1/users -> getUsersV1
[api] GET /v1/users/{id} -> getUserV1
[api] GET /v2/users -> getUsersV2
[api] GET /v2/users/{id} -> getUserV2
[api] GET /admin/stats -> getStats
[api] GET /admin/users -> listUsers
[api] POST /admin/users -> createUser
[api] DELETE /admin/users/{id} -> deleteUser
```

---

## 📋 Common Patterns

### Pattern 1: Multi-Version API
```go
v1 := r.AddGroup("/v1")
v2 := r.AddGroup("/v2")
v3 := r.AddGroup("/v3")
```

---

### Pattern 2: Public vs Protected
```go
public := r.AddGroup("/public")
public.GET("/products", getProducts)

protected := r.AddGroup("/api")
protected.Use(authMiddleware)  // All routes require auth
protected.GET("/users", getUsers)
```

---

### Pattern 3: Resource Grouping
```go
api := r.AddGroup("/api/v1")

users := api.AddGroup("/users")
users.GET("", listUsers)
users.POST("", createUser)
users.GET("/{id}", getUser)
users.PUT("/{id}", updateUser)
users.DELETE("/{id}", deleteUser)

products := api.AddGroup("/products")
products.GET("", listProducts)
// ...
```

---

**Next**: [04 - Handler Forms](../04-handler-forms/) - Explore different handler patterns
