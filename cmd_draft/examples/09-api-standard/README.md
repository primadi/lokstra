# 09 - API Standard Example

This example demonstrates the **Lokstra API Standard** using `PagingRequest` and `ApiResponse` for consistent, predictable API development.

## Features Demonstrated

### 1. Standardized Request Structure
- **PagingRequest** with pagination, filtering, searching, and ordering
- Query parameter binding with automatic validation
- Support for different data formats (JSON, CSV, Excel)

### 2. Standardized Response Structure  
- **ApiResponse[T]** with consistent success/error format
- **ListResponse[T]** for paginated data
- Structured error responses with field-level validation
- Metadata for pagination and request tracing

### 3. Real-world Handlers
- `GET /users` - List users with full pagination and filtering
- `GET /users/{id}` - Single user retrieval
- `POST /users` - User creation with validation

## API Examples

### List Users with Pagination
```bash
# Basic pagination
curl "http://localhost:8080/users?page=1&page_size=10"

# With filters and search
curl "http://localhost:8080/users?filter=status:active&filter=role:admin&search=john"

# With ordering
curl "http://localhost:8080/users?order_by=name,-created_at&page=1"

# Export as CSV
curl "http://localhost:8080/users?data_format=csv&download=true&all=true"

# Get all data without pagination
curl "http://localhost:8080/users?all=true"
```

### Single User
```bash
curl "http://localhost:8080/users/1"
```

### Create User
```bash
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com", 
    "role": "admin"
  }'
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
      "role": "admin",
      "status": "active",
      "created_at": "2024-01-01"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 5,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false,
    "filters": {
      "status": "active"
    },
    "order_by": ["name", "-created_at"],
    "data_type": "list"
  }
}
```

### Error Response
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
        "field": "role",
        "code": "INVALID_VALUE",
        "message": "Role must be 'admin' or 'user'",
        "value": "invalid_role"
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
    "id": 6,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "admin",
    "status": "active",
    "created_at": "2024-10-01"
  }
}
```

## Query Parameters Reference

### PagingRequest Parameters
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)
- `order_by` - Ordering fields, prefix `-` for descending (e.g. `name,-created_at`)
- `all` - Skip pagination, return all data (boolean)
- `fields` - Select specific fields (e.g. `id,name,email`)
- `search` - Global keyword search
- `filter` - Field filters in format `field:value` (multiple allowed)
- `data_type` - Response format: `list` (default) or `table`
- `data_format` - Output format: `json`, `csv`, `xlsx`, `json_download`
- `download` - Force download attachment (boolean)

### Filter Examples
```bash
# Single filter
?filter=status:active

# Multiple filters (AND logic)  
?filter=status:active&filter=role:admin

# Combined with other params
?filter=status:active&search=john&order_by=name&page=1
```

## Key Implementation Details

### 1. Request Binding
```go
var req request.PagingRequest
if err := c.Req.BindQuery(&req); err != nil {
    return c.Resp.WithStatus(400).Json(response.NewError("INVALID_QUERY", err.Error()))
}
req.SetDefaults()
```

### 2. Response Building
```go
// Success response
resp := response.NewSuccess(data)
return c.Resp.WithStatus(200).Json(resp)

// List response with pagination
meta := response.CalculateListMeta(req.Page, req.PageSize, total)
resp := response.NewListResponse(users, meta)
return c.Resp.WithStatus(200).Json(resp)

// Error response
return c.Resp.WithStatus(400).Json(response.NewValidationError("Validation failed", fieldErrors))
```

### 3. Export Handling
```go
switch req.DataFormat {
case "csv":
    return exportUsersCSV(c, users)
case "xlsx": 
    return exportUsersExcel(c, users)
default:
    return c.Resp.WithStatus(200).Json(resp)
}
```

## Running the Example

```bash
cd cmd/examples/09-api-standard
go run main.go
```

This example showcases the **complete Lokstra API standard** for building consistent, developer-friendly APIs with built-in pagination, filtering, validation, and export capabilities.

## Standards Benefits

1. **Consistency** - All APIs follow the same request/response patterns
2. **Developer Experience** - Predictable query parameters and response structure  
3. **Auto-UI Ready** - Metadata supports automatic frontend generation
4. **Flexibility** - Support multiple output formats and data types
5. **Validation** - Structured error responses with field-level details
6. **Tracing** - Request metadata for debugging and analytics