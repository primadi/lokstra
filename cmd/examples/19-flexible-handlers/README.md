# Example 19: Flexible Handler Patterns

This example demonstrates the enhanced `adaptSmart` function that automatically adapts various handler signatures without requiring service router conventions.

## Key Features

### ✨ **Multiple String Path Parameters** (NEW!)
```go
// GET /departments/{dep}/users/{id}
func GetUser(dep, id string) (*User, error) {
    // dep from path, id from path
}
```

### 🎯 **Automatic Parameter Detection**
The router automatically detects and extracts parameters based on their types:
- **String parameters** → Path parameters (in order)
- **Struct pointers** → Bind using tags (`path`, `query`, `header`, `json`)
- **Integer parameters** → Path parameters converted to int
- **Context** → Optional first parameter for HTTP context access

## Supported Handler Patterns

### Pattern A: Pure Business Logic with Multiple Path Params
```go
func (s *Service) GetUser(dep, id string) (*User, error)
//                        ↑    ↑
//                      path  path
```
- No HTTP context
- Multiple string parameters from URL path
- Returns data and error
- Framework wraps response automatically

### Pattern B: Context + Multiple Path Params
```go
func (s *Service) DeleteUser(ctx *request.Context, dep, id string) error
//                           ↑                     ↑    ↑
//                        context                path  path
```
- Has HTTP context access
- Multiple string path parameters
- Manual response control
- Return error only

### Pattern C: Struct with Tags
```go
type SearchRequest struct {
    DepartmentID string `path:"dep"`      // From URL path
    Query        string `query:"q"`       // From query string
    Page         int    `query:"page"`    // From query string
}

func (s *Service) SearchUsers(req *SearchRequest) ([]*User, error)
```
- Struct with flexible tag-based binding
- Can combine path, query, header, body sources
- Pure business logic (no context)

### Pattern D: Context + Mixed Parameters
```go
func (s *Service) CreateUser(ctx *request.Context, dep string, req *CreateUserRequest) (*User, error)
//                           ↑                     ↑          ↑
//                        context                path       body struct
```
- Has context + path param + body struct
- Mix of simple and complex parameters
- Full flexibility

### Pattern E: Traditional Context-Only
```go
func (s *Service) ListDepartments(ctx *request.Context) error
```
- Traditional Lokstra handler
- Manual parameter extraction
- Manual response control

## How It Works

### Parameter Extraction by Type

```go
// adaptSmart automatically detects parameter types:

func Handler(dep, id string) {
    // String params → path params by position
    // dep = ctx.Req.PathParam("dep" or "param0")
    // id = ctx.Req.PathParam("id" or "param1")
}

func Handler(req *Struct) {
    // Struct pointer → BindAll with tags
    // req.Field = from `path:"..."` or `query:"..."` etc.
}

func Handler(ctx *request.Context, id string, req *Struct) {
    // Mixed: context + path + body
    // ctx = request context
    // id = from path
    // req = from body with tags
}
```

### Path Parameter Naming

For string parameters, the router tries these names in order:
1. **Common names**: `id`, `dep`, `category`, `type`, `name`
2. **Indexed names**: `param0`, `param1`, `param2`, etc.
3. **Numeric**: `0`, `1`, `2`, etc.

Example route patterns:
```go
// Handler: func(dep, id string)
GET /departments/{dep}/users/{id}     // ✅ Works with {dep} and {id}
GET /departments/{param0}/users/{param1}  // ✅ Works with indexed names
```

## Running the Example

```bash
go run main.go
```

Server starts on `:3000` with these endpoints:

```
GET    http://localhost:3000/departments
GET    http://localhost:3000/departments/{dep}/users/{id}
DELETE http://localhost:3000/departments/{dep}/users/{id}
GET    http://localhost:3000/departments/{dep}/users/search?q=xxx
POST   http://localhost:3000/departments/{dep}/users
```

## Testing

Use the provided `test.http` file with REST Client extension:

```http
### Get user from engineering department
GET http://localhost:3000/departments/engineering/users/1

### Create new user in sales department
POST http://localhost:3000/departments/sales/users
Content-Type: application/json

{
  "name": "Eve",
  "email": "eve@example.com"
}

### Search users in department
GET http://localhost:3000/departments/sales/users/search?q=eve
```

## Comparison with Service Router

### Service Router (Example 18):
- Uses naming conventions: `GetUser` → `GET /users/{id}`
- Automatic route generation from method names
- Less control over route patterns
- Best for REST APIs with standard patterns

### Flexible Handlers (Example 19):
- Manual route registration with flexible handlers
- Any route pattern you want
- More control, less convention
- Best for custom APIs or mixed patterns

## Benefits

1. **🚀 Zero Boilerplate**: No wrapper functions needed
2. **🎯 Type Safety**: Compile-time parameter type checking
3. **🔄 Flexible**: Mix different parameter styles in one API
4. **📝 Clean Code**: Business logic separated from HTTP concerns
5. **⚡ Performance**: Reflection done once at registration time
6. **🛡️ Safe**: Panics early if handler signature is invalid

## Error Handling

If handler signature is not recognized, router panics with clear message:

```
Invalid handler type for path [/users/{id}]. Supported signatures:
  - func(*Context) error
  - func(*Context) (data, error)
  - func(*Context, *Struct) error
  - func(*Context, *Struct) (data, error)
  - func(*Context, string, ...) error
  - func(*Context, string, ...) (data, error)
  - func(*Struct) error
  - func(*Struct) (data, error)
  - func(string, ...) error
  - func(string, ...) (data, error)
  - request.HandlerFunc
  - http.HandlerFunc
  - http.Handler
```

## When to Use

### ✅ Use Flexible Handlers When:
- Need custom route patterns (e.g., `/departments/{dep}/users/{id}`)
- Want maximum control over routes
- Building non-REST APIs
- Migrating from other frameworks

### ✅ Use Service Router When:
- Building standard REST APIs
- Want minimal code (convention over configuration)
- Prefer automatic route generation
- Team follows REST conventions

### 💡 Best Practice:
You can **mix both approaches** in the same application! Use service router for standard CRUD, flexible handlers for special cases.

## Next Steps

1. Try different parameter combinations
2. Add validation to struct tags
3. Combine with middleware
4. Use in production APIs

See also:
- Example 18: Service Router (Convention-based)
- Example 14: Client Response Parsing
- Example 09: API Standard
