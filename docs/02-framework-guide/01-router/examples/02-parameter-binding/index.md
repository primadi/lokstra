# Parameter Binding Deep Dive

> **Master parameter extraction from paths, queries, headers, and bodies**

This example demonstrates all parameter binding capabilities in Lokstra.

## Parameter Types

### 1. Path Parameters
```go
type Params struct {
    ID       int    `path:"id"`
    Category string `path:"category"`
}
```

**Route**: `/path/:id/:category`  
**Usage**: Extract values from URL path

---

### 2. Query Parameters
```go
type Params struct {
    Page  int    `query:"page"`
    Limit int    `query:"limit"`
    Sort  string `query:"sort"`
}
```

**Route**: `/items?page=1&limit=10&sort=name`  
**Usage**: Optional filters, pagination, sorting

---

### 3. Header Parameters
```go
// Use ctx.Req.HeaderParam()
apiKey := ctx.Req.HeaderParam("X-API-Key", "default")
```

**Usage**: Authentication, API keys, custom headers

---

### 4. Body Binding
```go
type CreateRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func Create(ctx *request.Context, body CreateRequest) (Response, error)
```

**Usage**: POST/PUT/PATCH requests with JSON body

---

### 5. Partial Updates (Pointer Fields)
```go
type UpdateRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}
```

**Usage**: PATCH requests where only provided fields are updated

---

## Advanced Patterns

### Array Parameters
```go
type FilterParams struct {
    IDs  []int    `query:"ids"`   // ?ids=1,2,3
    Tags []string `query:"tags"`  // ?tags=go,rust
}
```

### Date Range
```go
type DateParams struct {
    StartDate string `query:"start_date"` // 2025-01-01
    EndDate   string `query:"end_date"`
    TimeZone  string `query:"timezone"`
}
```

### Combined Parameters
```go
func Handler(
    ctx *request.Context,
    pathParams PathParams,    // :id, :category
    body CreateRequest,       // JSON body
) (Response, error) {
    // Access all: path, query, headers, body
}
```

---

## Validation

Lokstra supports struct validation tags:

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"required,gte=18,lte=100"`
    Score    int    `json:"score" validate:"gt=0,lt=100"`
    Status   string `json:"status" validate:"required,oneof=active inactive pending"`
}

type UpdateUserRequest struct {
    Email *string `json:"email,omitempty" validate:"email"`
    Age   *int    `json:"age,omitempty" validate:"gte=18,lte=100"`
}
```

### Built-in Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must be present and non-zero | `validate:"required"` |
| `email` | Valid email format | `validate:"email"` |
| `min=N` | Minimum value (numbers) or length (strings) | `validate:"min=8"` |
| `max=N` | Maximum value (numbers) or length (strings) | `validate:"max=100"` |
| `gt=N` | Greater than (exclusive) | `validate:"gt=0"` |
| `gte=N` | Greater than or equal to | `validate:"gte=18"` |
| `lt=N` | Less than (exclusive) | `validate:"lt=100"` |
| `lte=N` | Less than or equal to | `validate:"lte=100"` |
| `oneof` | Value must be one of the specified options | `validate:"oneof=active inactive"` |

**Combining tags**: Use commas to combine multiple validations
```go
Age int `json:"age" validate:"required,gte=18,lte=100"`
```

### Custom Validators

You can register custom validation functions:

```go
import "github.com/primadi/lokstra/common/validator"

func init() {
    // Register custom validator
    validator.RegisterValidator("username", validateUsername)
}

func validateUsername(value any, param string) error {
    username, ok := value.(string)
    if !ok {
        return fmt.Errorf("username must be a string")
    }
    
    if len(username) < 3 {
        return fmt.Errorf("username must be at least 3 characters")
    }
    
    if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username) {
        return fmt.Errorf("username can only contain letters, numbers, and underscores")
    }
    
    return nil
}

// Use in struct
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,username"`
}
```

### Validation Error Response

When validation fails, Lokstra automatically returns a structured error:

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Email format is invalid"
      },
      {
        "field": "age",
        "code": "MIN_VALUE",
        "message": "Age must be at least 18"
      }
    ]
  }
}
```

---

## Best Practices

### ✅ Do
```go
// Use pointers for optional updates
type UpdateRequest struct {
    Name *string `json:"name,omitempty"`
}

// Set defaults in handler
if params.Page == 0 {
    params.Page = 1
}

// Validate business logic
if params.ID < 1 {
    return nil, ctx.Api.BadRequest("Invalid ID")
}
```

### ❌ Don't
```go
// Don't use pointers for required fields
type CreateRequest struct {
    Name *string `json:"name"` // Should be string, not *string
}

// Don't mix concerns
type Params struct {
    ID int `path:"id"`
    // Don't put header params in struct - use ctx.Req.HeaderParam()
}
```

---

## Parameter Binding Order

1. **Path parameters** → Extracted from URL path
2. **Query parameters** → Extracted from query string
3. **Body** → Parsed from JSON body (POST/PUT/PATCH)
4. **Validation** → Runs on struct tags
5. **Handler** → Called with validated parameters

---

## Running

```bash
go run main.go

# Test with test.http file
```

---

## Key Takeaways

✅ Path params: `path:"id"`  
✅ Query params: `query:"page"`  
✅ Headers: `ctx.Req.HeaderParam()`  
✅ Body: JSON binding with validation  
✅ Partial updates: Use pointer fields with `omitempty`  
✅ Arrays: Comma-separated values automatically parsed
