---
name: lokstra-api-specification
description: Generate OpenAPI-style API specifications for Lokstra modules. Creates detailed endpoint definitions with request/response schemas, validation rules, error codes, and examples. Use after module requirements are approved.
phase: design
order: 4
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Lokstra API Specification Generation

## When to use this skill

Use this skill when:
- ✅ Module requirements document is approved
- ✅ Database schema design is complete
- ✅ Need detailed API endpoint specifications
- ✅ Defining request/response contracts before implementation
- ✅ Establishing validation rules and error handling
- ✅ Creating documentation for frontend/mobile developers

**Prerequisites**:
- Approved BRD (Business Requirements Document)
- Completed module requirements specification
- Database schema design finalized
- Understanding of multi-tenant architecture

**Next Steps After This Skill**:
- Create SQL migrations (implementation-lokstra-create-migrations)
- Implement @Handler services (implementation-lokstra-create-handler)
- Generate .http test files (implementation-lokstra-generate-http-files)

---

## Overview

This skill generates comprehensive API specifications that serve as contracts between frontend and backend teams. Specifications include:

1. **Complete Endpoint Definitions** - HTTP methods, paths, authentication, authorization
2. **Multi-Tenant Architecture** - Tenant isolation, header validation, JWT claims
3. **Request/Response Schemas** - JSON structures with validation rules
4. **Error Handling** - All HTTP status codes with error response formats
5. **Security Specifications** - JWT tokens, permissions, rate limiting
6. **Performance Patterns** - Pagination, filtering, sorting, caching
7. **Lokstra Integration** - @Handler, @Route, @Inject mappings
8. **Testing Criteria** - Functional, security, performance tests

**Output Location**: `docs/modules/{module-name}/API_SPEC.md`

---

## Specification Process

### Step 1: Understand Module Context (5 min)

**Read Module Requirements**:
```
docs/modules/{module-name}/MODULE_REQUIREMENTS.md
```

**Extract Key Information**:
- Module name and purpose
- Functional requirements list
- Use cases
- Integration points
- Security requirements
- Multi-tenant considerations

**Questions to Answer**:
- What resources does this module expose?
- What operations are supported (CRUD, search, export)?
- Who can access which endpoints (RBAC)?
- What are the tenant isolation requirements?
- What are the performance requirements?

### Step 2: Define API Structure (10 min)

**Base URL Structure**:
```
/api/v{version}/{module}/{resource}

Example:
/api/v1/auth/login
/api/v1/patients/123
/api/v1/appointments?date=2024-01-20
```

**Resource Hierarchy**:
```yaml
Primary Resources:
  - /patients
  - /appointments
  - /prescriptions

Nested Resources:
  - /patients/{id}/appointments
  - /appointments/{id}/prescriptions

Action Endpoints:
  - /auth/login
  - /patients/bulk
  - /reports/export
```

**Endpoint Naming Conventions**:
- Use plural nouns for resources: `/patients`, `/appointments`
- Use lowercase: `/patients` not `/Patients`
- Use hyphens for multi-word: `/patient-records` not `/patientRecords`
- Use clear action verbs for non-CRUD: `/auth/login`, `/reports/export`
- Path parameters use singular: `/patients/{id}` not `/patients/{patient_id}`

### Step 3: Design Each Endpoint (20 min per endpoint)

For each endpoint, specify:

#### 3.1 Basic Information
```yaml
Endpoint ID: EP-001
Summary: Short description (5-10 words)
Method: GET | POST | PUT | PATCH | DELETE
Path: /api/v1/resource/:param
Multi-Tenant: Yes | No
Authentication: Required | Optional | No
Authorization: Permissions required
Handler Mapping: @Route annotation
```

#### 3.2 Request Specification

**Headers**:
```yaml
X-Tenant-ID:
  required: true (for multi-tenant endpoints)
  type: string
  validation: Must match JWT tenant_id claim
  example: "clinic_001"

Authorization:
  required: true (for protected endpoints)
  type: string
  format: "Bearer {jwt_token}"
  validation: Valid JWT token with required claims
```

**Path Parameters**:
```yaml
id:
  type: string
  format: UUID or ULID
  required: true
  description: Resource identifier
  example: "pat_abc123def456"
```

**Query Parameters**:
```yaml
page:
  type: integer
  default: 1
  min: 1
  description: Page number for pagination
  
limit:
  type: integer
  default: 20
  min: 1
  max: 100
  description: Items per page

status:
  type: string
  enum: [active, inactive, pending]
  description: Filter by status
```

**Request Body Schema**:
```json
{
  "type": "object",
  "required": ["field1", "field2"],
  "properties": {
    "field1": {
      "type": "string",
      "minLength": 3,
      "maxLength": 255,
      "example": "John Doe"
    },
    "field2": {
      "type": "string",
      "format": "email",
      "example": "john@example.com"
    }
  }
}
```

**Validation Rules Table**:
| Field | Type | Required | Validation Rules | Error Message |
|-------|------|----------|------------------|---------------|
| name | string | Yes | min=3, max=255 | "Name must be 3-255 characters" |
| email | string | Yes | email format | "Invalid email format" |
| age | integer | No | min=0, max=150 | "Age must be between 0 and 150" |

**Lokstra DTO Mapping**:
```go
type CreatePatientRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=255"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
}
```

#### 3.3 Response Specification

**Success Responses**:
```yaml
200 OK:
  description: Successful operation
  schema:
    {
      "status": "success",
      "data": { ... },
      "meta": { "requestId": "req_123" }
    }

201 Created:
  description: Resource created successfully
  headers:
    Location: /api/v1/resource/{id}
  schema:
    {
      "status": "success",
      "data": { "id": "...", ... }
    }

204 No Content:
  description: Successful deletion (no body)
```

**Error Responses**:
```yaml
400 Bad Request:
  code: VALIDATION_ERROR
  message: "Invalid request data"
  details: [
    { "field": "email", "message": "Invalid email format" }
  ]

401 Unauthorized:
  code: TOKEN_EXPIRED | TOKEN_INVALID | MISSING_TOKEN
  message: "Authentication required"

403 Forbidden:
  code: INSUFFICIENT_PERMISSIONS | TENANT_MISMATCH
  message: "Access denied"

404 Not Found:
  code: RESOURCE_NOT_FOUND
  message: "Resource not found"

409 Conflict:
  code: DUPLICATE_ENTRY | CONSTRAINT_VIOLATION
  message: "Resource already exists"

429 Too Many Requests:
  code: RATE_LIMIT_EXCEEDED
  message: "Rate limit exceeded"
  headers:
    Retry-After: 60
    X-RateLimit-Limit: 100
    X-RateLimit-Remaining: 0
    X-RateLimit-Reset: 1705752000

500 Internal Server Error:
  code: INTERNAL_ERROR
  message: "Internal server error"
```

#### 3.4 Multi-Tenant Specifications

**Tenant Isolation**:
```yaml
Database Queries:
  - All queries MUST include: WHERE tenant_id = $1
  - Foreign keys MUST include: tenant_id
  - Composite indexes: (tenant_id, other_columns)

API Validation:
  - Extract X-Tenant-ID from header
  - Extract tenant_id from JWT claims
  - Verify both match (return 403 if not)
  - Pass tenant_id to all service methods

Cache Keys:
  - Format: tenant:{tenant_id}:resource:{resource_type}:{id}
  - Example: tenant:clinic_001:patient:pat_123
  - Never share cache across tenants

Error Handling:
  - NEVER reveal data from other tenants
  - Generic 404 for cross-tenant access attempts
  - Log security violations with tenant context
```

**Example Query**:
```sql
-- ✅ CORRECT: With tenant filtering
SELECT * FROM patients 
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- ❌ WRONG: Missing tenant filter (SECURITY BUG!)
SELECT * FROM patients 
WHERE id = $1 AND deleted_at IS NULL;
```

#### 3.5 Security Specifications

**JWT Token Requirements**:
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "tenant_id": "clinic_001",
  "role": "doctor",
  "permissions": ["patient:read", "patient:write"],
  "iat": 1705315800,
  "exp": 1705319400,
  "iss": "lokstra-auth"
}
```

**Permission Checks**:
```yaml
Endpoint: POST /patients
Required Permission: patient:create
Allowed Roles: [admin, doctor]
Tenant Scope: Current user's tenant only
```

**Rate Limiting**:
```yaml
Endpoint: POST /auth/login
Limit: 5 requests per 15 minutes per IP
Scope: Global (all tenants)

Endpoint: GET /patients
Limit: 100 requests per minute per user
Scope: Per tenant
```

#### 3.6 Performance Specifications

**Pagination** (use cursor-based for large datasets):
```yaml
Request:
  GET /patients?limit=20&cursor=eyJpZCI6InBhdF8xMDAifQ==

Response:
  {
    "data": [...],
    "pagination": {
      "limit": 20,
      "hasNext": true,
      "hasPrev": false,
      "nextCursor": "eyJpZCI6InBhdF8xMjAifQ==",
      "prevCursor": null
    }
  }
```

**Filtering**:
```yaml
Query Parameters:
  status: Filter by status (enum)
  dateFrom: Start date (ISO 8601)
  dateTo: End date (ISO 8601)
  search: Full-text search query

SQL Implementation:
  WHERE tenant_id = $1
    AND status = ANY($2::text[])
    AND created_at BETWEEN $3 AND $4
    AND search_vector @@ to_tsquery($5)
```

**Sorting**:
```yaml
Query Parameter:
  sort: Comma-separated fields with direction
  example: sort=createdAt:desc,name:asc

SQL Implementation:
  ORDER BY created_at DESC, name ASC
```

**Caching**:
```yaml
Strategy: Cache per tenant with TTL
Cache Key: tenant:{tenant_id}:patients:{id}
TTL: 10 minutes
Invalidation: On UPDATE/DELETE operations
ETag Support: Yes (for conditional requests)
```

#### 3.7 Lokstra Handler Mapping

**@Handler Annotation**:
```go
// @Handler name="patient-handler", prefix="/api/v1/patients"
type PatientHandler struct {
    patientService *PatientService  // @Inject "patient-service"
    cache          redis.Client     // @Inject "redis"
}
```

**@Route Annotation**:
```go
// @Route "POST /", middlewares=["auth", "tenant-validation", "permission:patient:create"]
func (h *PatientHandler) CreatePatient(
    ctx *request.Context, 
    params *CreatePatientRequest,
) error {
    tenantID := ctx.Auth.TenantID()
    
    patient, err := h.patientService.CreatePatient(ctx.Context(), tenantID, params)
    if err != nil {
        return ctx.Api.InternalServerError("Failed to create patient")
    }
    
    return ctx.Api.Created(patient)
}
```

**Handler Signature Selection**:
```go
// Pattern 1: Context + DTO → Error
func (h *Handler) Method(ctx *request.Context, dto *RequestDTO) error

// Pattern 2: Context + PathParam + DTO → Error
func (h *Handler) Method(ctx *request.Context, id string, dto *RequestDTO) error

// Pattern 3: Context + PathParam → Error
func (h *Handler) Method(ctx *request.Context, id string) error

// Pattern 4: Context Only → Error
func (h *Handler) Method(ctx *request.Context) error

// Pattern 5: No Context → (Data, Error)
func (h *Handler) Method() (Data, error)
```

#### 3.8 Testing Criteria

**Functional Tests**:
- [ ] Create with valid data returns 201
- [ ] Create with invalid data returns 400
- [ ] Get existing resource returns 200
- [ ] Get non-existent resource returns 404
- [ ] Update with valid data returns 200
- [ ] Delete existing resource returns 204
- [ ] Pagination works correctly
- [ ] Filtering works correctly
- [ ] Sorting works correctly

**Multi-Tenant Tests**:
- [ ] User cannot access data from other tenants
- [ ] X-Tenant-ID and JWT tenant_id mismatch returns 403
- [ ] Missing X-Tenant-ID returns 400
- [ ] Cross-tenant foreign key constraints enforced

**Security Tests**:
- [ ] Missing JWT token returns 401
- [ ] Expired JWT token returns 401
- [ ] Insufficient permissions return 403
- [ ] Rate limiting works correctly
- [ ] SQL injection attempts blocked
- [ ] XSS attempts sanitized

**Performance Tests**:
- [ ] Response time < target (e.g., p95 < 200ms)
- [ ] Handles concurrent requests (e.g., 100 concurrent)
- [ ] Pagination performance acceptable for large datasets
- [ ] Caching reduces database load

### Step 4: Document Examples (10 min per endpoint)

**Provide Complete Examples**:

**Request Example**:
```bash
curl -X POST https://api.example.com/api/v1/patients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "X-Tenant-ID: clinic_001" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+6281234567890",
    "dateOfBirth": "1990-01-15"
  }'
```

**Success Response Example**:
```json
{
  "status": "success",
  "message": "Patient created successfully",
  "data": {
    "id": "pat_abc123def456",
    "tenantId": "clinic_001",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+6281234567890",
    "dateOfBirth": "1990-01-15",
    "createdAt": "2024-01-20T10:30:00Z",
    "updatedAt": "2024-01-20T10:30:00Z"
  },
  "meta": {
    "requestId": "req_xyz789"
  }
}
```

**Error Response Example**:
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": [
      {
        "field": "email",
        "message": "Email already registered",
        "code": "DUPLICATE_EMAIL"
      }
    ]
  },
  "meta": {
    "requestId": "req_xyz789"
  }
}
```

### Step 5: Validate with Checklist (15 min)

**Use the validation checklist**: [API_VALIDATION_CHECKLIST.md](references/API_VALIDATION_CHECKLIST.md)

**Critical Checks**:
- ❌ FAIL if any query missing tenant_id filter
- ❌ FAIL if tenant_id optional or from request body
- ❌ FAIL if cache keys don't include tenant_id
- ❌ FAIL if errors leak cross-tenant information
- ❌ FAIL if missing required sections

**Minimum Quality Score**: 85% with no FAIL items

### Step 6: Generate Final Document (5 min)

**Use template**: [API_SPEC_TEMPLATE.md](references/API_SPEC_TEMPLATE.md)

**Document Structure**:
```markdown
# {Module Name} API Specification

## Overview
- Module purpose
- Base URL
- Multi-tenant strategy
- Dependencies

## Common Headers
- X-Tenant-ID
- Authorization
- Content-Type

## Endpoints
### 1. POST /resource
- Request specification
- Response specification
- Examples
- Lokstra mapping

### 2. GET /resource
...

## Security Specifications
- JWT requirements
- Permissions
- Rate limiting

## Multi-Tenant Implementation
- Tenant isolation
- Validation rules
- Cache strategy

## Error Response Format
- Standard structure
- HTTP status codes
- Error code hierarchy

## Testing Checklist
- Functional tests
- Security tests
- Performance tests

## Integration Points
- Dependencies
- Events published
- External APIs
```

---

## Multi-Tenant Checklist (Critical)

Before finalizing API specification, verify:

### Database Level
- [ ] All tables include `tenant_id` column
- [ ] All foreign keys include `tenant_id`
- [ ] All queries filter by `tenant_id`
- [ ] Composite unique indexes include `tenant_id`
- [ ] Row-level security policies (optional but recommended)

### API Level
- [ ] X-Tenant-ID header required for all non-public endpoints
- [ ] JWT token includes `tenant_id` claim
- [ ] Validate header tenant_id matches JWT tenant_id
- [ ] Return 403 on tenant mismatch
- [ ] All service methods accept tenantID parameter

### Cache Level
- [ ] All cache keys include tenant_id prefix
- [ ] Format: `tenant:{tenant_id}:resource:{type}:{id}`
- [ ] Cache invalidation tenant-scoped
- [ ] No shared cache across tenants

### Error Handling
- [ ] Generic 404 for cross-tenant access (don't reveal existence)
- [ ] No tenant-specific information in error messages
- [ ] Audit logs include tenant_id
- [ ] Security violations logged with tenant context

### Testing
- [ ] Cross-tenant isolation tests defined
- [ ] Tenant mismatch tests defined
- [ ] Performance isolation tests defined
- [ ] Security boundary tests defined

---

## Error Code Standards

### Format
```
{DOMAIN}_{CATEGORY}_{DETAIL}

Examples:
- AUTH_VALIDATION_EMAIL_INVALID
- PATIENT_AUTHORIZATION_INSUFFICIENT_PERMISSION
- APPOINTMENT_CONFLICT_TIME_OVERLAP
- TENANT_VALIDATION_MISMATCH
```

### Categories
- **VALIDATION**: Input validation errors (400)
- **AUTHORIZATION**: Permission/access errors (403)
- **AUTHENTICATION**: Token/credential errors (401)
- **NOT_FOUND**: Resource not found (404)
- **CONFLICT**: Duplicate/constraint violations (409)
- **RATE_LIMIT**: Too many requests (429)
- **INTERNAL**: Server errors (500)

### Common Error Codes
```yaml
Authentication:
  - AUTH_TOKEN_MISSING
  - AUTH_TOKEN_EXPIRED
  - AUTH_TOKEN_INVALID
  - AUTH_CREDENTIALS_INVALID

Authorization:
  - AUTH_INSUFFICIENT_PERMISSIONS
  - AUTH_ACCOUNT_LOCKED
  - AUTH_ACCOUNT_DISABLED

Tenant:
  - TENANT_ID_MISSING
  - TENANT_MISMATCH
  - TENANT_NOT_FOUND
  - TENANT_DISABLED

Validation:
  - VALIDATION_REQUIRED_FIELD
  - VALIDATION_INVALID_FORMAT
  - VALIDATION_OUT_OF_RANGE
  - VALIDATION_INVALID_ENUM

Conflict:
  - CONFLICT_DUPLICATE_ENTRY
  - CONFLICT_CONSTRAINT_VIOLATION
  - CONFLICT_RESOURCE_IN_USE

Rate Limit:
  - RATE_LIMIT_EXCEEDED
  - RATE_LIMIT_TENANT_EXCEEDED
```

---

## Performance Guidelines

### Response Time Targets
```yaml
Simple GET (by ID):
  Target: p95 < 100ms
  Max: p99 < 200ms

List/Search Operations:
  Target: p95 < 200ms
  Max: p99 < 500ms

Create/Update Operations:
  Target: p95 < 200ms
  Max: p99 < 500ms

Complex Reports:
  Target: p95 < 1s
  Max: p99 < 3s
```

### Pagination Limits
```yaml
Default Page Size: 20
Maximum Page Size: 100
Minimum Page Size: 1

For large datasets (>10K records):
  Use cursor-based pagination
  
For small datasets (<10K records):
  Offset pagination acceptable
```

### Caching Strategy
```yaml
Static Data:
  TTL: 1 hour
  Examples: Config, lookup tables

Volatile Data:
  TTL: 5-15 minutes
  Examples: User profiles, patient records

Real-time Data:
  No caching
  Examples: Live appointment status

Cache Key Pattern:
  tenant:{tenant_id}:resource:{type}:{id}
```

---

## Resources & References

### Core Documentation
- **Template**: [API_SPEC_TEMPLATE.md](references/API_SPEC_TEMPLATE.md) - Full API specification template (656 lines)
- **Validation Checklist**: [API_VALIDATION_CHECKLIST.md](references/API_VALIDATION_CHECKLIST.md) - Comprehensive quality gates (13 KB, 15 sections)
- **Lokstra Integration**: [LOKSTRA_INTEGRATION_GUIDE.md](references/LOKSTRA_INTEGRATION_GUIDE.md) - @Handler, @Route, @Inject patterns (18 KB)

### Multi-Tenant Patterns
- **Multi-Tenant Guide**: [MULTI_TENANT_API_PATTERNS.md](references/MULTI_TENANT_API_PATTERNS.md) - Tenant isolation, security, performance (18 KB)
- **API Patterns**: [COMMON_API_PATTERNS.md](references/COMMON_API_PATTERNS.md) - Pagination, filtering, sorting, bulk ops (16 KB)

### Real Examples
- **Auth API Example**: [AUTH_API_EXAMPLE.md](references/AUTH_API_EXAMPLE.md) - Complete multi-tenant authentication API (19 KB, 6 endpoints)

### Framework Documentation
- Lokstra Quick Reference: `.github/copilot-instructions.md`
- Handler Signatures: 29+ supported patterns
- Validation Tags: Go struct tag format

---

## Tips & Best Practices

### DO
✅ Use clear, descriptive endpoint names  
✅ Include complete request/response examples  
✅ Document ALL possible error scenarios  
✅ Specify multi-tenant requirements explicitly  
✅ Map to Lokstra @Handler and @Route annotations  
✅ Use cursor-based pagination for large datasets  
✅ Include tenant_id in all queries and cache keys  
✅ Document rate limiting per endpoint  
✅ Specify validation rules for all fields  
✅ Include Lokstra DTO structs with validation tags

### DON'T
❌ Leave tenant_id filtering optional  
❌ Trust client-provided tenant_id (use JWT)  
❌ Share cache keys across tenants  
❌ Reveal cross-tenant information in errors  
❌ Use offset pagination for large datasets  
❌ Omit error response examples  
❌ Skip security specifications  
❌ Forget to document permissions  
❌ Miss rate limiting specifications  
❌ Use inconsistent naming conventions

### Common Mistakes
1. **Missing tenant_id in queries** → Add `WHERE tenant_id = $1` to ALL queries
2. **Vague validation rules** → Specify min/max, patterns, formats
3. **Incomplete error handling** → Document all 4xx and 5xx scenarios
4. **No Lokstra mapping** → Include @Handler, @Route, DTO structs
5. **Missing examples** → Provide curl commands and JSON examples
6. **No performance specs** → Define pagination, caching, rate limits
7. **Generic descriptions** → Be specific about behavior and constraints

---

## Quality Checklist

Before submitting API specification:

### Completeness (Required)
- [ ] All endpoints documented
- [ ] All request parameters defined
- [ ] All response scenarios covered
- [ ] All error codes documented
- [ ] Examples for all endpoints
- [ ] Lokstra annotations mapped
- [ ] Multi-tenant specifications complete

### Security (Critical)
- [ ] Authentication requirements specified
- [ ] Authorization/permissions documented
- [ ] JWT token structure defined
- [ ] Rate limiting configured
- [ ] tenant_id filtering enforced
- [ ] No cross-tenant data leakage

### Quality (Important)
- [ ] Validation rules complete
- [ ] Error messages user-friendly
- [ ] Naming conventions consistent
- [ ] Response formats standardized
- [ ] Performance targets defined
- [ ] Testing criteria specified

### Lokstra Integration (Required)
- [ ] @Handler annotations documented
- [ ] @Route annotations documented
- [ ] Handler signatures correct
- [ ] @Inject dependencies specified
- [ ] DTO structs with validation tags
- [ ] Response helpers mapped

**Minimum Score**: 85% (all sections passing)

---

## Examples

### Example 1: Simple CRUD Endpoint

```yaml
Endpoint: GET /patients/{id}
Method: GET
Path: /api/v1/patients/:id
Multi-Tenant: Yes
Authentication: Required
Permissions: patient:read
```

**Request**:
```
GET /api/v1/patients/pat_abc123
Headers:
  Authorization: Bearer eyJhbGc...
  X-Tenant-ID: clinic_001
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "data": {
    "id": "pat_abc123",
    "tenantId": "clinic_001",
    "name": "John Doe",
    "email": "john@example.com",
    "createdAt": "2024-01-20T10:30:00Z"
  }
}
```

**Lokstra Mapping**:
```go
// @Route "GET /:id", middlewares=["auth", "permission:patient:read"]
func (h *PatientHandler) GetPatient(ctx *request.Context, id string) error {
    tenantID := ctx.Auth.TenantID()
    patient, err := h.patientService.GetPatient(ctx.Context(), tenantID, id)
    if err != nil {
        return ctx.Api.NotFound("Patient not found")
    }
    return ctx.Api.Ok(patient)
}
```

### Example 2: Complex Search Endpoint

```yaml
Endpoint: GET /patients
Method: GET
Path: /api/v1/patients
Multi-Tenant: Yes
Authentication: Required
Permissions: patient:read
```

**Request**:
```
GET /api/v1/patients?status=active&limit=20&cursor=eyJ...&sort=name:asc
Headers:
  Authorization: Bearer eyJhbGc...
  X-Tenant-ID: clinic_001
```

**Lokstra DTO**:
```go
type ListPatientsQuery struct {
    Status string `query:"status" validate:"omitempty,oneof=active inactive"`
    Limit  int    `query:"limit" validate:"omitempty,min=1,max=100" default:"20"`
    Cursor string `query:"cursor" validate:"omitempty,base64"`
    Sort   string `query:"sort" validate:"omitempty"`
}
```

---

## Workflow Summary

```
1. Read Module Requirements (5 min)
   ↓
2. Define API Structure (10 min)
   ↓
3. Design Each Endpoint (20 min per endpoint)
   - Basic info
   - Request spec
   - Response spec
   - Multi-tenant specs
   - Security specs
   - Performance specs
   - Lokstra mapping
   - Testing criteria
   ↓
4. Document Examples (10 min per endpoint)
   ↓
5. Validate with Checklist (15 min)
   ↓
6. Generate Final Document (5 min)
   ↓
7. Peer Review (30 min)
   ↓
8. Approval & Lock Specification
```

**Total Time**: ~1-2 hours per module (depends on complexity)

---

**Last Updated**: 2024-01-20  
**Version**: 2.0.0  
**Phase**: Design  
**Next Skill**: implementation-lokstra-create-handler
