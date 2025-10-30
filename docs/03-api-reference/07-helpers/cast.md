# Cast Package

The `cast` package provides type-safe conversion utilities with error handling for converting between Go types, especially useful when working with `any` (interface{}) values from JSON, maps, or dynamic data sources.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Basic Conversions](#basic-conversions)
- [Generic Conversion](#generic-conversion)
- [Struct Conversion](#struct-conversion)
- [Slice Conversion](#slice-conversion)
- [Utility Functions](#utility-functions)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/cast`

**Key Features:**

```
✓ Type-Safe Conversion    - Convert with compile-time safety
✓ Error Handling          - Explicit error returns
✓ Struct Mapping          - Map to struct with caching
✓ Generic Support         - ToType[T] for any type
✓ Nested Structures       - Handle complex nested data
✓ Performance Optimized   - Field mapping cache
```

## Installation

```go
import "github.com/primadi/lokstra/common/cast"
```

## Basic Conversions

### ToInt

Convert any value to `int`:

```go
// From various numeric types
age, err := cast.ToInt(42)           // int: 42
age, err = cast.ToInt(int64(100))    // int: 100
age, err = cast.ToInt(uint8(25))     // int: 25

// From nil (returns 0)
age, err = cast.ToInt(nil)           // int: 0, nil

// From incompatible type (returns error)
age, err = cast.ToInt("not a number") // int: 0, error
```

**Supported Types:**
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `nil` → `0`

### ToFloat64

Convert any value to `float64`:

```go
// From various numeric types
price, err := cast.ToFloat64(99.99)      // float64: 99.99
price, err = cast.ToFloat64(100)         // float64: 100.0
price, err = cast.ToFloat64(float32(50)) // float64: 50.0

// From nil (returns 0)
price, err = cast.ToFloat64(nil)         // float64: 0.0, nil
```

**Supported Types:**
- `float32`, `float64`
- `int`, `int8`, `int16`, `int32`, `int64`
- `nil` → `0.0`

### ToTime

Convert any value to `time.Time`:

```go
// From string (multiple formats supported)
created, err := cast.ToTime("2024-01-15 14:30:00")  // time.DateTime
created, err = cast.ToTime("2024-01-15")            // time.DateOnly
created, err = cast.ToTime("14:30:00")              // time.TimeOnly
created, err = cast.ToTime("2024-01-15T14:30:00Z")  // time.RFC3339

// From Unix timestamp
created, err = cast.ToTime(int64(1705328400))       // Unix seconds
created, err = cast.ToTime(float64(1705328400))     // Unix seconds

// From time.Time (no conversion)
created, err = cast.ToTime(time.Now())              // Pass through

// From nil (returns zero time)
created, err = cast.ToTime(nil)                     // time.Time{}, nil
```

**Supported Formats:**
- `time.DateTime` - "2006-01-02 15:04:05"
- `time.DateOnly` - "2006-01-02"
- `time.TimeOnly` - "15:04:05"
- `time.RFC3339` - "2006-01-02T15:04:05Z07:00"
- `time.RFC3339Nano` - "2006-01-02T15:04:05.999999999Z07:00"
- Unix timestamp (`int64`, `float64`)

## Generic Conversion

### ToType[T]

Convert any value to a specific type using generics:

```go
// Basic types
age, err := cast.ToType[int](userData["age"], false)
price, err := cast.ToType[float64](productData["price"], false)
created, err := cast.ToType[time.Time](data["created_at"], false)

// Structs
user, err := cast.ToType[User](userData, false)
userPtr, err := cast.ToType[*User](userData, false)

// Slices
users, err := cast.ToType[[]User](usersData, false)
```

**Parameters:**
- `val any` - Value to convert
- `strict bool` - If `true`, fails on unknown fields in structs

**Strict Mode:**

```go
// Non-strict (default) - ignores unknown fields
user, err := cast.ToType[User](data, false)  // OK, extra fields ignored

// Strict - fails on unknown fields
user, err := cast.ToType[User](data, true)   // Error if data has extra fields
```

## Struct Conversion

### ToStruct

Convert `map[string]any` to struct:

```go
type User struct {
    ID       int       `json:"id"`
    Username string    `json:"username"`
    Email    string    `json:"email"`
    Age      int       `json:"age"`
    Active   bool      `json:"active"`
    Created  time.Time `json:"created_at"`
}

data := map[string]any{
    "id":         123,
    "username":   "john_doe",
    "email":      "john@example.com",
    "age":        30,
    "active":     true,
    "created_at": "2024-01-15T10:30:00Z",
}

var user User
err := cast.ToStruct(data, &user, false)
if err != nil {
    log.Fatal(err)
}

// user.ID = 123
// user.Username = "john_doe"
// user.Email = "john@example.com"
// user.Age = 30
// user.Active = true
// user.Created = parsed time
```

### Nested Structs

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    Address Address `json:"address"`
}

data := map[string]any{
    "id":   123,
    "name": "John Doe",
    "address": map[string]any{
        "street":  "123 Main St",
        "city":    "New York",
        "country": "USA",
    },
}

var user User
err := cast.ToStruct(data, &user, false)

// user.Address.Street = "123 Main St"
// user.Address.City = "New York"
// user.Address.Country = "USA"
```

### Pointer Fields

```go
type User struct {
    ID      int      `json:"id"`
    Name    string   `json:"name"`
    Address *Address `json:"address,omitempty"`
}

// With address
data := map[string]any{
    "id":   123,
    "name": "John",
    "address": map[string]any{
        "city": "NYC",
    },
}
var user1 User
cast.ToStruct(data, &user1, false)
// user1.Address != nil

// Without address
data = map[string]any{
    "id":   123,
    "name": "John",
}
var user2 User
cast.ToStruct(data, &user2, false)
// user2.Address == nil
```

### Slice Fields

```go
type User struct {
    ID    int      `json:"id"`
    Name  string   `json:"name"`
    Tags  []string `json:"tags"`
    Roles []Role   `json:"roles"`
}

type Role struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

data := map[string]any{
    "id":   123,
    "name": "John",
    "tags": []any{"admin", "verified", "premium"},
    "roles": []any{
        map[string]any{"id": 1, "name": "admin"},
        map[string]any{"id": 2, "name": "user"},
    },
}

var user User
err := cast.ToStruct(data, &user, false)

// user.Tags = ["admin", "verified", "premium"]
// user.Roles = [{ID: 1, Name: "admin"}, {ID: 2, Name: "user"}]
```

### Field Mapping

The package uses JSON tags for field mapping and caches field information for performance:

```go
type User struct {
    ID       int    `json:"id"`           // Maps to "id"
    Username string `json:"username"`     // Maps to "username"
    FullName string `json:"full_name"`    // Maps to "full_name"
    Internal string                       // Maps to "Internal" (field name)
    Ignored  string `json:"-"`            // Ignored
}
```

**Caching:**
- Field mappings are cached per struct type
- First call builds the cache
- Subsequent calls use cached mappings
- Thread-safe with sync.Map

## Slice Conversion

### SliceConvert

Convert slice of one type to slice of another:

```go
// Convert []any to []string
source := []any{"one", "two", "three"}
result, err := cast.SliceConvert[[]string](source)
// result = []string{"one", "two", "three"}

// Convert []any to []int
source := []any{1, 2, 3}
result, err := cast.SliceConvert[[]int](source)
// result = []int{1, 2, 3}

// Type mismatch (error)
source := []any{"one", 2, "three"}
result, err := cast.SliceConvert[[]string](source)
// err: cannot assign value of type int to string at index 1
```

**Requirements:**
- Input must be a slice
- Output type parameter must be a slice
- Each element must be assignable to output element type

## Utility Functions

### IsEmpty

Check if a value is considered empty:

```go
// Strings
cast.IsEmpty("")           // true
cast.IsEmpty("hello")      // false

// Numbers
cast.IsEmpty(0)            // true
cast.IsEmpty(42)           // false
cast.IsEmpty(0.0)          // true

// Booleans
cast.IsEmpty(false)        // true
cast.IsEmpty(true)         // false

// Slices
cast.IsEmpty([]int{})      // true
cast.IsEmpty([]int{1, 2})  // false

// Maps
cast.IsEmpty(map[string]int{})           // true
cast.IsEmpty(map[string]int{"a": 1})     // false

// Pointers
var ptr *int
cast.IsEmpty(ptr)          // true
ptr = new(int)
cast.IsEmpty(ptr)          // false

// Nil
cast.IsEmpty(nil)          // true

// Structs (all fields empty)
type User struct {
    Name string
    Age  int
}
cast.IsEmpty(User{})       // true
cast.IsEmpty(User{Name: "John"})  // false
```

## Best Practices

### Error Handling

```go
✓ DO: Always check errors from conversion functions
age, err := cast.ToInt(value)
if err != nil {
    return fmt.Errorf("invalid age: %w", err)
}

✗ DON'T: Ignore errors
age, _ := cast.ToInt(value)  // BAD: Silent failure
```

### Struct Conversion

```go
✓ DO: Use strict mode for API validation
err := cast.ToStruct(requestData, &request, true)
if err != nil {
    return BadRequest("Invalid request structure")
}

✓ DO: Use non-strict mode for flexible data sources
err := cast.ToStruct(configData, &config, false)

✗ DON'T: Use ToStruct without pointer
err := cast.ToStruct(data, user, false)  // BAD: Pass &user instead
```

### Type Safety

```go
✓ DO: Use ToType[T] for type-safe conversions
age, err := cast.ToType[int](value, false)

✗ DON'T: Use type assertions directly
age := value.(int)  // BAD: Panics on wrong type
```

### Performance

```go
✓ DO: Reuse struct types to benefit from caching
var user User
cast.ToStruct(data1, &user, false)
cast.ToStruct(data2, &user, false)  // Uses cached field mapping

✓ DO: Use appropriate conversion functions
age, err := cast.ToInt(value)  // More efficient than ToType[int]

✗ DON'T: Build structs in loops without consideration
for _, data := range hugeDataset {
    cast.ToStruct(data, &result, false)  // OK: Cache makes this efficient
}
```

### Nil Handling

```go
✓ DO: Handle nil values appropriately
value, err := cast.ToInt(nil)  // Returns 0, nil

✓ DO: Check for zero values after conversion
if value == 0 && originalValue == nil {
    // Handle nil case
}

✗ DON'T: Assume non-nil returns
value, _ := cast.ToInt(possiblyNil)  // Could be 0 from nil or actual 0
```

## Examples

### HTTP Request Parsing

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]any
    json.NewDecoder(r.Body).Decode(&requestData)
    
    type CreateUserRequest struct {
        Username string `json:"username" validate:"required"`
        Email    string `json:"email" validate:"required,email"`
        Age      int    `json:"age" validate:"gte=18"`
    }
    
    var req CreateUserRequest
    err := cast.ToStruct(requestData, &req, true)
    if err != nil {
        http.Error(w, "Invalid request structure", http.StatusBadRequest)
        return
    }
    
    // Validate
    fieldErrors, err := validator.ValidateStruct(&req)
    if len(fieldErrors) > 0 {
        // Return validation errors
        return
    }
    
    // Process request
    user := createUser(req)
    json.NewEncoder(w).Encode(user)
}
```

### Configuration Parsing

```go
func LoadConfig(configData map[string]any) (*Config, error) {
    type DatabaseConfig struct {
        Host     string `json:"host"`
        Port     int    `json:"port"`
        Database string `json:"database"`
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    type Config struct {
        AppName  string         `json:"app_name"`
        Port     int            `json:"port"`
        Database DatabaseConfig `json:"database"`
    }
    
    var config Config
    err := cast.ToStruct(configData, &config, false)
    if err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &config, nil
}
```

### Database Result Mapping

```go
func GetUsers(ctx context.Context) ([]User, error) {
    // Get rows from database as []map[string]any
    rows, err := conn.SelectManyRowMap(ctx, "SELECT * FROM users")
    if err != nil {
        return nil, err
    }
    
    users := make([]User, 0, len(rows))
    for _, row := range rows {
        var user User
        if err := cast.ToStruct(row, &user, false); err != nil {
            return nil, fmt.Errorf("failed to map user: %w", err)
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

### Dynamic Type Conversion

```go
func ProcessField(fieldType string, value any) (any, error) {
    switch fieldType {
    case "int":
        return cast.ToType[int](value, false)
    case "float":
        return cast.ToType[float64](value, false)
    case "time":
        return cast.ToType[time.Time](value, false)
    case "string":
        return fmt.Sprintf("%v", value), nil
    default:
        return nil, fmt.Errorf("unsupported type: %s", fieldType)
    }
}
```

### Service Factory Pattern

```go
func ServiceFactory(params map[string]any) (Service, error) {
    type ServiceConfig struct {
        Host    string        `json:"host"`
        Port    int           `json:"port"`
        Timeout time.Duration `json:"timeout"`
        Options map[string]any `json:"options"`
    }
    
    var config ServiceConfig
    err := cast.ToStruct(params, &config, false)
    if err != nil {
        return nil, fmt.Errorf("invalid service config: %w", err)
    }
    
    return NewService(&config), nil
}
```

### Form Data Processing

```go
func ProcessForm(formData map[string][]string) (*FormData, error) {
    // Convert form values (first value of each field)
    data := make(map[string]any)
    for key, values := range formData {
        if len(values) > 0 {
            data[key] = values[0]
        }
    }
    
    type FormData struct {
        Name    string `json:"name"`
        Email   string `json:"email"`
        Age     int    `json:"age"`
        Message string `json:"message"`
    }
    
    var form FormData
    err := cast.ToStruct(data, &form, false)
    if err != nil {
        return nil, fmt.Errorf("invalid form data: %w", err)
    }
    
    return &form, nil
}
```

## Related Documentation

- [Helpers Overview](index) - All helper packages
- [Utils Package](utils) - General utilities
- [Validator Package](validator) - Struct validation
- [Custom Types](customtype) - Custom type implementations

---

**Next:** [Utils Package](utils) - General utility functions
