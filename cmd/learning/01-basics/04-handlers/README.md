# 04-handlers - Request Parameter Handling

This example demonstrates **three different approaches** to handling request parameters in Lokstra.

## The Three Approaches

### 1. 🔧 Manual Parameter Access

**When to use:** Simple cases, optional parameters with defaults

```go
func(c *lokstra.RequestContext) error {
    id := c.Req.PathParam("id", "0")
    format := c.Req.QueryParam("format", "json")
    // Manual validation required
}
```

**Pros:**
- Simple and straightforward
- Easy to provide default values
- Good for optional parameters

**Cons:**
- No automatic validation
- Verbose for complex structures
- No type safety
- Manual error handling

---

### 2. 📦 Manual Binding

**When to use:** Structured data with validation requirements

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func(c *lokstra.RequestContext) error {
    var req CreateUserRequest
    if err := c.Req.BindBody(&req); err != nil {
        // Check if it's a validation error
        if valErr, ok := err.(*request.ValidationError); ok {
            // Return field-level validation errors
            return c.Api.ValidationError("Validation failed", valErr.FieldErrors)
        }
        return c.Api.BadRequest("INVALID_REQUEST", err.Error())
    }
    // Now req is populated and validated
}
```

**Available bind methods:**
- `c.Req.BindBody(&struct)` - JSON request body
- `c.Req.BindQuery(&struct)` - Query parameters
- `c.Req.BindPath(&struct)` - URL path parameters

**NEW! Automatic Validation:**
- All bind methods now automatically validate using `validate` tags
- Returns `request.ValidationError` with field-level error details
- Use `c.Api.ValidationError()` to return structured error response

**Pros:**
- ✅ Automatic validation using struct tags (NEW!)
- ✅ Field-level error details
- Type-safe
- Reusable struct definitions
- Clear data structure

**Cons:**
- Requires explicit binding call
- Extra error handling code
- Slightly more verbose

---

### 3. ⚡ Smart Binding (RECOMMENDED!)

**When to use:** ALWAYS (best developer experience)

**IMPORTANT RULE:** Only **ONE struct parameter** allowed (besides context)!

**But** that single struct can combine multiple sources:
- `path:"param"` - URL path parameters
- `query:"param"` - Query string parameters
- `header:"Header-Name"` - HTTP headers
- `json:"field"` or `body:"field"` - Request body

```go
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"gt=0"`
}

// Correct: ONE struct parameter
func(c *lokstra.RequestContext, req *CreateProductRequest) error {
    // req is already populated and validated!
    return c.Api.Created(req, "Product created")
}
```

**Combining multiple sources in ONE struct:**
```go
type UpdateProductRequest struct {
    // Path parameter
    ID string `path:"id" validate:"required"`
    
    // Body fields (JSON)
    Name  *string  `json:"name"`
    Price *float64 `json:"price" validate:"omitempty,gt=0"`
    
    // Query parameter
    Notify bool `query:"notify"`
}

// Still just ONE struct!
func(c *lokstra.RequestContext, req *UpdateProductRequest) error {
    // req.ID comes from URL path
    // req.Name and req.Price come from JSON body
    // req.Notify comes from query string
    return c.Api.Ok(req)
}
```

**❌ WRONG - Multiple struct parameters not allowed:**
```go
// This will NOT work!
func(c *lokstra.RequestContext, 
     pathParams *PathParams,      // ❌ 
     body *BodyParams,             // ❌
     query *QueryParams) error {   // ❌
    // Error: Too many struct parameters
}
```

**Pros:**
- ✅ Cleanest code
- ✅ Automatic validation before handler execution
- ✅ Type-safe
- ✅ Can combine path, query, header, and body in one struct
- ✅ No manual binding code needed
- ✅ No error handling boilerplate

**Cons:**
- ⚠️ Only ONE struct parameter allowed (but this is usually enough!)

---

## Comparison Table

| Feature | Manual Params | Manual Binding | Smart Binding |
|---------|--------------|----------------|---------------|
| Code clarity | 😐 Moderate | 🙂 Good | 😍 Excellent |
| Type safety | ❌ No | ✅ Yes | ✅ Yes |
| Auto validation | ❌ No | ✅ Yes | ✅ Yes |
| Boilerplate | 😔 High | 😐 Medium | 😊 Minimal |
| Error handling | 😔 Manual | 😐 Required | 😍 Automatic |
| Best for | Simple cases | Structured data | Everything! |

---

## Validation Tags Reference

Lokstra uses the `validate` tag for automatic validation. **NEW!** Validation is now built-in and automatic:

```go
type Example struct {
    // Path parameter
    ID string `path:"id" validate:"required"`
    
    // Query parameters
    Page int `query:"page" validate:"min=1"`
    Sort string `query:"sort" validate:"oneof=name email date"`
    
    // Header
    Authorization string `header:"Authorization" validate:"required"`
    
    // Body fields (JSON)
    Name string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    
    // Number ranges
    Age int `json:"age" validate:"min=1,max=120"`
    Price float64 `json:"price" validate:"gt=0"`
    
    // String validation
    Role string `json:"role" validate:"oneof=admin user guest"`
    
    // Optional with validation (only validate if provided)
    Phone *string `json:"phone" validate:"omitempty,min=10"`
}
```

**Available tags:**
- `path:"param_name"` - Extract from URL path
- `query:"param_name"` - Extract from query string
- `header:"Header-Name"` - Extract from HTTP header
- `json:"field_name"` or `body:"field_name"` - Extract from request body

**Validation happens automatically** after binding! No need to call validator manually.

**Supported validators:**
- `required` - Field must be present and not empty
- `email` - Valid email format (contains @ and .)
- `min=N` - Minimum value (numbers) or length (strings/slices)
- `max=N` - Maximum value or length
- `gt=N` - Greater than (numbers only)
- `gte=N` - Greater than or equal
- `lt=N` - Less than
- `lte=N` - Less than or equal
- `oneof=a b c` - Must be one of the space-separated values
- `omitempty` - Only validate if not empty/nil (for optional fields)

**Error Handling:**

When validation fails, bind methods return `request.ValidationError` with structured field errors:

```go
if err := c.Req.BindBody(&req); err != nil {
    if valErr, ok := err.(*request.ValidationError); ok {
        // valErr.FieldErrors contains []api_formatter.FieldError
        // Each FieldError has: Field (name) and Message (error message)
        return c.Api.ValidationError("Validation failed", valErr.FieldErrors)
    }
    return c.Api.BadRequest("INVALID_REQUEST", err.Error())
}
```

The `c.Api.ValidationError()` method returns a structured response like:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {"field": "email", "message": "email must be a valid email address"},
      {"field": "age", "message": "age must be at least 1"}
    ]
  }
}
```

---

## Running the Example

```bash
go run main.go
```

Then use the test.http file to try all three approaches.

---

## Key Takeaways

1. **Smart Binding is the recommended approach** for 99% of use cases
2. **Important:** Only ONE struct parameter allowed in Smart Binding
3. **Powerful:** That one struct can combine path, query, header, and body tags
4. **NEW! Automatic Validation:** All binding methods now validate automatically
5. **Field-level Errors:** Get detailed validation errors for each field
6. Manual parameter access is fine for very simple cases
7. Manual binding gives you control over error handling
8. Always use validation tags to ensure data quality
9. Use `c.Api.ValidationError()` for structured error responses

---

## Next Steps

See `05-config` to learn about YAML-based configuration and deployment patterns.
