# Example 21: Type Alias Context Detection

## 🎯 Problem

When using type aliases for `*request.Context`, name-based detection fails:

```go
// Type alias
type RequestContext = request.Context

// Service method
func (s *Service) ListUsers(ctx *RequestContext) ([]*User, error)

// ❌ OLD: Name-based detection
if elemType.Name() == "Context" {  // Fails! Name is still "Context"
    // But the package path might be different
}
```

## ✅ Solution: Type Comparison

Use **type comparison** instead of name comparison:

```go
// Define reference type
var typeOfContextPtr = reflect.TypeOf((*request.Context)(nil))

// Compare types directly
if paramType == typeOfContextPtr {
    // ✅ Works with type aliases!
    continue
}
```

## 🔬 How Type Comparison Works

```go
// Original type
type Context struct { ... }

// Type alias (same underlying type)
type RequestContext = request.Context
type MyContext = request.Context

// Type comparison
reflect.TypeOf((*request.Context)(nil)) == reflect.TypeOf((*RequestContext)(nil))  // ✅ true
reflect.TypeOf((*request.Context)(nil)) == reflect.TypeOf((*MyContext)(nil))       // ✅ true

// Name comparison (FAILS)
elemType.Name() == "Context"  // ✅ true for request.Context
elemType.Name() == "Context"  // ❌ Still true for RequestContext, but not what we want!
```

## 📋 Test Cases

This example tests 4 different signatures:

### 1. Type Alias Context Only
```go
type RequestContext = request.Context

func (s *Service) ListUsers(ctx *RequestContext) ([]*User, error)
// Should generate: GET /users (no path params)
// Context detected correctly via type comparison
```

### 2. Original Context + Struct Param
```go
func (s *Service) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error)
// Should generate: GET /users/{id}
// Context skipped, struct param detected
```

### 3. Type Alias Context + Struct Param
```go
func (s *Service) CreateUser(ctx *RequestContext, req *CreateUserRequest) (*User, error)
// Should generate: POST /users
// Context skipped (type alias detected!), struct param detected
```

### 4. No Context, Just Struct Param
```go
func (s *Service) DeleteUser(req *GetUserRequest) error
// Should generate: DELETE /users/{id}
// No context, struct param detected
```

## 🚀 Running the Example

```bash
# Build
go build ./cmd/examples/21-type-alias-context

# Run
go run ./cmd/examples/21-type-alias-context/main.go

# Test
curl http://localhost:3001/users          # ListUsers (type alias context)
curl http://localhost:3001/users/1        # GetUser (original context)
curl -X POST http://localhost:3001/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
curl -X DELETE http://localhost:3001/users/1
```

## ✅ Expected Routes

```
GET    /users       -> ListUsers      ✅ Type alias detected as context
GET    /users/{id}  -> GetUser        ✅ Struct param detected
POST   /users       -> CreateUser     ✅ Type alias + struct param
DELETE /users/{id}  -> DeleteUser     ✅ Struct param only
```

## 🔧 Implementation

### Before (Name-Based Detection)
```go
// ❌ BROKEN: Name-based detection
if elemType.Name() == "Context" && elemType.Kind() == reflect.Struct {
    continue
}
// Problem: Doesn't handle type aliases correctly
// Problem: Package path might differ
```

### After (Type Comparison)
```go
// ✅ FIXED: Type comparison
var typeOfContextPtr = reflect.TypeOf((*request.Context)(nil))

if paramType == typeOfContextPtr {
    continue
}
// Works with type aliases: type MyContext = request.Context
// Works across package boundaries
// Exact type matching
```

## 📊 Why Type Comparison is Better

| Aspect | Name-Based | Type Comparison |
|--------|------------|----------------|
| **Type Aliases** | ❌ Fails | ✅ Works |
| **Package Path** | ❌ Must match exact name | ✅ Compares underlying type |
| **Reliability** | ❌ Fragile | ✅ Robust |
| **Performance** | 🟡 String comparison | ✅ Pointer comparison |
| **Maintainability** | ❌ Magic strings | ✅ Type-safe |

## 🎓 Key Takeaways

1. **Always use type comparison** for detecting known types
2. **Type aliases are transparent** - `reflect.TypeOf` sees the same type
3. **Name comparison is fragile** - can break with aliases or similar names
4. **Define reference types once** - reuse across codebase

## 📚 See Also

- `core/router/service_router.go` - Type comparison implementation
- `core/router/helper.go` - Also uses `typeOfContext` for detection
- Example 20 - Struct-based parameters
