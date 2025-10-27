# CRUD API Example - Two Approaches!

> **Complete REST API demonstrating BOTH manual and config-based service management**

🎯 **This example runs in TWO MODES:**
- **Mode 1: By Code** - Manual instantiation (simple, explicit)
- **Mode 2: By Config** - YAML + Lazy DI (scalable, production-ready)

Related to: [Architecture - Service Component](../../architecture.md#component-4-service)

---

## 🎭 Choose Your Approach!

### Mode 1: Run by Code (Manual)
```bash
go run main.go --mode=code
# or just: go run main.go
```

**What you'll see:**
```
🚀 Starting CRUD API in 'code' mode...
📝 APPROACH 1: Manual instantiation (run by code)
```

**How it works:**
- Services created manually in code
- Direct instantiation: `db := NewDatabase()`
- Simple and explicit
- Good for learning and small apps

---

### Mode 2: Run by Config (YAML + Lazy DI)
```bash
go run main.go --mode=config
```

**What you'll see:**
```
🚀 Starting CRUD API in 'config' mode...
⚙️ APPROACH 2: YAML Configuration + Lazy DI (run by config)
✅ Services loaded from YAML config
```

**How it works:**
- Services defined in `config.yaml`
- Factory pattern with lazy loading
- Declarative configuration
- Production-ready with validation
- **Path rewrite**: Routes use `/api/v2/*` instead of `/api/v1/*` (configured in YAML)

---

## 📖 What This Example Shows

### Core Features:
- ✅ Complete CRUD operations (Create, Read, Update, Delete)
- ✅ Service layer with business logic
- ✅ Thread-safe in-memory database
- ✅ Request binding (path, body)
- ✅ Custom error responses (404, 409, 500)

### TWO Service Management Approaches:
1. **Manual** - Direct instantiation (code mode)
2. **Factory + YAML** - Lazy DI with config (config mode)

### Lazy Dependency Injection:
- ✅ `service.Cached[T]` for type-safe lazy loading
- ✅ Database only created when UserService needs it
- ✅ No initialization order issues
- ✅ MustGet() for fail-fast error handling

---

## 🚀 Quick Start

```bash
# From this directory
go run main.go
```

Server will start on `http://localhost:3002`

---

## 🧪 Test the Endpoints

Use the `test.http` file in VS Code with REST Client extension.

**Note**: In config mode, paths are rewritten from `/api/v1/*` to `/api/v2/*` via YAML configuration.

### List all users

```bash
# Code mode:
curl http://localhost:3002/api/v1/users

# Config mode (path rewrite enabled):
curl http://localhost:3002/api/v2/users
```

### Get user by ID

```bash
# Code mode:
curl http://localhost:3002/api/v1/users/1

# Config mode:
curl http://localhost:3002/api/v2/users/1
```

### Create user

```bash
# Code mode:
curl -X POST http://localhost:3002/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'

# Config mode:
curl -X POST http://localhost:3002/api/v2/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

### Update user

```bash
# Code mode:
curl -X PUT http://localhost:3002/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'

# Config mode:
curl -X PUT http://localhost:3002/api/v2/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'
```

### Delete user

```bash
# Code mode:
curl -X DELETE http://localhost:3002/api/v1/users/2

# Config mode:
curl -X DELETE http://localhost:3002/api/v2/users/2
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

### 6. Path Rewrite (Config Mode Only)

**In config mode**, the YAML configuration includes path rewrite rules that transform routes:

```yaml
router-definitions:
  api:
    path-rewrites:
      - pattern: "^/api/v1/(.*)$"
        replacement: "/api/v2/$1"
```

**What this does:**
- All routes originally defined as `/api/v1/*` are rewritten to `/api/v2/*`
- Uses regex pattern matching for flexible transformations
- Applied at router build time (zero runtime overhead)
- Route names remain unchanged (only HTTP paths change)

**Example transformation:**
```
Original code:    r.GET("/api/v1/users", ...)
After rewrite:    GET /api/v2/users  (accessible via this path)
Internal name:    GET_/api/v1/users  (unchanged, for tracking)
```

**Use cases:**
- API versioning without code changes
- Migrating from old to new URL structure
- A/B testing different path conventions
- Multi-tenant path prefixing

**Note**: This is an alternative to `path-prefix`. Choose based on your needs:
- `path-prefix`: Adds prefix to all routes (simpler)
- `path-rewrites`: Regex-based transformation (more flexible)

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

## � Comparing Both Approaches

### Side-by-Side:

| Aspect | Mode: Code | Mode: Config |
|--------|------------|--------------|
| **Setup** | Very simple | More setup (factories) |
| **Services** | Manual instantiation | YAML definition |
| **Dependencies** | Manual wiring | Auto lazy-load |
| **Config** | Hardcoded | YAML + env vars |
| **Multi-env** | Manual flags | Built-in deployments |
| **Validation** | None | JSON Schema |
| **Best for** | Learning, small apps | Production, teams |

### Try Both!

```bash
# Run in CODE mode - see manual approach
go run main.go --mode=code

# Run in CONFIG mode - see YAML approach
go run main.go --mode=config

# Both produce identical API behavior!
# Test with: curl http://localhost:3002/api/v1/users
```

**Read detailed comparison:** [MIGRATION.md](./MIGRATION.md)

---

## 📚 Learn More

### Understanding the Patterns:
- **[CODE-VS-CONFIG.md](../../CODE-VS-CONFIG.md)** - Parallel structure explanation (NEW!)
- **[MIGRATION.md](./MIGRATION.md)** - Detailed comparison of both approaches
- **[config.yaml](./config.yaml)** - Example YAML configuration
- **Service Factories** - See `DatabaseFactory` and `UserServiceFactory` in main.go

### When to Use Which:
- **Use "code" mode** when:
  - Learning Lokstra
  - Prototyping quickly
  - Simple apps (1-3 services)
  - Want full explicit control

- **Use "config" mode** when:
  - Building production apps
  - Multiple environments (dev/staging/prod)
  - Complex dependencies (5+ services)
  - Team development

---

## �🔍 What's Next?

Try modifying:
- Add pagination to list endpoint
- Add search/filter functionality
- Implement authentication
- Add more validation rules
- Switch between both modes

See more examples:
- [Multi-Deployment](../04-multi-deployment/) - Monolith vs microservices with YAML

---

**Questions?** Check [MIGRATION.md](./MIGRATION.md) for detailed explanations!
