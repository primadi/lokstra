# RPC Service Module - Manual Implementation

This module provides a simple and efficient RPC client for calling remote services over HTTP with msgpack serialization.

## Design Philosophy

We use **manual implementation** approach for RPC clients because:

- ✅ **Type Safety**: Full compile-time type checking
- ✅ **IDE Support**: Autocompletion and error checking  
- ✅ **Performance**: No reflection overhead during calls
- ✅ **Debugging**: Clear stack traces and error messages
- ✅ **Maintainability**: Explicit, readable code

## Core Components

### 1. RpcClient

The basic HTTP client for making RPC calls:

```go
client := NewRpcClient("http://localhost:8080/api")
result, err := client.Call("GetUser", []any{123})
```

### 2. Manual Interface Implementation

For each remote service interface, create a client struct:

```go
// 1. Define your service interface
type UserService interface {
    GetUser(id int) (*User, error)
    CreateUser(name, email string) (*User, error)
    DeleteUser(id int) error
    ListUsers() ([]*User, error)
}

// 2. Create client implementation
type UserServiceClient struct {
    client *RpcClient
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
    return &UserServiceClient{
        client: NewRpcClient(baseURL),
    }
}

// 3. Implement each method manually
func (c *UserServiceClient) GetUser(id int) (*User, error) {
    result, err := c.client.Call("GetUser", []any{id})
    if err != nil {
        return nil, err
    }
    return result.(*User), nil
}

// ... implement other methods
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/primadi/lokstra/modules/rpc_service"
)

func main() {
    // Create client
    client := rpc_service.NewUserServiceClient("http://localhost:8080/api/users")
    
    // Use like any normal interface
    user, err := client.GetUser(123)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("User: %+v\n", user)
    
    // Create new user
    newUser, err := client.CreateUser("John Doe", "john@example.com")
    if err != nil {
        fmt.Printf("Error creating user: %v\n", err)
        return
    }
    
    fmt.Printf("Created user: %+v\n", newUser)
}
```

## Protocol Details

### HTTP Transport
- **Method**: POST
- **Content-Type**: `application/octet-stream`
- **Encoding**: MessagePack for request/response bodies
- **URL Format**: `{baseURL}/{methodname}` (lowercase)

### Error Handling
- **HTTP 4xx/5xx**: Returns error with response message
- **Network errors**: Returns wrapped error
- **Encoding errors**: Returns detailed error message

### Example Request Flow

1. **Encode**: Arguments → MessagePack
2. **HTTP**: POST to `{baseURL}/getuser`
3. **Decode**: Response → Go types
4. **Type Conversion**: any → Concrete types

## Why Not Dynamic Implementation?

Go doesn't support dynamic interface implementation at runtime because:

1. **Compile-time method sets**: Interface satisfaction checked at compile time
2. **Type system design**: Go prioritizes type safety and performance
3. **Reflection limitations**: Can create functions but not struct methods

## Best Practices

1. **Keep it simple**: One client struct per interface
2. **Error handling**: Always check and wrap errors appropriately
3. **Type safety**: Use concrete types, avoid any when possible
4. **Testing**: Mock the RpcClient for unit tests

```go
// Good: Type-safe client
type UserServiceClient struct {
    client *RpcClient
}

// Bad: Generic client without type safety
type GenericClient struct {
    client *RpcClient
}
```

## Testing

```go
func TestUserServiceClient(t *testing.T) {
    // Create mock or test server
    server := httptest.NewServer(...)
    defer server.Close()
    
    client := NewUserServiceClient(server.URL)
    
    user, err := client.GetUser(123)
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
}
```

This approach gives you the best balance of simplicity, type safety, and performance for RPC clients in Go.
