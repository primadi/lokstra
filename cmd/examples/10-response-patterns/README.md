# 10 - Response Patterns Example

This example demonstrates the **dual response approach** in Lokstra:
- **`c.Resp`** - Base response (no opinions, full control)
- **`c.Api`** - Opinionated API response (wrapped in ApiResponse structure)

## Key Concepts

### Base Response (`c.Resp`)
- **Direct control** over response structure
- **No wrapping** - returns exactly what you send
- **Flexible** - can return any JSON structure, text, or binary data
- **Lightweight** - minimal overhead

### API Response (`c.Api`)
- **Consistent structure** wrapped in `ApiResponse`
- **Opinionated** - follows API standards automatically
- **Rich metadata** - supports pagination, validation errors, etc.
- **Configurable** - can be customized globally at startup

## Usage Patterns

### 1. Direct JSON Response
```go
// Raw response - returns array directly
func GetUsers(c *request.Context) error {
    return c.Resp.WithStatus(200).Json(users)
}
// Output: [{"id":1,"name":"John"}]

// API response - wrapped in ApiResponse
func GetUsers(c *request.Context) error {
    return c.Api.Ok(users)
}
// Output: {"status":"success","data":[{"id":1,"name":"John"}]}
```

### 2. Custom vs Standard Structure
```go
// Custom structure
func GetHealth(c *request.Context) error {
    health := map[string]any{
        "server": "online",
        "uptime": "24h",
        "custom_field": "value",
    }
    return c.Resp.WithStatus(200).Json(health)
}

// Standard API structure
func GetHealth(c *request.Context) error {
    health := map[string]string{"status": "healthy"}
    return c.Api.Ok(health)
}
```

### 3. Error Handling
```go
// Custom error format
func HandleError(c *request.Context) error {
    return c.Resp.WithStatus(404).Json(map[string]string{
        "error": "Not found",
        "code": "404",
    })
}

// Standard API error format
func HandleError(c *request.Context) error {
    return c.Api.NotFound("Resource not found")
}
```

## Response Format Comparison

### Raw Response Examples

#### Simple Data
```json
[
  {"id": 1, "name": "John Doe"},
  {"id": 2, "name": "Jane Smith"}
]
```

#### Custom Structure
```json
{
  "users": [...],
  "count": 2,
  "timestamp": "2024-10-01T10:00:00Z"
}
```

#### Custom Error
```json
{
  "error": "User not found",
  "code": "USER_NOT_FOUND"
}
```

### API Response Examples

#### Success Response
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe"},
    {"id": 2, "name": "Jane Smith"}
  ]
}
```

#### Success with Message
```json
{
  "status": "success",
  "message": "Users retrieved successfully",
  "data": [...]
}
```

#### Paginated List Response
```json
{
  "status": "success",
  "data": [...],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total": 2,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

#### Structured Error Response
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

#### Validation Error Response
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "fields": [
      {
        "field": "name",
        "code": "REQUIRED",
        "message": "Name is required"
      }
    ]
  }
}
```

## API Helper Methods

### Success Methods
- `c.Api.Ok(data)` - Success response with data
- `c.Api.OkWithMessage(data, message)` - Success with message
- `c.Api.Created(data, message)` - 201 Created response
- `c.Api.OkList(data, meta)` - Paginated list response

### Error Methods
- `c.Api.BadRequest(code, message)` - 400 Bad Request
- `c.Api.Unauthorized(message)` - 401 Unauthorized
- `c.Api.Forbidden(message)` - 403 Forbidden
- `c.Api.NotFound(message)` - 404 Not Found
- `c.Api.InternalError(message)` - 500 Internal Server Error
- `c.Api.ValidationError(message, fields)` - 400 with field errors

## When to Use Which

### Use `c.Resp` when:
- Building **microservices** with custom protocols
- **Integrating** with existing systems with specific formats
- Need **full control** over response structure
- Building **non-REST** APIs (GraphQL, RPC, etc.)
- **Performance critical** endpoints with minimal overhead
- **File downloads**, streaming, or binary responses

### Use `c.Api` when:
- Building **REST APIs** with consistent structure
- Need **frontend-friendly** responses with metadata
- Want **automatic validation** error formatting
- Building **paginated lists** with metadata
- Need **standardized error** responses across team
- Want **API documentation** generation support

## Flexibility Benefits

1. **Migration Path**: Existing code using `c.Resp` continues to work
2. **Team Choice**: Different teams can choose different approaches
3. **Mixed Usage**: Can use both in same application as needed
4. **Global Configuration**: API response structure can be customized
5. **Backward Compatibility**: No breaking changes to existing code

## Running the Example

```bash
cd cmd/examples/10-response-patterns
go run main.go
```

This example showcases both response patterns, demonstrating how developers can choose the approach that best fits their needs while maintaining consistency where desired.