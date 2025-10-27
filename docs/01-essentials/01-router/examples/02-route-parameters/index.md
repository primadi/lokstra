# Example 02: Route Parameters

> **Learn path and query parameters**  
> **Time**: 7 minutes • **Concepts**: Path params, query params, combined params

---

## 🎯 What You'll Learn

- Extracting path parameters from URL (`/users/{id}`)
- Reading query parameters (`?category=electronics&max_price=100`)
- Combining path and query parameters
- Default values for query params
- Type conversion (string → int, float64)

---

## 🚀 Run It

```bash
cd docs/01-essentials/01-router/examples/02-route-parameters
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 🧪 Test It

### 1. Path Parameter - Get User by ID
```bash
curl http://localhost:3000/users/1
```

**Response**:
```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com"
}
```

---

### 2. Path Parameter - Update User
```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Smith","email":"alice.smith@example.com"}'
```

**Response**:
```json
{
  "id": 1,
  "name": "Alice Smith",
  "email": "alice.smith@example.com"
}
```

---

### 3. Query Parameters - Filter Products
```bash
# By category
curl "http://localhost:3000/products?category=electronics"

# By price range
curl "http://localhost:3000/products?min_price=100&max_price=500"

# Combined filters
curl "http://localhost:3000/products?category=furniture&max_price=250"
```

**Response**:
```json
[
  {"id": 3, "name": "Desk", "price": 299.99, "category": "furniture"},
  {"id": 4, "name": "Chair", "price": 199.99, "category": "furniture"}
]
```

---

### 4. Combined - Path + Query Parameters
```bash
# Get user's electronics
curl "http://localhost:3000/users/1/products?category=electronics"

# Get user's products above $100
curl "http://localhost:3000/users/1/products?min_price=100"
```

**Response**:
```json
{
  "user": {"id": 1, "name": "Alice", "email": "alice@example.com"},
  "products": [
    {"id": 1, "name": "Laptop", "price": 999.99, "category": "electronics"}
  ],
  "count": 1
}
```

---

## 📝 Key Concepts

### 1. Path Parameters
Extract dynamic values from URL path:

```go
type GetUserRequest struct {
    ID int `path:"id"`  // Extracts from /users/{id}
}

r.GET("/users/{id}", func(req *GetUserRequest) (*User, error) {
    // req.ID contains the ID from URL
})
```

**URL**: `/users/123` → `req.ID = 123`

---

### 2. Query Parameters
Extract from query string:

```go
type SearchRequest struct {
    Category string  `query:"category"`
    MinPrice float64 `query:"min_price"`
    MaxPrice float64 `query:"max_price"`
    Limit    int     `query:"limit" default:"10"`
}
```

**URL**: `/products?category=electronics&max_price=500` 
- `req.Category = "electronics"`
- `req.MaxPrice = 500.0`

---

### 3. Combined Parameters
Mix path, query, and body:

```go
type UpdateUserRequest struct {
    ID    int    `path:"id"`      // From URL path
    Name  string `json:"name"`    // From JSON body
    Email string `json:"email"`   // From JSON body
}

r.PUT("/users/{id}", updateUser)
```

**Request**:
- URL: `/users/1`
- Body: `{"name":"Alice Smith"}`
- Result: `req.ID=1, req.Name="Alice Smith"`

---

### 4. Default Values
Set defaults for optional query params:

```go
type SearchRequest struct {
    Limit int `query:"limit" default:"10"`
}
```

- URL: `/products` → `req.Limit = 10` (default)
- URL: `/products?limit=5` → `req.Limit = 5`

---

### 5. Automatic Type Conversion
Lokstra converts string parameters to correct types:

```go
type Request struct {
    ID       int     `path:"id"`       // "123" → 123
    Price    float64 `query:"price"`   // "99.99" → 99.99
    Active   bool    `query:"active"`  // "true" → true
    Page     int     `query:"page"`    // "2" → 2
}
```

---

## 🎓 What You Learned

- ✅ Path parameters with `path:"name"` tag
- ✅ Query parameters with `query:"name"` tag
- ✅ Default values with `default:"value"`
- ✅ Automatic type conversion (string → int/float64/bool)
- ✅ Combining multiple parameter sources
- ✅ RESTful patterns (GET, PUT, DELETE)

---

## 💡 Tips

### Multiple Path Parameters
```go
r.GET("/users/{userId}/posts/{postId}", handler)

type Request struct {
    UserID int `path:"userId"`
    PostID int `path:"postId"`
}
```

---

### Optional vs Required
```go
type Request struct {
    Required string `query:"q" validate:"required"`  // Must be present
    Optional string `query:"sort"`                    // Can be empty
}
```

---

### Common Query Patterns
```go
// Pagination
Page  int `query:"page" default:"1"`
Limit int `query:"limit" default:"20"`

// Sorting
Sort  string `query:"sort" default:"created_at"`
Order string `query:"order" default:"desc"`

// Filtering
Search   string `query:"q"`
Category string `query:"category"`
Status   string `query:"status"`
```

---

**Next**: [03 - Route Groups](../03-route-groups/) - Learn API versioning with groups
