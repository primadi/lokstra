
# Lokstra API Standards

This document defines the **opinionated standard for API Request and Response** in Lokstra.  
The goal is to ensure **consistency**, **developer experience (DX)**, and **auto-UI readiness** across all services.

---

## 📌 API Request Standard

### Struct Definition

```go
type PagingRequest struct {
    Page       int      `query:"page"`        // default: 1
    PageSize   int      `query:"page_size"`   // default: 20, max: 100
    OrderBy    []string `query:"order_by"`    // e.g. order_by=id,-name
    QueryAll   bool     `query:"all"`         // true → ignore paging
    Fields     []string `query:"fields"`      // e.g. fields=id,name,email
    Search     string   `query:"search"`      // global keyword search
    Filters    []string `query:"filter"`      // e.g. filter=status:active
    DataType   string   `query:"data_type"`   // "list" | "table", default "list"
    DataFormat string   `query:"data_format"` // "json", "json_download", "csv", "xlsx"
    Download   bool     `query:"download"`    // true = force download, false = inline
}
```

### Rules

- **page/page_size** → standard pagination, default and max limit applied.  
- **order_by** → multi-field allowed, prefix `-` means descending.  
- **all** → if true, paging is ignored. Backend may limit with `MaxData`.  
- **fields** → subset of columns to return, default is all fields.  
- **filters** → `field:value` format, multi = AND, comma = IN.  
- **data_type** → `"list"` returns JSON object array, `"table"` returns 2D array + headers.  
- **data_format** →  
  - `"json"` → wrapped `ApiResponse`.  
  - `"json_download"` → raw JSON data, served as file.  
  - `"csv"` → CSV file.  
  - `"xlsx"` → Excel file.  
- **download** → if true, response is forced as `attachment`. If false, served inline.  

---

## 📌 API Response Standard

### Base Struct

```go
// ApiResponse standardizes API response structure
type ApiResponse[T any] struct {
    Status    string `json:"status"`              // "success" | "error"
    Message   string `json:"message,omitempty"`   // Human readable message
    Data      T      `json:"data,omitempty"`      // Response data
    Error     *Error `json:"error,omitempty"`     // Error details if status = "error"
    Meta      *Meta  `json:"meta,omitempty"`      // Metadata for lists/pagination
    RequestID string `json:"request_id,omitempty"` // For tracing
}

// ListResponse is specialized for list/paginated data
type ListResponse[T any] struct {
    ApiResponse[[]T]
}
```

### Error Structure

```go
type Error struct {
    Code    string                 `json:"code"`              // Error code (e.g. "VALIDATION_ERROR")
    Message string                 `json:"message"`           // Error message
    Details map[string]any `json:"details,omitempty"` // Additional error details
    Fields  []FieldError          `json:"fields,omitempty"`  // Validation field errors
}

type FieldError struct {
    Field   string `json:"field"`   // Field name
    Code    string `json:"code"`    // Error code (e.g. "REQUIRED")
    Message string `json:"message"` // Error message
    Value   any    `json:"value,omitempty"` // Invalid value provided
}
```

### Metadata Structure

```go
type Meta struct {
    *ListMeta     `json:",omitempty"`
    *RequestMeta  `json:",omitempty"`
    *ResponseMeta `json:",omitempty"`
}

type ListMeta struct {
    Page       int  `json:"page"`        // Current page
    PageSize   int  `json:"page_size"`   // Items per page
    Total      int  `json:"total"`       // Total items
    TotalPages int  `json:"total_pages"` // Total pages
    HasNext    bool `json:"has_next"`    // Has next page
    HasPrev    bool `json:"has_prev"`    // Has previous page
}

type RequestMeta struct {
    Filters  map[string]string `json:"filters,omitempty"`   // Applied filters
    OrderBy  []string         `json:"order_by,omitempty"`  // Applied ordering
    Fields   []string         `json:"fields,omitempty"`    // Selected fields
    Search   string           `json:"search,omitempty"`    // Search query
    DataType string           `json:"data_type,omitempty"` // Response format type
}
```

### Rules

- **Status** → "success" or "error", mutually exclusive with Error field.  
- **Meta** is always included for list responses with ListMeta.  
- **RequestMeta** echoes back applied filters, ordering, search terms.  
- **ResponseMeta** includes processing time, cache status, etc.  
- **Error** contains structured error information with field-level validation errors.  
- **RequestID** supports distributed tracing and debugging.  

---

## 📌 Response Helpers

Lokstra provides helper functions to build standardized responses.

```go
// ✅ Success Responses
func NewSuccess[T any](data T) *ApiResponse[T]
func NewSuccessWithMessage[T any](data T, message string) *ApiResponse[T]

// ✅ Error Responses  
func NewError(code, message string) *ApiResponse[any]
func NewErrorWithDetails(code, message string, details map[string]any) *ApiResponse[any]
func NewValidationError(message string, fields []FieldError) *ApiResponse[any]

// ✅ List Responses
func NewListResponse[T any](data []T, meta *ListMeta) *ListResponse[T]
func CalculateListMeta(page, pageSize, total int) *ListMeta

// ✅ PagingRequest Helpers
func (p *PagingRequest) SetDefaults()
func (p *PagingRequest) GetOffset() int
func (p *PagingRequest) GetLimit() int
func (p *PagingRequest) ParseFilters() map[string]string
func (p *PagingRequest) ParseOrderBy() []OrderField
```

### Usage in Handlers

```go
func GetUsers(c *request.Context) error {
    var req request.PagingRequest
    if err := c.Req.BindQuery(&req); err != nil {
        return c.Resp.JSON(400, response.NewValidationError("Invalid query", nil))
    }
    
    req.SetDefaults()
    users, total, err := userService.GetUsers(req.GetOffset(), req.GetLimit())
    if err != nil {
        return c.Resp.JSON(500, response.NewError("DATABASE_ERROR", err.Error()))
    }
    
    meta := response.CalculateListMeta(req.Page, req.PageSize, total)
    resp := response.NewListResponse(users, meta)
    return c.Resp.JSON(200, resp)
}
```

---

## 📌 Example

### Request Examples
```bash
# Basic pagination
GET /users?page=1&page_size=20

# With ordering and filters  
GET /users?page=1&order_by=-created_at&filter=status:active

# Export as CSV
GET /users?data_format=csv&download=true&all=true
```

### Response Examples

#### JSON List Response
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John", "email": "john@example.com"}
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false,
    "filters": {"status": "active"},
    "order_by": ["-created_at"]
  }
}
```

#### Error Response  
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
        "message": "Email format is invalid"
      }
    ]
  }
}
```

#### CSV Response
```
Content-Type: text/csv
Content-Disposition: attachment; filename="users.csv"

id,name,email
1,John,john@example.com
2,Ana,ana@example.com
```

---

## ✅ Summary

- **Request**: standardized query for list operations, including paging, ordering, filtering, field selection, data type, and data format.  
- **Response**: standardized JSON structure with optional raw outputs (JSON/CSV/XLSX).  
- **ApiHelper**: convenience wrapper for building responses consistently.  

This standard ensures **predictable APIs**, **easy integration**, and **auto-UI rendering** support.
