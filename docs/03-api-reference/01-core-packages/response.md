# Response

> Response building and formatting helpers

## Overview

Lokstra provides two response helpers:
1. **`Response`** - Low-level response builder (flexible)
2. **`ApiHelper`** - High-level API response formatter (opinionated, recommended)

Both are available through the request context (`c.Resp` and `c.Api`) or via helper constructors.

## Import Path

```go
import "github.com/primadi/lokstra/core/response"

// Method 1: Helper constructors (quick one-liner)
func handler(params *Params) *response.Response {
    return response.NewJsonResponse(data)
}

func apiHandler(params *Params) *response.ApiHelper {
    return response.NewApiOk(data)
}

// Method 2: Context access (chainable methods)
func handler(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Json(data)  // Low-level
}

func apiHandler(c *lokstra.RequestContext) error {
    return c.Api.Ok(data)  // High-level
}
```

---

## ApiHelper (Recommended)

High-level helper that formats responses according to a standard API structure.

### Usage Methods

**Method 1: Helper Constructors (Quick One-Liner)**
```go
func getUser(params *getUserParams) *response.ApiHelper {
    user := getUserFromDB(params.id)
    return response.NewApiOk(user)  // Returns *ApiHelper directly
}
```

**Method 2: Context Methods (Chainable)**
```go
func getUser(c *lokstra.RequestContext) error {
    user := getUserFromDB(c.Req.Param("id"))
    return c.Api.Ok(user)  // Returns error
}
```

**Method 3: Manual Creation**
```go
func getUser() *response.ApiHelper {
    user := getUserFromDB(id)
    api := response.NewApiHelper()
    api.Ok(user)
    return api
}
```

> **Recommendation**: Use **Method 1** (helper constructors) for clean one-liner returns, or **Method 2** (context methods) when you need request context.

---

### Success Responses

#### Ok
Sends successful response with data.

**Signatures:**
```go
// Constructor (returns *ApiHelper)
func NewApiOk(data any) *ApiHelper

// Context method (returns error)
func (a *ApiHelper) Ok(data any) error
```

**Parameters:**
- `data` - Response data (any JSON-serializable type)

**Examples:**
```go
// Using constructor
func getUser(params *getUserParams) *response.ApiHelper {
    user := getUserFromDB(params.id)
    return response.NewApiOk(user)
}

// Using context
func getUser(c *lokstra.RequestContext) error {
    user := getUserFromDB(c.Req.Param("id"))
    return c.Api.Ok(user)
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

**Signatures:**
```go
// Constructor
func NewApiOkWithMessage(data any, message string) *ApiHelper

// Context method
func (a *ApiHelper) OkWithMessage(data any, message string) error
```

**Examples:**
```go
// Using constructor
func updateUser(params *updateUserParams) *response.ApiHelper {
    user := updateUserInDB(params.id, params.input)
    return response.NewApiOkWithMessage(user, "User updated successfully")
}

// Using context
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

**Signatures:**
```go
// Constructor
func NewApiCreated(data any, message string) *ApiHelper

// Context method
func (a *ApiHelper) Created(data any, message string) error
```

**Examples:**
```go
// Using constructor
func createUser(input *CreateUserInput) *response.ApiHelper {
    user := createUserInDB(input)
    return response.NewApiCreated(user, "User created successfully")
}

// Using context
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

---

### List Responses

#### OkList
Sends paginated list response.

**Signatures:**
```go
// Constructor
func NewApiOkList(data any, meta *api_formatter.ListMeta) *ApiHelper

// Context method
func (a *ApiHelper) OkList(data any, meta *api_formatter.ListMeta) error
```

**Parameters:**
- `data` - List data (slice)
- `meta` - Pagination metadata

**Examples:**
```go
// Using constructor
func listUsers(params *listUsersParams) *response.ApiHelper {
    users := getUsersFromDB(params.page, params.limit)
    total := countUsers()
    
    meta := &api_formatter.ListMeta{
        Page:       params.page,
        Limit:      params.limit,
        Total:      total,
        TotalPages: (total + params.limit - 1) / params.limit,
    }
    
    return response.NewApiOkList(users, meta)
}

// Using context
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

**Signatures:**
```go
// Constructor
func NewApiBadRequest(code, message string) *ApiHelper

// Context method
func (a *ApiHelper) BadRequest(code, message string) error
```

**Examples:**
```go
// Using constructor
func validateInput(input *Input) *response.ApiHelper {
    if input.Amount <= 0 {
        return response.NewApiBadRequest("INVALID_AMOUNT", "Amount must be positive")
    }
    // ...
}

// Using context
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

**Signatures:**
```go
// Constructor
func NewApiUnauthorized(message string) *ApiHelper

// Context method
func (a *ApiHelper) Unauthorized(message string) error
```

**Examples:**
```go
// Using constructor
func checkAuth(token string) *response.ApiHelper {
    if token == "" {
        return response.NewApiUnauthorized("Missing authorization token")
    }
    // ...
}

// Using context
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

**Signatures:**
```go
// Constructor
func NewApiForbidden(message string) *ApiHelper

// Context method
func (a *ApiHelper) Forbidden(message string) error
```

**Examples:**
```go
// Using constructor
func checkPermission(user *User) *response.ApiHelper {
    if !user.IsAdmin {
        return response.NewApiForbidden("Admin access required")
    }
    // ...
}

// Using context
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

**Signatures:**
```go
// Constructor
func NewApiNotFound(message string) *ApiHelper

// Context method
func (a *ApiHelper) NotFound(message string) error
```

**Examples:**
```go
// Using constructor
func getUser(params *getUserParams) *response.ApiHelper {
    user, err := getUserFromDB(params.id)
    if err != nil {
        return response.NewApiNotFound("User not found")
    }
    return response.NewApiOk(user)
}

// Using context
func getUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    user, err := getUserFromDB(id)
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    return c.Api.Ok(user)
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

**Signatures:**
```go
// Constructor
func NewApiInternalError(message string) *ApiHelper

// Context method
func (a *ApiHelper) InternalError(message string) error
```

**Examples:**
```go
// Using constructor
func processData(data *Data) *response.ApiHelper {
    if err := processingLogic(data); err != nil {
        log.Printf("Processing error: %v", err)
        return response.NewApiInternalError("Failed to process data")
    }
    return response.NewApiOk(result)
}

// Using context
func handler(c *lokstra.RequestContext) error {
    user, err := getUserFromDB(id)
    if err != nil {
        log.Printf("Database error: %v", err)
        return c.Api.InternalError("Failed to fetch user")
    }
    return c.Api.Ok(user)
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

**Signatures:**
```go
// Constructor
func NewApiValidationError(message string, fields []api_formatter.FieldError) *ApiHelper

// Context method
func (a *ApiHelper) ValidationError(message string, fields []api_formatter.FieldError) error
```

**Examples:**
```go
// Using constructor
func validateUser(input *CreateUserInput) *response.ApiHelper {
    var fieldErrors []api_formatter.FieldError
    
    if !isValidEmail(input.Email) {
        fieldErrors = append(fieldErrors, api_formatter.FieldError{
            Field:   "email",
            Code:    "INVALID_FORMAT",
            Message: "Email format is invalid",
        })
    }
    
    if len(fieldErrors) > 0 {
        return response.NewApiValidationError("Validation failed", fieldErrors)
    }
    // ...
}

// Using context (automatic via BindJSON)
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

**Signatures:**
```go
// Constructor
func NewApiError(statusCode int, code, message string) *ApiHelper

// Context method
func (a *ApiHelper) Error(statusCode int, code, message string) error
```

**Examples:**
```go
// Using constructor
func checkQuota() *response.ApiHelper {
    if quota.Exceeded() {
        return response.NewApiError(429, "QUOTA_EXCEEDED", "API quota exceeded")
    }
    // ...
}

// Using context
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

### Helper Constructors

Quick one-liner constructors for common response types:

```go
// JSON response
func NewJsonResponse(data any) *Response

// HTML response
func NewHtmlResponse(html string) *Response

// Plain text response
func NewTextResponse(text string) *Response

// Custom content-type (CSV, XML, PDF, etc.)
func NewRawResponse(contentType string, b []byte) *Response

// Streaming response (SSE, chunked transfer)
func NewStreamResponse(contentType string, fn func(w http.ResponseWriter) error) *Response
```

**Examples:**
```go
// JSON
func getUsers() *response.Response {
    users := getUsersFromDB()
    return response.NewJsonResponse(users)
}

// HTML
func homepage() *response.Response {
    html := "<html><body><h1>Welcome</h1></body></html>"
    return response.NewHtmlResponse(html)
}

// Text
func healthCheck() *response.Response {
    return response.NewTextResponse("OK")
}

// CSV
func exportData() *response.Response {
    csvData := generateCSV()
    return response.NewRawResponse("text/csv", csvData)
}

// Stream
func streamFile() *response.Response {
    return response.NewStreamResponse("application/octet-stream", func(w http.ResponseWriter) error {
        return streamFileContent(w)
    })
}
```

---

### Chainable Methods

#### WithStatus
Sets HTTP status code.

**Signature:**
```go
func (r *Response) WithStatus(code int) *Response
```

**Example:**
```go
// Chainable method
func handler(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Json(data)
}

// Or manual creation
func handler() *response.Response {
    r := response.NewResponse()
    return r.WithStatus(201).Json(data)
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

#### Json
Sends JSON response.

**Signature:**
```go
func (r *Response) Json(data any) error
```

**Examples:**
```go
// Using helper constructor (recommended)
func handler(params *Params) *response.Response {
    data := map[string]any{
        "message": "Hello",
        "users":   users,
    }
    return response.NewJsonResponse(data)
}

// Using context (chainable)
func handler(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Json(data)
}
```

---

#### Html
Sends HTML response.

**Signature:**
```go
func (r *Response) Html(html string) error
```

**Examples:**
```go
// Using helper constructor (recommended)
func homepage() *response.Response {
    html := `<html><body><h1>Welcome</h1></body></html>`
    return response.NewHtmlResponse(html)
}

// Using context (chainable)
func homepage(c *lokstra.RequestContext) error {
    html := `<html><body><h1>Welcome</h1></body></html>`
    return c.Resp.WithStatus(200).Html(html)
}
```

---

#### Text
Sends plain text response.

**Signature:**
```go
func (r *Response) Text(text string) error
```

**Examples:**
```go
// Using helper constructor (recommended)
func healthCheck() *response.Response {
    return response.NewTextResponse("OK")
}

// Using context (chainable)
func healthCheck(c *lokstra.RequestContext) error {
    return c.Resp.WithStatus(200).Text("OK")
}
```

---

#### Raw
Sends raw bytes with custom content type.

**Signature:**
```go
func (r *Response) Raw(contentType string, b []byte) error
```

**Examples:**
```go
// Using helper constructor (recommended)
func downloadFile(filename string) *response.Response {
    data := readFileBytes(filename)
    return response.NewRawResponse("application/pdf", data)
}

// Using context (chainable)
func downloadFile(c *lokstra.RequestContext) error {
    data := readFileBytes(filename)
    return c.Resp.WithStatus(200).Raw("application/pdf", data)
}
```

---

#### Stream
Streams response using custom writer function.

**Signature:**
```go
func (r *Response) Stream(contentType string, fn func(w http.ResponseWriter) error) error
```

**Examples:**
```go
// Using helper constructor (recommended)
func streamLargeFile(filename string) *response.Response {
    return response.NewStreamResponse("application/octet-stream", func(w http.ResponseWriter) error {
        file, err := os.Open(filename)
        if err != nil {
            return err
        }
        defer file.Close()
        
        _, err = io.Copy(w, file)
        return err
    })
}

// Using context (chainable)
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

#### Using Helper Constructors (Recommended for handlers without context)
```go
func listUsers(params *listUsersParams) *response.ApiHelper {
    users := getUsersFromDB(params.page, params.limit)
    total := countUsers()
    
    meta := &api_formatter.ListMeta{
        Page:       params.page,
        Limit:      params.limit,
        Total:      total,
        TotalPages: (total + params.limit - 1) / params.limit,
    }
    
    return response.NewApiOkList(users, meta)
}

func getUser(params *getUserParams) *response.ApiHelper {
    user, err := getUserFromDB(params.id)
    if err != nil {
        return response.NewApiNotFound("User not found")
    }
    return response.NewApiOk(user)
}

func createUser(input *CreateUserInput) *response.ApiHelper {
    user, err := createUserInDB(input)
    if err != nil {
        return response.NewApiInternalError("Failed to create user")
    }
    return response.NewApiCreated(user, "User created successfully")
}

func updateUser(params *updateUserParams, input *UpdateUserInput) *response.ApiHelper {
    user, err := updateUserInDB(params.id, input)
    if err != nil {
        return response.NewApiInternalError("Failed to update user")
    }
    return response.NewApiOkWithMessage(user, "User updated successfully")
}

func deleteUser(params *deleteUserParams) *response.ApiHelper {
    if err := deleteUserFromDB(params.id); err != nil {
        return response.NewApiInternalError("Failed to delete user")
    }
    // Note: NoContent doesn't have a constructor, use context method or manual
    api := response.NewApiHelper()
    api.resp.WithStatus(http.StatusNoContent)
    return api
}
```

#### Using Context Methods (When you need request context)
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
    return c.Api.Ok(user)
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
```

### Custom Response Format (Low-Level)

#### Using Helper Constructors (Recommended)
```go
func customResponse() *response.Response {
    // Custom JSON structure
    data := map[string]any{
        "version":   "1.0",
        "timestamp": time.Now().Unix(),
        "data": map[string]any{
            "users": users,
            "count": len(users),
        },
    }
    
    return response.NewJsonResponse(data)
}

func xmlResponse() *response.Response {
    xml := `<?xml version="1.0"?>
    <users>
        <user id="1">John</user>
        <user id="2">Jane</user>
    </users>`
    
    return response.NewRawResponse("application/xml", []byte(xml))
}

func csvExport() *response.Response {
    csv := generateCSVBytes()
    return response.NewRawResponse("text/csv", csv)
}
```

#### Using Context Methods
```go
func customResponse(c *lokstra.RequestContext) error {
    data := map[string]any{
        "version":   "1.0",
        "timestamp": time.Now().Unix(),
        "data": map[string]any{
            "users": users,
            "count": len(users),
        },
    }
    
    return c.Resp.WithStatus(200).Json(data)
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

#### Using Helper Constructor
```go
func downloadReport(reportID string) *response.Response {
    // Generate or fetch report
    data := generateReport(reportID)
    
    // Create response with helper
    r := response.NewRawResponse("application/pdf", data)
    
    // Set headers for download
    r.RespHeaders = map[string][]string{
        "Content-Disposition": {fmt.Sprintf("attachment; filename=report-%s.pdf", reportID)},
    }
    
    return r
}
```

#### Using Context Method
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

#### Using Helper Constructor
```go
func streamEvents() *response.Response {
    return response.NewStreamResponse("text/event-stream", func(w http.ResponseWriter) error {
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

#### Using Context Method
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

### 1. Use Helper Constructors for Clean Code
```go
// ✅ Recommended: Helper constructors (one-liner)
func getUser(params *getUserParams) *response.ApiHelper {
    return response.NewApiOk(user)
}

func homepage() *response.Response {
    return response.NewHtmlResponse(html)
}

// ✅ Good: Context methods when you need request context
func getUser(c *lokstra.RequestContext) error {
    return c.Api.Ok(user)
}

// 🚫 Avoid: Manual creation when helpers exist
func getUser() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Ok(user)
    return api
}
```

### 2. Use ApiHelper for REST APIs
```go
// ✅ Recommended: ApiHelper for structured API responses
return response.NewApiOk(user)
return c.Api.Ok(user)

// 🚫 Avoid: Manual JSON structure for standard APIs
return response.NewJsonResponse(map[string]any{"data": user})
```

### 3. Choose the Right Response Type
```go
// ✅ Good: Use Response for custom formats
func homepage() *response.Response {
    return response.NewHtmlResponse(html)
}

func exportCSV() *response.Response {
    return response.NewRawResponse("text/csv", csvData)
}

// ✅ Good: Use ApiHelper for REST APIs
func getUser(params *getUserParams) *response.ApiHelper {
    return response.NewApiOk(user)
}
```

### 4. Consistent Error Codes
```go
// ✅ Good: Use consistent error codes
const (
    ErrInvalidInput   = "INVALID_INPUT"
    ErrUserNotFound   = "USER_NOT_FOUND"
    ErrUnauthorized   = "UNAUTHORIZED"
)

return response.NewApiError(400, ErrInvalidInput, "Invalid user data")

// 🚫 Avoid: Random error messages
return response.NewApiBadRequest("error123", "something went wrong")
```

### 5. Don't Log Sensitive Data in Responses
```go
// ✅ Good
log.Printf("Failed to authenticate user: %v", err)
return response.NewApiUnauthorized("Authentication failed")

// 🚫 Avoid
return response.NewApiUnauthorized(fmt.Sprintf("Auth failed: %v", err))
```

---

## See Also

- **[Request Context](request)** - Request handling
- **[Router](router)** - Handler registration
- **[API Formatter](../08-advanced/api-formatter)** - Custom formatters

---

## Related Guides

- **[Router Essentials](../../01-router-guide/01-router/)** - Handler basics
- **[API Design](../../02-deep-dive/router/)** - Best practices
