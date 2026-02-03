---
name: implementation-lokstra-generate-http-files
description: Generate .http client files for testing API endpoints. Create executable HTTP requests with environment variables, request/response examples, and curl equivalents. Use after API specification is finalized for quick endpoint testing.
phase: implementation
order: 6
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Implementation: Generate HTTP Client Files

## When to Use

Use this skill when:
- Creating executable HTTP requests for endpoint testing
- Documenting API usage with live examples
- Setting up development testing without external tools
- Sharing API examples with team
- Testing validation rules and error handling
- Verifying multi-tenant isolation

Prerequisites:
- ✅ API specification finalized (see: design-lokstra-api-specification)
- ✅ Handler implementation complete (see: implementation-lokstra-create-handler)
- ✅ Application running locally or on server
- ✅ Environment variables configured
- ✅ Database migrated and seeded (optional)

**Next Steps After This Skill**:
- Write unit tests (see: advanced-lokstra-tests)
- Validate consistency (see: advanced-lokstra-validate-consistency)

---

## Overview

This skill generates `.http` files (VS Code REST Client format) for:

1. **Complete CRUD Testing** - All endpoints per module
2. **Validation Testing** - Error cases and edge conditions
3. **Authentication Testing** - JWT, API keys, Bearer tokens
4. **Multi-Tenant Testing** - Tenant isolation verification
5. **Workflow Testing** - Sequential operations with response capture
6. **Performance Testing** - Pagination, filtering, sorting

**Output Location**: `api/{module}.http` or `test/{module}.http`

---

## Tool Requirements

VS Code extension: [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

Install via VS Code Extensions panel or command line:
```bash
code --install-extension humao.rest-client
```

---

## HTTP File Basics

### Simple GET Request

File: `api/user.http`

```http
### Get user by ID
GET http://localhost:8080/api/users/550e8400-e29b-41d4-a716-446655440000
Accept: application/json
```

### POST with Request Body

```http
### Create user
POST http://localhost:8080/api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

### Path Parameters and Variables

```http
@baseUrl = http://localhost:8080
@userId = 550e8400-e29b-41d4-a716-446655440000

### Get user
GET {{baseUrl}}/api/users/{{userId}}

### Update user
PUT {{baseUrl}}/api/users/{{userId}}
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com"
}

### Delete user
DELETE {{baseUrl}}/api/users/{{userId}}
```

---

## Environment Configuration

### File: .http-client.env.json

```json
{
  "development": {
    "baseUrl": "http://localhost:8080",
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "authToken": "dev-jwt-token"
  },
  "production": {
    "baseUrl": "https://api.myapp.com",
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "authToken": "prod-jwt-token"
  }
}
```

### Using Environment Variables

```http
### Using variables
GET {{baseUrl}}/api/users/{{userId}}
Authorization: Bearer {{authToken}}
```

---

## Complete API Test Suite

### File: api/users.http

```http
@baseUrl = http://localhost:8080
@contentType = application/json

### ============================================
### USER API - Test Suite
### ============================================

### 1. Create User (POST)
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "name": "John Doe",
  "email": "john@example.com"
}

> {%
  client.global.set("userId", response.body.id);
  client.log("Saved userId: " + client.global.get("userId"));
%}

###

### 2. Get User (GET)
GET {{baseUrl}}/api/users/{{userId}}
Accept: {{contentType}}

###

### 3. Update User (PUT)
PUT {{baseUrl}}/api/users/{{userId}}
Content-Type: {{contentType}}

{
  "name": "Jane Doe",
  "email": "jane@example.com"
}

###

### 4. Delete User (DELETE)
DELETE {{baseUrl}}/api/users/{{userId}}

###

### 5. List Users (GET)
GET {{baseUrl}}/api/users
Accept: {{contentType}}
```

---

## Authentication Patterns

### Bearer Token

```http
@baseUrl = http://localhost:8080
@token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

### Authenticated request
GET {{baseUrl}}/api/users/profile
Authorization: Bearer {{token}}
```

### Basic Auth

```http
### Basic authentication
GET {{baseUrl}}/api/users
Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

### API Key

```http
### API key in header
GET {{baseUrl}}/api/users
X-API-Key: your-api-key-here
```

### Login and Get Token

```http
@baseUrl = http://localhost:8080

### 1. Login
POST {{baseUrl}}/api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

> {%
  client.global.set("token", response.body.token);
%}

###

### 2. Use token in subsequent requests
GET {{baseUrl}}/api/users/profile
Authorization: Bearer {{token}}
```

---

## Request/Response Examples

### With Headers

```http
### Request with custom headers
POST {{baseUrl}}/api/users
Content-Type: application/json
Accept: application/json
X-Request-ID: req-12345
X-Client-Version: 1.0.0

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

### Query Parameters

```http
@baseUrl = http://localhost:8080

### List with pagination
GET {{baseUrl}}/api/users?page=1&limit=10&sort=created_at&order=desc

###

### Filter by status
GET {{baseUrl}}/api/users?status=active&role=admin
```

### File Upload

```http
### Upload file
POST {{baseUrl}}/api/users/{{userId}}/avatar
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="avatar.jpg"
Content-Type: image/jpeg

< ./avatar.jpg
------WebKitFormBoundary--
```

---

## Response Handling

### Capture Values from Response

```http
### Create and capture ID
POST {{baseUrl}}/api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}

> {%
  // Extract and save response fields
  client.global.set("userId", response.body.id);
  client.global.set("email", response.body.email);
  
  // Log to console
  client.log("Created user: " + response.body.id);
  
  // Assert response status
  client.assert(response.status === 201, "Expected 201 Created");
%}

###

### Use captured value
GET {{baseUrl}}/api/users/{{userId}}
```

### Response Assertions

```http
### Test response
GET {{baseUrl}}/api/users/{{userId}}

> {%
  // Check status code
  client.assert(response.status === 200, "Expected 200 OK");
  
  // Check response body
  client.assert(response.body.id !== null, "User ID should not be null");
  client.assert(response.body.email === "john@example.com", "Email mismatch");
  
  // Check headers
  client.assert(response.headers.valueOf("content-type").includes("application/json"), "Invalid content type");
  
  // Log results
  client.log("All assertions passed!");
%}
```

---

## Module-Organized Structure

```
api/
├── user.http            # User endpoints
├── auth.http            # Authentication endpoints
├── order.http           # Order endpoints
├── notification.http    # Notification endpoints
├── variables.http       # Shared variables
└── .http-client.env.json  # Environment configs
```

### File: api/variables.http

```http
### Shared variables and test data
@baseUrl = http://localhost:8080
@contentType = application/json

@testUserName = John Doe
@testUserEmail = john@example.com

@adminToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
@userToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Curl Equivalents

Each `.http` file can be converted to curl:

```http
### Create user
POST {{baseUrl}}/api/users
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

Curl equivalent:

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"John Doe","email":"john@example.com"}'
```

---

## Multi-Tenant Testing

For Lokstra applications with multi-tenant architecture, include tenant-specific headers and validation.

### Multi-Tenant Environment Configuration

File: `.http-client.env.json`

```json
{
  "tenant_a": {
    "baseUrl": "http://localhost:3000",
    "tenantId": "tenant_a",
    "adminToken": "eyJhbGc..._tenant_a_admin",
    "userToken": "eyJhbGc..._tenant_a_user"
  },
  "tenant_b": {
    "baseUrl": "http://localhost:3000",
    "tenantId": "tenant_b",
    "adminToken": "eyJhbGc..._tenant_b_admin",
    "userToken": "eyJhbGc..._tenant_b_user"
  }
}
```

### Tenant-Isolated Requests

File: `api/tenant-users.http`

```http
@baseUrl = http://localhost:3000
@contentType = application/json
@tenantId = tenant_a

### ============================================
### MULTI-TENANT USER API - Test Suite
### ============================================

### 1. Create User for Tenant A
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}
Authorization: Bearer {{adminToken}}
X-Tenant-ID: {{tenantId}}

{
  "name": "John Doe",
  "email": "john@tenant-a.com"
}

> {%
  client.global.set("tenantAUserId", response.body.data.id);
  client.assert(response.status === 201, "Expected 201 Created");
  client.assert(response.body.data.tenant_id === "tenant_a", "Tenant ID mismatch");
%}

###

### 2. Verify Tenant Isolation - Tenant B cannot access Tenant A's user
GET {{baseUrl}}/api/users/{{tenantAUserId}}
Authorization: Bearer {{tenant_b_token}}
X-Tenant-ID: tenant_b

> {%
  // Should return 403 Forbidden or 404 Not Found
  client.assert(response.status === 403 || response.status === 404, 
    "Expected tenant isolation - 403 or 404");
%}

###

### 3. List Users - Should only return current tenant's users
GET {{baseUrl}}/api/users
Authorization: Bearer {{adminToken}}
X-Tenant-ID: {{tenantId}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  // All returned users should belong to current tenant
  const users = response.body.data || [];
  users.forEach(user => {
    client.assert(user.tenant_id === client.global.get("tenantId"), 
      "User from different tenant found!");
  });
%}
```

---

## Validation Testing

Test all validation rules defined in your DTOs.

### Validation Test Cases

File: `api/users-validation.http`

```http
@baseUrl = http://localhost:3000
@contentType = application/json

### ============================================
### USER VALIDATION TESTS
### ============================================

### Test: Missing required field (name)
# Expected: 400 Bad Request with validation error
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "email": "john@example.com"
}

> {%
  client.assert(response.status === 400, "Expected 400 for missing required field");
  client.assert(response.body.message.includes("validation") || 
                response.body.error.includes("required"), 
                "Expected validation error message");
%}

###

### Test: Invalid email format
# Expected: 400 Bad Request with email validation error
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "name": "John Doe",
  "email": "not-an-email"
}

> {%
  client.assert(response.status === 400, "Expected 400 for invalid email");
%}

###

### Test: Name too short (min=3)
# Expected: 400 Bad Request
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "name": "Jo",
  "email": "john@example.com"
}

> {%
  client.assert(response.status === 400, "Expected 400 for name too short");
%}

###

### Test: Name too long (max=50)
# Expected: 400 Bad Request
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "name": "This name is way too long and exceeds the maximum allowed characters for a user name",
  "email": "john@example.com"
}

> {%
  client.assert(response.status === 400, "Expected 400 for name too long");
%}

###

### Test: Duplicate email (unique constraint)
# Expected: 409 Conflict or 400 Bad Request
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{
  "name": "Jane Doe",
  "email": "existing@example.com"
}

> {%
  client.assert(response.status === 409 || response.status === 400, 
    "Expected conflict or bad request for duplicate email");
%}

###

### Test: Invalid status value (oneof validation)
# Expected: 400 Bad Request
PUT {{baseUrl}}/api/users/{{userId}}
Content-Type: {{contentType}}

{
  "name": "John Doe",
  "status": "invalid_status"
}

> {%
  client.assert(response.status === 400, "Expected 400 for invalid enum value");
%}
```

---

## Error Response Testing

Test all error scenarios and HTTP status codes.

### Error Cases Test Suite

File: `api/users-errors.http`

```http
@baseUrl = http://localhost:3000
@contentType = application/json

### ============================================
### ERROR RESPONSE TESTS
### ============================================

### 400 Bad Request - Malformed JSON
POST {{baseUrl}}/api/users
Content-Type: {{contentType}}

{invalid json}

> {%
  client.assert(response.status === 400, "Expected 400 for malformed JSON");
%}

###

### 401 Unauthorized - Missing token
GET {{baseUrl}}/api/users/profile

> {%
  client.assert(response.status === 401, "Expected 401 for missing auth");
%}

###

### 401 Unauthorized - Invalid token
GET {{baseUrl}}/api/users/profile
Authorization: Bearer invalid-token-here

> {%
  client.assert(response.status === 401, "Expected 401 for invalid token");
%}

###

### 403 Forbidden - Insufficient permissions
DELETE {{baseUrl}}/api/admin/users/{{userId}}
Authorization: Bearer {{userToken}}

> {%
  client.assert(response.status === 403, "Expected 403 for insufficient permissions");
%}

###

### 404 Not Found - Non-existent resource
GET {{baseUrl}}/api/users/non-existent-uuid-here

> {%
  client.assert(response.status === 404, "Expected 404 for non-existent resource");
%}

###

### 405 Method Not Allowed
PATCH {{baseUrl}}/api/health
Content-Type: {{contentType}}

{}

> {%
  client.assert(response.status === 405 || response.status === 404, 
    "Expected 405 Method Not Allowed or 404");
%}

###

### 415 Unsupported Media Type
POST {{baseUrl}}/api/users
Content-Type: text/plain

name=John&email=john@example.com

> {%
  client.assert(response.status === 415 || response.status === 400, 
    "Expected 415 or 400 for wrong content type");
%}
```

---

## Pagination and Filtering Tests

Test query parameter handling for list endpoints.

### Pagination Test Cases

File: `api/users-pagination.http`

```http
@baseUrl = http://localhost:3000
@contentType = application/json

### ============================================
### PAGINATION & FILTERING TESTS
### ============================================

### Default pagination
GET {{baseUrl}}/api/users
Accept: {{contentType}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  client.assert(response.body.meta.page === 1, "Default page should be 1");
  client.assert(response.body.meta.limit <= 20, "Default limit should be <= 20");
%}

###

### Custom page and limit
GET {{baseUrl}}/api/users?page=2&limit=5
Accept: {{contentType}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  client.assert(response.body.meta.page === 2, "Page should be 2");
  client.assert(response.body.meta.limit === 5, "Limit should be 5");
%}

###

### Filter by status
GET {{baseUrl}}/api/users?status=active
Accept: {{contentType}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  const users = response.body.data || [];
  users.forEach(user => {
    client.assert(user.status === "active", "All users should be active");
  });
%}

###

### Sort by created_at descending
GET {{baseUrl}}/api/users?sort=created_at&order=desc
Accept: {{contentType}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
%}

###

### Search by name
GET {{baseUrl}}/api/users?q=John
Accept: {{contentType}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
%}

###

### Invalid pagination parameters
GET {{baseUrl}}/api/users?page=-1&limit=1000
Accept: {{contentType}}

> {%
  // Should either normalize values or return 400
  client.assert(response.status === 200 || response.status === 400, 
    "Expected 200 (normalized) or 400 (validation error)");
%}
```

---

## Complete Module Template

Use this template when generating .http files for a new module.

### Template: `api/{module}.http`

```http
### ============================================
### {MODULE_NAME} API - Test Suite
### Generated from API Specification
### ============================================

@baseUrl = http://localhost:3000
@contentType = application/json
@authToken = {{$dotenv AUTH_TOKEN}}

### ============================================
### HEALTH CHECK
### ============================================

### Health check
GET {{baseUrl}}/health

> {%
  client.assert(response.status === 200, "Health check failed");
%}

###

### ============================================
### CRUD OPERATIONS
### ============================================

### 1. CREATE - POST /{resource}
# @name create{Resource}
POST {{baseUrl}}/api/{resources}
Content-Type: {{contentType}}
Authorization: Bearer {{authToken}}

{
  "field1": "value1",
  "field2": "value2"
}

> {%
  client.global.set("{resource}Id", response.body.data.id);
  client.assert(response.status === 201, "Expected 201 Created");
%}

###

### 2. READ - GET /{resource}/{id}
GET {{baseUrl}}/api/{resources}/{{{resource}Id}}
Accept: {{contentType}}
Authorization: Bearer {{authToken}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  client.assert(response.body.data.id === client.global.get("{resource}Id"), "ID mismatch");
%}

###

### 3. LIST - GET /{resources}
GET {{baseUrl}}/api/{resources}?page=1&limit=10
Accept: {{contentType}}
Authorization: Bearer {{authToken}}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
  client.assert(Array.isArray(response.body.data), "Expected array of {resources}");
%}

###

### 4. UPDATE - PUT /{resource}/{id}
PUT {{baseUrl}}/api/{resources}/{{{resource}Id}}
Content-Type: {{contentType}}
Authorization: Bearer {{authToken}}

{
  "field1": "updated_value1",
  "field2": "updated_value2"
}

> {%
  client.assert(response.status === 200, "Expected 200 OK");
%}

###

### 5. DELETE - DELETE /{resource}/{id}
DELETE {{baseUrl}}/api/{resources}/{{{resource}Id}}
Authorization: Bearer {{authToken}}

> {%
  client.assert(response.status === 200 || response.status === 204, 
    "Expected 200 or 204");
%}

###

### ============================================
### VALIDATION TESTS
### ============================================

### Validation: Missing required field
POST {{baseUrl}}/api/{resources}
Content-Type: {{contentType}}
Authorization: Bearer {{authToken}}

{
  "field2": "value2"
}

> {%
  client.assert(response.status === 400, "Expected 400 for missing required field");
%}

###

### ============================================
### ERROR CASES
### ============================================

### Error: Not Found
GET {{baseUrl}}/api/{resources}/non-existent-id
Authorization: Bearer {{authToken}}

> {%
  client.assert(response.status === 404, "Expected 404 Not Found");
%}

###

### Error: Unauthorized
GET {{baseUrl}}/api/{resources}

> {%
  client.assert(response.status === 401, "Expected 401 Unauthorized");
%}
```

---

## Testing Workflow

### Step 1: Start Application

```bash
go run .
```

### Step 2: Open HTTP File

Open `api/{module}.http` in VS Code with REST Client extension

### Step 3: Execute Requests

Click "Send Request" above each request block

### Step 4: Review Responses

- Check status code (shown in response header)
- Verify response body matches expectations
- Review response headers

### Step 5: Capture Values

Use response handlers (`> {% ... %}`) to save values for chained requests

### Step 6: Run Full Test Suite

Execute all requests sequentially to test complete workflow

---

## AI Agent Instructions

When generating HTTP files, follow this process:

### Step 1: Analyze API Specification (Required)

Read the API specification document to understand:
- All endpoints (method, path, description)
- Request body schemas with validation rules
- Response schemas and status codes
- Authentication requirements
- Multi-tenant requirements (X-Tenant-ID header)

### Step 2: Create File Structure

```
api/
├── {module}.http                 # Main CRUD tests
├── {module}-validation.http      # Validation test cases  
├── {module}-errors.http          # Error response tests
└── .http-client.env.json         # Environment configuration
```

### Step 3: Generate Content

For each endpoint in the API specification:

1. **Create basic request** with all required headers
2. **Add request body** matching DTO schema
3. **Add response assertions** for expected status code
4. **Add response capture** for dynamic values (IDs, tokens)
5. **Add descriptive comment** explaining the test

### Step 4: Include Test Categories

Ensure coverage of:

| Category | Required Tests |
|----------|----------------|
| **CRUD** | Create, Read, Update, Delete, List |
| **Validation** | Missing required fields, invalid formats, boundary values |
| **Auth** | Missing token, invalid token, expired token |
| **Permissions** | Forbidden actions, role-based access |
| **Multi-tenant** | Tenant isolation, cross-tenant access denied |
| **Pagination** | Default values, custom page/limit, invalid params |
| **Search/Filter** | Query params, sorting, filtering |

### Step 5: Naming Conventions

```http
### [Action] [Resource] [Condition]
### Create User with valid data
### Get User by ID
### Update User with invalid email
### Delete User without permission
### List Users with pagination
```

### DO's and DON'Ts

#### DO's ✅
- Include descriptive comments for every request
- Add response assertions for critical validations
- Use variables for reusable values (baseUrl, tokens)
- Test both success and error scenarios
- Capture dynamic values (IDs) for sequential requests
- Match validation rules from DTOs (min, max, required)
- Include multi-tenant headers when applicable
- Use environment files for sensitive data

#### DON'Ts ❌
- Hardcode authentication tokens in files
- Skip error case testing
- Ignore multi-tenant isolation tests
- Forget to add Content-Type headers
- Use invalid/placeholder UUIDs without noting
- Skip pagination boundary tests
- Ignore validation constraint testing
- Leave requests without descriptions

---

## Best Practices

### 1. Organize by Module

```
api/
├── auth.http              # Authentication endpoints
├── users.http             # User CRUD + profile
├── users-validation.http  # User validation tests
├── orders.http            # Order management
└── products.http          # Product catalog
```

### 2. Use Descriptive Request Names

```http
### ✅ Good - Clear and descriptive
### Create new user with email validation
POST {{baseUrl}}/api/users

### ❌ Bad - No description
POST {{baseUrl}}/api/users
```

### 3. Document Test Scenarios

```http
### Create new user with validation
# Tests:
# - Required fields: name, email
# - Email format: validate:"email"
# - Name length: min=3, max=50
# Expected: 201 Created with user object
POST {{baseUrl}}/api/users
```

### 4. Keep Sensitive Data in Environment

```json
// .http-client.env.json
{
  "development": {
    "adminToken": "dev-token-here",
    "dbPassword": "never-commit-this"
  }
}
```

**Important**: Add `.http-client.env.json` to `.gitignore`!

### 5. Group Related Tests

```http
### ============================================
### USER CRUD OPERATIONS
### ============================================

### Create User
...

### Get User
...

### ============================================
### USER VALIDATION TESTS  
### ============================================

### Invalid email format
...
```

---

## Summary Checklist

When generating HTTP files for a module:

### File Structure
- [ ] Main CRUD file: `api/{module}.http`
- [ ] Validation tests: `api/{module}-validation.http`
- [ ] Error tests: `api/{module}-errors.http`
- [ ] Environment config: `.http-client.env.json`

### Test Coverage
- [ ] All CRUD endpoints (Create, Read, Update, Delete, List)
- [ ] All validation rules from DTOs tested
- [ ] All error status codes (400, 401, 403, 404, 409, 500)
- [ ] Authentication scenarios (valid, invalid, missing)
- [ ] Authorization scenarios (permitted, forbidden)
- [ ] Pagination and filtering
- [ ] Multi-tenant isolation (if applicable)

### Request Quality
- [ ] Descriptive comment for every request
- [ ] Correct HTTP method and path
- [ ] Required headers (Content-Type, Authorization)
- [ ] Valid request body matching DTO schema
- [ ] Response assertions where appropriate

### Variables and Environment
- [ ] Base URL as variable
- [ ] Auth tokens in environment file
- [ ] Dynamic values captured from responses
- [ ] Tenant IDs configured (if multi-tenant)

### Documentation
- [ ] Section headers for organization
- [ ] Test scenario descriptions
- [ ] Expected behavior documented
- [ ] Curl equivalents for sharing (optional)

---

## Next Steps

After generating HTTP files:

1. **Run Application**: `go run .`
2. **Execute Test Suite**: Run all requests in sequence
3. **Verify Responses**: Check status codes and body content
4. **Document Issues**: Note any failures for handler fixes
5. **Create Unit Tests**: See [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)
6. **Validate Consistency**: See [advanced-lokstra-validate-consistency](../advanced-lokstra-validate-consistency/SKILL.md)

---

## Related Skills

- [design-lokstra-api-specification](../design-lokstra-api-specification/SKILL.md) - API design and contracts
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - Handler implementation with DTOs
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Configuration and service definitions
- [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md) - Automated unit and integration tests
- [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md) - Authentication and authorization middleware

---

**Last Updated**: 2024-01-20  
**Version**: 2.0.0  
**Phase**: Implementation  
**Previous Skill**: implementation-lokstra-create-migrations  
**Next Skill**: advanced-lokstra-tests
