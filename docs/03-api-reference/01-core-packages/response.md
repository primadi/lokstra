# Response

> Response building and formatting helpers

## Overview

Lokstra provides two response helpers:
1. **`Response`** - Low-level response builder (flexible)
2. **`ApiHelper`** - High-level API response formatter (opinionated, recommended)

Both are available through the request context (`c.Resp` and `c.Api`).

## Import Path

```go
import "github.com/primadi/lokstra/core/response"

// Access via context (recommended)
func handler(c *lokstra.RequestContext) error {
    return c.Api.Success(data)  // High-level
    // or
    return c.Resp.WithStatus(200).Json(data)  // Low-level
}
```

---

## ApiHelper (Recommended)

High-level helper that formats responses according to a standard API structure.

### Success Responses

#### Success / Ok
Sends successful response with data.

**Signature:**
```go
func (a *ApiHelper) Success(data any) error
func (a *ApiHelper) Ok(data any) error  // Alias
```

**Parameters:**
- `data` - Response data (any JSON-serializable type)

**Returns:**
- `error` - Always returns `nil` (framework requirement)

**Example:**
```go
func getUser(c *lokstra.RequestContext) error {
    user := getUserFromDB(id)
    return c.Api.Success(user)
}

// Response (HTTP 200):
// {
//   "status": "success",
//   "data": { "id": 1, "name": "John" }
// }
```

---

#### OkWithMessage
Sends successful response with message and data.

**Signature:**
```go
func (a *ApiHelper) OkWithMessage(data any, message string) error
```

**Example:**
```go
func updateUser(c *lokstra.RequestContext) error {
    user := updateUserInDB(id, input)
    return c.Api.OkWithMessage(user, "User updated successfully")
}

// Response (HTTP 200):
// {
//   "status": "success",
//   "message": "User updated successfully",
//   "data": { "id": 1, "name": "John" }
// }
```

---

#### Created
Sends 201 Created response.

**Signature:**
```go
func (a *ApiHelper) Created(data any, message string) error
```

**Example:**
```go
func createUser(c *lokstra.RequestContext, input *CreateUserInput) error {
    user := createUserInDB(input)
    return c.Api.Created(user, "User created successfully")
}

// Response (HTTP 201):
// {
//   "status": "success",
//   "message": "User created successfully",
//   "data": { "id": 1, "name": "John" }
// }
```

---

#### NoContent
Sends 204 No Content response.

**Signature:**
```go
func (a *ApiHelper) NoContent() error
```

**Example:**
```go
func deleteUser(c *lokstra.RequestContext) error {
    deleteUserFromDB(id)
    return c.Api.NoContent()
}

// Response (HTTP 204):
// (empty body)
```

---

### List Responses

#### OkList
Sends paginated list response.

**Signature:**
```go
func (a *ApiHelper) OkList(data any, meta *api_formatter.ListMeta) error
```

**Parameters:**
- `data` - List data (slice)
- `meta` - Pagination metadata

**Example:**
```go
func listUsers(c *lokstra.RequestContext) error {
    users := getUsersFromDB(page, limit)
    total := countUsers()
    
    meta := &api_formatter.ListMeta{
        Page:       page,
        Limit:      limit,
        Total:      total,
        TotalPages: (total + limit - 1) / limit,
    }
    
    return c.Api.OkList(users, meta)
}

// Response (HTTP 200):
// {
//   "status": "success",
//   "data": [
//     { "id": 1, "name": "John" },
//     { "id": 2, "name": "Jane" }
//   ],
//   "meta": {
//     "page": 1,
//     "limit": 20,
//     "total": 42,
//     "total_pages": 3
//   }
// }
```

---

### Error Responses

#### BadRequest
Sends 400 Bad Request error.

**Signature:**
```go
func (a *ApiHelper) BadRequest(code, message string) error
```

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    if input.Amount <= 0 {
        return c.Api.BadRequest("INVALID_AMOUNT", "Amount must be positive")
    }
    // ...
}

// Response (HTTP 400):
// {
//   "status": "error",
//   "error": {
//     "code": "INVALID_AMOUNT",
//     "message": "Amount must be positive"
//   }
// }
```

---

#### Unauthorized
Sends 401 Unauthorized error.

**Signature:**
```go
func (a *ApiHelper) Unauthorized(message string) error
```

**Example:**
```go
func authMiddleware(c *lokstra.RequestContext) error {
    token := c.Req.Header("Authorization")
    if token == "" {
        return c.Api.Unauthorized("Missing authorization token")
    }
    return c.Next()
}

// Response (HTTP 401):
// {
//   "status": "error",
//   "error": {
//     "code": "UNAUTHORIZED",
//     "message": "Missing authorization token"
//   }
// }
```

---

#### Forbidden
Sends 403 Forbidden error.

**Signature:**
```go
func (a *ApiHelper) Forbidden(message string) error
```

**Example:**
```go
func deleteUser(c *lokstra.RequestContext) error {
    user := c.Get("user").(*User)
    if !user.IsAdmin {
        return c.Api.Forbidden("Admin access required")
    }
    // ...
}

// Response (HTTP 403):
// {
//   "status": "error",
//   "error": {
//     "code": "FORBIDDEN",
//     "message": "Admin access required"
//   }
// }
```

---

#### NotFound
Sends 404 Not Found error.

**Signature:**
```go
func (a *ApiHelper) NotFound(message string) error
```

**Example:**
```go
func getUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    user, err := getUserFromDB(id)
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    return c.Api.Success(user)
}

// Response (HTTP 404):
// {
//   "status": "error",
//   "error": {
//     "code": "NOT_FOUND",
//     "message": "User not found"
//   }
// }
```

---

#### InternalError
Sends 500 Internal Server Error.

**Signature:**
```go
func (a *ApiHelper) InternalError(message string) error
```

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    user, err := getUserFromDB(id)
    if err != nil {
        log.Printf("Database error: %v", err)
        return c.Api.InternalError("Failed to fetch user")
    }
    return c.Api.Success(user)
}

// Response (HTTP 500):
// {
//   "status": "error",
//   "error": {
//     "code": "INTERNAL_ERROR",
//     "message": "Failed to fetch user"
//   }
// }
```

---

#### ValidationError
Sends 400 validation error with field details.

**Signature:**
```go
func (a *ApiHelper) ValidationError(message string, fields []api_formatter.FieldError) error
```

**Example:**
```go
func createUser(c *lokstra.RequestContext) error {
    var input CreateUserInput
    if err := c.Req.BindJSON(&input); err != nil {
        // BindJSON automatically returns ValidationError
        return err
    }
    // ...
}

// Response (HTTP 400):
// {
//   "status": "error",
//   "error": {
//     "code": "VALIDATION_ERROR",
//     "message": "Validation failed",
//     "fields": [
//       {
//         "field": "email",
//         "code": "INVALID_FORMAT",
//         "message": "Email format is invalid"
//       },
//       {
//         "field": "age",
//         "code": "MIN_VALUE",
//         "message": "Age must be at least 18"
//       }
//     ]
//   }
// }
```

---

#### Error (Generic)
Sends error response with custom status code and error code.

**Signature:**
```go
func (a *ApiHelper) Error(statusCode int, code, message string) error
```

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    if quota.Exceeded() {
        return c.Api.Error(429, "QUOTA_EXCEEDED", "API quota exceeded")
    }
    // ...
}

// Response (HTTP 429):
// {
//   "status": "error",
//   "error": {
//     "code": "QUOTA_EXCEEDED",
//     "message": "API quota exceeded"
//   }
// }
```

---

## Response (Low-Level)

Low-level response builder for custom response formats.

### WithStatus
Sets HTTP status code.

**Signature:**
```go
func (r *Response) WithStatus(code int) *Response
```

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Json(data)
}
```

**Status Code Constants:**
```go
http.StatusOK                  // 200
http.StatusCreated             // 201
http.StatusAccepted            // 202
http.StatusNoContent           // 204
http.StatusBadRequest          // 400
http.StatusUnauthorized        // 401
http.StatusForbidden           // 403
http.StatusNotFound            // 404
http.StatusInternalServerError // 500
// ... see net/http package for full list
```

---

### Json
Sends JSON response.

**Signature:**
```go
func (r *Response) Json(data any) error
```

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    data := map[string]any{
        "message": "Hello",
        "users":   users,
    }
    return c.Resp.WithStatus(200).Json(data)
}
```

---

### Html
Sends HTML response.

**Signature:**
```go
func (r *Response) Html(html string) error
```

**Example:**
```go
func homepage(c *lokstra.RequestContext) error {
    html := `<html><body><h1>Welcome</h1></body></html>`
    return c.Resp.WithStatus(200).Html(html)
}
```

---

### Text
Sends plain text response.

**Signature:**
```go
func (r *Response) Text(text string) error
```

**Example:**
```go
func healthCheck(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Text("OK")
}
```

---

### Raw
Sends raw bytes with custom content type.

**Signature:**
```go
func (r *Response) Raw(contentType string, b []byte) error
```

**Example:**
```go
func downloadFile(c *lokstra.RequestContext) error {
    data := readFileBytes(filename)
    return c.Resp.WithStatus(200).Raw("application/pdf", data)
}
```

---

### Stream
Streams response using custom writer function.

**Signature:**
```go
func (r *Response) Stream(contentType string, fn func(w http.ResponseWriter) error) error
```

**Example:**
```go
func streamLargeFile(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Stream("application/octet-stream", func(w http.ResponseWriter) error {
        file, err := os.Open(filename)
        if err != nil {
            return err
        }
        defer file.Close()
        
        _, err = io.Copy(w, file)
        return err
    })
}
```

---

## Complete Examples

### CRUD API with ApiHelper
```go
func listUsers(c *lokstra.RequestContext) error {
    page := getIntParam(c.Req.Query("page", "1"))
    limit := getIntParam(c.Req.Query("limit", "20"))
    
    users := getUsersFromDB(page, limit)
    total := countUsers()
    
    meta := &api_formatter.ListMeta{
        Page:       page,
        Limit:      limit,
        Total:      total,
        TotalPages: (total + limit - 1) / limit,
    }
    
    return c.Api.OkList(users, meta)
}

func getUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    user, err := getUserFromDB(id)
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    return c.Api.Success(user)
}

func createUser(c *lokstra.RequestContext, input *CreateUserInput) error {
    user, err := createUserInDB(input)
    if err != nil {
        return c.Api.InternalError("Failed to create user")
    }
    return c.Api.Created(user, "User created successfully")
}

func updateUser(c *lokstra.RequestContext, input *UpdateUserInput) error {
    id := c.Req.Param("id")
    user, err := updateUserInDB(id, input)
    if err != nil {
        return c.Api.InternalError("Failed to update user")
    }
    return c.Api.OkWithMessage(user, "User updated successfully")
}

func deleteUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    if err := deleteUserFromDB(id); err != nil {
        return c.Api.InternalError("Failed to delete user")
    }
    return c.Api.NoContent()
}
```

### Custom Response Format (Low-Level)
```go
func customResponse(c *lokstra.RequestContext) error {
    // Custom JSON structure
    response := map[string]any{
        "version": "1.0",
        "timestamp": time.Now().Unix(),
        "data": map[string]any{
            "users": users,
            "count": len(users),
        },
    }
    
    return c.Resp.WithStatus(200).Json(response)
}

func xmlResponse(c *lokstra.RequestContext) error {
    xml := `<?xml version="1.0"?>
    <users>
        <user id="1">John</user>
        <user id="2">Jane</user>
    </users>`
    
    return c.Resp.WithStatus(200).Raw("application/xml", []byte(xml))
}
```

### File Download
```go
func downloadReport(c *lokstra.RequestContext) error {
    reportID := c.Req.Param("id")
    
    // Generate or fetch report
    data := generateReport(reportID)
    
    // Set headers for download
    c.Resp.RespHeaders = map[string][]string{
        "Content-Disposition": {fmt.Sprintf("attachment; filename=report-%s.pdf", reportID)},
    }
    
    return c.Resp.WithStatus(200).Raw("application/pdf", data)
}
```

### Streaming Response
```go
func streamEvents(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Stream("text/event-stream", func(w http.ResponseWriter) error {
        flusher, ok := w.(http.Flusher)
        if !ok {
            return fmt.Errorf("streaming not supported")
        }
        
        for i := 0; i < 10; i++ {
            fmt.Fprintf(w, "data: Event %d\n\n", i)
            flusher.Flush()
            time.Sleep(1 * time.Second)
        }
        
        return nil
    })
}
```

---

## Response Formatters

### Custom Response Format
You can customize the API response format:

```go
import "github.com/primadi/lokstra/core/response"

// Set custom formatter globally
response.SetApiResponseFormatter(myCustomFormatter)

// Or use built-in formatters by name
response.SetApiResponseFormatterByName("default")
response.SetApiResponseFormatterByName("custom-name")
```

---

## Best Practices

### 1. Use ApiHelper for REST APIs
```go
// ✅ Recommended
return c.Api.Success(user)

// 🚫 Avoid (unless custom format needed)
return c.Resp.WithStatus(200).Json(map[string]any{"data": user})
```

### 2. Consistent Error Codes
```go
// ✅ Good: Use consistent error codes
const (
    ErrInvalidInput   = "INVALID_INPUT"
    ErrUserNotFound   = "USER_NOT_FOUND"
    ErrUnauthorized   = "UNAUTHORIZED"
)

return c.Api.Error(400, ErrInvalidInput, "Invalid user data")

// 🚫 Avoid: Random error messages
return c.Api.BadRequest("error123", "something went wrong")
```

### 3. Don't Log Sensitive Data in Responses
```go
// ✅ Good
log.Printf("Failed to authenticate user: %v", err)
return c.Api.Unauthorized("Authentication failed")

// 🚫 Avoid
return c.Api.Unauthorized(fmt.Sprintf("Auth failed: %v", err))
```

---

## See Also

- **[Request Context](request.md)** - Request handling
- **[Router](router.md)** - Handler registration
- **[API Formatter](../08-advanced/api-formatter.md)** - Custom formatters

---

## Related Guides

- **[Router Essentials](../../01-essentials/01-router/)** - Handler basics
- **[API Design](../../02-deep-dive/router/)** - Best practices
