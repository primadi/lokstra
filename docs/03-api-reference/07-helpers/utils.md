# Utils Package

The `utils` package provides general-purpose utility functions for common programming tasks including map operations, slice manipulations, string processing, and more.

## Table of Contents

- [Overview](#overview)
- [Map Operations](#map-operations)
- [Slice Operations](#slice-operations)
- [String Operations](#string-operations)
- [Security Utilities](#security-utilities)
- [Type Checking](#type-checking)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/utils`

**Key Features:**

```
✓ Type-Safe Map Access    - GetValueFromMap with generics
✓ Duration Parsing        - Flexible duration extraction
✓ Slice Operations        - Concat, convert, sorted insert
✓ String Utilities        - Case conversion, parsing
✓ Security Helpers        - Password hashing, IP extraction
✓ Type Checking           - Nil checking utilities
```

## Map Operations

### GetValueFromMap

Extract typed values from `map[string]any` with defaults:

```go
config := map[string]any{
    "host":    "localhost",
    "port":    8080,
    "timeout": "30s",
    "enabled": true,
}

// String value
host := utils.GetValueFromMap(config, "host", "127.0.0.1")
// host = "localhost"

// Int value
port := utils.GetValueFromMap(config, "port", 3000)
// port = 8080

// Bool value
enabled := utils.GetValueFromMap(config, "enabled", false)
// enabled = true

// Missing key (returns default)
retries := utils.GetValueFromMap(config, "retries", 3)
// retries = 3

// Wrong type (returns default)
invalidPort := utils.GetValueFromMap[string](config, "port", "3000")
// invalidPort = "3000" (port is int, not string)
```

**Type Safety:**

```go
// Type parameter ensures compile-time safety
port := utils.GetValueFromMap[int](config, "port", 8080)  // Correct
name := utils.GetValueFromMap[string](config, "name", "") // Correct

// Compiler catches type mismatches
port := utils.GetValueFromMap[string](config, "port", 8080)  // Compile error
```

**Pointer Values:**

```go
config := map[string]any{
    "timeout": new(int),
}

// Handles both direct values and pointers
timeout := utils.GetValueFromMap(config, "timeout", 30)
// Dereferences pointer automatically
```

### GetDurationFromMap

Extract `time.Duration` from various formats:

```go
config := map[string]any{
    "timeout1": "30s",           // String format
    "timeout2": 45,              // Int (seconds)
    "timeout3": 60.5,            // Float (seconds)
    "timeout4": 2 * time.Minute, // Duration
}

// From string
timeout1 := utils.GetDurationFromMap(config, "timeout1", 10*time.Second)
// timeout1 = 30 * time.Second

// From int (interpreted as seconds)
timeout2 := utils.GetDurationFromMap(config, "timeout2", time.Minute)
// timeout2 = 45 * time.Second

// From float (interpreted as seconds)
timeout3 := utils.GetDurationFromMap(config, "timeout3", time.Minute)
// timeout3 = 60.5 * time.Second

// From time.Duration
timeout4 := utils.GetDurationFromMap(config, "timeout4", time.Minute)
// timeout4 = 2 * time.Minute

// Missing key (returns default)
timeout5 := utils.GetDurationFromMap(config, "timeout5", time.Minute)
// timeout5 = time.Minute
```

**Default Value Formats:**

```go
// String default
timeout := utils.GetDurationFromMap(config, "timeout", "1m30s")

// Int default (seconds)
timeout := utils.GetDurationFromMap(config, "timeout", 90)

// Duration default
timeout := utils.GetDurationFromMap(config, "timeout", 90*time.Second)
```

### CloneMap

Create a shallow copy of a map:

```go
original := map[string]int{
    "a": 1,
    "b": 2,
    "c": 3,
}

// Clone map
cloned := utils.CloneMap(original)

// Modify clone (doesn't affect original)
cloned["a"] = 100
cloned["d"] = 4

// original = {"a": 1, "b": 2, "c": 3}
// cloned = {"a": 100, "b": 2, "c": 3, "d": 4}
```

**Use Cases:**

```go
✓ Passing maps to untrusted functions
safeCopy := utils.CloneMap(sensitive Data)
process(safeCopy)

✓ Creating base configurations
baseConfig := map[string]any{"port": 8080}
devConfig := utils.CloneMap(baseConfig)
devConfig["debug"] = true

prodConfig := utils.CloneMap(baseConfig)
prodConfig["debug"] = false
```

## Slice Operations

### ToAnySlice

Convert typed slice to `[]any`:

```go
// From []string
names := []string{"Alice", "Bob", "Charlie"}
anySlice := utils.ToAnySlice(names)
// []any{"Alice", "Bob", "Charlie"}

// From []int
numbers := []int{1, 2, 3, 4, 5}
anySlice := utils.ToAnySlice(numbers)
// []any{1, 2, 3, 4, 5}

// From custom types
type User struct {
    Name string
}
users := []User{{Name: "Alice"}, {Name: "Bob"}}
anySlice := utils.ToAnySlice(users)
// []any{User{Name: "Alice"}, User{Name: "Bob"}}
```

### SlicesConcat

Concatenate multiple slices:

```go
slice1 := []int{1, 2, 3}
slice2 := []int{4, 5}
slice3 := []int{6, 7, 8}

// Concatenate all
result := utils.SlicesConcat(slice1, slice2, slice3)
// []int{1, 2, 3, 4, 5, 6, 7, 8}

// With empty slices
result := utils.SlicesConcat([]int{1, 2}, []int{}, []int{3})
// []int{1, 2, 3}

// No slices
result := utils.SlicesConcat[int]()
// nil

// All empty slices
result := utils.SlicesConcat([]int{}, []int{}, []int{})
// []int{} (non-nil empty slice)
```

### AppendSorted

Insert element into sorted slice:

```go
type Person struct {
    Name string
    Age  int
}

// Less function for sorting by age
byAge := func(a, b Person) bool {
    return a.Age < b.Age
}

// Start with sorted slice
people := []Person{
    {Name: "Alice", Age: 25},
    {Name: "Bob", Age: 30},
    {Name: "Charlie", Age: 35},
}

// Insert new person (maintains sort order)
people = utils.AppendSorted(people, Person{Name: "David", Age: 28}, byAge)
// [{Alice 25}, {David 28}, {Bob 30}, {Charlie 35}]

// Insert at beginning
people = utils.AppendSorted(people, Person{Name: "Eve", Age: 20}, byAge)
// [{Eve 20}, {Alice 25}, {David 28}, {Bob 30}, {Charlie 35}]

// Insert at end
people = utils.AppendSorted(people, Person{Name: "Frank", Age: 40}, byAge)
// [{Eve 20}, {Alice 25}, {David 28}, {Bob 30}, {Charlie 35}, {Frank 40}]
```

**Simple Types:**

```go
// Sort integers
numbers := []int{1, 3, 5, 7}
numbers = utils.AppendSorted(numbers, 4, func(a, b int) bool {
    return a < b
})
// []int{1, 3, 4, 5, 7}

// Sort strings
names := []string{"Alice", "Charlie", "David"}
names = utils.AppendSorted(names, "Bob", func(a, b string) bool {
    return a < b
})
// []string{"Alice", "Bob", "Charlie", "David"}
```

### AppendSortedOptimize

Optimized sorted insertion using binary search for large slices (>= 16 elements):

```go
// For small slices (< 16), uses linear search
small := []int{1, 2, 3, 4, 5}
small = utils.AppendSortedOptimize(small, 3, func(a, b int) bool {
    return a < b
})

// For large slices (>= 16), uses binary search
large := make([]int, 100)
for i := range large {
    large[i] = i * 2
}
large = utils.AppendSortedOptimize(large, 77, func(a, b int) bool {
    return a < b
})
```

**Performance:**
- Linear search: O(n) - Better for small slices
- Binary search: O(log n) - Better for large slices
- Threshold: 16 elements

## String Operations

### CamelToSnake

Convert camelCase to snake_case:

```go
// Simple conversion
result := utils.CamelToSnake("userId")       // "user_id"
result = utils.CamelToSnake("firstName")     // "first_name"
result = utils.CamelToSnake("isActive")      // "is_active"

// Multiple capitals
result = utils.CamelToSnake("HTTPRequest")   // "http_request"
result = utils.CamelToSnake("XMLParser")     // "xml_parser"

// Already snake_case (no change)
result = utils.CamelToSnake("user_id")       // "user_id"

// PascalCase
result = utils.CamelToSnake("UserProfile")   // "user_profile"
```

**Use Cases:**

```go
// Database column names
type User struct {
    UserID    int    `db:"user_id"`
    FirstName string `db:"first_name"`
}

// Generate column name from field
fieldName := "UserID"
columnName := utils.CamelToSnake(fieldName)  // "user_id"

// API field name conversion
jsonField := "createdAt"
dbField := utils.CamelToSnake(jsonField)  // "created_at"
```

### ParseClientIP

Extract client IP address from HTTP request:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ip := utils.ParseClientIP(r)
    
    // Log client IP
    log.Printf("Request from: %s", ip)
    
    // Use for rate limiting
    if rateLimiter.IsBlocked(ip) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
}
```

**Priority Order:**
1. `X-Forwarded-For` header (first IP)
2. `X-Real-IP` header
3. `RemoteAddr` from request

**With Proxy:**

```
Client → Proxy → Server

X-Forwarded-For: client_ip, proxy1_ip, proxy2_ip
Result: client_ip (first IP in chain)
```

## Security Utilities

### HashPassword

Hash password using bcrypt:

```go
// Hash password for storage
password := "mySecretPassword123"
hashedPassword, err := utils.HashPassword(password)
if err != nil {
    return fmt.Errorf("failed to hash password: %w", err)
}

// Store hashedPassword in database
user.PasswordHash = hashedPassword

// Later, verify password
err = bcrypt.CompareHashAndPassword(
    []byte(user.PasswordHash),
    []byte(password),
)
if err != nil {
    // Invalid password
    return ErrInvalidCredentials
}
// Password valid
```

**Complete Authentication Flow:**

```go
// Registration
func Register(username, password string) error {
    hashedPassword, err := utils.HashPassword(password)
    if err != nil {
        return err
    }
    
    user := &User{
        Username:     username,
        PasswordHash: hashedPassword,
    }
    
    return userRepo.Create(user)
}

// Login
func Login(username, password string) (*User, error) {
    user, err := userRepo.GetByUsername(username)
    if err != nil {
        return nil, ErrInvalidCredentials
    }
    
    err = bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(password),
    )
    if err != nil {
        return nil, ErrInvalidCredentials
    }
    
    return user, nil
}
```

## Type Checking

### IsNil

Check if a value is nil (handles interfaces correctly):

```go
// Regular nil check
var ptr *int
utils.IsNil(ptr)  // true

ptr = new(int)
utils.IsNil(ptr)  // false

// Interface nil check (tricky case)
var service Service
utils.IsNil(service)  // true

service = (*MyService)(nil)
utils.IsNil(service)  // true (nil interface value)

service = &MyService{}
utils.IsNil(service)  // false
```

**Why IsNil is Needed:**

```go
// Regular nil check fails on interfaces
var service Service = (*MyService)(nil)

// This is true (interface contains nil value)
fmt.Println(service == nil)  // false ❌ (interface type is not nil)

// IsNil checks the underlying value
utils.IsNil(service)  // true ✅
```

**Use Cases:**

```go
// Service initialization check
if utils.IsNil(service) {
    service = NewService()
}

// Optional dependency
if !utils.IsNil(logger) {
    logger.Log("Message")
}

// Error handling with custom errors
if !utils.IsNil(err) {
    return fmt.Errorf("operation failed: %w", err)
}
```

## Best Practices

### Map Access

```go
✓ DO: Use GetValueFromMap with appropriate defaults
port := utils.GetValueFromMap(config, "port", 8080)
host := utils.GetValueFromMap(config, "host", "localhost")

✓ DO: Use type parameters for safety
port := utils.GetValueFromMap[int](config, "port", 8080)

✗ DON'T: Access maps directly without type checking
port := config["port"].(int)  // BAD: Panics if wrong type
```

### Duration Parsing

```go
✓ DO: Provide sensible defaults
timeout := utils.GetDurationFromMap(config, "timeout", 30*time.Second)

✓ DO: Support multiple formats
// Config can use any format:
// timeout: "30s"    (string)
// timeout: 30       (int seconds)
// timeout: 30.5     (float seconds)

✗ DON'T: Assume specific format
timeout := config["timeout"].(time.Duration)  // BAD: Only works for Duration type
```

### Slice Operations

```go
✓ DO: Use SlicesConcat for multiple slices
all := utils.SlicesConcat(slice1, slice2, slice3)

✓ DO: Use AppendSorted for maintaining order
sorted = utils.AppendSorted(sorted, newItem, lessFunc)

✗ DON'T: Manually implement concatenation
// BAD: Inefficient
result := make([]int, 0)
result = append(result, slice1...)
result = append(result, slice2...)
result = append(result, slice3...)
```

### String Conversion

```go
✓ DO: Use CamelToSnake for consistent naming
dbColumn := utils.CamelToSnake(structField)

✗ DON'T: Implement custom conversion
// BAD: Incomplete implementation
func toSnake(s string) string {
    return strings.ToLower(s)  // Doesn't handle capitals
}
```

### Password Security

```go
✓ DO: Always hash passwords before storage
hashedPassword, err := utils.HashPassword(password)
if err != nil {
    return err
}
user.PasswordHash = hashedPassword

✗ DON'T: Store plaintext passwords
user.Password = password  // BAD: Security vulnerability

✗ DON'T: Use weak hashing
user.PasswordHash = md5(password)  // BAD: MD5 is not secure
```

## Examples

### Configuration Loading

```go
func LoadConfig(data map[string]any) *Config {
    return &Config{
        AppName:     utils.GetValueFromMap(data, "app_name", "myapp"),
        Port:        utils.GetValueFromMap(data, "port", 8080),
        Host:        utils.GetValueFromMap(data, "host", "0.0.0.0"),
        Debug:       utils.GetValueFromMap(data, "debug", false),
        ReadTimeout: utils.GetDurationFromMap(data, "read_timeout", 30*time.Second),
        WriteTimeout: utils.GetDurationFromMap(data, "write_timeout", 30*time.Second),
        MaxConnections: utils.GetValueFromMap(data, "max_connections", 100),
    }
}
```

### Service Configuration

```go
func ServiceFactory(params map[string]any) Service {
    host := utils.GetValueFromMap(params, "host", "localhost")
    port := utils.GetValueFromMap(params, "port", 5432)
    timeout := utils.GetDurationFromMap(params, "timeout", 30*time.Second)
    poolSize := utils.GetValueFromMap(params, "pool_size", 10)
    
    return &MyService{
        Host:     host,
        Port:     port,
        Timeout:  timeout,
        PoolSize: poolSize,
    }
}
```

### Middleware with IP Logging

```go
func IPLoggerMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := utils.ParseClientIP(r)
        log.Printf("[%s] %s %s", ip, r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiter

```go
type RateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
}

func (rl *RateLimiter) Allow(r *http.Request) bool {
    ip := utils.ParseClientIP(r)
    
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    window := now.Add(-time.Minute)
    
    // Get recent requests for this IP
    requests := rl.requests[ip]
    
    // Filter old requests
    recent := make([]time.Time, 0)
    for _, t := range requests {
        if t.After(window) {
            recent = append(recent, t)
        }
    }
    
    // Check limit
    if len(recent) >= 60 {
        return false
    }
    
    // Add new request
    recent = append(recent, now)
    rl.requests[ip] = recent
    
    return true
}
```

### Sorted Insert Example

```go
type Event struct {
    Name      string
    Timestamp time.Time
}

type EventLog struct {
    events []Event
    mu     sync.Mutex
}

func (el *EventLog) Add(event Event) {
    el.mu.Lock()
    defer el.mu.Unlock()
    
    // Maintain chronological order
    el.events = utils.AppendSorted(el.events, event, func(a, b Event) bool {
        return a.Timestamp.Before(b.Timestamp)
    })
}

func (el *EventLog) GetRecent(n int) []Event {
    el.mu.Lock()
    defer el.mu.Unlock()
    
    if len(el.events) <= n {
        return el.events
    }
    
    return el.events[len(el.events)-n:]
}
```

### User Registration

```go
func Register(username, email, password string) error {
    // Hash password
    hashedPassword, err := utils.HashPassword(password)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Create user
    user := &User{
        ID:           uuid.New().String(),
        Username:     username,
        Email:        email,
        PasswordHash: hashedPassword,
        CreatedAt:    time.Now(),
    }
    
    // Save to database
    return userRepo.Create(user)
}
```

## Related Documentation

- [Helpers Overview](README.md) - All helper packages
- [Cast Package](cast.md) - Type conversion utilities
- [Validator Package](validator.md) - Struct validation

---

**Next:** [Validator Package](validator.md) - Struct validation with tags
