# API Error Handling

## Overview

When making remote service calls via `FetchAndCast` or `CallRemoteService`, errors are returned as `*api_client.ApiError` which preserves HTTP status code information from the downstream service.

## ApiError Structure

```go
type ApiError struct {
    StatusCode int               // HTTP status code (400, 401, 404, 500, etc.)
    Code       string            // Error code (e.g., "VALIDATION_ERROR", "NOT_FOUND")
    Message    string            // Human-readable error message
    Details    map[string]any    // Optional additional error details
}
```

## Key Design Decisions

### Why Not Use `c.Api.XXX()` in FetchAndCast?

**Problem:** Initially, `FetchAndCast` used `c.Api.BadRequest()`, `c.Api.NotFound()` etc. to return errors. This was **incorrect** because:

1. ❌ `c.Api.XXX()` methods are **response terminators** for HTTP handlers
2. ❌ They return `nil` as the data value and a special error that triggers response writing
3. ❌ In service-to-service calls, this causes errors to be "lost" 
4. ❌ The caller receives `(nil, nil)` instead of proper error information

**Solution:** Return `*ApiError` which:

1. ✅ Implements standard Go `error` interface
2. ✅ Preserves HTTP status code information
3. ✅ Allows caller to decide how to handle the error
4. ✅ Decouples client layer from response layer

### Removed `request.Context` Dependency

**Before:**
```go
func FetchAndCast[T any](c *request.Context, client *ClientRouter, path string, opts ...FetchOption) (T, error)
```

**After:**
```go
func FetchAndCast[T any](client *ClientRouter, path string, opts ...FetchOption) (T, error)
```

**Reason:** `FetchAndCast` is a client helper function that should not be tightly coupled to HTTP request context. It should be usable in any context (background jobs, CLI tools, etc.).

## Usage Patterns

### Pattern 1: Direct Error Propagation (Recommended)

Simply return the error as-is. Let the framework's error handler deal with it:

```go
func (s *orderServiceLocal) CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
    // Call user service (may be local or remote)
    user, err := s.userService.Get().GetUser(ctx, &GetUserRequest{UserID: req.UserID})
    if err != nil {
        // ApiError will be automatically handled by the framework
        return nil, fmt.Errorf("user verification failed: %w", err)
    }
    
    // Continue with order creation...
}
```

### Pattern 2: Conditional Error Handling

Handle specific error types differently:

```go
func (s *orderServiceLocal) CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
    user, err := s.userService.Get().GetUser(ctx, &GetUserRequest{UserID: req.UserID})
    if err != nil {
        // Check if it's an ApiError
        if apiErr, ok := err.(*api_client.ApiError); ok {
            switch apiErr.StatusCode {
            case 404:
                return nil, fmt.Errorf("user does not exist")
            case 401:
                return nil, fmt.Errorf("unauthorized access")
            default:
                return nil, fmt.Errorf("user service error: %s", apiErr.Message)
            }
        }
        // Not an ApiError - some other error type
        return nil, err
    }
    
    // Continue...
}
```

### Pattern 3: HTTP Handler with ApiError

In HTTP handlers, you can convert `ApiError` back to proper HTTP response:

```go
func createOrderHandler(c *request.Context, req *CreateOrderRequest) error {
    order, err := orderService.CreateOrder(c, req)
    if err != nil {
        // Check if it's an ApiError from downstream service
        if apiErr, ok := err.(*api_client.ApiError); ok {
            // Forward the exact status code and message
            return c.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
        
        // Generic error - return 500
        return c.Api.InternalError(err.Error())
    }
    
    return c.Api.Created(order, "Order created successfully")
}
```

### Pattern 4: Using Helper Methods

`ApiError` provides convenient helper methods:

```go
user, err := userService.GetUser(ctx, req)
if err != nil {
    if apiErr, ok := err.(*api_client.ApiError); ok {
        if apiErr.IsNotFound() {
            // Handle 404
        } else if apiErr.IsUnauthorized() {
            // Handle 401
        } else if apiErr.IsServerError() {
            // Handle 5xx
        }
    }
}
```

Available helper methods:
- `IsClientError()` - Returns true for 4xx errors
- `IsServerError()` - Returns true for 5xx errors
- `IsBadRequest()` - Returns true for 400
- `IsUnauthorized()` - Returns true for 401
- `IsForbidden()` - Returns true for 403
- `IsNotFound()` - Returns true for 404

## Error Flow Example

```
┌─────────────────────┐
│  HTTP Handler       │
│  (order endpoint)   │
└──────────┬──────────┘
           │ calls
           ▼
┌─────────────────────┐
│  Order Service      │
│  (business logic)   │
└──────────┬──────────┘
           │ calls
           ▼
┌─────────────────────┐
│  User Service       │
│  (may be remote)    │
└──────────┬──────────┘
           │ if remote
           ▼
┌─────────────────────┐
│  CallRemoteService  │
└──────────┬──────────┘
           │ uses
           ▼
┌─────────────────────┐
│  FetchAndCast       │
│  Makes HTTP call    │
└──────────┬──────────┘
           │ returns ApiError on failure
           ▼
┌─────────────────────┐
│  ApiError           │
│  StatusCode: 404    │
│  Code: "NOT_FOUND"  │
│  Message: "user..." │
└─────────────────────┘
```

## Best Practices

1. **Service Layer**: Don't check `ApiError` unless you need to handle specific status codes differently. Just propagate errors up.

2. **HTTP Handler Layer**: Convert `ApiError` to HTTP response using `c.Api.Error()`.

3. **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` to preserve the original `ApiError` while adding context.

4. **Logging**: Log `ApiError` details for debugging:
   ```go
   if apiErr, ok := err.(*api_client.ApiError); ok {
       log.Printf("Remote API error: [%d] %s: %s", 
           apiErr.StatusCode, apiErr.Code, apiErr.Message)
   }
   ```

5. **Testing**: When testing services that call remote services, you can create mock `ApiError`:
   ```go
   mockError := &api_client.ApiError{
       StatusCode: 404,
       Code:       "NOT_FOUND",
       Message:    "user not found",
   }
   ```

## Migration from Old Code

**Before:**
```go
// ❌ OLD - FetchAndCast required request.Context
user, err := api_client.FetchAndCast[*User](ctx, client, "/users/123")
```

**After:**
```go
// ✅ NEW - No request.Context needed
user, err := api_client.FetchAndCast[*User](client, "/users/123")
```

The `request.Context` parameter has been completely removed from `FetchAndCast` and `CallRemoteService` signatures.
