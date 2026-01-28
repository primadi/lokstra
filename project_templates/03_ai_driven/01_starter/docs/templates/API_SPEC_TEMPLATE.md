# API Specification: [Module Name]
## [Project Name]

**Version:** 1.0.0  
**Status:** draft  
**Requirements Reference:** [Module] Requirements v[version]  
**Last Updated:** [Date]  
**Owner:** [Name/Team]  

---

## 1. Overview

**Base URL:** `/api/[module-name]`

**Authentication:** JWT Bearer token (unless specified otherwise)

**Content Type:** `application/json`

**Versioning:** URL-based (e.g., `/api/v1/[module]`)

---

## 2. Common Response Structures

### Success Response
```json
{
  "data": { ... },
  "message": "Success message",
  "timestamp": "2026-01-28T10:30:00Z"
}
```

### Error Response
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "User-friendly error message",
    "details": "Technical details (optional)",
    "field": "fieldName (for validation errors)"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

### Paginated Response
```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

---

## 3. Endpoints

### EP-001: [Endpoint Name]

**Purpose:** [Brief description]

**HTTP Method:** `GET/POST/PUT/DELETE`

**Path:** `/api/[module]/[resource]`

**Authentication:** Required/Optional

**Authorization:** Roles: `[role1, role2]` or Public

---

#### Request

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Path Parameters:**
- `id` (string, UUID): [Description]

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Items per page (default: 20, max: 100)
- `sort` (string, optional): Sort field (e.g., "created_at:desc")
- `filter` (string, optional): Filter criteria

**Request Body:**
```json
{
  "field1": "value1",
  "field2": "value2",
  "nested": {
    "field3": "value3"
  }
}
```

**Validation Rules:**
| Field     | Type    | Required | Constraints              | Error Message               |
|-----------|---------|----------|--------------------------|----------------------------|
| `field1`  | string  | Yes      | Min 3, Max 50            | "Field1 must be 3-50 chars"|
| `field2`  | string  | Yes      | Email format             | "Invalid email format"     |
| `field3`  | number  | No       | >= 0                     | "Must be non-negative"     |

---

#### Response

**Success (200/201):**
```json
{
  "data": {
    "id": "uuid",
    "field1": "value1",
    "field2": "value2",
    "created_at": "2026-01-28T10:30:00Z",
    "updated_at": "2026-01-28T10:30:00Z"
  },
  "message": "Resource created successfully",
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**Error Responses:**

**400 Bad Request:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": "Field1 must be 3-50 characters",
    "field": "field1"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**401 Unauthorized:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required",
    "details": "JWT token is missing or invalid"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**403 Forbidden:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions",
    "details": "Requires admin role"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "details": "Resource with id 'uuid' does not exist"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**409 Conflict:**
```json
{
  "error": {
    "code": "CONFLICT",
    "message": "Resource already exists",
    "details": "Resource with this identifier already exists"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

**500 Internal Server Error:**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "details": "Please try again later or contact support"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

---

#### Example

**Request:**
```bash
curl -X POST https://api.example.com/api/v1/[module]/[resource] \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "field1": "example value",
    "field2": "user@example.com"
  }'
```

**Response:**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "field1": "example value",
    "field2": "user@example.com",
    "created_at": "2026-01-28T10:30:00Z",
    "updated_at": "2026-01-28T10:30:00Z"
  },
  "message": "Resource created successfully",
  "timestamp": "2026-01-28T10:30:00Z"
}
```

---

### EP-002: [Next Endpoint]
[Repeat structure above for each endpoint]

---

## 4. Data Models

### [Model Name 1]

```json
{
  "id": "uuid",
  "field1": "string",
  "field2": "string",
  "field3": "number",
  "nested": {
    "field4": "string"
  },
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Field Descriptions:**
| Field       | Type      | Required | Description                    |
|-------------|-----------|----------|--------------------------------|
| `id`        | UUID      | Yes      | Unique identifier              |
| `field1`    | string    | Yes      | [Description]                  |
| `field2`    | string    | Yes      | [Description]                  |
| `created_at`| timestamp | Yes      | Creation timestamp (ISO 8601)  |

---

## 5. Error Codes

| Code                  | HTTP Status | Description                    | Resolution                       |
|-----------------------|-------------|--------------------------------|----------------------------------|
| `VALIDATION_ERROR`    | 400         | Input validation failed        | Check request body               |
| `UNAUTHORIZED`        | 401         | Authentication required        | Provide valid JWT token          |
| `FORBIDDEN`           | 403         | Insufficient permissions       | Request admin access             |
| `NOT_FOUND`           | 404         | Resource not found             | Check resource ID                |
| `CONFLICT`            | 409         | Resource already exists        | Use different identifier         |
| `INTERNAL_ERROR`      | 500         | Server error                   | Retry or contact support         |

---

## 6. Rate Limiting

**Limits:**
- Authenticated users: 1000 requests/minute
- Unauthenticated: 100 requests/minute

**Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

**429 Too Many Requests:**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "details": "Rate limit will reset in 60 seconds"
  },
  "timestamp": "2026-01-28T10:30:00Z"
}
```

---

## 7. Pagination

**Query Parameters:**
- `page`: Page number (default: 1, min: 1)
- `limit`: Items per page (default: 20, min: 1, max: 100)
- `sort`: Sort field with direction (e.g., "created_at:desc")

**Response:**
```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5,
    "hasNext": true,
    "hasPrev": false
  }
}
```

---

## 8. Filtering & Searching

**Query Parameters:**
- `search`: Full-text search across searchable fields
- `filter[field]`: Filter by specific field value
- `filter[field][op]`: Filter with operator (gt, lt, gte, lte, ne)

**Examples:**
```
GET /api/v1/products?search=laptop
GET /api/v1/products?filter[price][gte]=100&filter[price][lte]=500
GET /api/v1/products?filter[category]=electronics
```

---

## 9. Versioning Strategy

**Current Version:** v1

**Version Changes:**
- Breaking changes will increment version (v1 → v2)
- Non-breaking changes will be added to current version
- Old versions supported for 12 months after deprecation

---

## 10. Testing

### Test Cases

| Test Case ID | Description                    | Method | Path              | Expected Status |
|--------------|--------------------------------|--------|-------------------|-----------------|
| TC-001       | Create with valid data         | POST   | /api/v1/resource  | 201             |
| TC-002       | Create with invalid data       | POST   | /api/v1/resource  | 400             |
| TC-003       | Get without authentication     | GET    | /api/v1/resource  | 401             |
| TC-004       | Get non-existent resource      | GET    | /api/v1/resource/x| 404             |

---

## Appendix

### A. Change Log

| Version | Date   | Author | Changes                |
|---------|--------|--------|------------------------|
| 1.0.0   | [Date] | [Name] | Initial specification  |

### B. References
- Module Requirements: [Link]
- Database Schema: [Link]
