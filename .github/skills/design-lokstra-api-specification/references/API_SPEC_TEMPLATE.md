---
module: [MODULE_NAME]
version: [VERSION]
status: draft
based_on: Requirements v[VERSION]
---

# [Module Name] - API Specification

## API Information

| Field | Value |
|-------|-------|
| Module | [module_name] |
| Version | [VERSION] |
| Status | draft / approved / implemented |
| Base URL | [/api/module-path] |
| Authentication | Required / Optional / None |
| Last Updated | [DATE] |

---

## 1. Overview

### 1.1 Purpose
[Brief description of the API purpose]

### 1.2 Base URL
```
Development:  http://localhost:3000/api/[module]
Staging:      https://staging.example.com/api/[module]
Production:   https://api.example.com/api/[module]
```

### 1.3 Authentication
[Describe authentication method]
- **Method:** JWT Bearer Token / API Key / OAuth2 / None
- **Header:** `Authorization: Bearer <token>`
- **Token Expiry:** [Duration]

### 1.4 Common Response Format

**Success Response:**
```json
{
  "data": { ... },
  "message": "Success message",
  "timestamp": "2026-01-27T10:00:00Z"
}
```

**Error Response:**
```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": [
    {
      "field": "field_name",
      "error": "validation_error",
      "message": "Specific error message"
    }
  ],
  "timestamp": "2026-01-27T10:00:00Z"
}
```

### 1.5 HTTP Status Codes
| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET, PUT, DELETE |
| 201 | Created | Successful POST (resource created) |
| 400 | Bad Request | Validation error, malformed request |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (duplicate, etc.) |
| 422 | Unprocessable Entity | Business logic validation failed |
| 500 | Internal Server Error | Server error |

---

## 2. Endpoints

### 2.1 Create [Entity]

**Endpoint:** `POST /api/[module]/[entities]`

**Authorization:** Required (Roles: [admin, user])

**Description:**
[Detailed description of what this endpoint does]

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "field1": "string (required, min=3, max=100)",
  "field2": "string (required, email format)",
  "field3": "integer (required, min=0, max=150)",
  "field4": "date (required, ISO 8601 format)",
  "field5": "enum (optional, values: value1|value2|value3)",
  "nested_object": {
    "subfield1": "string (required)",
    "subfield2": "number (optional)"
  }
}
```

**Example Request:**
```json
{
  "field1": "Example Value",
  "field2": "user@example.com",
  "field3": 25,
  "field4": "2026-01-27",
  "field5": "value1",
  "nested_object": {
    "subfield1": "Nested Value",
    "subfield2": 100.50
  }
}
```

**Validation Rules:**
| Field | Rule | Error Message |
|-------|------|---------------|
| field1 | required, min=3, max=100 | "Field1 is required and must be between 3-100 characters" |
| field2 | required, email | "Valid email address is required" |
| field3 | required, min=0, max=150 | "Field3 must be between 0 and 150" |
| field4 | required, date, past | "Valid past date is required" |

**Business Rules:**
1. [Business rule 1]
2. [Business rule 2]
3. [Business rule 3]

**Success Response:** `201 Created`
```json
{
  "data": {
    "id": "uuid-or-id",
    "field1": "Example Value",
    "field2": "user@example.com",
    "field3": 25,
    "field4": "2026-01-27",
    "created_at": "2026-01-27T10:00:00Z",
    "updated_at": "2026-01-27T10:00:00Z"
  },
  "message": "Entity created successfully"
}
```

**Error Responses:**

**400 Bad Request - Validation Error**
```json
{
  "error": "validation_error",
  "message": "Invalid input data",
  "details": [
    {
      "field": "field2",
      "error": "invalid_email",
      "message": "Valid email address is required"
    }
  ]
}
```

**409 Conflict - Duplicate**
```json
{
  "error": "duplicate_entity",
  "message": "Entity with this identifier already exists"
}
```

**422 Unprocessable Entity - Business Rule Violation**
```json
{
  "error": "business_rule_violation",
  "message": "Cannot create entity: [specific reason]"
}
```

**Integration:**
- [External System 1]: Sync created entity within 1 minute
- [External System 2]: Validate field3 against external database

**Performance:**
- Expected response time: < 500ms (95th percentile)

---

### 2.2 Get [Entity] by ID

**Endpoint:** `GET /api/[module]/[entities]/{id}`

**Authorization:** Required (Roles: [admin, user, guest])

**Description:**
Retrieve a single entity by its unique identifier.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Unique identifier of the entity |

**Example Request:**
```
GET /api/[module]/[entities]/uuid-123
Authorization: Bearer <token>
```

**Success Response:** `200 OK`
```json
{
  "data": {
    "id": "uuid-123",
    "field1": "Example Value",
    "field2": "user@example.com",
    "field3": 25,
    "field4": "2026-01-27",
    "status": "active",
    "created_at": "2026-01-27T10:00:00Z",
    "updated_at": "2026-01-27T10:00:00Z"
  }
}
```

**Error Responses:**

**404 Not Found**
```json
{
  "error": "entity_not_found",
  "message": "Entity with id 'uuid-123' not found"
}
```

**Performance:**
- Expected response time: < 100ms (95th percentile)

---

### 2.3 List [Entities]

**Endpoint:** `GET /api/[module]/[entities]`

**Authorization:** Required (Roles: [admin, user])

**Description:**
Retrieve a paginated list of entities with optional filtering and sorting.

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | integer | No | 1 | Page number (1-indexed) |
| limit | integer | No | 20 | Items per page (max: 100) |
| sort | string | No | created_at | Sort field |
| order | string | No | desc | Sort order: asc/desc |
| status | string | No | - | Filter by status |
| search | string | No | - | Search in specific fields |

**Example Requests:**
```
GET /api/[module]/[entities]?page=1&limit=20
GET /api/[module]/[entities]?status=active&sort=field1&order=asc
GET /api/[module]/[entities]?search=john
```

**Success Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid-123",
      "field1": "Example 1",
      "field2": "user1@example.com",
      "status": "active",
      "created_at": "2026-01-27T10:00:00Z"
    },
    {
      "id": "uuid-456",
      "field1": "Example 2",
      "field2": "user2@example.com",
      "status": "active",
      "created_at": "2026-01-27T09:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

**Performance:**
- Expected response time: < 300ms (95th percentile)

---

### 2.4 Update [Entity]

**Endpoint:** `PUT /api/[module]/[entities]/{id}`

**Authorization:** Required (Roles: [admin, owner])

**Description:**
Update an existing entity. Only provided fields will be updated (partial update).

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Unique identifier of the entity |

**Request Body:**
```json
{
  "field1": "Updated Value",
  "field2": "updated@example.com",
  "field3": 30
}
```

**Validation Rules:**
[Same as Create endpoint for provided fields]

**Business Rules:**
1. User must own the entity or be admin
2. Cannot update if entity status is 'deleted'
3. [Other rules]

**Success Response:** `200 OK`
```json
{
  "data": {
    "id": "uuid-123",
    "field1": "Updated Value",
    "field2": "updated@example.com",
    "field3": 30,
    "updated_at": "2026-01-27T11:00:00Z"
  },
  "message": "Entity updated successfully"
}
```

**Error Responses:**

**403 Forbidden - Not Owner**
```json
{
  "error": "forbidden",
  "message": "You don't have permission to update this entity"
}
```

**404 Not Found**
```json
{
  "error": "entity_not_found",
  "message": "Entity with id 'uuid-123' not found"
}
```

**422 Unprocessable Entity**
```json
{
  "error": "invalid_status",
  "message": "Cannot update deleted entity"
}
```

---

### 2.5 Delete [Entity]

**Endpoint:** `DELETE /api/[module]/[entities]/{id}`

**Authorization:** Required (Roles: [admin, owner])

**Description:**
Soft delete an entity (sets deleted_at timestamp, entity not removed from database).

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Unique identifier of the entity |

**Business Rules:**
1. Soft delete only (entity not physically removed)
2. User must own the entity or be admin
3. Cannot delete if entity has active dependencies

**Success Response:** `200 OK`
```json
{
  "message": "Entity deleted successfully",
  "data": {
    "id": "uuid-123",
    "deleted_at": "2026-01-27T11:00:00Z"
  }
}
```

**Error Responses:**

**403 Forbidden**
```json
{
  "error": "forbidden",
  "message": "You don't have permission to delete this entity"
}
```

**404 Not Found**
```json
{
  "error": "entity_not_found",
  "message": "Entity with id 'uuid-123' not found"
}
```

**422 Unprocessable Entity - Has Dependencies**
```json
{
  "error": "has_dependencies",
  "message": "Cannot delete entity with active dependencies"
}
```

---

### 2.6 [Custom Endpoint]

**Endpoint:** `[METHOD] /api/[module]/[path]`

**Authorization:** [Requirements]

**Description:**
[Description of custom business operation]

**Request:**
[Request details]

**Response:**
[Response details]

---

## 3. Data Models

### 3.1 [Entity] Object

```json
{
  "id": "string (UUID format)",
  "field1": "string",
  "field2": "string (email format)",
  "field3": "integer",
  "field4": "date (ISO 8601)",
  "field5": "enum (value1|value2|value3)",
  "status": "enum (active|inactive|deleted)",
  "created_at": "timestamp (ISO 8601)",
  "updated_at": "timestamp (ISO 8601)",
  "deleted_at": "timestamp (ISO 8601, nullable)"
}
```

**Field Descriptions:**
- `id`: Unique identifier, auto-generated
- `field1`: [Description]
- `field2`: [Description]
- `status`: Current status of the entity
- `created_at`: Creation timestamp
- `updated_at`: Last modification timestamp
- `deleted_at`: Soft delete timestamp (null if active)

---

### 3.2 [Nested Object] Object

```json
{
  "subfield1": "string",
  "subfield2": "number"
}
```

---

## 4. Enumerations

### 4.1 Status Enum
| Value | Description |
|-------|-------------|
| active | Entity is active and usable |
| inactive | Entity is inactive but not deleted |
| deleted | Entity is soft-deleted |

### 4.2 [Custom Enum]
| Value | Description |
|-------|-------------|
| value1 | [Description] |
| value2 | [Description] |

---

## 5. Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| validation_error | 400 | Input validation failed |
| unauthorized | 401 | Authentication required |
| forbidden | 403 | Insufficient permissions |
| entity_not_found | 404 | Entity does not exist |
| duplicate_entity | 409 | Entity already exists |
| business_rule_violation | 422 | Business logic constraint violated |
| internal_error | 500 | Server error |

---

## 6. Rate Limiting

**Limits:**
- Authenticated users: 1000 requests per hour
- Anonymous: 100 requests per hour

**Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1643292000
```

**Error Response (429 Too Many Requests):**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Please try again later.",
  "retry_after": 3600
}
```

---

## 7. Versioning

**Current Version:** v1

**Version Header:**
```
Accept: application/vnd.api+json; version=1
```

**Deprecation Notice:**
Deprecated endpoints will be marked and supported for 6 months before removal.

---

## 8. Examples

### 8.1 Complete CRUD Example

```bash
# 1. Create entity
curl -X POST http://localhost:3000/api/[module]/[entities] \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "field1": "Example",
    "field2": "user@example.com",
    "field3": 25
  }'

# Response: 201 Created
# { "data": { "id": "uuid-123", ... } }

# 2. Get entity
curl -X GET http://localhost:3000/api/[module]/[entities]/uuid-123 \
  -H "Authorization: Bearer <token>"

# Response: 200 OK

# 3. Update entity
curl -X PUT http://localhost:3000/api/[module]/[entities]/uuid-123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "field1": "Updated"
  }'

# Response: 200 OK

# 4. Delete entity
curl -X DELETE http://localhost:3000/api/[module]/[entities]/uuid-123 \
  -H "Authorization: Bearer <token>"

# Response: 200 OK
```

---

## 9. Testing

### 9.1 Test Data
```json
{
  "valid_entity_1": {
    "field1": "Test Entity 1",
    "field2": "test1@example.com",
    "field3": 25
  },
  "invalid_entity_1": {
    "field1": "Te",
    "field2": "invalid-email",
    "field3": -5
  }
}
```

### 9.2 Test Scenarios
- [ ] Create with valid data
- [ ] Create with invalid data (validation)
- [ ] Create duplicate (conflict)
- [ ] Get existing entity
- [ ] Get non-existent entity
- [ ] List with pagination
- [ ] List with filters
- [ ] Update with valid data
- [ ] Update non-existent entity
- [ ] Delete existing entity
- [ ] Delete non-existent entity

---

## 10. Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| [VERSION] | [DATE] | [AUTHOR] | [CHANGES] |

---

**End of API Specification**
