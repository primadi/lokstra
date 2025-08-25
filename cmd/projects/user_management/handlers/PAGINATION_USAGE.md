# User Management Pagination Implementation

## Overview

Implementasi pagination untuk ListUser menggunakan `BindPaginationQuery` dari request.Context untuk memberikan pengalaman yang konsisten dan powerful.

## API Usage

### Basic Pagination
```
GET /users?page=1&pageSize=10
```

### With Filters
```
GET /users?page=1&pageSize=10&filter[username]=john&filter[is_active]=true
```

### With Sorting
```
GET /users?page=1&pageSize=10&sort[username]=asc&sort[email]=desc
```

### With Field Selection
```
GET /users?page=1&pageSize=10&fields=id,username,email
```

## Response Format

### Standardized Pagination Response
```json
{
  "data": [
    {
      "id": "user-123",
      "tenant_id": "default",
      "username": "john_doe",
      "email": "john@example.com",
      "is_active": true,
      "metadata": {}
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 45,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  },
  "filters": {
    "applied": {
      "username": "john",
      "is_active": "true"
    }
  }
}
```

## Supported Query Parameters

### Pagination
- `page` - Page number (default: 1)
- `pageSize` - Items per page (default: 10, max: 100)

### Filters
- `filter[username]` - Search by username (ILIKE pattern matching)
- `filter[email]` - Search by email (ILIKE pattern matching)  
- `filter[is_active]` - Filter by active status (true/false)

### Sorting
- `sort[username]` - Sort by username (asc/desc)
- `sort[email]` - Sort by email (asc/desc)

### Field Selection
- `fields` - Comma-separated list of fields to return

## Implementation Details

### Repository Layer
```go
// New method added to UserRepository
func (u *UserRepository) ListUsersWithPagination(
    ctx context.Context, 
    tenantID string, 
    page, pageSize int, 
    filters map[string]string
) ([]*auth.User, int, error)
```

Features:
- Dynamic SQL building with filters
- Total count query for pagination info
- LIMIT/OFFSET for pagination
- ILIKE for case-insensitive search
- SQL injection protection with parameterized queries

### Handler Layer
```go
func listUsersAction(fctx *flow.Context[ListUserRequestDTO]) error {
    // Uses BindPaginationQuery for automatic parameter binding
    pagination, err := fctx.BindPaginationQuery()
    
    // Returns standardized PaginatedResponse[T]
    response := PaginatedResponse[*auth.User]{
        Data: users,
        Pagination: PaginationInfo{...},
        Filters: map[string]interface{}{...},
    }
}
```

Features:
- Automatic query parameter binding
- Input validation and sanitization
- Standardized response format
- Type-safe generic response structure

### DTO Simplification
```go
// Simplified - no manual pagination fields needed
type ListUserRequestDTO struct {
}

// Reusable pagination response for any resource
type PaginatedResponse[T any] struct {
    Data       []T                    `json:"data"`
    Pagination PaginationInfo         `json:"pagination"`
    Filters    map[string]interface{} `json:"filters,omitempty"`
}
```

## Migration Benefits

### Before
```go
// Manual pagination handling
type ListUserRequestDTO struct {
    Page     *int    `json:"page,omitempty"`
    PageSize *int    `json:"page_size,omitempty"`
    Search   *string `json:"search,omitempty"`
    IsActive *bool   `json:"is_active,omitempty"`
}

// Custom response format
return fctx.Ok(map[string]interface{}{
    "users":     users,
    "page":      page,
    "page_size": pageSize,
    "total":     len(users),
})
```

### After
```go
// Automatic binding via BindPaginationQuery
type ListUserRequestDTO struct {
}

// Standardized response with full pagination info
return fctx.Ok(PaginatedResponse[*auth.User]{
    Data:       users,
    Pagination: PaginationInfo{...},
    Filters:    map[string]interface{}{...},
})
```

## Key Improvements

1. **Automatic Parameter Binding**: Uses `BindPaginationQuery()` untuk parsing otomatis
2. **Rich Filtering Support**: Multiple filter types dengan SQL pattern matching
3. **Flexible Sorting**: Multiple field sorting dengan direction control
4. **Field Selection**: Optimize response size dengan field selection
5. **Standardized Response**: Consistent pagination format across all endpoints
6. **SQL Performance**: Proper pagination queries dengan COUNT dan LIMIT/OFFSET
7. **Type Safety**: Generic `PaginatedResponse[T]` untuk compile-time safety

## Usage Example

```bash
# Get second page with 5 users per page, filter active users named 'john'
curl "http://localhost:8080/users?page=2&pageSize=5&filter[username]=john&filter[is_active]=true&sort[username]=asc"

# Get only specific fields
curl "http://localhost:8080/users?fields=id,username,email&pageSize=20"
```

This implementation provides a robust, performant, and user-friendly pagination system that scales well and maintains consistency across the application! 🚀
