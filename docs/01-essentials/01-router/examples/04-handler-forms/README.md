# Example 04: Handler Forms

> **Explore 4 essential handler patterns**  
> **Time**: 10 minutes • **Concepts**: Handler signatures, when to use each form

---

## 🎯 What You'll Learn

- 4 essential handler forms (out of 29 total!)
- When to use each form
- Request binding patterns
- Context access
- Custom responses

---

## 🚀 Run It

```bash
cd docs/01-essentials/01-router/examples/04-handler-forms
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 🧪 Test It

### Form 1: Simple Return Value
```bash
curl http://localhost:3000/ping
```

**Response**:
```json
"pong"
```

```bash
curl http://localhost:3000/time
```

**Response**:
```json
{
  "current_time": "2025-10-22T10:30:00Z"
}
```

---

### Form 2: Return with Error (Most Common!)
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

### Form 3: Request Binding with Error
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

### Form 4: Context + Request with Error
```bash
curl http://localhost:3000/users/1/details -H "User-Agent: MyApp/1.0"
```

**Server logs**:
```
Request from: MyApp/1.0
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

### Form 5: Custom Response (Full Control)
```bash
curl -i http://localhost:3000/users/1/custom
```

**Response** (with custom headers):
```
HTTP/1.1 200 OK
Content-Type: application/json
X-User-ID: 1
X-Response-Time: 2025-10-22T10:30:00Z

{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com"
}
```

**Error response**:
```bash
curl -i http://localhost:3000/users/999/custom
```

```
HTTP/1.1 404 Not Found
X-Error-Code: USR404

{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User does not exist"
  }
}
```

---

## 📝 Handler Forms Explained

### Form 1: Simple Return Value
**Signature**: `func() T`

**When to use**:
- ✅ Simple data, no errors possible
- ✅ Static responses
- ✅ Health checks, ping endpoints

**Example**:
```go
r.GET("/ping", func() string {
    return "pong"
})

r.GET("/config", func() map[string]string {
    return map[string]string{
        "version": "1.0",
        "env": "production",
    }
})
```

---

### Form 2: Return with Error (90% of Cases!)
**Signature**: `func() (T, error)`

**When to use**:
- ✅ Database queries
- ✅ File operations
- ✅ External API calls
- ✅ Any operation that can fail

**Example**:
```go
r.GET("/users", func() ([]User, error) {
    users, err := db.GetUsers()
    if err != nil {
        return nil, err  // Lokstra handles error response
    }
    return users, nil  // Auto 200 OK with JSON
})
```

**Why most common?**
- Lokstra automatically converts errors to HTTP responses
- Clean code: just return the error
- No need to manually set status codes

---

### Form 3: Request Binding with Error
**Signature**: `func(req *RequestType) (T, error)`

**When to use**:
- ✅ POST/PUT endpoints (need body data)
- ✅ Path parameters
- ✅ Query parameters
- ✅ Need validation

**Example**:
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    // req is auto-bound from JSON body
    // auto-validated
    user, err := db.CreateUser(req.Name, req.Email)
    return user, err
})
```

**Binding sources**:
- `json:"name"` - From JSON body
- `path:"id"` - From URL path
- `query:"page"` - From query string
- `header:"Authorization"` - From headers

---

### Form 4: Context + Request with Error
**Signature**: `func(ctx *request.Context, req *RequestType) (T, error)`

**When to use**:
- ✅ Need to read headers
- ✅ Need to access raw request
- ✅ Logging with request metadata
- ✅ Custom authentication logic

**Example**:
```go
r.GET("/users/{id}", func(ctx *request.Context, req *GetUserRequest) (*User, error) {
    // Access headers
    authToken := ctx.R.Header.Get("Authorization")
    userAgent := ctx.R.Header.Get("User-Agent")
    
    // Log request
    log.Printf("User %d requested by %s", req.ID, userAgent)
    
    // Still have request binding
    user, err := db.GetUser(req.ID)
    return user, err
})
```

**What ctx provides**:
- `ctx.R` - Raw `*http.Request`
- `ctx.W` - Response writer
- `ctx.PathParams` - Path parameters map
- `ctx.Query()` - Query parameters

---

### Form 5: Custom Response (Full Control)
**Signature**: `func(...) (*response.Response, error)`

**When to use**:
- ✅ Custom status codes
- ✅ Custom headers (CORS, caching, etc)
- ✅ Multiple response types
- ✅ Fine-grained error responses

**Example**:
```go
r.GET("/users/{id}", func(ctx *request.Context, req *GetUserRequest) (*response.Response, error) {
    user, err := db.GetUser(req.ID)
    if err != nil {
        // Custom error response
        return response.Error(404, "USER_NOT_FOUND", "User does not exist").
            WithHeader("X-Error-ID", generateErrorID()), nil
    }
    
    // Custom success response
    return response.Success(user).
        WithHeader("X-User-ID", fmt.Sprintf("%d", user.ID)).
        WithHeader("Cache-Control", "max-age=3600").
        WithStatus(200), nil
})
```

---

## 📊 Comparison Table

| Form | Parameters | Return | Use Case | % Usage |
|------|------------|--------|----------|---------|
| **1** | None | `T` | Simple, static | 5% |
| **2** | None | `(T, error)` | Can fail | 60% |
| **3** | Request | `(T, error)` | Need input | 30% |
| **4** | Context + Request | `(T, error)` | Need headers | 4% |
| **5** | Context (optional) | `(*Response, error)` | Custom control | 1% |

**Recommendation**: 
- Start with Form 2 or 3 (90% of cases)
- Use Form 4 when you need headers
- Use Form 5 only when necessary

---

## 🎓 What You Learned

- ✅ Form 1: Simple return (no errors)
- ✅ Form 2: Return with error (most common!)
- ✅ Form 3: Request binding (POST/PUT)
- ✅ Form 4: Context access (headers)
- ✅ Form 5: Custom responses (full control)
- ✅ When to use each form
- ✅ Automatic JSON conversion
- ✅ Automatic error handling

---

## 💡 Decision Guide

```
Do you need request data (body/path/query)?
├─ NO  → Do errors happen?
│       ├─ NO  → Form 1 (simple return)
│       └─ YES → Form 2 (return with error)
│
└─ YES → Do you need headers/cookies?
         ├─ NO  → Form 3 (request binding)
         ├─ YES → Do you need custom status/headers?
                  ├─ NO  → Form 4 (context + request)
                  └─ YES → Form 5 (custom response)
```

---

## 🔗 Additional Forms

Lokstra supports **29 total handler forms**! 

**Other useful forms**:
- `func(ctx *request.Context) error` - Context only, no return
- `func(req *Request) error` - Request only, no data return
- `func() error` - No params, just error
- And 21 more variations!

**See**: [Deep Dive: All 29 Handler Forms](../../../02-deep-dive/router/handler-forms.md)

---

**Next**: Try combining these patterns in a [complete application](../../06-putting-it-together/)!

**Back**: [03 - Route Groups](../03-route-groups/)
