---
name: implementation-lokstra-create-handler
description: Create @Handler annotated HTTP endpoint services. Generate handler structs with @Route decorators, dependency injection, request/response handling, and auto-validation. Use after API specifications are approved to implement business logic endpoints.
phase: implementation
order: 3
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Implementation: @Handler Creation

## When to Use

Use this skill when:
- Implementing HTTP REST endpoints from API specification
- Creating handler services with multiple routes
- Setting up dependency injection for handlers
- Adding automatic request validation and error handling
- Building CRUD operations with path parameters

Prerequisites:
- ✅ API specification finalized (see: design-lokstra-api-specification)
- ✅ Domain models created (DTOs with validation tags)
- ✅ Repository interfaces defined (see: design-lokstra-module-requirements)
- ✅ config.yaml with service definitions ready (see: implementation-lokstra-yaml-config)
- ✅ Framework initialized in main.go (see: implementation-lokstra-init-framework)

## Basic Handler Structure

### Example: User Handler

**File:** `modules/user/application/user_handler.go`

```go
package application

import (
	"myapp/modules/user/domain"
	"github.com/primadi/lokstra/core/request"
)

// @Handler annotation defines:
// - name: unique service identifier for DI
// - prefix: base URL path for all routes in this handler
// - middlewares (optional): global middlewares for all routes
//
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
	// @Inject dependency injection
	// References service name from config.yaml or another @Service
	//
	// @Inject "user-repository"
	UserRepo domain.UserRepository

	// @Inject config values using "cfg:" prefix
	// @Inject "cfg:app.timeout"
	Timeout time.Duration
}

// @Route "GET /{id}"
// Returns user or error (Lokstra auto-converts to JSON response)
func (h *UserHandler) GetByID(id string) (*domain.User, error) {
	return h.UserRepo.GetByID(id)
}

// @Route "GET /"
// List all users
func (h *UserHandler) List() ([]*domain.User, error) {
	return h.UserRepo.List()
}

// @Route "POST /", middlewares=["auth"]
// Create user with request validation
// params are auto-validated based on struct tags
func (h *UserHandler) Create(params *domain.CreateUserRequest) (*domain.User, error) {
	user := &domain.User{
		Name:  params.Name,
		Email: params.Email,
	}
	return h.UserRepo.Create(user)
}

// @Route "PUT /{id}", middlewares=["auth"]
// Update user with path parameter and request body
func (h *UserHandler) Update(id string, params *domain.UpdateUserRequest) (*domain.User, error) {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	user.Name = params.Name
	user.Email = params.Email
	return h.UserRepo.Update(user)
}

// @Route "DELETE /{id}", middlewares=["auth", "admin"]
// Delete user (admin only)
func (h *UserHandler) Delete(id string) error {
	return h.UserRepo.Delete(id)
}
```

**File:** `modules/user/domain/user_dto.go`

```go
package domain

// Request DTOs with validation tags
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"omitempty,min=0,max=150"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}

// Domain model
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

---

## Handler Components Deep Dive

### 1. @Handler Annotation

Defines the handler service and its configuration:

```go
// Basic handler
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct { }

// With global middlewares
// @Handler name="admin-handler", prefix="/api/admin", middlewares=["auth", "admin"]
type AdminHandler struct { }

// Multiple handlers in same module
// @Handler name="user-public-handler", prefix="/api/public/users"
type UserPublicHandler struct { }

// @Handler name="user-admin-handler", prefix="/api/admin/users", middlewares=["auth", "admin"]
type UserAdminHandler struct { }
```

**Annotation Parameters:**

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `name` | Yes | Unique service identifier for DI | `"user-handler"` |
| `prefix` | Yes | Base URL path for all routes | `"/api/users"` |
| `middlewares` | No | Global middlewares for all routes | `["auth", "logger"]` |

**Key Rules:**
- Handler names must be unique across the application
- Prefix should follow REST conventions (`/api/resource`)
- Middlewares are applied to ALL routes in this handler
- Use empty prefix `prefix=""` to mount routes at root

### 2. @Inject Dependency Injection

Inject services and configuration into handler fields:

```go
type UserHandler struct {
	// Pattern 1: Direct service injection
	// @Inject "user-repository"
	UserRepo domain.UserRepository

	// Pattern 2: Service from config (interface selection)
	// References service name from config.yaml
	// @Inject "@repositories.user"
	UserRepo domain.UserRepository

	// Pattern 3: Config value injection
	// @Inject "cfg:app.timeout"
	Timeout time.Duration
	
	// @Inject "cfg:app.name"
	AppName string
	
	// @Inject "cfg:features.email_enabled"
	EmailEnabled bool

	// Pattern 4: Indirect config reference
	// Looks up value at key specified in another config
	// @Inject "cfg:@database.connection_string"
	DatabaseURL string
	
	// Multiple injections
	// @Inject "logger-service"
	Logger domain.Logger
	
	// @Inject "cache-service"
	Cache domain.CacheService
}
```

**Injection Patterns:**

| Pattern | Syntax | Use Case | Example |
|---------|--------|----------|---------|
| Direct service | `"service-name"` | Inject registered service | `"user-repository"` |
| Config service ref | `"@config.key"` | Service name from config | `"@repository.impl"` |
| Config value | `"cfg:key"` | Direct config value | `"cfg:app.timeout"` |
| Indirect config | `"cfg:@key"` | Config value reference | `"cfg:@jwt.secret_path"` |

### 3. @Route HTTP Endpoints

Define HTTP methods and paths for handler functions:

```go
// Basic routes
// @Route "GET /"                         // List all
// @Route "GET /{id}"                     // Get by ID
// @Route "POST /"                        // Create
// @Route "PUT /{id}"                     // Update
// @Route "PATCH /{id}"                   // Partial update
// @Route "DELETE /{id}"                  // Delete

// With route-specific middlewares
// @Route "POST /", middlewares=["auth"]
// @Route "DELETE /{id}", middlewares=["auth", "admin"]

// Multiple path parameters
// @Route "GET /{userId}/posts/{postId}"
// @Route "PUT /{teamId}/members/{memberId}"

// Complex paths
// @Route "POST /{userId}/activate"
// @Route "POST /{userId}/suspend"
// @Route "GET /{id}/export"

// Query parameters (use struct with query tags)
// @Route "GET /search"
func (h *UserHandler) Search(params *SearchParams) ([]*User, error) { }
```

**Route Parameters:**

| Component | Description | Example |
|-----------|-------------|---------|
| HTTP Method | GET, POST, PUT, PATCH, DELETE | `"GET"` |
| Path | URL path (relative to handler prefix) | `"/{id}"` |
| Path params | `{paramName}` in path | `"/{userId}/posts/{postId}"` |
| Middlewares | Route-specific middleware list | `middlewares=["auth", "rate-limit"]` |

**Important Notes:**
- Paths are relative to handler's `prefix`
- Path parameters must match function parameter names
- Middlewares in `@Route` are added to handler's global middlewares
- Route methods generate: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `HEAD`

---

## Handler Function Signatures

Lokstra supports **29+ handler signatures**. Choose the signature that best fits your needs.

### 1. Simple Return (No Parameters)

```go
// @Route "GET /ping"
func (h *UserHandler) Ping() string {
	return "pong"
}

// @Route "GET /status"
func (h *UserHandler) Status() map[string]interface{} {
	return map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now(),
	}
}
```

**Use Case:** Simple endpoints, health checks, static responses

### 2. Return with Error

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// @Route "GET /"
func (h *UserHandler) List() ([]*domain.User, error) {
	return h.UserRepo.List()
}
```

**Use Case:** Most common pattern, automatic JSON response, error handling

### 3. Request Body with Auto-Validation

```go
// @Route "POST /"
func (h *UserHandler) Create(req *domain.CreateUserRequest) (*domain.User, error) {
	// req is already validated by framework
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}
	return h.UserRepo.Create(user)
}
```

**Validation Tags (in DTO):**

```go
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"omitempty,min=0,max=150"`
	Website  string `json:"website" validate:"omitempty,url"`
	Password string `json:"password" validate:"required,min=8"`
}
```

**Use Case:** POST/PUT endpoints with request body validation

### 4. Path Parameters

```go
// Single parameter
// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) {
	return h.UserRepo.GetByID(id)
}

// Multiple parameters
// @Route "PUT /{id}"
func (h *UserHandler) Update(id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Name = req.Name
	user.Email = req.Email
	return h.UserRepo.Update(user)
}

// Multiple path parameters
// @Route "GET /{userId}/posts/{postId}"
func (h *UserHandler) GetPost(userId, postId string) (*domain.Post, error) {
	return h.PostRepo.GetByUserAndPost(userId, postId)
}
```

**Use Case:** RESTful CRUD operations with resource identifiers

### 5. With Context (Custom Response Control)

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	return ctx.Api.Ok(user)
}

// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *domain.CreateUserRequest) error {
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}
	
	saved, err := h.UserRepo.Create(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to create user")
	}
	
	// Return 201 Created
	return ctx.Api.Created(saved)
}

// @Route "DELETE /{id}"
func (h *UserHandler) Delete(ctx *request.Context, id string) error {
	err := h.UserRepo.Delete(id)
	if err != nil {
		return ctx.Api.InternalServerError("failed to delete user")
	}
	
	// Return 204 No Content
	return ctx.Api.NoContent()
}
```

**Available Response Helpers:**

```go
ctx.Api.Ok(data)                      // 200 OK
ctx.Api.Created(data)                 // 201 Created
ctx.Api.NoContent()                   // 204 No Content
ctx.Api.BadRequest(message)           // 400 Bad Request
ctx.Api.Unauthorized(message)         // 401 Unauthorized
ctx.Api.Forbidden(message)            // 403 Forbidden
ctx.Api.NotFound(message)             // 404 Not Found
ctx.Api.Conflict(message)             // 409 Conflict
ctx.Api.InternalServerError(message)  // 500 Internal Server Error
```

**Use Case:** Custom HTTP status codes, advanced error handling, accessing request context

### 6. Query Parameters

```go
// Define struct with query tags
type SearchParams struct {
	Query  string `query:"q" validate:"required"`
	Page   int    `query:"page" validate:"min=1"`
	Limit  int    `query:"limit" validate:"min=1,max=100"`
	SortBy string `query:"sort" validate:"omitempty,oneof=name email created_at"`
}

// @Route "GET /search"
func (h *UserHandler) Search(params *SearchParams) ([]*domain.User, error) {
	return h.UserRepo.Search(params.Query, params.Page, params.Limit, params.SortBy)
}

// With context
// @Route "GET /search"
func (h *UserHandler) SearchWithContext(ctx *request.Context, params *SearchParams) error {
	users, err := h.UserRepo.Search(params.Query, params.Page, params.Limit, params.SortBy)
	if err != nil {
		return ctx.Api.InternalServerError("search failed")
	}
	return ctx.Api.Ok(users)
}
```

**Use Case:** Search endpoints, filtering, pagination

### 7. Combined Parameters (Context + Path + Body)

```go
// @Route "PUT /{id}"
func (h *UserHandler) Update(ctx *request.Context, id string, req *domain.UpdateUserRequest) error {
	// Validate ID
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	// Check existence
	existing, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	// Update fields
	existing.Name = req.Name
	existing.Email = req.Email
	
	// Save
	updated, err := h.UserRepo.Update(existing)
	if err != nil {
		return ctx.Api.InternalServerError("update failed")
	}
	
	return ctx.Api.Ok(updated)
}

// @Route "POST /{teamId}/members"
func (h *TeamHandler) AddMember(ctx *request.Context, teamId string, req *domain.AddMemberRequest) error {
	member, err := h.TeamRepo.AddMember(teamId, req.UserID, req.Role)
	if err != nil {
		return ctx.Api.InternalServerError("failed to add member")
	}
	return ctx.Api.Created(member)
}
```

**Use Case:** Complex operations requiring fine-grained control

---

## request.Context: Abstraction Layers

The `request.Context` is the core API for handler functions. It provides multiple layers of abstraction over Go's standard HTTP primitives, allowing you to choose the level of abstraction that fits your needs.

### Architecture: Four-Layer Abstraction

```
Layer 4 (Business Logic)
    ↓
ctx.Api              // Opinionated API responses (wrapped in standard format)
    ↓
Layer 3 (Structured Responses)
    ↓
ctx.Resp             // Generic response builder (JSON, HTML, streaming)
    ↓
Layer 2 (Low-Level HTTP)
    ↓
ctx.W                // ResponseWriter  |  ctx.R.Method, ctx.R.URL, etc.
    ↓
Layer 1 (Go Standard Library)
    ↓
http.ResponseWriter  |  http.Request
```

### Context Fields Overview

```go
type Context struct {
	context.Context        // Embedded - standard Go context (for transactions, values)
	
	// Layer 4: High-level, opinionated API responses
	Api  *ApiHelper        // Best for: REST APIs, standard response format
	
	// Layer 3: Generic, low-level response builder
	Resp *Response         // Best for: Custom responses, non-JSON formats
	
	// Layer 2: Request/Response primitives + helpers
	Req  *RequestHelper    // Request data binding & validation (high-level)
	W    *writerWrapper    // ResponseWriter wrapper (raw HTTP output)
	R    *http.Request     // Go's standard http.Request (raw HTTP input)
}
```

---

## Layer 1: Go Standard Library

Raw HTTP primitives from Go's standard library. Use when you need maximum control.

### http.Request (ctx.R)

Access the raw HTTP request:

```go
// Basic information
method := ctx.R.Method                              // "GET", "POST", etc.
url := ctx.R.URL.String()                           // Full request URL
path := ctx.R.URL.Path                              // Just the path
query := ctx.R.URL.Query()                          // Query parameters as map
headers := ctx.R.Header                             // All HTTP headers

// Specific values
contentType := ctx.R.Header.Get("Content-Type")
userAgent := ctx.R.Header.Get("User-Agent")

// Request body (use ctx.Req.RawRequestBody() instead)
body := ctx.R.Body

// Client address
remoteAddr := ctx.R.RemoteAddr                      // "192.168.1.1:12345"

// Standard context (for database, timeouts, cancellation)
goContext := ctx.R.Context()                        // context.Context
```

### http.ResponseWriter (ctx.W)

Write raw HTTP responses (use ctx.Resp or ctx.Api instead):

```go
// ⚠️ Low-level API - prefer ctx.Resp or ctx.Api
ctx.W.Header().Set("X-Custom-Header", "value")      // Set header
ctx.W.WriteHeader(http.StatusCreated)               // Set status code
ctx.W.Write([]byte("raw data"))                     // Write body
```

---

## Layer 2: Request/Response Helpers (High-Level Primitives)

Type-safe, validated access to request data and response building.

### RequestHelper (ctx.Req)

Extract and bind request data with validation:

```go
// Single value getters (safe, with defaults)
value := ctx.Req.QueryParam("q", "")               // Query: ?q=value
value := ctx.Req.PathParam("id", "")               // Path: /users/{id}
value := ctx.Req.HeaderParam("X-API-Key", "")      // Header: X-API-Key
value := ctx.Req.FormParam("field", "")            // Form field

// Multiple value getters
values := ctx.Req.QueryParams("tags")              // []string - multiple query values
values := ctx.Req.HeaderValues("Accept")           // []string - header values

// All at once
all := ctx.Req.AllQueryParams()                    // map[string][]string
all := ctx.Req.AllHeaders()                        // map[string][]string

// Raw request body
body, err := ctx.Req.RawRequestBody()              // []byte + error

// Binding with automatic validation
err := ctx.Req.BindPath(&params)                   // Bind path params + validate
err := ctx.Req.BindQuery(&params)                  // Bind query params + validate
err := ctx.Req.BindHeader(&params)                 // Bind headers + validate
err := ctx.Req.BindBody(&params)                   // Bind JSON body + validate
err := ctx.Req.BindAll(&params)                    // Bind all sources + validate
err := ctx.Req.BindAllAuto(&params)                // Auto-detect content-type + validate
```

**Example:**

```go
// @Route "GET /users"
func (h *UserHandler) List(params *ListUsersRequest) ([]*User, error) {
	// params auto-bound and validated from:
	// - Query: ?page=1&limit=10&status=active
	// - Already validated by framework before entering handler
	
	page := params.Page   // Already converted to int, validated min=1
	limit := params.Limit // Already converted to int, validated min=1,max=100
	
	return h.repo.List(page, limit, params.Status)
}

type ListUsersRequest struct {
	Page   int    `query:"page" validate:"omitempty,min=1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Status string `query:"status" validate:"omitempty,oneof=active inactive"`
}
```

### Response (ctx.Resp)

Generic response builder supporting multiple content types:

```go
// JSON (most common)
ctx.Resp.WithStatus(http.StatusOK).Json(map[string]any{
	"id": 1,
	"name": "John",
})

// HTML
ctx.Resp.WithStatus(http.StatusOK).Html("<h1>Welcome</h1>")

// Plain text
ctx.Resp.WithStatus(http.StatusOK).Text("Hello, World!")

// Raw bytes with content-type
ctx.Resp.WithStatus(http.StatusOK).Raw("text/csv", csvBytes)

// Streaming (for large files, SSE, etc.)
ctx.Resp.Stream("text/event-stream", func(w http.ResponseWriter) error {
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "data: event %d\n\n", i)
	}
	return nil
})

// Custom headers
ctx.Resp.RespHeaders = map[string][]string{
	"X-Custom": {"value1", "value2"},
}
ctx.Resp.WithStatus(http.StatusOK).Json(data)
```

---

## Layer 3: Direct Response (ctx.Resp)

Build responses with full control over status codes and content types. Use this when `ctx.Api` doesn't fit your needs.

### Common Patterns

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	user, err := h.repo.GetByID(id)
	if err != nil {
		// Return 404 with custom message
		return ctx.Resp.WithStatus(http.StatusNotFound).Json(map[string]string{
			"error": "user not found",
		})
	}
	
	// Return 200 with data
	return ctx.Resp.WithStatus(http.StatusOK).Json(user)
}

// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *CreateUserRequest) error {
	user := &User{Name: req.Name}
	
	if err := h.repo.Create(user); err != nil {
		// Return 500 with error
		return ctx.Resp.WithStatus(http.StatusInternalServerError).Json(map[string]string{
			"error": "failed to create user",
		})
	}
	
	// Return 201 Created with location header
	ctx.Resp.RespHeaders = map[string][]string{
		"Location": {fmt.Sprintf("/users/%s", user.ID)},
	}
	return ctx.Resp.WithStatus(http.StatusCreated).Json(user)
}

// Streaming large file
// @Route "GET /export"
func (h *UserHandler) ExportCSV(ctx *request.Context) error {
	return ctx.Resp.Stream("text/csv", func(w http.ResponseWriter) error {
		w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
		
		// Write CSV header
		fmt.Fprintln(w, "ID,Name,Email")
		
		// Stream users
		users, _ := h.repo.List()
		for _, user := range users {
			fmt.Fprintf(w, "%s,%s,%s\n", user.ID, user.Name, user.Email)
		}
		return nil
	})
}
```

---

## Layer 4: API Helper (ctx.Api) - Recommended for REST APIs

Opinionated, structured API responses wrapped in a standard format. Automatically handles status codes, error formatting, and response wrapping.

### Success Responses

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	user, err := h.repo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	return ctx.Api.Ok(user)  // 200 OK, wrapped in standard format
}

// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *CreateUserRequest) error {
	user := &User{Name: req.Name, Email: req.Email}
	if err := h.repo.Create(user); err != nil {
		return ctx.Api.InternalError("failed to create user")
	}
	return ctx.Api.Created(user, "User created successfully")  // 201 Created
}

// List with pagination
// @Route "GET /"
func (h *UserHandler) List(params *ListRequest) error {
	users, total := h.repo.List(params.Page, params.Limit)
	return ctx.Api.OkList(users, &api_formatter.ListMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: (total + params.Limit - 1) / params.Limit,
	})
}
```

### Error Responses

```go
// Validation errors (automatic - framework handles)
// 400 Bad Request with field-level errors
ctx.Api.ValidationError("Validation failed", []api_formatter.FieldError{
	{Field: "email", Code: "INVALID_FORMAT", Message: "invalid email"},
	{Field: "age", Code: "OUT_OF_RANGE", Message: "must be 18-100"},
})

// Client errors
ctx.Api.BadRequest("INVALID_INPUT", "email format is invalid")       // 400
ctx.Api.Unauthorized("token is expired")                             // 401
ctx.Api.Forbidden("insufficient permissions")                        // 403
ctx.Api.NotFound("user not found")                                   // 404

// Server error
ctx.Api.InternalError("database connection failed")                  // 500

// Custom status code
ctx.Api.Error(http.StatusConflict, "DUPLICATE_EMAIL", "email already exists")  // 409
```

### Response Format

All `ctx.Api.*` responses wrap data in a standard structure (configurable via `api_formatter`):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "User retrieved",
  "data": {
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

Error response:
```json
{
  "success": false,
  "code": "NOT_FOUND",
  "message": "user not found",
  "data": null
}
```

Validation error:
```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "code": "INVALID_FORMAT",
      "message": "invalid email format"
    }
  ]
}
```

---

## Context Utilities: Values & Transactions

### Context Values Storage

Store and retrieve request-scoped values:

```go
// Simple storage (request-scoped only, not propagated)
ctx.Set("user_id", "user-123")
ctx.Set("tenant_id", "tenant-abc")

userID := ctx.Get("user_id").(string)

// Standard context values (propagated to child contexts)
ctx.SetContextValue("tracing_id", "trace-xyz")
traceID := ctx.GetContextValue("tracing_id").(string)

// Pass to services
user, err := s.repo.FindByID(ctx, userID)  // ctx embeds context.Context
```

### Transaction Management

Manage database transactions with automatic commit/rollback:

```go
// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *CreateRequest) error {
	// Begin transaction (auto-managed)
	ctx.BeginTransaction("postgres-db")
	
	// All repository calls within same handler use transaction
	user := &User{Name: req.Name}
	if err := h.repo.Create(ctx, user); err != nil {
		return err  // Auto-rollback on error
	}
	
	// Auto-commit on successful return (2xx status)
	return ctx.Api.Created(user)
}

// Multiple transactions (finalized in reverse order - LIFO)
ctx.BeginTransaction("postgres-primary")
ctx.BeginTransaction("postgres-replica")
// ... operations ...
// Auto-finalize: replica first, then primary
```

**Transaction Rules:**
- ✅ Auto-commit on 2xx status (200, 201, etc.)
- ✅ Auto-rollback on error or 4xx/5xx status
- ✅ Multiple transactions supported (LIFO finalization)
- ⚠️ Manual commit/rollback only safe in single handler (no nested calls)
- ✅ Context propagates to all repository calls

---

## Choosing the Right Layer

| Scenario | Layer | Example |
|----------|-------|---------|
| Standard REST API | Layer 4: `ctx.Api` | `ctx.Api.Ok(user)` |
| Custom status codes | Layer 3: `ctx.Resp` | `ctx.Resp.WithStatus(409).Json(data)` |
| Streaming/File download | Layer 3: `ctx.Resp` | `ctx.Resp.Stream(...)` |
| HTML/Text response | Layer 3: `ctx.Resp` | `ctx.Resp.Html(...)` |
| Raw HTTP control | Layer 1: `ctx.R`/`ctx.W` | Direct HTTP calls |
| Request data extraction | Layer 2: `ctx.Req` | `ctx.Req.QueryParam()` |
| Get client IP | Layer 2: Helper | `utils.ClientIP(ctx.R)` |
| Database context | Layer 2: `ctx.R` | `ctx.R.Context()` |

### Quick Decision Tree

```
Is it a REST API?
  ↓ Yes
  Use ctx.Api (Layer 4)
    ✓ Standard response format
    ✓ Automatic error handling
    ✓ Built-in validation error formatting
  
  ↓ No
  Do you need custom response format?
    ↓ Yes
    Use ctx.Resp (Layer 3)
      ✓ Full control over status & headers
      ✓ Multiple content types
      ✓ Streaming support
    
    ↓ No
    Use standard Go HTTP
      ✓ ctx.R for request
      ✓ ctx.W for response
```

---

## Complete Example: Multi-Layer Usage

```go
// @Route "POST /{tenantId}/users"
func (h *UserHandler) CreateUser(
	ctx *request.Context,
	tenantId string,
	req *CreateUserRequest,
) error {
	// Layer 2: Extract client IP using helper
	import "github.com/primadi/lokstra/common/utils"
	clientIP := utils.ClientIP(ctx.R)
	
	// Layer 2: Extract other headers
	userAgent := ctx.Req.HeaderParam("User-Agent", "unknown")
	apiKey := ctx.Req.HeaderParam("X-API-Key", "")
	
	// Validate API key
	if apiKey == "" {
		return ctx.Api.Unauthorized("missing API key")  // Layer 4
	}
	
	// Layer 1: Use standard context for database timeout
	dbCtx, cancel := context.WithTimeout(ctx.R.Context(), 5*time.Second)
	defer cancel()
	
	// Create transaction
	ctx.BeginTransaction("postgres-db")
	
	// Create user with metadata
	user := &User{
		TenantID:  tenantId,
		Name:      req.Name,
		Email:     req.Email,
		CreatedBy: apiKey,
		ClientIP:  clientIP,
	}
	
	if err := h.repo.Create(ctx, user); err != nil {
		// If business logic error, use custom layer 3
		if errors.Is(err, ErrEmailExists) {
			return ctx.Resp.WithStatus(http.StatusConflict).Json(map[string]string{
				"error": "email already exists",
				"code":  "DUPLICATE_EMAIL",
			})
		}
		// Generic error - use layer 4
		return ctx.Api.InternalError("failed to create user")
	}
	
	// Success - layer 4 opinionated response
	return ctx.Api.Created(user, "User created successfully")
	// Transaction auto-commits because status is 201 (2xx)
}

// Service layer example
func (s *AuthService) Register(
	ctx *request.Context,  // Can be used as context.Context
	tenantID string,
	params *RegisterRequest,
) (*RegisterResponse, error) {
	// Pass ctx directly to repository (it embeds context.Context)
	existingUser, err := s.repo.FindByUsername(ctx, params.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}
	
	// ... more service logic ...
	
	return &RegisterResponse{...}, nil
}
```

### Service Implementation Best Practices

When implementing `@Service` annotated services that receive `request.Context`:

```go
// @Service "user-service"
type UserService struct {
	// @Inject "user-repository"
	repo UserRepository
}

// Service method receiving request.Context from handler
func (s *UserService) Create(ctx *request.Context, req *CreateUserRequest) (*User, error) {
	// Pass ctx directly - it embeds context.Context
	// Repository expects context.Context
	user := &User{Name: req.Name, Email: req.Email}
	
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

// Repository interface (uses standard context.Context)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

// Repository implementation
type PostgresUserRepository struct {
	db *sql.DB
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
	// ctx can be passed directly to database queries
	// Works with both context.Context and request.Context (embedding)
	return r.db.ExecContext(ctx, "INSERT INTO users ...", user.Name)
}
```

**Key Points:**
- ✅ `request.Context` embeds `context.Context` - passes anywhere context is expected
- ✅ Repositories should use `context.Context`, not `request.Context`
- ✅ Services can accept either `request.Context` or `context.Context`
- ✅ Pass `ctx` directly to repositories - no need for `ctx.R.Context()`
- ✅ Transaction context automatically propagated through embedded context

---

## Error Handling Patterns

### 1. Automatic Error Response (Return error)

Framework automatically converts errors to 500 Internal Server Error:

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) {
	if id == "" {
		// Returns 500 with error message
		return nil, fmt.Errorf("id required")
	}
	
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		// Returns 500 with error message
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}
```

**Use Case:** Simple error handling, when 500 is acceptable for all errors

### 2. Custom Error Responses (With Context)

Use `request.Context` to return specific HTTP status codes:

```go
// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	// Validation error - 400
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	// Resource not found - 404
	user, err := h.UserRepo.GetByID(id)
	if err == sql.ErrNoRows {
		return ctx.Api.NotFound(fmt.Sprintf("user %s not found", id))
	}
	if err != nil {
		// Database error - 500
		return ctx.Api.InternalServerError("database error")
	}
	
	// Success - 200
	return ctx.Api.Ok(user)
}
```

### 3. Domain-Specific Errors

Create custom error types for better error handling:

```go
// domain/errors.go
type UserError struct {
	Code    string
	Message string
	Status  int
}

func (e *UserError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound     = &UserError{"USER_NOT_FOUND", "user not found", 404}
	ErrInvalidEmail     = &UserError{"INVALID_EMAIL", "invalid email format", 400}
	ErrDuplicateEmail   = &UserError{"DUPLICATE_EMAIL", "email already exists", 409}
	ErrUnauthorized     = &UserError{"UNAUTHORIZED", "unauthorized access", 401}
)

// Handler
// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *domain.CreateUserRequest) error {
	// Check for duplicate
	existing, _ := h.UserRepo.GetByEmail(req.Email)
	if existing != nil {
		return ctx.Api.Conflict("email already exists")
	}
	
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}
	
	created, err := h.UserRepo.Create(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to create user")
	}
	
	return ctx.Api.Created(created)
}
```

### 4. Validation Errors

Validation errors are automatically handled by framework:

```go
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"omitempty,min=18,max=150"`
	Password string `json:"password" validate:"required,min=8,containsany=!@#$%"`
}

// @Route "POST /"
func (h *UserHandler) Create(req *CreateUserRequest) (*domain.User, error) {
	// If validation fails, framework returns 400 Bad Request with validation errors
	// This code only runs if validation passes
	
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}
	return h.UserRepo.Create(user)
}
```

**Validation Error Response Format:**

```json
{
  "error": "validation failed",
  "details": [
    {
      "field": "email",
      "message": "invalid email format"
    },
    {
      "field": "password",
      "message": "must be at least 8 characters"
    }
  ]
}
```

---

## Middleware Configuration

### Global Handler Middlewares

Applied to all routes in the handler:

```go
// @Handler name="user-handler", prefix="/api/users", middlewares=["auth", "logger"]
type UserHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// All routes below inherit ["auth", "logger"] middlewares
// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) { }

// @Route "POST /"
func (h *UserHandler) Create(req *domain.CreateUserRequest) (*domain.User, error) { }
```

### Route-Specific Middlewares

Add additional middlewares to specific routes:

```go
// @Handler name="user-handler", prefix="/api/users", middlewares=["logger"]
type UserHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// Only logger middleware
// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) { }

// logger + auth middlewares
// @Route "POST /", middlewares=["auth"]
func (h *UserHandler) Create(req *domain.CreateUserRequest) (*domain.User, error) { }

// logger + auth + admin middlewares
// @Route "DELETE /{id}", middlewares=["auth", "admin"]
func (h *UserHandler) Delete(id string) error { }
```

### Middleware Registration in main.go

Middlewares must be registered before bootstrap:

```go
package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"myapp/middleware/auth"
	"myapp/middleware/admin"
	
	_ "myapp/modules/user/application" // Import handlers
)

func main() {
	// Register middlewares
	middlewares := map[string]func(http.Handler) http.Handler{
		"recovery":       recovery.Middleware(nil),
		"logger":         request_logger.Middleware(nil),
		"auth":           auth.Middleware(),
		"admin":          admin.AdminMiddleware(),
		"rate-limit":     ratelimit.Middleware(100, time.Minute),
	}
	
	// Bootstrap with middlewares
	lokstra_init.BootstrapAndRun(
		lokstra_init.WithConfigFiles("configs/config.yaml"),
		lokstra_init.WithMiddlewares(middlewares),
	)
}
```

### Custom Middleware Example

```go
// middleware/auth/auth.go
package auth

import (
	"net/http"
	"strings"
	"github.com/primadi/lokstra/core/request"
)

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			token := r.Header.Get("Authorization")
			if token == "" {
				ctx := request.NewContext(w, r)
				ctx.Api.Unauthorized("missing authorization token")
				return
			}
			
			// Validate token
			token = strings.TrimPrefix(token, "Bearer ")
			if !validateToken(token) {
				ctx := request.NewContext(w, r)
				ctx.Api.Unauthorized("invalid token")
				return
			}
			
			// Add user to context
			userID := extractUserID(token)
			r = r.WithContext(context.WithValue(r.Context(), "user_id", userID))
			
			next.ServeHTTP(w, r)
		})
	}
}
```

**Middleware Order:**
1. Handler global middlewares
2. Route-specific middlewares
3. Handler function

Example: `["logger"]` (handler) + `["auth", "admin"]` (route) = `["logger", "auth", "admin"]`

---

## Project Structure & Organization

### Recommended Structure (DDD Bounded Context)

```
modules/
├── user/
│   ├── domain/                      # Business logic & interfaces
│   │   ├── user.go                  # Domain model
│   │   ├── user_dto.go              # Request/Response DTOs
│   │   ├── user_repository.go       # Repository interface
│   │   └── user_service.go          # Service interface (optional)
│   ├── application/                 # Handlers (@Handler)
│   │   ├── user_handler.go          # Main CRUD handler
│   │   ├── user_admin_handler.go    # Admin operations (optional)
│   │   └── zz_generated.lokstra.go  # Auto-generated
│   └── infrastructure/              # Data access (@Service)
│       ├── postgres_user_repo.go    # PostgreSQL implementation
│       └── zz_generated.lokstra.go  # Auto-generated
├── order/
│   ├── domain/
│   │   ├── order.go
│   │   ├── order_dto.go
│   │   └── order_repository.go
│   ├── application/
│   │   ├── order_handler.go
│   │   └── zz_generated.lokstra.go
│   └── infrastructure/
│       ├── postgres_order_repo.go
│       └── zz_generated.lokstra.go
└── ...
```

### Handler File Organization Options

#### Option 1: Single Handler Per Module (Recommended)

Best for small to medium modules with 5-15 endpoints:

```
modules/user/application/
├── user_handler.go           # All user CRUD operations
└── zz_generated.lokstra.go   # Auto-generated
```

**user_handler.go:**

```go
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// @Route "GET /"
func (h *UserHandler) List() ([]*domain.User, error) { }

// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) { }

// @Route "POST /"
func (h *UserHandler) Create(req *domain.CreateUserRequest) (*domain.User, error) { }

// @Route "PUT /{id}"
func (h *UserHandler) Update(id string, req *domain.UpdateUserRequest) (*domain.User, error) { }

// @Route "DELETE /{id}"
func (h *UserHandler) Delete(id string) error { }
```

#### Option 2: Multiple Handlers (Public/Admin Separation)

Best for modules with different access levels:

```
modules/user/application/
├── user_public_handler.go    # Public endpoints
├── user_admin_handler.go     # Admin endpoints
└── zz_generated.lokstra.go
```

**user_public_handler.go:**

```go
// @Handler name="user-public-handler", prefix="/api/users"
type UserPublicHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// @Route "GET /"
func (h *UserPublicHandler) List() ([]*domain.User, error) { }

// @Route "GET /{id}"
func (h *UserPublicHandler) GetByID(id string) (*domain.User, error) { }
```

**user_admin_handler.go:**

```go
// @Handler name="user-admin-handler", prefix="/api/admin/users", middlewares=["auth", "admin"]
type UserAdminHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// @Route "POST /"
func (h *UserAdminHandler) Create(req *domain.CreateUserRequest) (*domain.User, error) { }

// @Route "PUT /{id}"
func (h *UserAdminHandler) Update(id string, req *domain.UpdateUserRequest) (*domain.User, error) { }

// @Route "DELETE /{id}"
func (h *UserAdminHandler) Delete(id string) error { }

// @Route "POST /{id}/suspend"
func (h *UserAdminHandler) Suspend(id string) error { }
```

#### Option 3: Feature-Based Handlers

Best for large modules with distinct feature sets:

```
modules/user/application/
├── user_crud_handler.go          # Basic CRUD
├── user_auth_handler.go          # Login/logout/password
├── user_profile_handler.go       # Profile management
└── zz_generated.lokstra.go
```

### Naming Conventions

**Handler Names:**
- Use descriptive, unique names: `"user-handler"`, `"order-admin-handler"`
- Follow kebab-case: `"user-profile-handler"` not `"UserProfileHandler"`
- Include module name for clarity: `"user-"`, `"order-"`, `"payment-"`

**File Names:**
- Use snake_case: `user_handler.go`, `order_admin_handler.go`
- Suffix with `_handler.go` for handlers
- Match the domain: `user_*.go` for user module

**Struct Names:**
- Use PascalCase: `UserHandler`, `OrderAdminHandler`
- Suffix with `Handler`: `UserPublicHandler`, `AdminHandler`
- Descriptive: `UserCRUDHandler` better than `Handler1`

**Function Names:**
- Use PascalCase (Go standard): `GetByID`, `Create`, `Update`
- Action-based: `Suspend`, `Activate`, `Archive`
- RESTful: `List`, `Create`, `Update`, `Delete`, `GetByID`

---

## Request Validation

### Validation Tags Reference

Use standard validator tags in struct fields:

```go
type CreateUserRequest struct {
	// Required field
	Name string `json:"name" validate:"required"`
	
	// Length constraints
	Username string `json:"username" validate:"required,min=3,max=20"`
	
	// Email validation
	Email string `json:"email" validate:"required,email"`
	
	// URL validation
	Website string `json:"website" validate:"omitempty,url"`
	
	// Numeric range
	Age int `json:"age" validate:"required,min=18,max=150"`
	
	// One of values
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
	
	// Pattern matching
	Phone string `json:"phone" validate:"omitempty,e164"` // E.164 format
	
	// Custom patterns
	ZipCode string `json:"zip" validate:"required,len=5,numeric"`
	
	// Password requirements
	Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
	
	// Optional with constraints
	Bio string `json:"bio" validate:"omitempty,max=500"`
	
	// Array validation
	Tags []string `json:"tags" validate:"required,min=1,max=5,dive,min=2,max=20"`
	
	// Nested struct validation
	Address *Address `json:"address" validate:"required"`
}

type Address struct {
	Street  string `json:"street" validate:"required,min=5"`
	City    string `json:"city" validate:"required"`
	Country string `json:"country" validate:"required,iso3166_1_alpha2"`
	ZipCode string `json:"zip_code" validate:"required"`
}
```

### Common Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must not be empty | `validate:"required"` |
| `omitempty` | Skip validation if empty | `validate:"omitempty,url"` |
| `min=N` | Minimum value/length | `validate:"min=3"` |
| `max=N` | Maximum value/length | `validate:"max=100"` |
| `len=N` | Exact length | `validate:"len=5"` |
| `email` | Valid email format | `validate:"email"` |
| `url` | Valid URL format | `validate:"url"` |
| `numeric` | Numeric string | `validate:"numeric"` |
| `alpha` | Alphabetic only | `validate:"alpha"` |
| `alphanum` | Alphanumeric only | `validate:"alphanum"` |
| `oneof=val1 val2` | One of allowed values | `validate:"oneof=red blue green"` |
| `e164` | E.164 phone format | `validate:"e164"` |
| `iso3166_1_alpha2` | ISO country code | `validate:"iso3166_1_alpha2"` |
| `uuid` | Valid UUID | `validate:"uuid"` |
| `contains=text` | Contains substring | `validate:"contains=@"` |
| `containsany=chars` | Contains any char | `validate:"containsany=!@#"` |
| `dive` | Validate array elements | `validate:"dive,min=2"` |

### Struct Tags for Parameter Binding

Lokstra supports multiple struct tags to bind request data to handler parameters. Only use the tags you need for each field:

```go
type CreateUserRequest struct {
	// JSON body binding (default for POST/PUT/PATCH request body)
	Email string `json:"email" validate:"required,email"`
	
	// Path parameter binding (from URL path {paramName})
	UserID string `path:"userID" validate:"required,uuid"`
	
	// Header binding
	TenantID string `header:"X-Tenant-ID" validate:"required,uuid"`
	
	// Query parameter binding
	Format string `query:"format" validate:"omitempty,oneof=json xml"`
	
	// Combined with validation
	Age int `json:"age" validate:"omitempty,min=0,max=150"`
}
```

**Important:** Each field should use only ONE binding tag:
- Use `json` for request body fields (POST/PUT/PATCH)
- Use `path` for URL path parameters (e.g., `{userID}`)
- Use `header` for HTTP headers
- Use `query` for URL query parameters

❌ **Wrong** - Multiple binding tags on same field:
```go
type Request struct {
	// DON'T do this - field will be in body OR header, not both
	TenantID string `json:"tenant_id" header:"X-Tenant-ID"`
}
```

✅ **Correct** - Use separate fields if needed from multiple sources:
```go
type Request struct {
	Email    string `json:"email" validate:"required,email"`          // From body
	UserID   string `path:"userID" validate:"required,uuid"`          // From path
	TenantID string `header:"X-Tenant-ID" validate:"required,uuid"`   // From header
	Format   string `query:"format" validate:"omitempty"`             // From query
}
```

### Path Parameter Binding

Extract URL path parameters into request struct:

```go
type GetUserRequest struct {
	// Bind from {userID} path parameter
	UserID string `path:"userID" validate:"required,uuid"`
}

// @Route "GET /{userID}"
func (h *UserHandler) GetByID(req *GetUserRequest) (*domain.User, error) {
	// req.UserID is automatically extracted from URL path
	return h.UserRepo.GetByID(req.UserID)
}

type UpdateUserRequest struct {
	// Path parameters
	UserID string `path:"userID" validate:"required,uuid"`
	
	// Body fields
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}

// @Route "PUT /{userID}"
func (h *UserHandler) Update(req *UpdateUserRequest) (*domain.User, error) {
	user := &domain.User{
		ID:    req.UserID,
		Name:  req.Name,
		Email: req.Email,
	}
	return h.UserRepo.Update(user)
}
```

**Path Tag Rules:**
- Use `path:"paramName"` to bind URL path parameters
- Path parameter name must match the placeholder in @Route (e.g., `{userID}`)
- Path values are always strings; use `validate` tags for format validation
- Multiple path parameters: `{userID}/posts/{postID}`

### Query Parameter Validation

```go
type SearchParams struct {
	Query    string `query:"q" validate:"required,min=2"`
	Page     int    `query:"page" validate:"omitempty,min=1"`
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	SortBy   string `query:"sort" validate:"omitempty,oneof=name email created_at"`
	SortDir  string `query:"dir" validate:"omitempty,oneof=asc desc"`
}

// @Route "GET /search"
func (h *UserHandler) Search(params *SearchParams) ([]*domain.User, error) {
	// params are validated automatically
	return h.UserRepo.Search(params)
}
```

### Header Parameter Binding

Extract HTTP headers into request struct:

```go
type RegisterRequest struct {
	// Extract X-Tenant-ID header (only header tag, no json tag)
	TenantID string `header:"X-Tenant-ID" validate:"required,uuid"`
	
	// Body fields
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

// @Route "POST /register", middlewares=["auth", "tenant_admin"]
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) (*RegisterResponse, error) {
	// params.TenantID is automatically extracted from X-Tenant-ID header
	return h.authService.Register(ctx, params.TenantID, params)
}
```

**Header Tag Rules:**
- Use `header:"HeaderName"` to bind HTTP headers
- Do NOT use `json` tag on header fields (they're not in request body)
- Header values are always strings; use `validate` tags for conversion/validation
- Headers are case-insensitive (e.g., `X-Tenant-ID` same as `x-tenant-id`)

### Custom Validation Error Messages

While framework provides automatic validation, you can add business logic validation:

```go
// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, req *domain.CreateUserRequest) error {
	// Framework validates struct tags automatically
	
	// Additional business logic validation
	exists, _ := h.UserRepo.GetByEmail(req.Email)
	if exists != nil {
		return ctx.Api.Conflict("email already registered")
	}
	
	// Additional validation
	if req.Age < 18 {
		return ctx.Api.BadRequest("must be at least 18 years old")
	}
	
	// Create user
	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	
	created, err := h.UserRepo.Create(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to create user")
	}
	
	return ctx.Api.Created(created)
}
```

---

## Complete Example: CRUD Handler

This example shows a complete, production-ready user handler with all CRUD operations:

**File:** `modules/user/domain/user.go`

```go
package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"` // active, inactive, suspended
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**File:** `modules/user/domain/user_dto.go`

```go
package domain

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}

type ListUsersParams struct {
	Page   int    `query:"page" validate:"omitempty,min=1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Status string `query:"status" validate:"omitempty,oneof=active inactive suspended"`
}
```

**File:** `modules/user/domain/user_repository.go`

```go
package domain

type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	List(page, limit int, status string) ([]*User, error)
	Create(user *User) (*User, error)
	Update(user *User) (*User, error)
	Delete(id string) error
}
```

**File:** `modules/user/application/user_handler.go`

```go
package application

import (
	"fmt"
	"myapp/modules/user/domain"
	"github.com/primadi/lokstra/core/request"
)

// @Handler name="user-handler", prefix="/api/users", middlewares=["recovery", "logger"]
type UserHandler struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// List users with pagination and filtering
// GET /api/users?page=1&limit=10&status=active
// @Route "GET /"
func (h *UserHandler) List(params *domain.ListUsersParams) ([]*domain.User, error) {
	// Set defaults
	page := params.Page
	if page == 0 {
		page = 1
	}
	
	limit := params.Limit
	if limit == 0 {
		limit = 10
	}
	
	return h.UserRepo.List(page, limit, params.Status)
}

// Get user by ID
// GET /api/users/{id}
// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound(fmt.Sprintf("user %s not found", id))
	}
	
	return ctx.Api.Ok(user)
}

// Create new user
// POST /api/users
// @Route "POST /", middlewares=["auth"]
func (h *UserHandler) Create(ctx *request.Context, req *domain.CreateUserRequest) error {
	// Check for duplicate email
	existing, _ := h.UserRepo.GetByEmail(req.Email)
	if existing != nil {
		return ctx.Api.Conflict("email already registered")
	}
	
	// Create user
	user := &domain.User{
		Name:   req.Name,
		Email:  req.Email,
		Status: "active",
	}
	
	created, err := h.UserRepo.Create(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to create user")
	}
	
	return ctx.Api.Created(created)
}

// Update existing user
// PUT /api/users/{id}
// @Route "PUT /{id}", middlewares=["auth"]
func (h *UserHandler) Update(ctx *request.Context, id string, req *domain.UpdateUserRequest) error {
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	// Get existing user
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound(fmt.Sprintf("user %s not found", id))
	}
	
	// Check email uniqueness if changed
	if user.Email != req.Email {
		existing, _ := h.UserRepo.GetByEmail(req.Email)
		if existing != nil && existing.ID != id {
			return ctx.Api.Conflict("email already in use")
		}
	}
	
	// Update fields
	user.Name = req.Name
	user.Email = req.Email
	
	updated, err := h.UserRepo.Update(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to update user")
	}
	
	return ctx.Api.Ok(updated)
}

// Suspend user account
// POST /api/users/{id}/suspend
// @Route "POST /{id}/suspend", middlewares=["auth", "admin"]
func (h *UserHandler) Suspend(ctx *request.Context, id string) error {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	if user.Status == "suspended" {
		return ctx.Api.BadRequest("user already suspended")
	}
	
	user.Status = "suspended"
	_, err = h.UserRepo.Update(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to suspend user")
	}
	
	return ctx.Api.Ok(map[string]string{
		"message": "user suspended successfully",
	})
}

// Activate user account
// POST /api/users/{id}/activate
// @Route "POST /{id}/activate", middlewares=["auth", "admin"]
func (h *UserHandler) Activate(ctx *request.Context, id string) error {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	if user.Status == "active" {
		return ctx.Api.BadRequest("user already active")
	}
	
	user.Status = "active"
	_, err = h.UserRepo.Update(user)
	if err != nil {
		return ctx.Api.InternalServerError("failed to activate user")
	}
	
	return ctx.Api.Ok(map[string]string{
		"message": "user activated successfully",
	})
}

// Delete user
// DELETE /api/users/{id}
// @Route "DELETE /{id}", middlewares=["auth", "admin"]
func (h *UserHandler) Delete(ctx *request.Context, id string) error {
	if id == "" {
		return ctx.Api.BadRequest("id is required")
	}
	
	// Check existence
	_, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	// Delete
	err = h.UserRepo.Delete(id)
	if err != nil {
		return ctx.Api.InternalServerError("failed to delete user")
	}
	
	return ctx.Api.NoContent()
}
```

This handler demonstrates:
- ✅ Proper error handling with appropriate HTTP codes
- ✅ Request validation (struct tags + business logic)
- ✅ Middleware configuration (global + route-specific)
- ✅ Complete CRUD operations
- ✅ Custom actions (suspend, activate)
- ✅ Query parameters (pagination, filtering)
- ✅ Dependency injection
- ✅ Clean architecture (domain interfaces)

---

## Testing Handlers

### Unit Test Example

**File:** `modules/user/application/user_handler_test.go`

```go
package application_test

import (
	"myapp/modules/user/application"
	"myapp/modules/user/domain"
	"myapp/modules/user/infrastructure/mocks"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_GetByID(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	handler := &application.UserHandler{
		UserRepo: mockRepo,
	}
	
	// Test case: success
	expectedUser := &domain.User{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	mockRepo.On("GetByID", "123").Return(expectedUser, nil)
	
	user, err := handler.GetByID("123")
	
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Create(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	handler := &application.UserHandler{
		UserRepo: mockRepo,
	}
	
	req := &domain.CreateUserRequest{
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}
	
	mockRepo.On("GetByEmail", "jane@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything).Return(&domain.User{
		ID:    "456",
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}, nil)
	
	user, err := handler.Create(req)
	
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", user.Name)
	mockRepo.AssertExpectations(t)
}
```

For comprehensive testing guidance, see: [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)

---

## Common Patterns & Best Practices

### 1. Resource Ownership & Authorization

```go
// @Route "PUT /{id}", middlewares=["auth"]
func (h *UserHandler) Update(ctx *request.Context, id string, req *domain.UpdateUserRequest) error {
	// Get authenticated user ID from context
	authUserID := ctx.Req.Context().Value("user_id").(string)
	
	// Check if user owns the resource
	if authUserID != id {
		return ctx.Api.Forbidden("cannot update other user's data")
	}
	
	// Proceed with update
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	user.Name = req.Name
	updated, err := h.UserRepo.Update(user)
	if err != nil {
		return ctx.Api.InternalServerError("update failed")
	}
	
	return ctx.Api.Ok(updated)
}
```

### 2. Soft Delete Pattern

```go
// @Route "DELETE /{id}", middlewares=["auth"]
func (h *UserHandler) Delete(ctx *request.Context, id string) error {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	// Soft delete: set deleted_at instead of actual deletion
	user.DeletedAt = time.Now()
	user.Status = "deleted"
	
	_, err = h.UserRepo.Update(user)
	if err != nil {
		return ctx.Api.InternalServerError("delete failed")
	}
	
	return ctx.Api.NoContent()
}
```

### 3. Batch Operations

```go
type BatchDeleteRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// @Route "POST /batch-delete", middlewares=["auth", "admin"]
func (h *UserHandler) BatchDelete(ctx *request.Context, req *BatchDeleteRequest) error {
	deleted := 0
	failed := []string{}
	
	for _, id := range req.IDs {
		err := h.UserRepo.Delete(id)
		if err != nil {
			failed = append(failed, id)
		} else {
			deleted++
		}
	}
	
	return ctx.Api.Ok(map[string]interface{}{
		"deleted": deleted,
		"failed":  failed,
		"total":   len(req.IDs),
	})
}
```

### 4. File Upload

```go
// @Route "POST /{id}/avatar", middlewares=["auth"]
func (h *UserHandler) UploadAvatar(ctx *request.Context, id string) error {
	// Get multipart form file
	file, header, err := ctx.Req.FormFile("avatar")
	if err != nil {
		return ctx.Api.BadRequest("avatar file required")
	}
	defer file.Close()
	
	// Validate file type
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		return ctx.Api.BadRequest("file must be an image")
	}
	
	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		return ctx.Api.BadRequest("file size must be less than 5MB")
	}
	
	// Save file (implementation depends on your storage)
	avatarURL, err := h.StorageService.SaveFile(file, header.Filename)
	if err != nil {
		return ctx.Api.InternalServerError("failed to save avatar")
	}
	
	// Update user avatar URL
	user, _ := h.UserRepo.GetByID(id)
	user.AvatarURL = avatarURL
	h.UserRepo.Update(user)
	
	return ctx.Api.Ok(map[string]string{"avatar_url": avatarURL})
}
```

### 5. Response Transformation

```go
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	// Don't expose sensitive fields
}

func toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

// @Route "GET /{id}"
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	user, err := h.UserRepo.GetByID(id)
	if err != nil {
		return ctx.Api.NotFound("user not found")
	}
	
	// Transform before returning
	return ctx.Api.Ok(toUserResponse(user))
}
```

---

## Troubleshooting

### Issue: Handler not registered

**Error:** `handler not found: user-handler`

**Solutions:**
1. Ensure handler package is imported in main.go with `_` import:
   ```go
   import _ "myapp/modules/user/application"
   ```

2. Check annotation syntax (must have no extra spaces):
   ```go
   // ✅ Correct
   // @Handler name="user-handler", prefix="/api/users"
   
   // ❌ Wrong (space after @)
   // @ Handler name="user-handler", prefix="/api/users"
   ```

3. Run code generation:
   ```bash
   go run . --generate-only
   ```

### Issue: Dependency injection fails

**Error:** `service not found: user-repository`

**Solutions:**
1. Check service is registered in config.yaml or via @Service
2. Verify service name matches exactly (case-sensitive)
3. Ensure repository package is imported in main.go
4. Check `@Inject` annotation syntax

### Issue: Validation not working

**Problem:** Invalid data passes validation

**Solutions:**
1. Ensure struct has validation tags:
   ```go
   type CreateUserRequest struct {
       Name string `json:"name" validate:"required,min=3"`
   }
   ```

2. Check parameter is pointer type:
   ```go
   // ✅ Correct - validation works
   func (h *UserHandler) Create(req *CreateUserRequest) error
   
   // ❌ Wrong - validation may not work
   func (h *UserHandler) Create(req CreateUserRequest) error
   ```

### Issue: Middleware not applied

**Problem:** Middleware not executing

**Solutions:**
1. Register middleware in main.go:
   ```go
   middlewares := map[string]func(http.Handler) http.Handler{
       "auth": auth.Middleware(),
   }
   ```

2. Check middleware name matches exactly:
   ```go
   // @Route "POST /", middlewares=["auth"]  // Must match "auth" in map
   ```

3. Verify middleware order (global → route-specific)

### Issue: Path parameters not binding

**Problem:** Path parameter is empty

**Solutions:**
1. Ensure parameter name matches route:
   ```go
   // ✅ Correct - "id" matches {id}
   // @Route "GET /{id}"
   func (h *UserHandler) GetByID(id string) error
   
   // ❌ Wrong - "userId" doesn't match {id}
   // @Route "GET /{id}"
   func (h *UserHandler) GetByID(userId string) error
   ```

2. Check route path uses curly braces: `/{id}` not `/:id`

---

## Next Steps

After creating handlers, proceed with:

1. **Create Repository Implementation**
   - See: [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md)
   - Implement @Service for data persistence

2. **Create Database Migrations**
   - See: [implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md)
   - Set up database schema

3. **Generate HTTP Test Files**
   - See: [implementation-lokstra-generate-http-files](../implementation-lokstra-generate-http-files/SKILL.md)
   - Create .http files for endpoint testing

4. **Write Tests**
   - See: [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)
   - Add unit and integration tests

5. **Add Custom Middleware**
   - See: [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md)
   - Implement auth, logging, etc.

---

## Related Skills

- [design-lokstra-api-specification](../design-lokstra-api-specification/SKILL.md) - Define API endpoints first
- [implementation-lokstra-init-framework](../implementation-lokstra-init-framework/SKILL.md) - Framework setup
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Configuration management
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - Repository implementation
- [implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md) - Database schema
- [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md) - Testing handlers
- [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md) - Custom middleware

---

## Summary Checklist

When creating a handler, ensure:

- [ ] Handler has unique name and prefix
- [ ] All dependencies injected via @Inject
- [ ] Routes use proper HTTP methods
- [ ] Request DTOs have validation tags
- [ ] Error handling returns appropriate status codes
- [ ] Middlewares registered in main.go
- [ ] Handler package imported in main.go with `_` import
- [ ] Repository interface defined in domain
- [ ] Domain models properly structured
- [ ] Code generated: `go run . --generate-only`
- [ ] Tests created for critical paths
