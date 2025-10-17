# CRUD API Example

> **Complete REST API with Service pattern and Lazy Dependency Injection**

Related to: [Architecture - Service Component](../../architecture.md#component-4-service)

---

## 📖 What This Example Shows

### Service Patterns:
- ✅ Service layer with business logic
- ✅ **Cached service access** (`service.LazyLoad[T]`) - **Optimal pattern**
- ✅ Lazy dependency injection within services
- ✅ Service registry (Factory → Lazy Service)
- ✅ Service methods with struct parameters

### Service Access Patterns (Quick Reference):

| Pattern | Performance | Error Handling | Use Case |
|---------|-------------|----------------|----------|
| `GetService[T]()` | ⚠️ Slow (map lookup every call) | Returns nil (confusing) | Quick prototypes only |
| `MustGetService[T]()` | ⚠️ Slow (map lookup every call) | Panics (clear error) | Fail-fast prototypes |
| **`LazyLoad + MustGet()`** | ✅ **Fast (cached)** | ✅ **Panics (clear error)** | **Production (recommended)** |

```go
// ✅ Optimal: Package-level cached service with fail-fast
var userService = service.LazyLoad[*UserService]("users")

func handler() {
    userService.MustGet().DoSomething() // Fast + clear errors!
}
```

### Complete CRUD:
- ✅ Create (POST)
- ✅ Read (GET single & list)
- ✅ Update (PUT)
- ✅ Delete (DELETE)
- ✅ Validation & error handling
- ✅ In-memory database

### Concepts:
- ✅ Service-oriented architecture
- ✅ Thread-safe operations (mutex)
- ✅ Request binding (path, body)
- ✅ Custom error responses (404, 409, 500)

---

## 🚀 Run the Example

```bash
# From this directory
go run main.go
```

Server will start on `http://localhost:3002`

---

## 🧪 Test the Endpoints

Use the `test.http` file in VS Code with REST Client extension.

### List all users

```bash
curl http://localhost:3002/api/v1/users
```

### Get user by ID

```bash
curl http://localhost:3002/api/v1/users/1
```

### Create user

```bash
curl -X POST http://localhost:3002/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

### Update user

```bash
curl -X PUT http://localhost:3002/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'
```

### Delete user

```bash
curl -X DELETE http://localhost:3002/api/v1/users/2
```

---

## 📝 Response Examples

### GET /api/v1/users (List)

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "name": "Alice",
      "email": "alice@example.com",
      "created_at": "2025-10-14T10:00:00Z",
      "updated_at": "2025-10-14T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Bob",
      "email": "bob@example.com",
      "created_at": "2025-10-14T22:00:00Z",
      "updated_at": "2025-10-14T22:00:00Z"
    }
  ]
}
```

### POST /api/v1/users (Create)

**Request**:
```json
{
  "name": "Charlie",
  "email": "charlie@example.com"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": 3,
    "name": "Charlie",
    "email": "charlie@example.com",
    "created_at": "2025-10-15T10:30:00Z",
    "updated_at": "2025-10-15T10:30:00Z"
  }
}
```

### GET /api/v1/users/999 (Not Found)

**Response** (404):
```json
{
  "status": "error",
  "code": "NOT_FOUND",
  "message": "user with ID 999 not found"
}
```

### POST /api/v1/users (Duplicate Email)

**Request**:
```json
{
  "name": "Duplicate",
  "email": "alice@example.com"
}
```

**Response** (409 Conflict):
```json
{
  "status": "error",
  "code": "DUPLICATE",
  "message": "email already exists"
}
```

---

## 💡 Key Concepts

### 1. Service Layer Pattern

#### Service Definition
```go
type UserService struct {
    DB *service.Cached[*Database]  // Lazy-loaded dependency
}
```

**Why Lazy?**
- Services created only when first accessed
- Prevents circular dependency issues
- Memory efficient

#### Service Methods with Struct Parameters

**IMPORTANT**: Service methods that need request data **must use struct parameters**.

```go
// ✅ CORRECT - Method with input data uses struct
type CreateParams struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    return s.DB.MustGet().Create(p.Name, p.Email)
}

// ✅ CORRECT - Method without input data uses no parameters
func (s *UserService) GetAll() ([]*User, error) {
    return s.DB.MustGet().GetAll()
}

// ❌ WRONG - Don't use empty struct as parameter
type GetAllParams struct{} // Empty struct

func (s *UserService) GetAll(p *GetAllParams) ([]*User, error) {
    // This may cause issues with Lokstra's binding mechanism
    return s.DB.MustGet().GetAll()
}
```

**Why struct params for methods with input?**
- Lokstra can bind from path, query, body, headers
- Auto-validation via struct tags
- Type-safe parameter passing
- Clear API contract

**Why no params for methods without input?**
- Simpler, more idiomatic Go
- Avoids empty struct overhead
- No binding needed = no potential binding errors

### 2. Service Registry

#### Register Services

**Step 1: Register Factory** (blueprint)
```go
lokstra_registry.RegisterServiceFactory("dbFactory", NewDatabase)
lokstra_registry.RegisterServiceFactory("usersFactory", func() any {
    return &UserService{
        DB: service.LazyLoad[*Database]("db"),
    }
})
```

**Step 2: Register Lazy Service** (uses factory)
```go
lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
```

**Key Concepts**:
- **Factory**: Blueprint/constructor function (registered once)
- **Service**: Instance with specific config (can have multiple from one factory)
- **Lazy**: Service created on first access, not at registration

#### Use Services in Handlers (3 Patterns)

**Pattern 1: ❌ GetService (Not Optimal)**
```go
func handler(ctx *request.Context) error {
    // ⚠️ Registry lookup EVERY request!
    userService := lokstra_registry.GetService[*UserService]("users")
    
    users, err := userService.GetAll()
    // ...
}
```
- **Pros**: Simple, straightforward
- **Cons**: Map lookup every request (slower)
- **Risk**: Returns nil if service not found

**Pattern 2: ⚠️ MustGetService (Better, but still not optimal)**
```go
func handler(ctx *request.Context) error {
    // ⚠️ Registry lookup EVERY request, but panics if not found
    userService := lokstra_registry.MustGetService[*UserService]("users")
    
    users, err := userService.GetAll()
    // ...
}
```
- **Pros**: Fail-fast (panics on missing service)
- **Cons**: Still map lookup every request

**Pattern 3: ✅ service.LazyLoad (OPTIMAL - Recommended)**
```go
// Package-level or struct field (NOT function-local!)
var userService = service.LazyLoad[*UserService]("users")

func listUsersHandler(ctx *request.Context) error {
    // ✅ MustGet() panics with clear error if service not found (recommended)
    users, err := userService.MustGet().GetAll()
    if err != nil {
        return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
    }
    
    return ctx.Api.Ok(users)
}
```
- **Pros**: 
  - **Loaded once**, cached forever
  - **Zero lookup cost** after first access
  - **Thread-safe** via sync.Once
  - **Clear errors** with MustGet() (fail-fast)
  - Type-safe generics
- **Cons**: Requires understanding of caching concept
- **Best for**: Package-level vars or struct fields

**MustGet() vs Get()**:
```go
// ✅ Recommended: MustGet() - Fail-fast with clear error
users, err := userService.MustGet().GetAll()
// Panics: "service 'users' not found or not initialized"

// ⚠️ Not recommended: Get() - Returns nil (confusing errors)
users, err := userService.MustGet().GetAll()
// Panics: "nil pointer dereference" (unclear what's wrong!)
```

**📖 Deep Dive**: See [Essentials → Service Guide](../../01-essentials/02-service/) for comprehensive patterns and best practices.

### 3. Request Binding

#### Path Parameters
```go
type GetByIDParams struct {
    ID int `path:"id"`  // Binds from URL path
}

func getUserHandler(ctx *request.Context) error {
    var params GetByIDParams
    if err := ctx.Req.BindPath(&params); err != nil {
        return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
    }
    
    // params.ID is now populated from /users/{id}
}
```

#### JSON Body
```go
type CreateParams struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func createUserHandler(ctx *request.Context) error {
    var params CreateParams
    if err := ctx.Req.BindBody(&params); err != nil {
        return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
    }
    
    // params.Name and params.Email are now populated + validated
}
```

#### Combined (Path + Body)
```go
type UpdateParams struct {
    ID    int    `path:"id"`       // From URL
    Name  string `json:"name"`     // From body
    Email string `json:"email"`    // From body
}

func updateUserHandler(ctx *request.Context) error {
    var params UpdateParams
    ctx.Req.BindPath(&params)  // Bind path first
    ctx.Req.BindBody(&params)  // Then bind body
    
    // params has both ID (from URL) and Name/Email (from JSON)
}
```

### 4. Error Handling

#### Custom Error Codes
```go
// 404 Not Found
if err != nil {
    return ctx.Api.Error(404, "NOT_FOUND", err.Error())
}

// 409 Conflict (duplicate)
if err.Error() == "email already exists" {
    return ctx.Api.Error(409, "DUPLICATE", err.Error())
}

// 500 Internal Error
return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
```

#### Success Responses
```go
// 200 OK
return ctx.Api.Ok(user)

// 201 Created with message
return ctx.Api.Created(user, "User created successfully")

// 200 OK with custom message
return ctx.Api.OkWithMessage(nil, "User deleted successfully")
```

### 5. Thread-Safe Database

```go
type Database struct {
    users  map[int]*User
    nextID int
    mu     sync.RWMutex  // For concurrent access
}

func (db *Database) GetAll() ([]*User, error) {
    db.mu.RLock()         // Read lock
    defer db.mu.RUnlock()
    
    // Safe to read
}

func (db *Database) Create(...) (*User, error) {
    db.mu.Lock()          // Write lock
    defer db.mu.Unlock()
    
    // Safe to write
}
```

---

## 🎯 Architecture Flow

```
Request
  ↓
Handler (getUserHandler)
  ↓
Get Service from Registry
  ↓
Service Method (userService.GetByID)
  ↓
Lazy Load DB (first access only)
  ↓
Database Operation
  ↓
Return Result
  ↓
Format Response (ApiHelper)
  ↓
Response
```

---

## 🔍 What's Next?

Try modifying:
- Add pagination to list endpoint
- Add search/filter functionality
- Implement authentication
- Add more validation rules
- Use real database (PostgreSQL, MySQL)

See more examples:
- [Multi-Deployment](../04-multi-deployment/) - Run as monolith or microservices
- [Service as Router](../../01-essentials/02-service/) - Auto-generate routes from service

---

**Questions?** Check the [Architecture Guide](../../architecture.md)
