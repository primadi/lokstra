# Validator Package

Package `common/validator` provides struct validation based on `validate` tags.

## Overview

The validator package validates Go structs based on struct field tags. It's integrated into Lokstra's request binding system but can also be used standalone for validating any struct.

### Performance

The validator uses **metadata caching** for optimal performance:
- Struct validation metadata is cached after first use
- Thread-safe caching using `sync.Map`
- Minimal allocations: ~80 B/op, 2 allocs/op
- Very fast: ~250 ns/op for cached structs
- Excellent concurrent performance: ~70 ns/op with parallel requests
- Reduces GC pressure significantly

Benchmark results on AMD Ryzen 9 5900HX:
```
BenchmarkValidateStruct_CachedCall-16      4,481,756    252.1 ns/op    80 B/op    2 allocs/op
BenchmarkValidateStruct_Concurrent-16     19,137,138     70.1 ns/op    80 B/op    2 allocs/op
```

This makes it suitable for high-throughput request validation without performance concerns.

## Usage

### Automatic Validation (Recommended)

When using Lokstra's request binding methods, validation happens automatically:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=1,max=120"`
}

func handler(c *lokstra.RequestContext) error {
    var req CreateUserRequest
    
    // Validation happens automatically after binding!
    if err := c.Req.BindBody(&req); err != nil {
        if valErr, ok := err.(*request.ValidationError); ok {
            // Structured field-level errors
            return c.Api.ValidationError("Validation failed", valErr.FieldErrors)
        }
        return c.Api.BadRequest("INVALID_REQUEST", err.Error())
    }
    
    // req is now validated
}
```

### Manual Validation

You can also use the validator directly for any struct:

```go
import "github.com/primadi/lokstra/common/validator"

type Config struct {
    Port     int    `validate:"required,min=1,max=65535"`
    Host     string `validate:"required"`
    LogLevel string `validate:"oneof=debug info warn error"`
}

config := Config{Port: 8080, Host: "localhost", LogLevel: "info"}

// Validate the struct
fieldErrors, err := validator.ValidateStruct(&config)
if err != nil {
    // System error (e.g., nil pointer, not a struct)
    return err
}

if len(fieldErrors) > 0 {
    // Validation failed
    for _, fe := range fieldErrors {
        fmt.Printf("Field '%s': %s\n", fe.Field, fe.Message)
    }
}
```

## Supported Validators

### required
Field must be present and not empty.

```go
type Example struct {
    Name string `validate:"required"`
}
```

- Strings: must not be empty `""`
- Numbers: must not be `0`
- Slices/Maps/Arrays: must not be empty (len > 0)
- Booleans: always valid (any value)

### email
String must be a valid email format.

```go
type Example struct {
    Email string `validate:"required,email"`
}
```

Simple validation: must contain `@` and `.` with valid structure.

### min=N
Minimum value or length.

```go
type Example struct {
    Age      int    `validate:"min=18"`          // minimum value
    Name     string `validate:"min=3"`           // minimum length
    Tags     []string `validate:"min=1"`         // minimum items
}
```

- Numbers: minimum value
- Strings: minimum length
- Slices/Maps/Arrays: minimum number of items

### max=N
Maximum value or length.

```go
type Example struct {
    Age      int    `validate:"max=120"`         // maximum value
    Name     string `validate:"max=50"`          // maximum length
    Tags     []string `validate:"max=10"`        // maximum items
}
```

### gt=N
Greater than (numbers only).

```go
type Example struct {
    Price float64 `validate:"gt=0"`  // must be > 0, not >= 0
}
```

### gte=N
Greater than or equal to (numbers only).

```go
type Example struct {
    Score int `validate:"gte=0"`  // must be >= 0
}
```

### lt=N
Less than (numbers only).

```go
type Example struct {
    Discount float64 `validate:"lt=100"`  // must be < 100
}
```

### lte=N
Less than or equal to (numbers only).

```go
type Example struct {
    Discount float64 `validate:"lte=100"`  // must be <= 100
}
```

### oneof=a b c
String must be one of the specified values (space-separated).

```go
type Example struct {
    Role   string `validate:"oneof=admin user guest"`
    Status string `validate:"oneof=active inactive pending"`
}
```

### omitempty
Only validate if field is not empty. Use with pointer fields for optional validation.

```go
type Example struct {
    Name  string  `validate:"required"`
    Email *string `validate:"omitempty,email"`  // optional, but if present must be valid email
    Phone *string `validate:"omitempty,min=10"` // optional, but if present must be >= 10 chars
}
```

## Combining Validators

You can combine multiple validators with commas:

```go
type User struct {
    Name  string `validate:"required,min=3,max=50"`
    Email string `validate:"required,email"`
    Age   int    `validate:"required,min=18,max=120"`
    Role  string `validate:"required,oneof=admin user guest"`
}
```

## Field Names in Errors

The validator uses the `json` tag for field names in error messages:

```go
type Example struct {
    UserEmail string `json:"email" validate:"required,email"`
}
```

Error message will refer to field as "email" (not "UserEmail").

## Optional Fields

Use pointers for optional fields:

```go
type UpdateRequest struct {
    Name  *string  `json:"name" validate:"omitempty,min=3"`
    Email *string  `json:"email" validate:"omitempty,email"`
    Age   *int     `json:"age" validate:"omitempty,min=18"`
}
```

If the pointer is `nil`, validation is skipped. If it has a value, validation is applied.

## Error Response

When used with Lokstra's API helpers, validation errors return structured responses:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "email",
        "message": "email must be a valid email address"
      },
      {
        "field": "age",
        "message": "age must be at least 18"
      }
    ]
  }
}
```

## Testing

The validator package includes comprehensive tests and benchmarks. Run them with:

```bash
cd common/validator

# Run tests
go test -v

# Run benchmarks
go test -bench=. -benchmem
```

Benchmark targets:
- `BenchmarkValidateStruct_FirstCall` - Initial validation (builds cache)
- `BenchmarkValidateStruct_CachedCall` - Subsequent validations (uses cache)
- `BenchmarkValidateStruct_Invalid` - Validation with errors
- `BenchmarkValidateStruct_Complex` - Complex struct with many fields
- `BenchmarkValidateStruct_Concurrent` - Concurrent validation (realistic scenario)

## Implementation Details

- Uses Go reflection to inspect struct fields
- **Metadata caching**: Validation metadata is built once and cached per struct type
- **Thread-safe**: Uses `sync.Map` for concurrent access
- Validates fields with `validate` tags
- Respects `json` tags for field naming
- Handles pointer fields for optional validation
- Returns `[]api_formatter.FieldError` for structured errors
- Stops at first error per field (fail-fast)
- Skips unexported fields
- **Performance optimized**: Minimal allocations, suitable for high-throughput scenarios

### Cache Behavior

The validator maintains an internal cache (`sync.Map`) of validation metadata:
- First validation of a struct type builds and caches metadata
- Subsequent validations use cached metadata (no reflection overhead)
- Cache is global and shared across all goroutines
- No cache invalidation needed (struct types are immutable)
- Memory overhead is minimal (only metadata, not struct instances)

## Custom Validators

You can register your own validators using `RegisterValidator()`:

### Basic Custom Validator

```go
import (
    "fmt"
    "reflect"
    "regexp"
    "github.com/primadi/lokstra/common/validator"
)

func init() {
    // Register custom UUID validator
    validator.RegisterValidator("uuid", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
        if fieldValue.Kind() != reflect.String {
            return nil
        }
        
        value := fieldValue.String()
        if value == "" {
            return nil // Use 'required' tag for empty check
        }
        
        uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
        if !uuidRegex.MatchString(value) {
            return fmt.Errorf("%s must be a valid UUID", fieldName)
        }
        
        return nil
    })
}
```

### Using Custom Validators

```go
type CreateProductRequest struct {
    ID      string `json:"id" validate:"required,uuid"`
    Name    string `json:"name" validate:"required,min=3"`
}

func handler(rc *request.RequestContext) response.Response {
    var req CreateProductRequest
    if err := rc.BindBody(&req); err != nil {
        if validationErr, ok := err.(*request.ValidationError); ok {
            return response.BadRequestError(validationErr.Errors)
        }
        return response.BadRequestError(err.Error())
    }
    
    // req.ID is now validated as a UUID!
    return response.OK(req)
}
```

### Validator with Parameters

Custom validators can accept parameters via the `ruleValue` parameter:

```go
func init() {
    validator.RegisterValidator("startswith", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
        if fieldValue.Kind() != reflect.String {
            return nil
        }
        
        value := fieldValue.String()
        if value == "" {
            return nil
        }
        
        if !strings.HasPrefix(value, ruleValue) {
            return fmt.Errorf("%s must start with '%s'", fieldName, ruleValue)
        }
        
        return nil
    })
}

type Product struct {
    Code string `json:"code" validate:"required,startswith=PRD-"`
}
```

### More Examples

```go
func init() {
    // URL validator
    validator.RegisterValidator("url", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
        if fieldValue.Kind() != reflect.String {
            return nil
        }
        
        value := fieldValue.String()
        if value == "" {
            return nil
        }
        
        if !regexp.MustCompile(`^https?://`).MatchString(value) {
            return fmt.Errorf("%s must be a valid URL (http:// or https://)", fieldName)
        }
        
        return nil
    })
    
    // Alphanumeric validator
    validator.RegisterValidator("alphanum", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
        if fieldValue.Kind() != reflect.String {
            return nil
        }
        
        value := fieldValue.String()
        if value == "" {
            return nil
        }
        
        if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
            return fmt.Errorf("%s must contain only alphanumeric characters", fieldName)
        }
        
        return nil
    })
}

type User struct {
    Username string `json:"username" validate:"required,alphanum,min=3"`
    Website  string `json:"website" validate:"omitempty,url"`
}
```

### Custom Validator Guidelines

1. **ValidatorFunc Signature**:
   ```go
   type ValidatorFunc func(fieldName string, fieldValue reflect.Value, ruleValue string) error
   ```
   - `fieldName`: JSON field name for error messages
   - `fieldValue`: `reflect.Value` of the field being validated
   - `ruleValue`: Parameter from tag (e.g., "PRD-" from `validate:"startswith=PRD-"`)

2. **Return `nil` for Valid Values**:
   - Return `nil` if validation passes
   - Return `error` with descriptive message if validation fails

3. **Handle Empty Values**:
   - Return `nil` for empty values (use `required` tag separately)
   - Check `fieldValue.Kind()` before type assertions
   - This allows combining with `omitempty` or `required`

4. **Thread Safety**:
   - `RegisterValidator()` is thread-safe (uses `sync.RWMutex`)
   - Can register validators at runtime from multiple goroutines
   - Best practice: Register in `init()` function for package-level validators

5. **Override Built-in Validators**:
   - You can override built-in validators by registering with same name
   - Example: Register custom "email" validator for stricter validation

6. **Performance**:
   - Registered validators have no performance overhead
   - Validators are looked up from map (O(1) operation)
   - Keep validator functions lightweight
   - Avoid allocations where possible

### Testing Custom Validators

See `custom_validator_test.go` for comprehensive examples of:
- Basic custom validators (UUID, URL, alphanumeric)
- Validators with parameters (startswith)
- Overriding built-in validators
- Thread-safe registration
- Integration with request binding

## See Also

- [Request Handling](../../docs/request-handling.md)
- [API Response Patterns](../../docs/response-architecture.md)
- [Learning Example: 04-handlers](../../cmd/learning/01-basics/04-handlers/)
