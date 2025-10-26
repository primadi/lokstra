# API Standard Usage Examples

## Request Structure Examples

### 1. List Users with Pagination

```go
func GetUsers(c *request.Context) error {
    // Parse paging request
    var req request.PagingRequest
    if err := c.Req.BindQuery(&req); err != nil {
        return c.Resp.JSON(400, response.NewValidationError("Invalid query parameters", nil))
    }
    
    // Apply defaults
    req.SetDefaults()
    
    // Get data with filters
    filters := req.ParseFilters()
    orders := req.ParseOrderBy()
    
    users, total, err := userService.GetUsers(req.GetOffset(), req.GetLimit(), filters, orders)
    if err != nil {
        return c.Resp.JSON(500, response.NewError("DATABASE_ERROR", err.Error()))
    }
    
    // Build response with pagination meta
    meta := response.CalculateListMeta(req.Page, req.PageSize, total)
    resp := response.NewListResponse(users, meta)
    
    return c.Resp.JSON(200, resp)
}
```

### 2. Create User with Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Role     string `json:"role" validate:"required,oneof=admin user"`
}

func CreateUser(c *request.Context) error {
    var req CreateUserRequest
    if err := c.Req.BindBody(&req); err != nil {
        return c.Resp.JSON(400, response.NewError("INVALID_REQUEST", "Invalid JSON body"))
    }
    
    // Validate request
    if err := validator.Validate(&req); err != nil {
        fieldErrors := convertValidationErrors(err) // Convert to []response.FieldError
        return c.Resp.JSON(400, response.NewValidationError("Validation failed", fieldErrors))
    }
    
    user, err := userService.CreateUser(req)
    if err != nil {
        return c.Resp.JSON(500, response.NewError("CREATE_FAILED", err.Error()))
    }
    
    resp := response.NewSuccessWithMessage(user, "User created successfully")
    return c.Resp.JSON(201, resp)
}
```

### 3. Get Single User

```go
func GetUser(c *request.Context) error {
    userID := c.Req.PathParam("id")
    if userID == "" {
        return c.Resp.JSON(400, response.NewError("MISSING_PARAM", "User ID is required"))
    }
    
    user, err := userService.GetByID(userID)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return c.Resp.JSON(404, response.NewError("USER_NOT_FOUND", "User not found"))
        }
        return c.Resp.JSON(500, response.NewError("DATABASE_ERROR", err.Error()))
    }
    
    resp := response.NewSuccess(user)
    return c.Resp.JSON(200, resp)
}
```

### 4. Advanced Search with Multiple Parameters

```go
type UserSearchRequest struct {
    request.PagingRequest
    Name         string   `query:"name"`
    Email        string   `query:"email"`
    Roles        []string `query:"roles"`
    CreatedAfter string   `query:"created_after"`
    CreatedBefore string  `query:"created_before"`
    Status       string   `query:"status"`
}

func SearchUsers(c *request.Context) error {
    var req UserSearchRequest
    if err := c.Req.BindAll(&req); err != nil {
        return c.Resp.JSON(400, response.NewError("INVALID_QUERY", err.Error()))
    }
    
    req.SetDefaults()
    
    // Build search criteria
    criteria := buildSearchCriteria(&req)
    
    users, total, err := userService.Search(criteria, req.GetOffset(), req.GetLimit())
    if err != nil {
        return c.Resp.JSON(500, response.NewError("SEARCH_FAILED", err.Error()))
    }
    
    // Add request metadata
    meta := &response.Meta{
        ListMeta: response.CalculateListMeta(req.Page, req.PageSize, total),
        RequestMeta: &response.RequestMeta{
            Filters:  req.ParseFilters(),
            OrderBy:  req.OrderBy,
            Search:   req.Search,
            DataType: req.DataType,
        },
    }
    
    resp := &response.ListResponse[User]{
        ApiResponse: response.ApiResponse[[]User]{
            Status: "success",
            Data:   users,
            Meta:   meta,
        },
    }
    
    return c.Resp.JSON(200, resp)
}
```

### 5. Export Data (CSV/Excel)

```go
func ExportUsers(c *request.Context) error {
    var req request.PagingRequest
    if err := c.Req.BindQuery(&req); err != nil {
        return c.Resp.JSON(400, response.NewError("INVALID_QUERY", err.Error()))
    }
    
    req.SetDefaults()
    
    // For exports, usually we want all data
    if req.DataFormat == "csv" || req.DataFormat == "xlsx" {
        req.QueryAll = true
    }
    
    users, _, err := userService.GetUsers(req.GetOffset(), req.GetLimit(), nil, nil)
    if err != nil {
        return c.Resp.JSON(500, response.NewError("EXPORT_FAILED", err.Error()))
    }
    
    switch req.DataFormat {
    case "csv":
        return exportCSV(c, users)
    case "xlsx":
        return exportExcel(c, users)
    default:
        // Return as JSON with download headers
        c.Resp.Header("Content-Disposition", "attachment; filename=users.json")
        resp := response.NewSuccess(users)
        return c.Resp.JSON(200, resp)
    }
}
```

## URL Examples

```bash
# Basic pagination
GET /api/users?page=1&page_size=20

# With ordering
GET /api/users?page=1&page_size=10&order_by=name,-created_at

# With filters
GET /api/users?filter=status:active&filter=role:admin

# With search
GET /api/users?search=john&page=1

# Get all without pagination
GET /api/users?all=true

# Export as CSV
GET /api/users/export?data_format=csv&download=true

# Table format for DataTables
GET /api/users?data_type=table&page=1&page_size=10

# Select specific fields only
GET /api/users?fields=id,name,email&page=1
```

## Response Examples

### Successful List Response
```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "admin"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false,
    "filters": {
      "status": "active"
    },
    "order_by": ["name", "-created_at"]
  }
}
```

### Validation Error Response
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Email format is invalid",
        "value": "invalid-email"
      }
    ]
  }
}
```

### Single Entity Response
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```