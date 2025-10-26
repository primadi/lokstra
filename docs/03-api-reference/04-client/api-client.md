# API Client

> HTTP client utilities for calling Lokstra services locally and remotely

## Overview

The `api_client` package provides utilities for making HTTP requests to Lokstra services, with support for both local (in-process) and remote (HTTP) communication. It includes type-safe response parsing, error handling, and automatic request/response formatting.

## Import Path

```go
import "github.com/primadi/lokstra/api_client"
```

---

## Core Functions

### FetchAndCast
Type-safe fetch helper with flexible options.

**Signature:**
```go
func FetchAndCast[T any](
    client *ClientRouter,
    path string,
    opts ...FetchOption,
) (T, error)
```

**Type Parameters:**
- `T` - Expected response type

**Parameters:**
- `client` - ClientRouter instance
- `path` - Request path
- `opts` - Optional fetch options

**Returns:**
- `T` - Parsed response data
- `error` - Error if request fails or parsing fails

**Example:**
```go
// Simple GET request
user, err := api_client.FetchAndCast[*User](client, "/users/123")
if err != nil {
    log.Fatal(err)
}

// POST request with body
created, err := api_client.FetchAndCast[*User](client, "/users",
    api_client.WithMethod("POST"),
    api_client.WithBody(newUser),
)

// Custom headers
data, err := api_client.FetchAndCast[*Data](client, "/data",
    api_client.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
        "X-Request-ID":  requestID,
    }),
)
```

**Features:**
- ✅ Type-safe response parsing
- ✅ Automatic JSON marshaling/unmarshaling
- ✅ Error handling with status codes
- ✅ Custom formatters
- ✅ Minimal reflection overhead (~8ns)

---

## Fetch Options

### WithMethod
Sets the HTTP method for the request.

**Signature:**
```go
func WithMethod(method string) FetchOption
```

**Example:**
```go
api_client.FetchAndCast[*User](client, "/users",
    api_client.WithMethod("POST"))

api_client.FetchAndCast[*User](client, "/users/123",
    api_client.WithMethod("PUT"))
```

---

### WithBody
Sets the request body (auto-marshaled to JSON).

**Signature:**
```go
func WithBody(body any) FetchOption
```

**Example:**
```go
newUser := &User{
    Name:  "John Doe",
    Email: "john@example.com",
}

created, err := api_client.FetchAndCast[*User](client, "/users",
    api_client.WithMethod("POST"),
    api_client.WithBody(newUser),
)
```

---

### WithHeaders
Sets custom headers for the request.

**Signature:**
```go
func WithHeaders(headers map[string]string) FetchOption
```

**Example:**
```go
data, err := api_client.FetchAndCast[*Data](client, "/data",
    api_client.WithHeaders(map[string]string{
        "Authorization":  "Bearer " + token,
        "X-Request-ID":   requestID,
        "X-Custom-Header": "custom-value",
    }),
)
```

---

### WithFormatter
Sets a custom response formatter.

**Signature:**
```go
func WithFormatter(formatter api_formatter.ResponseFormatter) FetchOption
```

**Example:**
```go
customFormatter := &MyCustomFormatter{}

data, err := api_client.FetchAndCast[*Data](client, "/data",
    api_client.WithFormatter(customFormatter),
)
```

---

### WithCustomFunc
Custom handling of response with full control.

**Signature:**
```go
func WithCustomFunc(
    fn func(*http.Response, *api_formatter.ClientResponse) (any, error),
) FetchOption
```

**Example:**
```go
data, err := api_client.FetchAndCast[*Data](client, "/data",
    api_client.WithCustomFunc(func(resp *http.Response, clientResp *api_formatter.ClientResponse) (any, error) {
        // Custom validation
        if resp.StatusCode == 204 {
            return &Data{Empty: true}, nil
        }
        
        // Custom parsing
        if clientResp.Data != nil {
            return parseCustomData(clientResp.Data), nil
        }
        
        return nil, nil // Continue with default flow
    }),
)
```

---

## Error Handling

### ApiError
Structured error with HTTP status code information.

**Definition:**
```go
type ApiError struct {
    StatusCode int            // HTTP status code (400, 401, 404, 500, etc.)
    Code       string         // Error code (e.g., "VALIDATION_ERROR")
    Message    string         // Human-readable error message
    Details    map[string]any // Optional additional details
}
```

**Methods:**
```go
func (e *ApiError) Error() string
func (e *ApiError) IsClientError() bool     // 4xx
func (e *ApiError) IsServerError() bool     // 5xx
func (e *ApiError) IsBadRequest() bool      // 400
func (e *ApiError) IsUnauthorized() bool    // 401
func (e *ApiError) IsForbidden() bool       // 403
func (e *ApiError) IsNotFound() bool        // 404
```

**Example:**
```go
user, err := api_client.FetchAndCast[*User](client, "/users/123")
if err != nil {
    if apiErr, ok := err.(*api_client.ApiError); ok {
        switch {
        case apiErr.IsNotFound():
            return ctx.Api.NotFound("User not found")
        case apiErr.IsUnauthorized():
            return ctx.Api.Unauthorized("Authentication required")
        case apiErr.IsBadRequest():
            return ctx.Api.BadRequest(apiErr.Message)
        default:
            return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
    }
    return ctx.Api.InternalError(err.Error())
}
```

---

### NewApiError
Creates a new ApiError.

**Signature:**
```go
func NewApiError(statusCode int, code, message string) *ApiError
```

**Example:**
```go
err := api_client.NewApiError(404, "NOT_FOUND", "Resource not found")
```

---

### NewApiErrorWithDetails
Creates ApiError with additional details.

**Signature:**
```go
func NewApiErrorWithDetails(
    statusCode int,
    code, message string,
    details map[string]any,
) *ApiError
```

**Example:**
```go
err := api_client.NewApiErrorWithDetails(
    400,
    "VALIDATION_ERROR",
    "Invalid input",
    map[string]any{
        "fields": []string{"email", "password"},
        "constraints": map[string]string{
            "email":    "must be valid email",
            "password": "must be at least 8 characters",
        },
    },
)
```

---

## Complete Examples

### Simple GET Request
```go
package service

import (
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

type UserService struct {
    client *api_client.ClientRouter
}

func NewUserService() *UserService {
    return &UserService{
        client: lokstra_registry.GetClientRouter("user-service"),
    }
}

func (s *UserService) GetUser(id int) (*User, error) {
    path := fmt.Sprintf("/users/%d", id)
    return api_client.FetchAndCast[*User](s.client, path)
}

func (s *UserService) ListUsers() ([]*User, error) {
    return api_client.FetchAndCast[[]*User](s.client, "/users")
}
```

---

### POST Request with Body
```go
func (s *UserService) CreateUser(user *User) (*User, error) {
    return api_client.FetchAndCast[*User](s.client, "/users",
        api_client.WithMethod("POST"),
        api_client.WithBody(user),
    )
}

func (s *UserService) UpdateUser(id int, user *User) (*User, error) {
    path := fmt.Sprintf("/users/%d", id)
    return api_client.FetchAndCast[*User](s.client, path,
        api_client.WithMethod("PUT"),
        api_client.WithBody(user),
    )
}

func (s *UserService) DeleteUser(id int) error {
    path := fmt.Sprintf("/users/%d", id)
    _, err := api_client.FetchAndCast[any](s.client, path,
        api_client.WithMethod("DELETE"),
    )
    return err
}
```

---

### Request with Authentication
```go
type AuthenticatedService struct {
    client *api_client.ClientRouter
    token  string
}

func (s *AuthenticatedService) GetProtectedData() (*Data, error) {
    return api_client.FetchAndCast[*Data](s.client, "/protected/data",
        api_client.WithHeaders(map[string]string{
            "Authorization": "Bearer " + s.token,
        }),
    )
}

func (s *AuthenticatedService) CreateOrder(order *Order) (*Order, error) {
    return api_client.FetchAndCast[*Order](s.client, "/orders",
        api_client.WithMethod("POST"),
        api_client.WithBody(order),
        api_client.WithHeaders(map[string]string{
            "Authorization": "Bearer " + s.token,
            "X-Idempotency-Key": generateIdempotencyKey(),
        }),
    )
}
```

---

### Error Handling Pattern
```go
func (s *UserService) GetUser(ctx *request.Context, id int) error {
    path := fmt.Sprintf("/users/%d", id)
    user, err := api_client.FetchAndCast[*User](s.client, path)
    
    if err != nil {
        // Check if it's an API error
        if apiErr, ok := err.(*api_client.ApiError); ok {
            // Handle specific error types
            switch {
            case apiErr.IsNotFound():
                return ctx.Api.NotFound("User not found")
                
            case apiErr.IsUnauthorized():
                return ctx.Api.Unauthorized("Authentication required")
                
            case apiErr.IsBadRequest():
                return ctx.Api.BadRequest(apiErr.Message)
                
            case apiErr.IsServerError():
                log.Printf("Upstream server error: %v", apiErr)
                return ctx.Api.InternalError("Service temporarily unavailable")
                
            default:
                return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
            }
        }
        
        // Other error types (network, timeout, etc.)
        log.Printf("Request failed: %v", err)
        return ctx.Api.InternalError("Failed to fetch user")
    }
    
    return ctx.Api.Ok(user)
}
```

---

### Custom Response Parsing
```go
func (s *DataService) GetCustomData() (*CustomData, error) {
    return api_client.FetchAndCast[*CustomData](s.client, "/data",
        api_client.WithCustomFunc(func(resp *http.Response, clientResp *api_formatter.ClientResponse) (any, error) {
            // Handle empty response
            if resp.StatusCode == 204 {
                return &CustomData{Empty: true}, nil
            }
            
            // Validate custom headers
            if resp.Header.Get("X-Data-Version") != "v2" {
                return nil, fmt.Errorf("unsupported data version")
            }
            
            // Custom parsing logic
            if clientResp.Data != nil {
                data := &CustomData{}
                if err := parseCustomFormat(clientResp.Data, data); err != nil {
                    return nil, err
                }
                return data, nil
            }
            
            return nil, nil // Continue with default flow
        }),
    )
}
```

---

### Pagination Pattern
```go
type PaginatedResponse struct {
    Items      []*User `json:"items"`
    Page       int     `json:"page"`
    PageSize   int     `json:"page_size"`
    TotalItems int     `json:"total_items"`
    TotalPages int     `json:"total_pages"`
}

func (s *UserService) ListUsersPaginated(page, pageSize int) (*PaginatedResponse, error) {
    path := fmt.Sprintf("/users?page=%d&page_size=%d", page, pageSize)
    return api_client.FetchAndCast[*PaginatedResponse](s.client, path)
}

func (s *UserService) GetAllUsers() ([]*User, error) {
    var allUsers []*User
    page := 1
    pageSize := 100
    
    for {
        resp, err := s.ListUsersPaginated(page, pageSize)
        if err != nil {
            return nil, err
        }
        
        allUsers = append(allUsers, resp.Items...)
        
        if page >= resp.TotalPages {
            break
        }
        
        page++
    }
    
    return allUsers, nil
}
```

---

### Retry Pattern
```go
func (s *UserService) GetUserWithRetry(id int, maxRetries int) (*User, error) {
    path := fmt.Sprintf("/users/%d", id)
    
    var lastErr error
    for attempt := 0; attempt <= maxRetries; attempt++ {
        user, err := api_client.FetchAndCast[*User](s.client, path)
        if err == nil {
            return user, nil
        }
        
        // Check if error is retryable
        if apiErr, ok := err.(*api_client.ApiError); ok {
            if apiErr.IsClientError() {
                // Don't retry client errors (4xx)
                return nil, err
            }
        }
        
        lastErr = err
        
        // Wait before retry (exponential backoff)
        if attempt < maxRetries {
            time.Sleep(time.Duration(1<<attempt) * time.Second)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

### Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    failures     int
    lastFailTime time.Time
    threshold    int
    timeout      time.Duration
    mu           sync.Mutex
}

func (s *UserService) GetUserWithCircuitBreaker(id int) (*User, error) {
    s.cb.mu.Lock()
    
    // Check if circuit is open
    if s.cb.failures >= s.cb.threshold {
        if time.Since(s.cb.lastFailTime) < s.cb.timeout {
            s.cb.mu.Unlock()
            return nil, fmt.Errorf("circuit breaker open")
        }
        // Reset after timeout
        s.cb.failures = 0
    }
    
    s.cb.mu.Unlock()
    
    // Make request
    path := fmt.Sprintf("/users/%d", id)
    user, err := api_client.FetchAndCast[*User](s.client, path)
    
    if err != nil {
        s.cb.mu.Lock()
        s.cb.failures++
        s.cb.lastFailTime = time.Now()
        s.cb.mu.Unlock()
        return nil, err
    }
    
    // Reset on success
    s.cb.mu.Lock()
    s.cb.failures = 0
    s.cb.mu.Unlock()
    
    return user, nil
}
```

---

## Best Practices

### 1. Use Type-Safe FetchAndCast
```go
// ✅ Good: Type-safe
user, err := api_client.FetchAndCast[*User](client, "/users/123")

// 🚫 Avoid: Untyped response
resp, err := client.GET("/users/123")
var user User
json.Unmarshal(resp.Body, &user)
```

---

### 2. Handle ApiError Properly
```go
// ✅ Good: Check error type and handle appropriately
if apiErr, ok := err.(*api_client.ApiError); ok {
    switch {
    case apiErr.IsNotFound():
        return handleNotFound()
    case apiErr.IsUnauthorized():
        return handleUnauthorized()
    }
}

// 🚫 Avoid: Generic error handling
if err != nil {
    return err // Loses status code information
}
```

---

### 3. Use Options for Clarity
```go
// ✅ Good: Clear intent
api_client.FetchAndCast[*User](client, "/users",
    api_client.WithMethod("POST"),
    api_client.WithBody(user),
    api_client.WithHeaders(headers),
)

// 🚫 Avoid: Positional arguments
client.Request("POST", "/users", user, headers)
```

---

### 4. Don't Retry Client Errors
```go
// ✅ Good: Only retry server errors
if apiErr, ok := err.(*api_client.ApiError); ok {
    if apiErr.IsClientError() {
        return err // Don't retry 4xx
    }
}

// 🚫 Avoid: Retrying all errors
for i := 0; i < 3; i++ {
    _, err := fetch()
    if err == nil {
        break
    }
}
```

---

### 5. Set Appropriate Timeouts
```go
// ✅ Good: Configure timeout per service
client.Timeout = 10 * time.Second

// 🚫 Avoid: Using default timeout for all services
// (default is 30s, may be too long)
```

---

## See Also

- **[ClientRouter](./client-router.md)** - Client router management
- **[RemoteService](./remote-service.md)** - Remote service patterns
- **[Response](../01-core-packages/response.md)** - Response formatting

---

## Related Guides

- **[HTTP Clients](../../04-guides/http-clients/)** - HTTP client patterns
- **[Error Handling](../../04-guides/error-handling/)** - Error patterns
- **[Testing](../../04-guides/testing/)** - Testing remote services
