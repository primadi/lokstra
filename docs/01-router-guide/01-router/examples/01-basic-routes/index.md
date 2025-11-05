# Example 01: Basic Routes

> **Learn basic routing with GET and POST**  
> **Time**: 5 minutes • **Concepts**: HTTP methods, handler forms, auto JSON

---

## 🎯 What You'll Learn

- Creating a router
- GET and POST routes
- Simple return values (auto JSON conversion)
- Request binding from JSON body
- Basic validation

---

## 🚀 Run It

```bash
cd docs/01-router-guide/01-router/examples/01-basic-routes
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 🧪 Test It

### 1. Simple Ping
```bash
curl http://localhost:3000/ping
```

**Response**:
```json
"pong"
```

---

### 2. List Users
```bash
curl http://localhost:3000/users
```

**Response**:
```json
[
  {"id": 1, "name": "Alice", "email": "alice@example.com"},
  {"id": 2, "name": "Bob", "email": "bob@example.com"}
]
```

---

### 3. Create User
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

**Response**:
```json
{
  "id": 3,
  "name": "Charlie",
  "email": "charlie@example.com"
}
```

---

## 📝 Key Concepts

### 1. Simple Handler Form
```go
r.GET("/ping", func() string {
    return "pong"
})
```

- No parameters needed
- Return value auto-converted to JSON
- Perfect for simple endpoints

---

### 2. Return with Error
```go
r.GET("/users", func() ([]User, error) {
    return users, nil
})
```

- Most common form (90% of cases)
- Lokstra handles errors automatically
- Return type becomes JSON response

---

### 3. Request Binding
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    // req is auto-bound from JSON body
    // auto-validated
})
```

- Struct tags define binding source (`json:`)
- Validation with `validate:` tags
- Automatic error responses for invalid data

---

## 🎓 What You Learned

- ✅ Creating routers with `lokstra.NewRouter()`
- ✅ GET routes with simple returns
- ✅ POST routes with request binding
- ✅ Automatic JSON conversion
- ✅ Basic validation with struct tags

---

**Next**: [02 - Route Parameters](../02-route-parameters/) - Learn path and query parameters
