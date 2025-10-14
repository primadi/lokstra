# Remote Service Path Tag Detection

## Overview

`RemoteService` di Lokstra sekarang mendukung deteksi otomatis path parameters dari struct tags! Ini memungkinkan client-side API calls yang lebih natural dan type-safe tanpa perlu manual path construction.

## Features

### 1. **Automatic Path Parameter Detection**
Path parameters dideteksi otomatis dari struct tags `path:"param"`:

```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`  // /users/{dep}
    UserID       string `path:"id"`   // /users/{dep}/{id}
}

// Automatic mapping:
// GetUser(req) → GET /users/{dep}/{id}
```

### 2. **Strategy Support**
Mendukung berbagai strategi untuk path generation:

- **`rest`** (default): RESTful-style paths dengan parameter detection
- **`kebab-case`**: Legacy kebab-case paths dari method names

```go
client := api_client.NewRemoteService(clientRouter, "/users")
client.WithStrategy("rest")  // Default

// Or use kebab-case strategy:
client.WithStrategy("kebab-case")
```

### 3. **HTTP Method Inference**
HTTP method ditentukan otomatis dari method name prefix:

| Method Prefix | HTTP Method |
|---------------|-------------|
| `Get*`, `Find*`, `List*`, `Search*`, `Query*` | GET |
| `Create*`, `Add*`, `Post*` | POST |
| `Update*`, `Replace*`, `Put*` | PUT |
| `Modify*`, `Patch*` | PATCH |
| `Delete*`, `Remove*` | DELETE |

### 4. **Override Support** (Future Enhancement)
Struct-level method dan path override:

```go
// Coming soon:
type CustomRequest struct {
    Method_ string `value:"POST"`        // Override HTTP method
    Route_  string `value:"/custom/path"` // Override path
    // ... other fields
}
```

## Usage Examples

### Example 1: Simple GET with Path Parameters

**Server-side:**
```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
}

type UserService struct {}

func (s *UserService) GetUser(req *GetUserRequest) (*User, error) {
    // Implementation
}
```

**Client-side:**
```go
type RemoteUserService struct {
    client *api_client.RemoteService
}

func (s *RemoteUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
    // Automatically maps to: GET /users/{dep}/{id}
    return api_client.CallRemoteService[*User](s.client, "GetUser", ctx, req)
}

// Usage:
resp, err := remoteService.GetUser(ctx, &GetUserRequest{
    DepartmentID: "engineering",
    UserID: "123",
})
// → GET /users/engineering/123
```

### Example 2: POST with Path Parameter + Body

**Request Struct:**
```go
type CreateUserRequest struct {
    DepartmentID string `path:"dep"`     // Path param
    Name         string `json:"name"`    // Body param
    Email        string `json:"email"`   // Body param
}
```

**Client Call:**
```go
resp, err := remoteService.CreateUser(ctx, &CreateUserRequest{
    DepartmentID: "engineering",
    Name: "Alice",
    Email: "alice@example.com",
})
// → POST /users/engineering
// Body: {"name": "Alice", "email": "alice@example.com"}
```

### Example 3: GET with Path + Query Parameters

**Request Struct:**
```go
type ListUsersRequest struct {
    DepartmentID string `path:"dep"`     // Path param
    Query        string `query:"q"`      // Query param
    Page         int    `query:"page"`   // Query param
}
```

**Client Call:**
```go
resp, err := remoteService.ListUsers(ctx, &ListUsersRequest{
    DepartmentID: "engineering",
    Query: "alice",
    Page: 1,
})
// → GET /users/engineering?q=alice&page=1
```

### Example 4: PUT with Multiple Path Parameters + Body

**Request Struct:**
```go
type UpdateUserRequest struct {
    DepartmentID string `path:"dep"`     // Path param
    UserID       string `path:"id"`      // Path param
    Name         string `json:"name"`    // Body param
    Email        string `json:"email"`   // Body param
}
```

**Client Call:**
```go
resp, err := remoteService.UpdateUser(ctx, &UpdateUserRequest{
    DepartmentID: "engineering",
    UserID: "123",
    Name: "Alice Updated",
    Email: "alice.new@example.com",
})
// → PUT /users/engineering/123
// Body: {"name": "Alice Updated", "email": "alice.new@example.com"}
```

### Example 5: DELETE with Path Parameters

**Request Struct:**
```go
type DeleteUserRequest struct {
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
}
```

**Client Call:**
```go
err := remoteService.DeleteUser(ctx, &DeleteUserRequest{
    DepartmentID: "engineering",
    UserID: "123",
})
// → DELETE /users/engineering/123
```

## Implementation Details

### Path Detection Algorithm

```go
func (c *RemoteService) methodToHTTP(methodName string, req any) (httpMethod string, path string) {
    // 1. Extract struct metadata (path tags, overrides)
    pathParams, methodOverride, pathOverride := c.extractStructMetadata(req)
    
    // 2. Check for path override (highest priority)
    if pathOverride != "" {
        return methodOverride, pathOverride
    }
    
    // 3. Infer HTTP method from method name or override
    httpMethod = c.inferHTTPMethodFromMethodName(methodName, req)
    
    // 4. Build path based on strategy
    switch c.strategy {
    case "rest":
        path = c.buildRESTPath(methodName, pathParams)
    case "kebab-case":
        path = c.buildKebabPath(methodName)
    }
    
    return httpMethod, path
}
```

### Path Tag Extraction

```go
func (c *RemoteService) extractStructMetadata(req any) ([]string, string, string) {
    // Iterate through struct fields
    // Look for `path:"param"` tags
    // Collect them in order
    // Return path parameter names
}
```

### REST Path Building

```go
func (c *RemoteService) buildRESTPath(methodName string, pathParams []string) string {
    // Build path like: /users/{dep}/{id}
    basePath := c.basePath
    for _, param := range pathParams {
        basePath += "/{" + param + "}"
    }
    return basePath
}
```

## Comparison with Manual Approach

### Before (Manual Path Construction):
```go
func (s *RemoteUserService) GetUser(ctx *request.Context, dep, id string) (*User, error) {
    path := fmt.Sprintf("/users/%s/%s", dep, id)
    return api_client.FetchAndCast[*User](s.client, path, 
        api_client.WithMethod("GET"))
}
```

### After (Automatic Path Detection):
```go
func (s *RemoteUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
    return api_client.CallRemoteService[*User](s.client, "GetUser", ctx, req)
}
```

## Benefits

1. **✅ Type-Safe**: Struct-based requests ensure compile-time type checking
2. **✅ Self-Documenting**: Path parameters are visible in struct definition
3. **✅ Less Boilerplate**: No manual path construction needed
4. **✅ Consistent**: Same pattern for all CRUD operations
5. **✅ Flexible**: Supports path, query, and body parameters in one struct
6. **✅ Convention-Based**: Follows REST conventions automatically
7. **✅ Override-Friendly**: Can override method/path when needed

## Running the Example

```bash
cd cmd/examples/26-remote-service-path-detection
go run .
```

Expected output:
```
🚀 Starting server on :8080...
📡 Setting up client...

=== Testing Remote Calls with Path Tag Detection ===

1️⃣ Test GetUser (path params: dep=engineering, id=1)
   📡 [Client] GetUser: dep="engineering", id="1"
   🔍 [Server] GetUser: dep="engineering", id="1"
   ✅ Success: &{ID:1 DepartmentID:engineering Name:Alice Email:alice@example.com}

2️⃣ Test ListUsers (path: dep=engineering, query: page=1)
   📡 [Client] ListUsers: dep="engineering"
   🔍 [Server] ListUsers: dep="engineering", q="", page=1
   ✅ Success: Found 2 users
      - Alice (alice@example.com)
      - Bob (bob@example.com)

...
```

## Future Enhancements

1. **Struct-Level Tags**: Support `method:"POST"` and `route:"/custom"` at struct level
2. **Path Variable Resolution**: Auto-replace `{param}` with actual values
3. **Query Parameter Building**: Auto-build query string from `query:"name"` tags
4. **Header Support**: Support `header:"X-API-Key"` tags
5. **Validation Integration**: Auto-validate requests before sending
6. **OpenAPI Generation**: Generate OpenAPI spec from struct tags

## Related Documentation

- [Service Router](../20-service-router-struct-based/)
- [API Client](../../docs/api-client.md)
- [Remote Service](../../docs/remote-service.md)
- [Struct Tags](../../docs/struct-tags.md)
