# Request Parameter Binding Examples

This example demonstrates various approaches to binding request parameters in the Lokstra framework. It showcases manual binding, smart binding, and different techniques for handling dynamic data.

## 🎯 What You'll Learn

- **Manual Binding**: Step-by-step parameter binding with full control
- **Smart Binding**: Automatic binding using function signatures
- **Map Binding**: Using `map[string]any` for dynamic data
- **Hybrid Approach**: Combining structured and dynamic binding
- **Best Practices**: When to use each approach

## 🚀 Quick Start

```bash
# Run the example server
go run main.go

# In another terminal, run tests
go test -v

# Run benchmarks
go test -bench=.
```

The server will start on `http://localhost:8080` with the following endpoints available.

## 📋 Available Endpoints

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Manual Binding - Step by Step Control
```bash
curl "http://localhost:8080/users/user123?page=2&limit=10&tags=web&tags=api&active=true" \
  -H "Authorization: Bearer token123" \
  -H "User-Agent: TestClient/1.0"
```

**Use Case**: When you need full control over the binding process or want to handle errors for each parameter type separately.

### 3. Smart Binding - Automatic Magic
```bash
curl -X POST "http://localhost:8080/users/user456/smart?page=1&limit=5&tags=premium&active=false" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer smarttoken" \
  -H "User-Agent: SmartClient/2.0" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "preferences": {
      "theme": "dark",
      "language": "en",
      "notifications": true
    }
  }'
```

**Use Case**: When you have a well-defined struct and want automatic binding with minimal code.

### 4. BindBodySmart to Map - Dynamic Body Content
```bash
curl -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/json" \
  -d '{
    "dynamic_field_1": "value1",
    "nested_object": {
      "key": "value",
      "number": 42
    },
    "array_field": ["item1", "item2"],
    "boolean_field": true
  }'
```

**Use Case**: When you need to handle dynamic or unknown JSON structure in the request body.

### 5. BindAllSmart to Map - Limitation Demo
```bash
curl -X POST http://localhost:8080/users/user789/all-map \
  -H "Content-Type: application/json" \
  -d '{"name": "Charlie", "age": 35}'
```

**Use Case**: Demonstrates why `BindAllSmart` doesn't work well with `map[string]any` and shows the error.

### 6. Hybrid Approach - Recommended Pattern
```bash
curl -X POST "http://localhost:8080/users/hybrid123/hybrid?page=3&limit=20&tags=vip&tags=beta" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer hybridtoken" \
  -d '{
    "profile": {
      "firstName": "David",
      "lastName": "Smith",
      "avatar": "https://example.com/avatar.jpg"
    },
    "settings": {
      "notifications": {
        "email": true,
        "sms": false,
        "push": true
      },
      "privacy": {
        "showEmail": false,
        "showProfile": true
      }
    },
    "metadata": {
      "source": "api",
      "version": "2.0"
    }
  }'
```

**Use Case**: When you need structured parameters (path, query, headers) AND dynamic body content.

### 7. Complex Query Parameters
```bash
curl "http://localhost:8080/search?q=lokstra&filter=type:web&filter=lang:go&sort=name&page=1&limit=10&opt[format]=json&opt[include]=docs&date=2023-01-01&date=2023-12-31"
```

**Use Case**: Handling complex query parameters with arrays, maps, and multiple values.

## 🏗️ Binding Approaches Explained

### Manual Binding
```go
func manualBindingHandler(ctx *lokstra.Context) error {
    var req UserRequest
    
    if err := ctx.BindPath(&req); err != nil {
        return ctx.ErrorBadRequest("Path binding failed: " + err.Error())
    }
    if err := ctx.BindQuery(&req); err != nil {
        return ctx.ErrorBadRequest("Query binding failed: " + err.Error())
    }
    if err := ctx.BindHeader(&req); err != nil {
        return ctx.ErrorBadRequest("Header binding failed: " + err.Error())
    }
    
    return ctx.Ok(req)
}
```

**Pros:**
- Full control over binding process
- Individual error handling for each parameter type
- Explicit and clear code flow

**Cons:**
- More verbose
- Requires manual error handling

### Smart Binding
```go
func smartBindingHandler(ctx *lokstra.Context, req *UserRequest) error {
    // Request automatically bound by Lokstra!
    return ctx.Ok(req)
}
```

**Pros:**
- Minimal code
- Automatic binding of all parameter types
- Type-safe with struct definitions

**Cons:**
- Less control over binding process
- All-or-nothing error handling
- Requires well-defined structs

### BindBodySmart to Map
```go
func bindBodySmartToMapHandler(ctx *lokstra.Context) error {
    var bodyData map[string]any
    
    if err := ctx.BindBodySmart(&bodyData); err != nil {
        return ctx.ErrorBadRequest("Body binding failed: " + err.Error())
    }
    
    return ctx.Ok(bodyData)
}
```

**Pros:**
- Handles dynamic/unknown JSON structure
- Flexible for varying request formats
- Works with JSON, form data, etc.

**Cons:**
- No compile-time type safety
- Requires runtime type assertions
- More complex data processing

### Hybrid Approach (Recommended)
```go
func hybridBindingHandler(ctx *lokstra.Context) error {
    // Struct for known parameters
    var pathQuery struct {
        ID   string `path:"id"`
        Page int    `query:"page"`
        Auth string `header:"Authorization"`
    }
    
    // Map for dynamic body
    var bodyData map[string]any
    
    ctx.BindPath(&pathQuery)
    ctx.BindQuery(&pathQuery)
    ctx.BindHeader(&pathQuery)
    ctx.BindBodySmart(&bodyData)
    
    return ctx.Ok(map[string]any{
        "structured": pathQuery,
        "dynamic": bodyData,
    })
}
```

**Pros:**
- Best of both worlds
- Type safety for known parameters
- Flexibility for dynamic content
- Clear separation of concerns

**Cons:**
- Slightly more code
- Requires understanding of both approaches

## 🔖 Struct Tags Reference

Lokstra supports the following struct tags for automatic binding:

```go
type RequestExample struct {
    // Path parameters from URL segments
    ID string `path:"id"`
    
    // Query parameters from URL query string
    Page   int      `query:"page"`
    Tags   []string `query:"tags"`     // Multiple values: ?tags=a&tags=b
    Active bool     `query:"active"`   // Converts "true"/"false" strings
    
    // Headers from HTTP headers
    Auth      string `header:"Authorization"`
    UserAgent string `header:"User-Agent"`
    
    // Body from request body (JSON, form, etc.)
    Name  string                 `body:"name"`
    Data  map[string]any `body:"data"`
    Items []Item                 `body:"items"`
}
```

## 🧪 Testing

The example includes comprehensive tests demonstrating:

- **Unit Tests**: All endpoints with various input scenarios
- **Error Cases**: Invalid JSON, missing parameters, type mismatches
- **Edge Cases**: Empty values, special characters, nested data
- **Performance Tests**: Benchmarks comparing different binding approaches

```bash
# Run all tests with verbose output
go test -v

# Run specific test
go test -v -run TestSmartBinding

# Run benchmarks
go test -bench=. -benchmem

# Test with race detection
go test -race -v
```

## 📊 Performance Comparison

Based on the included benchmarks:

1. **Manual Binding**: Fastest for simple cases, more overhead for complex structs
2. **Smart Binding**: Good performance with automatic type conversion
3. **Map Binding**: Flexible but requires runtime type assertions
4. **Hybrid Approach**: Best balance of performance and flexibility

## 🎯 When to Use Each Approach

### Use Manual Binding When:
- You need fine-grained error handling
- Working with legacy code that requires explicit control
- Debugging binding issues
- Performance is critical and you want to optimize each step

### Use Smart Binding When:
- You have well-defined request structures
- You want minimal boilerplate code
- Type safety is important
- You're building CRUD APIs with consistent patterns

### Use BindBodySmart to Map When:
- Handling webhooks with varying payloads
- Building proxy or gateway services
- Working with dynamic configuration
- The request structure is unknown at compile time

### Use Hybrid Approach When:
- You need structured parameters AND dynamic body content
- Building flexible APIs that support multiple client types
- You want both type safety and flexibility
- This is the **recommended approach** for most applications

## 🔧 Form Data Support

The `BindBodySmart` method also works with form data:

```bash
# Test with form data
curl -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=FormUser&email=form@example.com&age=28"
```

## 🚨 Common Pitfalls

1. **BindAllSmart with Maps**: Doesn't work well because maps don't have struct tags
2. **Type Assertions**: When using maps, remember to assert types: `data["age"].(float64)`
3. **JSON Numbers**: JSON unmarshaling converts numbers to `float64`, not `int`
4. **Empty Query Arrays**: Empty query parameters might result in `nil` slices
5. **Header Case**: HTTP headers are case-insensitive, but Go struct tags are case-sensitive

## 📚 Related Documentation

- [Core Concepts - Request Binding](../../docs/core-concepts.md#request-binding)
- [Lokstra Framework Documentation](../../docs/README.md)
- [Middleware Documentation](../../docs/middleware.md)
- [Router Features](../02_router_features/README.md)

## 🤝 Contributing

This example is part of the Lokstra framework examples. Feel free to:

- Report issues or suggest improvements
- Add more test cases
- Contribute additional binding patterns
- Improve documentation clarity

---

**Framework**: [Lokstra](https://github.com/primadi/lokstra)  
**Documentation**: [/docs](../../docs/README.md)  
**License**: See [LICENSE](../../LICENSE) file