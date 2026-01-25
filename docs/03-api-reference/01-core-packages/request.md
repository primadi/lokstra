# Request Context

> Request handling and context API

## Overview

The `Context` type (often referred to as `RequestContext`) is passed to every handler and middleware. It provides access to the HTTP request, response helpers, parameter extraction, request binding, and context storage.

This is the central object for handling HTTP requests in Lokstra.

## Import Path

```go
import "github.com/primadi/lokstra/core/request"

// Or use the main package type alias
import "github.com/primadi/lokstra"

func handler(c *lokstra.RequestContext) error {
    // Use context
}
```

---

## Type Definition

```go
type Context struct {
    context.Context  // Embedded standard context
    
    // Helpers
    Req  *RequestHelper      // Request parameter extraction
    Resp *response.Response  // Response building
    Api  *response.ApiHelper // Opinionated API responses
    
    // Primitives (advanced usage)
    W *writerWrapper  // Response writer
    R *http.Request   // Raw HTTP request
}
```

**Fields:**
- `Context` - Embedded standard Go context
- `Req` - Helper for extracting parameters, headers, body
- `Resp` - Low-level response builder
- `Api` - High-level API response helper (recommended)
- `W` - Raw response writer (for advanced use)
- `R` - Raw HTTP request (for direct access)

---

## Context Methods

### Next
Calls the next handler in the middleware chain.

**Signature:**
```go
func (c *Context) Next() error
```

**Returns:**
- `error` - Error from next handler/middleware

**Example:**
```go
func loggingMiddleware(c *lokstra.RequestContext) error {
    start := time.Now()
    
    // Call next handler
    err := c.Next()
    
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
    
    return err
}
```

**Use Cases:**
- Middleware that wraps handler execution
- Pre/post processing
- Timing, logging, metrics

---

### Set
Repositorys a value in the request context.

**Signature:**
```go
func (c *Context) Set(key string, value any)
```

**Parameters:**
- `key` - Storage key
- `value` - Value to repository

**Example:**
```go
// In middleware
func authMiddleware(c *lokstra.RequestContext) error {
    user := authenticateUser(c)
    c.Set("user", user)
    c.Set("user_id", user.ID)
    return c.Next()
}

// In handler
func getProfile(c *lokstra.RequestContext) error {
    user := c.Get("user").(*User)
    return c.Api.Success(user)
}
```

---

### Get
Retrieves a value from the request context.

**Signature:**
```go
func (c *Context) Get(key string) any
```

**Parameters:**
- `key` - Storage key

**Returns:**
- `any` - Repositoryd value, or `nil` if not found

**Example:**
```go
userID := c.Get("user_id").(int)
user := c.Get("user").(*User)

// Safe retrieval
if val := c.Get("optional_key"); val != nil {
    data := val.(string)
}
```

---

### SetContextValue
Repositorys a value in the standard Go context.

**Signature:**
```go
func (c *Context) SetContextValue(key string, value any)
```

**Parameters:**
- `key` - Context key
- `value` - Value to repository

**Example:**
```go
c.SetContextValue("request_id", requestID)
c.SetContextValue("trace_id", traceID)
```

**Use Cases:**
- Passing values to downstream services
- Request tracing
- Correlation IDs

---

### GetContextValue
Retrieves a value from the standard Go context.

**Signature:**
```go
func (c *Context) GetContextValue(key string) any
```

**Parameters:**
- `key` - Context key

**Returns:**
- `any` - Repositoryd value, or `nil` if not found

**Example:**
```go
requestID := c.GetContextValue("request_id").(string)
```

---

## RequestHelper (c.Req)

The `RequestHelper` provides methods for extracting request data.

### Parameter Extraction

#### Param
Alias for `PathParam` - extracts path parameter.

**Signature:**
```go
func (h *RequestHelper) Param(name string) string
```

**Example:**
```go
// Route: GET /users/:id
id := c.Req.Param("id")
```

---

#### PathParam
Extracts path parameter with default value.

**Signature:**
```go
func (h *RequestHelper) PathParam(name string, defaultValue string) string
```

**Example:**
```go
// Route: GET /users/:id
id := c.Req.PathParam("id", "")
action := c.Req.PathParam("action", "view")
```

---

#### QueryParam
Extracts query parameter with default value.

**Signature:**
```go
func (h *RequestHelper) QueryParam(name string, defaultValue string) string
```

**Example:**
```go
// Request: GET /users?status=active&limit=10
status := c.Req.QueryParam("status", "all")
limit := c.Req.QueryParam("limit", "20")
page := c.Req.QueryParam("page", "1")
```

---

#### FormParam
Extracts form parameter with default value.

**Signature:**
```go
func (h *RequestHelper) FormParam(name string, defaultValue string) string
```

**Example:**
```go
// POST form data
username := c.Req.FormParam("username", "")
email := c.Req.FormParam("email", "")
```

---

#### HeaderParam
Extracts header value with default value.

**Signature:**
```go
func (h *RequestHelper) HeaderParam(name string, defaultValue string) string
```

**Example:**
```go
token := c.Req.HeaderParam("Authorization", "")
contentType := c.Req.HeaderParam("Content-Type", "application/json")
userAgent := c.Req.HeaderParam("User-Agent", "unknown")
```

---

### Multiple Values

#### QueryParams
Extracts all values for a query parameter.

**Signature:**
```go
func (h *RequestHelper) QueryParams(name string) []string
```

**Example:**
```go
// Request: GET /search?tags=go&tags=web&tags=api
tags := c.Req.QueryParams("tags")
// tags = ["go", "web", "api"]
```

---

#### HeaderValues
Extracts all values for a header.

**Signature:**
```go
func (h *RequestHelper) HeaderValues(name string) []string
```

**Example:**
```go
acceptValues := c.Req.HeaderValues("Accept")
// acceptValues = ["application/json", "text/html"]
```

---

#### AllQueryParams
Returns all query parameters as a map.

**Signature:**
```go
func (h *RequestHelper) AllQueryParams() map[string][]string
```

**Example:**
```go
params := c.Req.AllQueryParams()
for key, values := range params {
    log.Printf("%s: %v", key, values)
}
```

---

#### AllHeaders
Returns all headers as a map.

**Signature:**
```go
func (h *RequestHelper) AllHeaders() map[string][]string
```

**Example:**
```go
headers := c.Req.AllHeaders()
for name, values := range headers {
    log.Printf("%s: %v", name, values)
}
```

---

### Convenience Aliases

#### Query (Alias for QueryParam)
```go
func (h *RequestHelper) Query(name string, defaultValue string) string
```

#### Header (Alias for HeaderParam)
```go
func (h *RequestHelper) Header(name string) string
```

#### Form (Alias for FormParam)
```go
func (h *RequestHelper) Form(name string, defaultValue string) string
```

**Example:**
```go
// Shorter syntax
status := c.Req.Query("status", "all")
token := c.Req.Header("Authorization")
username := c.Req.Form("username", "")
```

---

### Request Body

#### RawRequestBody
Returns the raw request body as bytes.

**Signature:**
```go
func (h *RequestHelper) RawRequestBody() ([]byte, error)
```

**Returns:**
- `[]byte` - Request body
- `error` - Error if reading fails

**Example:**
```go
body, err := c.Req.RawRequestBody()
if err != nil {
    return err
}
log.Printf("Raw body: %s", string(body))
```

**Notes:**
- Body is cached automatically
- Can be called multiple times safely

---

#### BindJSON
Binds JSON request body to a struct.

**Signature:**
```go
func (h *RequestHelper) BindJSON(v any) error
```

**Parameters:**
- `v` - Pointer to struct to bind to

**Returns:**
- `error` - Binding or validation error

**Example:**
```go
type CreateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18"`
}

func createUser(c *lokstra.RequestContext) error {
    var input CreateUserInput
    if err := c.Req.BindJSON(&input); err != nil {
        return err // Auto-formatted as 400 Bad Request
    }
    
    // input is validated and ready to use
    user := saveUser(input)
    return c.Api.Created(user)
}
```

**Automatic Features:**
- JSON parsing
- Struct tag validation (`validate` tags)
- Friendly error messages
- Auto 400 response on error

---

### Binding Helpers

#### BindQuery
Binds query parameters to a struct.

**Example:**
```go
type ListUsersFilter struct {
    Status string   `query:"status"`
    Tags   []string `query:"tags"`
    Limit  int      `query:"limit"`
    Page   int      `query:"page"`
}

func listUsers(c *lokstra.RequestContext) error {
    var filter ListUsersFilter
    if err := c.Req.BindQuery(&filter); err != nil {
        return err
    }
    
    users := queryUsers(filter)
    return c.Api.Success(users)
}
```

**Struct Tags:**
- `query:"name"` - Query parameter name
- `validate:"required"` - Validation rules

---

#### BindPath
Binds path parameters to a struct.

**Example:**
```go
type UserParams struct {
    ID     string `path:"id" validate:"required"`
    Action string `path:"action"`
}

func userAction(c *lokstra.RequestContext) error {
    var params UserParams
    if err := c.Req.BindPath(&params); err != nil {
        return err
    }
    
    // params.ID and params.Action are populated
    return handleAction(params)
}
```

---

#### BindHeader
Binds headers to a struct.

**Example:**
```go
type RequestHeaders struct {
    Authorization string   `header:"Authorization" validate:"required"`
    ContentType   string   `header:"Content-Type"`
    Accept        []string `header:"Accept"`
}

func handler(c *lokstra.RequestContext) error {
    var headers RequestHeaders
    if err := c.Req.BindHeader(&headers); err != nil {
        return err
    }
    
    token := headers.Authorization
    // ...
}
```

---

## Complete Examples

### Basic Parameter Extraction
```go
func getUser(c *lokstra.RequestContext) error {
    // Path parameter
    id := c.Req.Param("id")
    
    // Query parameters
    fields := c.Req.Query("fields", "")
    includeDeleted := c.Req.Query("include_deleted", "false")
    
    // Headers
    token := c.Req.Header("Authorization")
    
    user, err := fetchUser(id, fields, includeDeleted == "true")
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    
    return c.Api.Success(user)
}
```

### JSON Request Binding
```go
type UpdateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"omitempty,min=18,max=120"`
}

func updateUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    
    var input UpdateUserInput
    if err := c.Req.BindJSON(&input); err != nil {
        return err // Auto-formatted as validation error
    }
    
    user, err := updateUserInDB(id, input)
    if err != nil {
        return c.Api.InternalError("Failed to update user")
    }
    
    return c.Api.Success(user)
}
```

### Middleware with Context Storage
```go
func authMiddleware(c *lokstra.RequestContext) error {
    token := c.Req.Header("Authorization")
    if token == "" {
        return c.Api.Unauthorized("Missing authorization token")
    }
    
    user, err := validateToken(token)
    if err != nil {
        return c.Api.Unauthorized("Invalid token")
    }
    
    // Repository user in context
    c.Set("user", user)
    c.Set("user_id", user.ID)
    c.SetContextValue("user_id", user.ID)
    
    return c.Next()
}

func getProfile(c *lokstra.RequestContext) error {
    // Retrieve user from context
    user := c.Get("user").(*User)
    return c.Api.Success(user)
}

func deleteAccount(c *lokstra.RequestContext) error {
    userID := c.Get("user_id").(int)
    if err := deleteUserAccount(userID); err != nil {
        return err
    }
    return c.Api.NoContent()
}
```

### Query Parameter Binding
```go
type SearchFilter struct {
    Query    string   `query:"q" validate:"required"`
    Tags     []string `query:"tags"`
    Category string   `query:"category"`
    Limit    int      `query:"limit" validate:"min=1,max=100"`
    Page     int      `query:"page" validate:"min=1"`
}

func search(c *lokstra.RequestContext) error {
    var filter SearchFilter
    if err := c.Req.BindQuery(&filter); err != nil {
        return err
    }
    
    // Set defaults
    if filter.Limit == 0 {
        filter.Limit = 20
    }
    if filter.Page == 0 {
        filter.Page = 1
    }
    
    results := performSearch(filter)
    return c.Api.Success(results)
}
```

### File Upload Handling
```go
func uploadFile(c *lokstra.RequestContext) error {
    // Parse multipart form
    if err := c.R.ParseMultipartForm(10 << 20); err != nil { // 10MB
        return c.Api.BadRequest("Failed to parse form")
    }
    
    file, header, err := c.R.FormFile("file")
    if err != nil {
        return c.Api.BadRequest("No file uploaded")
    }
    defer file.Close()
    
    // Save file
    filename := header.Filename
    savedPath, err := saveUploadedFile(file, filename)
    if err != nil {
        return c.Api.InternalError("Failed to save file")
    }
    
    return c.Api.Success(map[string]string{
        "filename": filename,
        "path":     savedPath,
    })
}
```

### Complex Request Processing
```go
func complexHandler(c *lokstra.RequestContext) error {
    // Path parameters
    id := c.Req.Param("id")
    
    // Query parameters
    includeRefs := c.Req.Query("include_refs", "false") == "true"
    fields := c.Req.QueryParams("fields")
    
    // Headers
    token := c.Req.Header("Authorization")
    acceptLang := c.Req.Header("Accept-Language")
    
    // Request body
    var input UpdateInput
    if err := c.Req.BindJSON(&input); err != nil {
        return err
    }
    
    // Context values
    userID := c.Get("user_id").(int)
    
    // Process
    result, err := processComplexRequest(id, input, fields, includeRefs, userID, acceptLang)
    if err != nil {
        return c.Api.InternalError(err.Error())
    }
    
    return c.Api.Success(result)
}
```

---

## See Also

- **[Response](response)** - Response building API
- **[Router](router)** - Handler registration
- **[Validation](../07-helpers/common-validator)** - Validation utilities

---

## Related Guides

- **[Router Essentials](../../01-router-guide/01-router/)** - Handler basics
- **[Middleware Guide](../../01-router-guide/03-middleware/)** - Middleware patterns
- **[Request Handling](../../02-deep-dive/router/)** - Advanced techniques
