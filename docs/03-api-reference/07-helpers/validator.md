# Validator Package

The `validator` package provides struct validation using struct tags with built-in and custom validators. It includes caching for high-performance validation.

## Table of Contents

- [Overview](#overview)
- [Basic Usage](#basic-usage)
- [Built-in Validators](#built-in-validators)
- [Custom Validators](#custom-validators)
- [Field Names](#field-names)
- [Pointer Fields](#pointer-fields)
- [Error Handling](#error-handling)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/validator`

**Key Features:**

```
✓ Tag-Based Validation   - Simple declarative syntax
✓ Built-in Validators    - required, email, min, max, gt, gte, lt, lte, oneof
✓ Custom Validators      - Register your own validators
✓ Performance Caching    - Metadata cached per type
✓ JSON Field Names       - Error messages use json tags
✓ Pointer Support        - Handles optional fields correctly
```

## Basic Usage

### Simple Validation

```go
import "github.com/primadi/lokstra/common/validator"

type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18"`
}

func CreateUser(data User) error {
    // Validate struct
    fieldErrors, err := validator.ValidateStruct(data)
    if err != nil {
        // System error (e.g., invalid input type)
        return err
    }
    
    if len(fieldErrors) > 0 {
        // Validation failed
        for _, fe := range fieldErrors {
            fmt.Printf("Field: %s, Error: %s\n", fe.Field, fe.Message)
        }
        return errors.New("validation failed")
    }
    
    // Validation passed
    return nil
}
```

### Validation in HTTP Handler

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Validate
    fieldErrors, err := validator.ValidateStruct(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if len(fieldErrors) > 0 {
        // Return validation errors
        response := map[string]any{
            "status": "error",
            "errors": fieldErrors,
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // Process valid user
    // ...
}
```

## Built-in Validators

### required

Ensures field is not empty:

```go
type Product struct {
    Name  string   `validate:"required"`           // Non-empty string
    Price float64  `validate:"required"`           // Non-zero number
    Tags  []string `validate:"required"`           // Non-empty slice
}

// Valid
product := Product{Name: "Book", Price: 9.99, Tags: []string{"new"}}

// Invalid - Name is empty
product := Product{Name: "", Price: 9.99, Tags: []string{"new"}}
// Error: "Name is required"

// Invalid - Price is zero
product := Product{Name: "Book", Price: 0, Tags: []string{"new"}}
// Error: "Price is required"

// Invalid - Tags is empty
product := Product{Name: "Book", Price: 9.99, Tags: []string{}}
// Error: "Tags is required"
```

**Empty Values by Type:**
- String: `""`
- Numbers: `0`
- Slices/Maps/Arrays: Empty/nil
- Bool: Always valid (can't be empty)

### email

Validates email address format:

```go
type Contact struct {
    Email string `json:"email" validate:"required,email"`
}

// Valid
contact := Contact{Email: "user@example.com"}

// Invalid
contact := Contact{Email: "invalid-email"}
// Error: "email must be a valid email address"

contact := Contact{Email: "@example.com"}
// Error: "email must be a valid email address"

contact := Contact{Email: "user@"}
// Error: "email must be a valid email address"
```

**Validation Rules:**
- Must contain `@`
- Must contain `.`
- Must have text before and after `@`

### min

Minimum value/length constraint:

```go
type Account struct {
    Username string   `json:"username" validate:"required,min=3"`     // Min 3 chars
    Age      int      `json:"age" validate:"min=18"`                  // Min value 18
    Balance  float64  `json:"balance" validate:"min=0"`               // Min 0.0
    Tags     []string `json:"tags" validate:"min=1"`                  // Min 1 item
}

// Valid
account := Account{
    Username: "john",
    Age:      25,
    Balance:  100.50,
    Tags:     []string{"vip", "active"},
}

// Invalid - Username too short
account := Account{Username: "jo"}
// Error: "username must be at least 3 characters"

// Invalid - Age too low
account := Account{Age: 15}
// Error: "age must be at least 18"

// Invalid - Balance negative
account := Account{Balance: -10.0}
// Error: "balance must be at least 0.00"

// Invalid - Tags empty
account := Account{Tags: []string{}}
// Error: "tags must have at least 1 items"
```

**Applies to:**
- String: Character length
- Numbers: Value
- Slices/Maps/Arrays: Item count

### max

Maximum value/length constraint:

```go
type Post struct {
    Title   string   `json:"title" validate:"required,max=100"`   // Max 100 chars
    Rating  int      `json:"rating" validate:"max=5"`             // Max value 5
    Price   float64  `json:"price" validate:"max=999.99"`         // Max 999.99
    Tags    []string `json:"tags" validate:"max=10"`              // Max 10 items
}

// Valid
post := Post{
    Title:  "Short Title",
    Rating: 5,
    Price:  99.99,
    Tags:   []string{"go", "web"},
}

// Invalid - Title too long
post := Post{Title: strings.Repeat("a", 101)}
// Error: "title must be at most 100 characters"

// Invalid - Rating too high
post := Post{Rating: 6}
// Error: "rating must be at most 5"

// Invalid - Too many tags
post := Post{Tags: make([]string, 11)}
// Error: "tags must have at most 10 items"
```

### gt (Greater Than)

Value must be strictly greater than specified value:

```go
type Order struct {
    Quantity int     `json:"quantity" validate:"gt=0"`     // > 0
    Total    float64 `json:"total" validate:"gt=0"`        // > 0.0
}

// Valid
order := Order{Quantity: 1, Total: 9.99}

// Invalid - Quantity is 0
order := Order{Quantity: 0}
// Error: "quantity must be greater than 0"

// Invalid - Quantity negative
order := Order{Quantity: -1}
// Error: "quantity must be greater than 0"

// Invalid - Total is 0
order := Order{Total: 0.0}
// Error: "total must be greater than 0.00"
```

### gte (Greater Than or Equal)

Value must be greater than or equal to specified value:

```go
type Rating struct {
    Score int `json:"score" validate:"gte=0,lte=100"`  // 0-100
}

// Valid
rating := Rating{Score: 0}   // Exactly 0 is valid
rating := Rating{Score: 50}
rating := Rating{Score: 100} // Exactly 100 is valid

// Invalid
rating := Rating{Score: -1}
// Error: "score must be greater than or equal to 0"

rating := Rating{Score: 101}
// Error: "score must be less than or equal to 100"
```

### lt (Less Than)

Value must be strictly less than specified value:

```go
type Temperature struct {
    Value float64 `json:"value" validate:"lt=100"`  // < 100
}

// Valid
temp := Temperature{Value: 99.9}

// Invalid - Value is 100
temp := Temperature{Value: 100.0}
// Error: "value must be less than 100.00"

// Invalid - Value greater than 100
temp := Temperature{Value: 150.0}
// Error: "value must be less than 100.00"
```

### lte (Less Than or Equal)

Value must be less than or equal to specified value:

```go
type Discount struct {
    Percentage int `json:"percentage" validate:"gte=0,lte=100"`  // 0-100%
}

// Valid
discount := Discount{Percentage: 0}    // Exactly 0 is valid
discount := Discount{Percentage: 50}
discount := Discount{Percentage: 100}  // Exactly 100 is valid

// Invalid
discount := Discount{Percentage: 101}
// Error: "percentage must be less than or equal to 100"
```

### oneof

Value must be one of specified options:

```go
type Status struct {
    Value string `json:"status" validate:"required,oneof=pending active inactive"`
}

// Valid
status := Status{Value: "pending"}
status := Status{Value: "active"}
status := Status{Value: "inactive"}

// Invalid
status := Status{Value: "unknown"}
// Error: "status must be one of: pending, active, inactive"

status := Status{Value: "Pending"}  // Case-sensitive
// Error: "status must be one of: pending, active, inactive"
```

## Custom Validators

### Register Custom Validator

```go
import (
    "fmt"
    "reflect"
    "strings"
    
    "github.com/primadi/lokstra/common/validator"
)

func init() {
    // Register UUID validator
    validator.RegisterValidator("uuid", validateUUID)
    
    // Register URL validator
    validator.RegisterValidator("url", validateURL)
    
    // Register phone validator
    validator.RegisterValidator("phone", validatePhone)
}

func validateUUID(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil
    }
    
    uuid := fieldValue.String()
    if uuid == "" {
        return nil // Use required tag for empty check
    }
    
    // Simple UUID format check (8-4-4-4-12)
    parts := strings.Split(uuid, "-")
    if len(parts) != 5 {
        return fmt.Errorf("%s must be a valid UUID", fieldName)
    }
    
    if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 ||
        len(parts[3]) != 4 || len(parts[4]) != 12 {
        return fmt.Errorf("%s must be a valid UUID", fieldName)
    }
    
    return nil
}

func validateURL(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil
    }
    
    url := fieldValue.String()
    if url == "" {
        return nil
    }
    
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        return fmt.Errorf("%s must be a valid URL (http or https)", fieldName)
    }
    
    return nil
}

func validatePhone(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil
    }
    
    phone := fieldValue.String()
    if phone == "" {
        return nil
    }
    
    // Simple check: only digits and common separators
    for _, r := range phone {
        if r != '+' && r != '-' && r != ' ' && r != '(' && r != ')' && (r < '0' || r > '9') {
            return fmt.Errorf("%s must be a valid phone number", fieldName)
        }
    }
    
    return nil
}
```

### Use Custom Validators

```go
type Resource struct {
    ID       string `json:"id" validate:"required,uuid"`
    Website  string `json:"website" validate:"url"`
    Phone    string `json:"phone" validate:"phone"`
}

// Valid
resource := Resource{
    ID:      "550e8400-e29b-41d4-a716-446655440000",
    Website: "https://example.com",
    Phone:   "+1-555-123-4567",
}

// Invalid - Bad UUID
resource := Resource{ID: "not-a-uuid"}
// Error: "id must be a valid UUID"

// Invalid - Bad URL
resource := Resource{Website: "not-a-url"}
// Error: "website must be a valid URL (http or https)"

// Invalid - Bad phone
resource := Resource{Phone: "abc-def-ghij"}
// Error: "phone must be a valid phone number"
```

### Validator with Parameters

```go
func init() {
    // Register length validator with exact length
    validator.RegisterValidator("len", validateLen)
}

func validateLen(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    expectedLen, err := strconv.Atoi(ruleValue)
    if err != nil {
        return nil // Invalid rule value
    }
    
    var actualLen int
    switch fieldValue.Kind() {
    case reflect.String:
        actualLen = len(fieldValue.String())
    case reflect.Slice, reflect.Map, reflect.Array:
        actualLen = fieldValue.Len()
    default:
        return nil
    }
    
    if actualLen != expectedLen {
        return fmt.Errorf("%s must be exactly %d characters/items", fieldName, expectedLen)
    }
    
    return nil
}
```

**Usage:**

```go
type Code struct {
    PIN     string   `json:"pin" validate:"required,len=4"`      // Exactly 4 chars
    Country string   `json:"country" validate:"required,len=2"`  // Exactly 2 chars
}

// Valid
code := Code{PIN: "1234", Country: "US"}

// Invalid
code := Code{PIN: "123"}
// Error: "pin must be exactly 4 characters/items"
```

## Field Names

Error messages use field names from JSON tags:

```go
type User struct {
    FirstName string `json:"first_name" validate:"required"`  // Uses "first_name"
    Email     string `json:"email" validate:"required,email"` // Uses "email"
    Age       int    `validate:"required,gte=18"`             // Uses "Age" (no json tag)
}

user := User{FirstName: ""}
fieldErrors, _ := validator.ValidateStruct(user)
// fieldErrors[0].Field = "first_name"  (from json tag)
// fieldErrors[0].Message = "first_name is required"

user := User{Age: 15}
fieldErrors, _ := validator.ValidateStruct(user)
// fieldErrors[0].Field = "Age"  (no json tag, uses field name)
// fieldErrors[0].Message = "Age must be greater than or equal to 18"
```

## Pointer Fields

Pointer fields are treated as optional:

```go
type Profile struct {
    Bio        *string `json:"bio" validate:"min=10"`        // Optional, but if provided min 10 chars
    Age        *int    `json:"age" validate:"gte=18"`        // Optional, but if provided >= 18
    Website    *string `json:"website" validate:"url"`       // Optional, but if provided must be URL
}

// Valid - All fields nil
profile := Profile{}

// Valid - Bio provided and valid
bio := "This is a long biography"
profile := Profile{Bio: &bio}

// Invalid - Bio provided but too short
bio := "Short"
profile := Profile{Bio: &bio}
// Error: "bio must be at least 10 characters"

// Valid - Age nil (optional)
profile := Profile{Age: nil}

// Invalid - Age provided but too low
age := 15
profile := Profile{Age: &age}
// Error: "age must be greater than or equal to 18"
```

**Required Pointers:**

```go
type Document struct {
    Title   *string `json:"title" validate:"required"`  // Pointer MUST be non-nil
    Content *string `json:"content" validate:"required"`
}

// Valid
title := "My Document"
content := "Content here"
doc := Document{Title: &title, Content: &content}

// Invalid - Title is nil
doc := Document{Title: nil, Content: &content}
// Error: "title is required"
```

## Error Handling

### FieldError Structure

```go
type FieldError struct {
    Field   string  // Field name (from json tag or field name)
    Message string  // Error message
}
```

### Processing Validation Errors

```go
func HandleValidation(data any) error {
    fieldErrors, err := validator.ValidateStruct(data)
    if err != nil {
        // System error (not validation error)
        return fmt.Errorf("validation system error: %w", err)
    }
    
    if len(fieldErrors) > 0 {
        // Build user-friendly error message
        var messages []string
        for _, fe := range fieldErrors {
            messages = append(messages, fe.Message)
        }
        return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
    }
    
    return nil
}
```

### HTTP Error Response

```go
func SendValidationError(w http.ResponseWriter, fieldErrors []api_formatter.FieldError) {
    response := map[string]any{
        "status": "error",
        "error": map[string]any{
            "code":    "VALIDATION_ERROR",
            "message": "Validation failed",
            "fields":  fieldErrors,
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(response)
}

// Example response:
// {
//   "status": "error",
//   "error": {
//     "code": "VALIDATION_ERROR",
//     "message": "Validation failed",
//     "fields": [
//       {"field": "email", "message": "email is required"},
//       {"field": "age", "message": "age must be greater than or equal to 18"}
//     ]
//   }
// }
```

## Performance

### Metadata Caching

Validation metadata is cached per struct type:

```go
// First validation - builds and caches metadata
user1 := User{Name: "Alice"}
validator.ValidateStruct(user1)  // Cache MISS - builds metadata

// Subsequent validations - uses cached metadata
user2 := User{Name: "Bob"}
validator.ValidateStruct(user2)  // Cache HIT - fast

user3 := User{Name: "Charlie"}
validator.ValidateStruct(user3)  // Cache HIT - fast
```

**Cache Behavior:**
- One cache entry per struct type
- Built on first validation
- Reused for all instances of that type
- Thread-safe (uses sync.Map)

### Benchmark Results

```
Operation                     Time         Allocations
First validation (cache miss) ~50 µs/op    5-10 allocs
Cached validation (cache hit) ~5 µs/op     2-3 allocs
Custom validator              ~10 µs/op    3-5 allocs
```

### Optimization Tips

```go
✓ DO: Reuse struct types for caching
type User struct { ... }
// All User instances use same cached metadata

✗ DON'T: Create struct types dynamically
// Each call creates new type, defeats caching
func validate(data map[string]any) {
    type DynamicStruct struct { ... }  // BAD: New type each call
}

✓ DO: Validate at API boundary
// Single validation per request
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)
    validator.ValidateStruct(user)  // Validate once
}

✗ DON'T: Validate repeatedly
// Multiple validations waste CPU
validator.ValidateStruct(user)
processUser(user)
validator.ValidateStruct(user)  // BAD: Redundant
```

## Best Practices

### Validation Strategy

```go
✓ DO: Validate at API boundaries
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)
    
    // Validate immediately after parsing
    if fieldErrors, _ := validator.ValidateStruct(user); len(fieldErrors) > 0 {
        SendValidationError(w, fieldErrors)
        return
    }
    
    // Continue with valid data
}

✗ DON'T: Skip validation
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)
    // BAD: No validation, invalid data propagates
    userRepo.Create(user)
}
```

### Struct Tag Organization

```go
✓ DO: Order tags logically
type User struct {
    Email string `json:"email" validate:"required,email"`  // required first, then type
    Age   int    `json:"age" validate:"required,gte=18,lte=120"`  // required, then range
}

✓ DO: Use meaningful constraints
type Product struct {
    Price float64 `json:"price" validate:"required,gt=0"`  // Price must be positive
    Stock int     `json:"stock" validate:"gte=0"`          // Stock can be zero
}

✗ DON'T: Use overly strict constraints
type Name struct {
    Value string `validate:"required,min=100"` // BAD: 100 chars is too long for names
}
```

### Error Messages

```go
✓ DO: Use json tags for user-friendly field names
type User struct {
    FirstName string `json:"first_name" validate:"required"`  // Error: "first_name is required"
}

✗ DON'T: Use Go field names in API
type User struct {
    FirstName string `validate:"required"`  // Error: "FirstName is required" (not user-friendly)
}

✓ DO: Provide clear validation requirements in API docs
// POST /users
// Body:
// - email: required, must be valid email
// - age: required, must be >= 18
// - username: required, 3-20 characters
```

### Custom Validators

```go
✓ DO: Make custom validators reusable
validator.RegisterValidator("uuid", validateUUID)  // Can be used in any struct

✓ DO: Handle edge cases
func validateEmail(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil  // Skip non-string fields
    }
    
    email := fieldValue.String()
    if email == "" {
        return nil  // Use required tag for empty check
    }
    
    // Validation logic...
}

✗ DON'T: Create validators for single-use cases
// BAD: Inline validation in handler instead
func CreateUser(user User) error {
    if !isValidEmail(user.Email) {
        return errors.New("invalid email")
    }
}
```

## Examples

### Complete User Registration

```go
type RegisterRequest struct {
    Username  string `json:"username" validate:"required,min=3,max=20"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    Age       int    `json:"age" validate:"required,gte=18"`
    Terms     bool   `json:"terms" validate:"required"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Validate request
    fieldErrors, err := validator.ValidateStruct(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if len(fieldErrors) > 0 {
        SendValidationError(w, fieldErrors)
        return
    }
    
    // All validation passed - create user
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashPassword(req.Password),
        Age:      req.Age,
    }
    
    if err := userRepo.Create(user); err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]any{
        "status": "success",
        "user":   user,
    })
}
```

### Product Creation with Optional Fields

```go
type CreateProductRequest struct {
    Name        string   `json:"name" validate:"required,min=1,max=100"`
    Description *string  `json:"description" validate:"max=500"`  // Optional, max 500 chars
    Price       float64  `json:"price" validate:"required,gt=0"`
    Stock       int      `json:"stock" validate:"gte=0"`
    Category    string   `json:"category" validate:"required,oneof=electronics clothing food"`
    Tags        []string `json:"tags" validate:"min=1,max=10"`
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Validate
    fieldErrors, _ := validator.ValidateStruct(req)
    if len(fieldErrors) > 0 {
        SendValidationError(w, fieldErrors)
        return
    }
    
    // Create product
    product := &Product{
        Name:        req.Name,
        Description: req.Description,  // May be nil
        Price:       req.Price,
        Stock:       req.Stock,
        Category:    req.Category,
        Tags:        req.Tags,
    }
    
    productRepo.Create(product)
    // ...
}
```

### Configuration Validation

```go
type AppConfig struct {
    Port         int           `json:"port" validate:"required,gte=1,lte=65535"`
    Host         string        `json:"host" validate:"required"`
    Debug        bool          `json:"debug"`
    ReadTimeout  int           `json:"read_timeout" validate:"gte=0"`
    WriteTimeout int           `json:"write_timeout" validate:"gte=0"`
    Database     DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
    Host     string `json:"host" validate:"required"`
    Port     int    `json:"port" validate:"required,gte=1,lte=65535"`
    User     string `json:"user" validate:"required"`
    Password string `json:"password" validate:"required,min=8"`
    Database string `json:"database" validate:"required"`
}

func LoadConfig(filename string) (*AppConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config AppConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    // Validate configuration
    fieldErrors, err := validator.ValidateStruct(config)
    if err != nil {
        return nil, err
    }
    
    if len(fieldErrors) > 0 {
        var messages []string
        for _, fe := range fieldErrors {
            messages = append(messages, fmt.Sprintf("%s: %s", fe.Field, fe.Message))
        }
        return nil, fmt.Errorf("invalid configuration:\n%s", strings.Join(messages, "\n"))
    }
    
    return &config, nil
}
```

### Custom Business Validator

```go
func init() {
    // Register business-specific validators
    validator.RegisterValidator("product_code", validateProductCode)
    validator.RegisterValidator("currency", validateCurrency)
}

func validateProductCode(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil
    }
    
    code := fieldValue.String()
    if code == "" {
        return nil
    }
    
    // Format: ABC-1234
    parts := strings.Split(code, "-")
    if len(parts) != 2 {
        return fmt.Errorf("%s must be in format ABC-1234", fieldName)
    }
    
    if len(parts[0]) != 3 || len(parts[1]) != 4 {
        return fmt.Errorf("%s must be in format ABC-1234", fieldName)
    }
    
    return nil
}

func validateCurrency(fieldName string, fieldValue reflect.Value, ruleValue string) error {
    if fieldValue.Kind() != reflect.String {
        return nil
    }
    
    currency := fieldValue.String()
    validCurrencies := []string{"USD", "EUR", "GBP", "JPY"}
    
    for _, valid := range validCurrencies {
        if currency == valid {
            return nil
        }
    }
    
    return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(validCurrencies, ", "))
}

type Order struct {
    ProductCode string  `json:"product_code" validate:"required,product_code"`
    Amount      float64 `json:"amount" validate:"required,gt=0"`
    Currency    string  `json:"currency" validate:"required,currency"`
}
```

## Related Documentation

- [Helpers Overview](index) - All helper packages
- [Cast Package](cast) - Type conversion utilities
- [Utils Package](utils) - General utilities

---

**Next:** [Custom Type Package](customtype) - DateTime, Date, Decimal types
