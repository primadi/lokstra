---
name: lokstra-api-specification
description: Generate OpenAPI-style API specifications for Lokstra modules. Creates detailed endpoint definitions with request/response schemas, validation rules, error codes, and examples. Use after module requirements are approved.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  phase: design
  order: 4
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---

# Lokstra API Specification Generation

## When to use this skill

Use this skill when:
- Module requirements are approved
- Need detailed API endpoint specifications
- Defining request/response contracts
- Before code implementation

## How to generate API specification

Save to: `docs/modules/{module-name}/API_SPEC.md`

### Key Sections Required

1. **Endpoints** - HTTP method, path, authentication
2. **Request/Response Schemas** - JSON structures
3. **Validation Rules** - Field constraints
4. **Error Responses** - 400, 401, 403, 404, 500
5. **Examples** - Request/response examples

### Endpoint Format

```markdown
### EP-001: [Endpoint Name]
**Method:** GET/POST/PUT/DELETE  
**Path:** `/api/{module}/{resource}`  
**Authentication:** Required/Optional  
**Authorization:** Roles: [role1, role2]

#### Request
**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Body:**
```json
{
  "field1": "value1",
  "field2": "value2"
}
```

**Validation:**
| Field | Type | Required | Rules | Error Message |
|-------|------|----------|-------|---------------|
| field1 | string | Yes | Min 3, Max 50 | "Field1 must be 3-50 chars" |

#### Response
**Success (200):**
```json
{
  "data": {...},
  "message": "Success"
}
```

**Error (400):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input"
  }
}
```
```

## Resources

- **Template:** [references/API_SPEC_TEMPLATE.md](references/API_SPEC_TEMPLATE.md)
- **Error Codes:** Standard HTTP status codes
- **Validation:** Use Go struct tags format
