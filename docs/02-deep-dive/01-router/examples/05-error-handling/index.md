# Error Handling Deep Dive

> **Master error handling patterns for production-ready APIs**

This example demonstrates comprehensive error handling strategies in Lokstra.

## Error Response Types

### 1. API Helper Errors (Recommended)

```go
func GetUser(params *getUserParams) *response.ApiHelper {
    user, err := db.GetUser(params.ID)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return response.NewApiNotFound("User not found")
        }
        return response.NewApiInternalError("Failed to fetch user")
    }
    return response.NewApiOk(user)
}
```

**Response** (404):
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

---

### 2. Validation Errors

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18"`
}

func CreateUser(req CreateUserRequest) *response.ApiHelper {
    // Validation automatically handled by framework
    // Returns ValidationError if validation fails
    return response.NewApiCreated(user, "User created")
}
```

**Response** (400):
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "email",
        "code": "REQUIRED",
        "message": "Email is required"
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

### 3. Custom Error Codes

```go
const (
    ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
    ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"
    ErrCodeDuplicateEntry    = "DUPLICATE_ENTRY"
)

func Transfer(req *TransferRequest) *response.ApiHelper {
    balance := getBalance(req.FromAccount)
    if balance < req.Amount {
        return response.NewApiBadRequest(
            ErrCodeInsufficientFunds,
            "Insufficient funds for transfer",
        )
    }
    // ... process transfer
    return response.NewApiOk(result)
}
```

---

## Error Handling Patterns

### Pattern 1: Early Return

```go
func ProcessOrder(req *OrderRequest) *response.ApiHelper {
    // Validate stock
    if !hasStock(req.ProductID, req.Quantity) {
        return response.NewApiBadRequest("OUT_OF_STOCK", "Product out of stock")
    }
    
    // Validate payment
    if !validatePayment(req.Payment) {
        return response.NewApiBadRequest("INVALID_PAYMENT", "Payment validation failed")
    }
    
    // Process order
    order, err := createOrder(req)
    if err != nil {
        return response.NewApiInternalError("Failed to create order")
    }
    
    return response.NewApiCreated(order, "Order created successfully")
}
```

---

### Pattern 2: Error Wrapping

```go
import "fmt"

func GetUserProfile(userID string) *response.ApiHelper {
    user, err := db.GetUser(userID)
    if err != nil {
        log.Printf("GetUserProfile: failed to fetch user %s: %v", userID, err)
        return response.NewApiInternalError("Failed to fetch user profile")
    }
    
    posts, err := db.GetUserPosts(userID)
    if err != nil {
        log.Printf("GetUserProfile: failed to fetch posts for user %s: %v", userID, err)
        // Non-critical error - continue with empty posts
        posts = []Post{}
    }
    
    return response.NewApiOk(map[string]any{
        "user":  user,
        "posts": posts,
    })
}
```

---

### Pattern 3: Error Middleware

```go
func ErrorRecoveryMiddleware(c *lokstra.RequestContext) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("PANIC: %v\n%s", r, debug.Stack())
            
            // Return 500 error
            c.Resp.WithStatus(500).Json(map[string]any{
                "status": "error",
                "error": map[string]any{
                    "code":    "INTERNAL_ERROR",
                    "message": "Internal server error",
                },
            })
        }
    }()
    
    return c.Next()
}
```

---

## HTTP Status Codes

| Status | Helper Method | Use Case |
|--------|--------------|----------|
| 200 | `NewApiOk()` | Successful operation |
| 201 | `NewApiCreated()` | Resource created |
| 400 | `NewApiBadRequest()` | Invalid input |
| 401 | `NewApiUnauthorized()` | Authentication required |
| 403 | `NewApiForbidden()` | Permission denied |
| 404 | `NewApiNotFound()` | Resource not found |
| 422 | `NewApiValidationError()` | Validation failed |
| 429 | `NewApiError(429, ...)` | Rate limit exceeded |
| 500 | `NewApiInternalError()` | Server error |
| 503 | `NewApiError(503, ...)` | Service unavailable |

---

## Best Practices

### ✅ Do

```go
// Use specific error codes
return response.NewApiBadRequest("INVALID_EMAIL", "Email format is invalid")

// Log internal errors, return generic message
log.Printf("Database error: %v", err)
return response.NewApiInternalError("Failed to process request")

// Differentiate between client and server errors
if validationFailed {
    return response.NewApiBadRequest("VALIDATION_ERROR", "Invalid input")
}
if databaseFailed {
    return response.NewApiInternalError("Database error")
}

// Use context errors for authentication
if !authenticated {
    return response.NewApiUnauthorized("Authentication required")
}
```

### ❌ Don't

```go
// Don't expose internal errors
return response.NewApiInternalError(err.Error())  // ❌ Leaks internal details

// Don't use generic error codes
return response.NewApiBadRequest("ERROR", "Something went wrong")  // ❌ Not helpful

// Don't ignore errors
result, _ := processPayment()  // ❌ Always handle errors

// Don't mix HTTP and application logic
if err != nil {
    return response.NewApiError(200, "ERROR", "Failed")  // ❌ Wrong status
}
```

---

## Error Categories

### Client Errors (4xx)

**User's fault** - Invalid input, missing auth, etc.

```go
// 400 - Bad input
response.NewApiBadRequest("INVALID_INPUT", message)

// 401 - Not authenticated
response.NewApiUnauthorized("Please login")

// 403 - Authenticated but no permission
response.NewApiForbidden("Admin access required")

// 404 - Resource doesn't exist
response.NewApiNotFound("User not found")

// 422 - Validation failed
response.NewApiValidationError("Validation failed", fields)
```

### Server Errors (5xx)

**Server's fault** - Database errors, external service failures, etc.

```go
// 500 - Generic server error
response.NewApiInternalError("Failed to process request")

// 503 - Service temporarily unavailable
response.NewApiError(503, "SERVICE_UNAVAILABLE", "Database maintenance")
```

---

## Error Logging

### Development

```go
func GetUser(id string) *response.ApiHelper {
    user, err := db.GetUser(id)
    if err != nil {
        // Verbose logging in development
        log.Printf("ERROR: GetUser(%s) failed: %v", id, err)
        log.Printf("Stack: %s", debug.Stack())
        return response.NewApiInternalError("Failed to fetch user")
    }
    return response.NewApiOk(user)
}
```

### Production

```go
func GetUser(id string) *response.ApiHelper {
    user, err := db.GetUser(id)
    if err != nil {
        // Structured logging in production
        log.Printf("GetUser error: user_id=%s error=%v", id, err)
        
        // Send to error tracking (Sentry, Rollbar, etc.)
        sentry.CaptureException(err)
        
        return response.NewApiInternalError("Failed to fetch user")
    }
    return response.NewApiOk(user)
}
```

---

## Running

```bash
go run main.go

# Test with test.http file
```

---

## Key Takeaways

✅ **Use ApiHelper methods** for consistent error responses  
✅ **Specific error codes** help clients handle errors  
✅ **Log internal errors**, return generic messages  
✅ **Differentiate 4xx (client) vs 5xx (server)** errors  
✅ **Validate early**, return errors immediately  
✅ **Never expose sensitive data** in error messages  
✅ **Use error middleware** for panic recovery

---

## Related Examples

- [02-parameter-binding](../02-parameter-binding/) - Validation errors
- [08-response-types](../08-response-types/) - Response patterns
- [03-lifecycle-hooks](../03-lifecycle-hooks/) - Error middleware
