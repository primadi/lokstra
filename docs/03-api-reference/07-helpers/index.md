# Helper Packages

Lokstra provides a comprehensive set of helper utilities in the `common` package to simplify common programming tasks like type conversion, validation, data manipulation, and more.

## Table of Contents

- [Overview](#overview)
- [Available Packages](#available-packages)
- [Quick Start](#quick-start)
- [Package Details](#package-details)
- [Best Practices](#best-practices)

## Overview

The `common` package contains utility functions and types that are used throughout Lokstra and are also available for your application code. These helpers are designed to be:

```
✓ Type-Safe         - Extensive use of generics
✓ Performance       - Optimized with caching where applicable
✓ Well-Tested       - Comprehensive test coverage
✓ Easy to Use       - Simple, intuitive APIs
✓ Documented        - Clear documentation and examples
```

## Available Packages

| Package | Purpose | Key Features |
|---------|---------|--------------|
| **[cast](cast.md)** | Type conversion utilities | ToInt, ToFloat64, ToTime, ToType[T], ToStruct |
| **[utils](utils.md)** | General utilities | Map helpers, slice helpers, hash, parsing |
| **[validator](validator.md)** | Struct validation | Tag-based validation, custom validators, caching |
| **[customtype](customtype.md)** | Custom types | DateTime, Date, Decimal with JSON support |
| **[json](json.md)** | JSON utilities | Parse with error recovery |
| **[response_writer](response-writer.md)** | HTTP response helpers | JSON responses, error handling |

## Quick Start

### Type Conversion

```go
import "github.com/primadi/lokstra/common/cast"

// Convert any type to int
age, err := cast.ToInt(userData["age"])

// Convert any type to time.Time
createdAt, err := cast.ToTime(userData["created_at"])

// Convert any type to struct
var user User
err := cast.ToStruct(userData, &user, false)

// Generic type conversion
count, err := cast.ToType[int](rawValue, false)
```

### Validation

```go
import "github.com/primadi/lokstra/common/validator"

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=100"`
}

// Validate struct
fieldErrors, err := validator.ValidateStruct(&request)
if len(fieldErrors) > 0 {
    // Return validation errors
    return api_formatter.ValidationError(fieldErrors)
}
```

### Map Helpers

```go
import "github.com/primadi/lokstra/common/utils"

// Extract value from map with default
host := utils.GetValueFromMap(config, "host", "localhost")
port := utils.GetValueFromMap(config, "port", 8080)
timeout := utils.GetDurationFromMap(config, "timeout", 30*time.Second)

// Clone map
cloned := utils.CloneMap(original)
```

### Custom Types

```go
import "github.com/primadi/lokstra/common/customtype"

type User struct {
    ID        string              `json:"id"`
    Name      string              `json:"name"`
    BirthDate customtype.Date     `json:"birth_date"`
    CreatedAt customtype.DateTime `json:"created_at"`
    Balance   customtype.Decimal  `json:"balance"`
}

// Custom types handle JSON marshaling/unmarshaling automatically
```

## Package Details

### cast Package

**Purpose:** Safe type conversion with error handling

**Key Functions:**
```go
// Primitive conversions
ToInt(val any) (int, error)
ToFloat64(val any) (float64, error)
ToTime(val any) (time.Time, error)

// Generic conversion
ToType[T any](val any, strict bool) (T, error)

// Struct conversion
ToStruct(source any, dest any, strict bool) error

// Slice conversion
SliceConvert[T any](slice any) (T, error)

// Check emptiness
IsEmpty(val any) bool
```

**Use Cases:**
- Converting request parameters to proper types
- Unmarshaling configuration values
- Converting database results to structs

[Full documentation →](cast.md)

### utils Package

**Purpose:** General utility functions for common tasks

**Key Functions:**
```go
// Map operations
GetValueFromMap[T any](map[string]any, string, T) T
GetDurationFromMap(map[string]any, string, any) time.Duration
CloneMap[K comparable, V any](map[K]V) map[K]V

// Slice operations
ToAnySlice[T any]([]T) []any
SlicesConcat[T any](...[]T) []T
AppendSorted[T any]([]T, T, func(a, b T) bool) []T

// String operations
CamelToSnake(string) string
ParseClientIP(*http.Request) string

// Security
HashPassword(string) (string, error)

// Type checking
IsNil(any) bool
```

[Full documentation →](utils.md)

### validator Package

**Purpose:** Struct validation with tag-based rules

**Key Features:**
- Tag-based validation rules
- Custom validator registration
- Performance optimization with caching
- Built-in validators (required, email, min, max, gt, gte, lt, lte, oneof)

**Built-in Validators:**
```go
`validate:"required"`             // Field must not be empty
`validate:"email"`                // Must be valid email
`validate:"min=5"`                // Min length/value
`validate:"max=100"`              // Max length/value
`validate:"gt=0"`                 // Greater than
`validate:"gte=18"`               // Greater than or equal
`validate:"lt=100"`               // Less than
`validate:"lte=200"`              // Less than or equal
`validate:"oneof=admin user"`     // Must be one of values
```

[Full documentation →](validator.md)

### customtype Package

**Purpose:** Custom types with special marshaling/unmarshaling

**Available Types:**
```go
customtype.DateTime    // ISO 8601 datetime with timezone
customtype.Date        // Date only (YYYY-MM-DD)
customtype.Decimal     // High-precision decimal numbers
```

**Features:**
- Automatic JSON marshaling/unmarshaling
- Database scanning support
- Null value handling
- Validation support

[Full documentation →](customtype.md)

### json Package

**Purpose:** Enhanced JSON operations with error recovery

**Key Features:**
- Parse JSON with better error messages
- Error recovery for malformed JSON
- Pretty printing

[Full documentation →](json.md)

### response_writer Package

**Purpose:** HTTP response helpers

**Key Features:**
- JSON response writing
- Error response formatting
- Content-Type handling
- Status code management

[Full documentation →](response-writer.md)

## Best Practices

### Type Conversion

```go
✓ DO: Always check errors from conversion functions
age, err := cast.ToInt(value)
if err != nil {
    return fmt.Errorf("invalid age: %w", err)
}

✓ DO: Use strict mode when data integrity is critical
err := cast.ToStruct(data, &user, true)  // Fails on unknown fields

✗ DON'T: Ignore conversion errors
age, _ := cast.ToInt(value)  // BAD: Silent failure

✗ DON'T: Use type assertion without checking
age := value.(int)  // BAD: Panics if wrong type
```

### Validation

```go
✓ DO: Validate at entry points (API handlers, etc.)
fieldErrors, err := validator.ValidateStruct(&request)
if len(fieldErrors) > 0 {
    return ValidationError(fieldErrors)
}

✓ DO: Use descriptive field names with json tags
type User struct {
    Email string `json:"email" validate:"required,email"`
}

✗ DON'T: Skip validation for untrusted input
// BAD: No validation
user := &User{Email: request.Email}

✗ DON'T: Use validation as business logic
// BAD: Validation should be format checking only
`validate:"required,min=18"`  // OK: Format validation
// Business rule checks should be in service layer
```

### Map Operations

```go
✓ DO: Use GetValueFromMap with appropriate defaults
port := utils.GetValueFromMap(config, "port", 8080)
timeout := utils.GetDurationFromMap(config, "timeout", 30*time.Second)

✓ DO: Clone maps when passing to untrusted code
safeCopy := utils.CloneMap(sensitiveData)
process(safeCopy)

✗ DON'T: Access maps directly without checking existence
port := config["port"].(int)  // BAD: Panics if missing or wrong type

✗ DON'T: Modify original maps unintentionally
// BAD: Modifies original
process(originalMap)
```

### Custom Types

```go
✓ DO: Use custom types for domain-specific values
type User struct {
    CreatedAt customtype.DateTime `json:"created_at"`
    BirthDate customtype.Date     `json:"birth_date"`
    Balance   customtype.Decimal  `json:"balance"`
}

✓ DO: Let custom types handle marshaling automatically
json.Marshal(user)  // DateTime/Date/Decimal handled correctly

✗ DON'T: Use string for dates/times in structs
type User struct {
    CreatedAt string `json:"created_at"`  // BAD: Use customtype.DateTime
}

✗ DON'T: Use float64 for money
type Product struct {
    Price float64 `json:"price"`  // BAD: Use customtype.Decimal
}
```

### Performance

```go
✓ DO: Reuse validator instances (cached automatically)
validator.ValidateStruct(&user)  // Metadata cached after first call

✓ DO: Use CloneMap instead of manual copying
cloned := utils.CloneMap(original)  // Efficient

✓ DO: Use AppendSorted for small slices
sorted := utils.AppendSorted(slice, value, less)

✗ DON'T: Build validation metadata repeatedly
// Validator automatically caches, but don't create new validators

✗ DON'T: Use inefficient algorithms for large slices
for i, item := range large_slice {
    // BAD: O(n²) for insertion sort
}
```

## Import Paths

```go
import (
    "github.com/primadi/lokstra/common/cast"
    "github.com/primadi/lokstra/common/customtype"
    "github.com/primadi/lokstra/common/json"
    "github.com/primadi/lokstra/common/response_writer"
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/common/validator"
)
```

## Related Documentation

- [Core Packages](../01-core-packages/README.md) - Core framework components
- [Services](../06-services/README.md) - Built-in services
- [Configuration](../03-configuration/README.md) - Configuration system

---

**Next:** Explore individual helper packages for detailed documentation and examples.
