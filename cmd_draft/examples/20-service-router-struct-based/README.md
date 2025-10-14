# Example 20: Service Router with Struct-Based Parameters

## 📋 Overview

This example demonstrates the **NEW and RECOMMENDED approach** for service router with **struct-based parameters**. All path, query, header, and body parameters are defined via struct tags, eliminating the reflection limitation and providing clear, type-safe routing.

## ✅ Why Struct-Based Parameters?

### Problem with Direct String Parameters
```go
// ❌ OLD: Direct string parameters
func (s *Service) GetUser(dep, id string) (*User, error)
// Problem: Go reflection can't get parameter names!
// Can only see types: (string, string) - no names
// Must guess param0, param1 or hardcode "id"
```

### Solution: Struct with Tags
```go
// ✅ NEW: Struct with explicit tags
type GetUserRequest struct {
    DepartmentID string `path:"dep"`   // Explicit name!
    UserID       string `path:"id"`    // Explicit name!
}
func (s *Service) GetUser(req *GetUserRequest) (*User, error)
// Path auto-generated: /users/{dep}/{id}
// Names extracted from tags: "dep", "id"
```

## 🎯 Benefits

1. **Type-Safe** - Struct fields are validated at compile time
2. **Self-Documenting** - Struct fields explain what parameters exist
3. **Flexible** - Mix path, query, header, body in one struct
4. **No Reflection Limitation** - Tag names are explicit, not guessed
5. **Consistent** - Same approach for all parameter types

## 📦 Supported Tag Types

```go
type CompleteRequest struct {
    // Path parameters
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
    
    // Query parameters
    SearchQuery  string `query:"q"`
    Page         int    `query:"page"`
    
    // Header parameters
    AuthToken    string `header:"Authorization"`
    
    // Body parameters (JSON)
    Name         string `json:"name" validate:"required"`
    Email        string `json:"email" validate:"required,email"`
}
```

## 🔄 Pattern Examples

### Pattern 1: Simple GET with Path Params
```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
}

func (s *Service) GetUser(req *GetUserRequest) (*User, error) {
    // req.DepartmentID = "engineering"
    // req.UserID = "1"
    return s.users[req.DepartmentID][req.UserID], nil
}

// Auto-generated route: GET /users/{dep}/{id}
// Test: GET /users/engineering/1
```

### Pattern 2: GET with Path + Query Params
```go
type ListUsersRequest struct {
    DepartmentID string `path:"dep"`
    Query        string `query:"q"`
    Page         int    `query:"page"`
}

func (s *Service) ListUsers(req *ListUsersRequest) ([]*User, error) {
    // req.DepartmentID = "engineering"
    // req.Query = "alice"
    // req.Page = 1
    return s.filterUsers(req.DepartmentID, req.Query), nil
}

// Auto-generated route: GET /users/{dep}
// Test: GET /users/engineering?q=alice&page=1
```

### Pattern 3: POST with Path + Body Params
```go
type CreateUserRequest struct {
    DepartmentID string `path:"dep"`            // From path
    Name         string `json:"name"`           // From body
    Email        string `json:"email"`          // From body
}

func (s *Service) CreateUser(req *CreateUserRequest) (*User, error) {
    // req.DepartmentID = "engineering" (from path)
    // req.Name = "David" (from JSON body)
    // req.Email = "david@example.com" (from JSON body)
    return s.create(req.DepartmentID, req.Name, req.Email), nil
}

// Auto-generated route: POST /users/{dep}
// Test: POST /users/engineering + JSON body
```

### Pattern 4: PUT/PATCH with Multiple Params
```go
type UpdateUserRequest struct {
    DepartmentID string `path:"dep"`     // Path
    UserID       string `path:"id"`      // Path
    Name         string `json:"name"`    // Body
    Email        string `json:"email"`   // Body
}

func (s *Service) UpdateUser(req *UpdateUserRequest) (*User, error) {
    return s.update(req.DepartmentID, req.UserID, req.Name, req.Email), nil
}

// Auto-generated route: PUT /users/{dep}/{id}
// Test: PUT /users/engineering/1 + JSON body
```

## 🚀 Path Generation Logic

The service router automatically generates paths from struct tags:

1. **Extract path tags** - Scan struct for `path:"xxx"` tags
2. **Generate path** - Build `/resource/{param1}/{param2}/...`
3. **Register route** - Connect HTTP method + path + handler

```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`   // Order matters!
    UserID       string `path:"id"`
}

// Extracted tags: ["dep", "id"] (in order)
// Convention: GetUser → GET
// Generated path: GET /users/{dep}/{id}
```

## 🧪 Running the Example

```bash
# Build
go build ./cmd/examples/20-service-router-struct-based

# Run
./20-service-router-struct-based

# Test with curl
curl http://localhost:3000/users/engineering/1
curl http://localhost:3000/users/engineering?q=alice
curl -X POST http://localhost:3000/users/engineering \
  -H "Content-Type: application/json" \
  -d '{"name":"David","email":"david@example.com"}'
```

Or use the `test.http` file with REST Client extension in VS Code.

## 📊 Generated Routes

```
GET    /users/{dep}/{id}     - GetUser
GET    /users/{dep}          - ListUsers (+ query params)
POST   /users/{dep}          - CreateUser
PUT    /users/{dep}/{id}     - UpdateUser
DELETE /users/{dep}/{id}     - DeleteUser
```

## 🎓 Key Takeaways

1. **Always use structs for parameters** - Don't use direct string params
2. **Tags are explicit** - No guessing param names
3. **Path generation is automatic** - Based on struct tag order
4. **Mix parameter types** - Path, query, header, body in one struct
5. **Type-safe and validated** - Compile-time checks

## 🔄 Migration from Old Approach

### Before (Direct Params - ❌ Not Recommended)
```go
func (s *Service) GetUser(dep, id string) (*User, error)
// Problem: Can't get param names from reflection
// Generated: /users/{param0}/{param1} ❌ Generic!
```

### After (Struct-Based - ✅ Recommended)
```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
}
func (s *Service) GetUser(req *GetUserRequest) (*User, error)
// Generated: /users/{dep}/{id} ✅ Explicit!
```

## 📚 See Also

- Example 18: Service Router Comparison (old approach)
- Example 19: Flexible Handlers (manual routing with adaptSmart)
- `core/router/convention.go` - Path generation logic
- `core/router/helper.go` - Smart handler adaptation
