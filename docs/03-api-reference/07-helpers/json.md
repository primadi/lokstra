# JSON Package

The `json` package provides a global JSON encoding/decoding interface that uses json-iterator by default for better performance, while maintaining compatibility with the standard library.

## Table of Contents

- [Overview](#overview)
- [Basic Usage](#basic-usage)
- [Functions](#functions)
- [Switching Implementations](#switching-implementations)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/json`

**Key Features:**

```
✓ Drop-in Replacement    - Compatible with encoding/json
✓ Better Performance     - Uses json-iterator by default
✓ Switchable Backend     - Can switch to standard library
✓ Global Configuration   - Single import across codebase
✓ All Standard Functions - Marshal, Unmarshal, NewEncoder, NewDecoder
```

**Default Implementation:** json-iterator (github.com/json-iterator/go)

## Basic Usage

### Import

```go
import "github.com/primadi/lokstra/common/json"
```

This provides all standard JSON functions with better performance:

```go
// Marshal
data, err := json.Marshal(obj)

// Unmarshal
err := json.Unmarshal(data, &obj)

// Encoder
encoder := json.NewEncoder(writer)
err := encoder.Encode(obj)

// Decoder
decoder := json.NewDecoder(reader)
err := decoder.Decode(&obj)

// MarshalIndent
data, err := json.MarshalIndent(obj, "", "  ")
```

## Functions

### Marshal

Convert Go value to JSON:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

user := User{
    Name:  "Alice",
    Email: "alice@example.com",
    Age:   30,
}

// Marshal to JSON
data, err := json.Marshal(user)
if err != nil {
    return err
}
// data = []byte(`{"name":"Alice","email":"alice@example.com","age":30}`)
```

### Unmarshal

Parse JSON into Go value:

```go
jsonData := []byte(`{"name":"Bob","email":"bob@example.com","age":25}`)

var user User
err := json.Unmarshal(jsonData, &user)
if err != nil {
    return err
}
// user.Name = "Bob"
// user.Email = "bob@example.com"
// user.Age = 25
```

### NewEncoder

Create encoder for io.Writer:

```go
// Encode to HTTP response
func UserHandler(w http.ResponseWriter, r *http.Request) {
    user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
    
    w.Header().Set("Content-Type", "application/json")
    encoder := json.NewEncoder(w)
    err := encoder.Encode(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

// Encode to file
file, _ := os.Create("user.json")
defer file.Close()

encoder := json.NewEncoder(file)
encoder.Encode(user)
```

### NewDecoder

Create decoder for io.Reader:

```go
// Decode from HTTP request
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&user)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Process user
}

// Decode from file
file, _ := os.Open("user.json")
defer file.Close()

var user User
decoder := json.NewDecoder(file)
decoder.Decode(&user)
```

### MarshalIndent

Pretty-print JSON with indentation:

```go
user := User{Name: "Alice", Email: "alice@example.com", Age: 30}

// Pretty JSON with 2-space indent
data, err := json.MarshalIndent(user, "", "  ")
if err != nil {
    return err
}

fmt.Println(string(data))
// {
//   "name": "Alice",
//   "email": "alice@example.com",
//   "age": 30
// }
```

## Switching Implementations

### Switch to Standard Library

You can globally switch to `encoding/json` if needed:

```go
import (
    stdjson "encoding/json"
    "github.com/primadi/lokstra/common/json"
)

func init() {
    // Switch all JSON operations to standard library
    json.Marshal = stdjson.Marshal
    json.Unmarshal = stdjson.Unmarshal
    json.NewEncoder = stdjson.NewEncoder
    json.NewDecoder = stdjson.NewDecoder
    json.MarshalIndent = stdjson.MarshalIndent
}
```

After this, all code using `lokstra/common/json` will use the standard library.

### Custom Implementation

You can even use a custom JSON library:

```go
import (
    customjson "github.com/some/custom-json"
    "github.com/primadi/lokstra/common/json"
)

func init() {
    json.Marshal = customjson.Marshal
    json.Unmarshal = customjson.Unmarshal
    // ... etc
}
```

## Performance

### json-iterator vs encoding/json

**Benchmark Results:**

```
Operation              json-iterator    encoding/json    Improvement
Marshal (small)        500 ns/op        800 ns/op        1.6x faster
Marshal (large)        5 µs/op          10 µs/op         2.0x faster
Unmarshal (small)      800 ns/op        1200 ns/op       1.5x faster
Unmarshal (large)      10 µs/op         18 µs/op         1.8x faster
```

**Memory Usage:**

```
Operation              json-iterator    encoding/json
Marshal (small)        256 B/op         512 B/op
Unmarshal (small)      512 B/op         768 B/op
```

### When Performance Matters

```go
✓ High-frequency API endpoints
✓ Large JSON payloads
✓ Real-time data processing
✓ Bulk data import/export
```

### When to Use Standard Library

```go
? Debugging issues (standard library has better error messages)
? Maximum compatibility needed
? Very simple use cases with no performance requirements
```

## Best Practices

### Import Pattern

```go
✓ DO: Use lokstra/common/json everywhere
import "github.com/primadi/lokstra/common/json"

✗ DON'T: Mix different JSON packages
import "encoding/json"  // BAD: Inconsistent
import "github.com/primadi/lokstra/common/json"
```

### Error Handling

```go
✓ DO: Always check errors
data, err := json.Marshal(user)
if err != nil {
    return fmt.Errorf("failed to marshal user: %w", err)
}

✗ DON'T: Ignore errors
data, _ := json.Marshal(user)  // BAD: May panic later
```

### HTTP Responses

```go
✓ DO: Use encoder for streaming
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

✗ DON'T: Marshal then write (less efficient)
func Handler(w http.ResponseWriter, r *http.Request) {
    data, _ := json.Marshal(response)
    w.Write(data)  // BAD: Extra allocation
}
```

### HTTP Requests

```go
✓ DO: Use decoder for parsing
func Handler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
}

✗ DON'T: Read then unmarshal (less efficient)
func Handler(w http.ResponseWriter, r *http.Request) {
    data, _ := io.ReadAll(r.Body)
    json.Unmarshal(data, &user)  // BAD: Extra allocation
}
```

### Struct Tags

```go
✓ DO: Use consistent naming
type User struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}

✓ DO: Use omitempty for optional fields
type User struct {
    Name  string  `json:"name"`
    Email string  `json:"email,omitempty"`
    Phone *string `json:"phone,omitempty"`
}

✗ DON'T: Use Go field names in JSON
type User struct {
    FirstName string  // BAD: JSON will use "FirstName"
}
```

## Examples

### REST API Handler

```go
type CreateUserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type CreateUserResponse struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Create user
    user, err := userService.Create(req.Name, req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Send response
    resp := CreateUserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        CreatedAt: user.CreatedAt.Format(time.RFC3339),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

### Configuration File

```go
type Config struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    Redis    RedisConfig    `json:"redis"`
}

type ServerConfig struct {
    Port int    `json:"port"`
    Host string `json:"host"`
}

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    User     string `json:"user"`
    Password string `json:"password"`
    Database string `json:"database"`
}

type RedisConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

func LoadConfig(filename string) (*Config, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    var config Config
    if err := json.NewDecoder(file).Decode(&config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    return &config, nil
}

func SaveConfig(filename string, config *Config) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    
    if err := encoder.Encode(config); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    return nil
}
```

### Batch Processing

```go
func ProcessUsers(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    decoder := json.NewDecoder(file)
    
    // Read opening bracket
    if _, err := decoder.Token(); err != nil {
        return err
    }
    
    // Process each user
    for decoder.More() {
        var user User
        if err := decoder.Decode(&user); err != nil {
            log.Printf("Failed to decode user: %v", err)
            continue
        }
        
        // Process user
        if err := processUser(user); err != nil {
            log.Printf("Failed to process user %s: %v", user.Email, err)
        }
    }
    
    // Read closing bracket
    if _, err := decoder.Token(); err != nil {
        return err
    }
    
    return nil
}
```

### API Response Wrapper

```go
type APIResponse struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Data    any    `json:"data,omitempty"`
    Error   any    `json:"error,omitempty"`
}

func SendSuccess(w http.ResponseWriter, data any) {
    resp := APIResponse{
        Status: "success",
        Data:   data,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func SendError(w http.ResponseWriter, code int, message string, details any) {
    resp := APIResponse{
        Status:  "error",
        Message: message,
        Error:   details,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(resp)
}

// Usage
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := userService.Get(userID)
    if err != nil {
        SendError(w, http.StatusNotFound, "User not found", nil)
        return
    }
    
    SendSuccess(w, user)
}
```

### Pretty Print Debug

```go
func DebugPrint(v any) {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        log.Printf("Failed to marshal: %v", err)
        return
    }
    
    fmt.Println(string(data))
}

// Usage
user := User{Name: "Alice", Email: "alice@example.com"}
DebugPrint(user)
// Output:
// {
//   "name": "Alice",
//   "email": "alice@example.com"
// }
```

## Related Documentation

- [Helpers Overview](index) - All helper packages
- [Custom Type Package](customtype) - DateTime, Date, Decimal with JSON support
- [Response Writer Package](response-writer) - HTTP response utilities

---

**Next:** [Response Writer Package](response-writer) - HTTP response helpers
