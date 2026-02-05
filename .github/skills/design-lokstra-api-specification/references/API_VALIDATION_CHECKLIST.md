# API Specification Validation Checklist

**Purpose**: Comprehensive quality gates for API specifications before implementation to ensure consistency, completeness, and best practices.

**Context**: Use this checklist during API specification review to catch issues early and ensure high-quality API design.

---

## How to Use This Checklist

1. **During Design**: Review each section as you write the API specification
2. **Before Review**: Self-check your specification against all items
3. **Peer Review**: Reviewer uses this to validate specification
4. **Before Implementation**: Final gate before handlers are created

**Rating System**:
- ✅ Pass: Requirement met
- ⚠️ Warning: Acceptable but could be improved
- ❌ Fail: Must be fixed before approval

---

## 1. API Metadata

### Basic Information
- [ ] API name clearly defined
- [ ] Module/domain specified
- [ ] Base URL/prefix documented
- [ ] API version specified (e.g., v1)
- [ ] Description of API purpose
- [ ] Related modules/dependencies listed
- [ ] Author/owner identified
- [ ] Last updated date included

### Documentation
- [ ] Purpose and scope clearly explained
- [ ] Target audience identified
- [ ] Prerequisites documented
- [ ] Known limitations listed
- [ ] Migration notes (if applicable)

---

## 2. Multi-Tenant Architecture

### Tenant Identification
- [ ] Tenant identification method specified (header/URL/subdomain)
- [ ] X-Tenant-ID header documented (if used)
- [ ] JWT tenant_id claim documented
- [ ] Tenant validation rules specified
- [ ] Cross-tenant access prevention described

### Tenant Isolation
- [ ] All queries include tenant_id filtering
- [ ] tenant_id in all database table designs
- [ ] Foreign keys include tenant_id
- [ ] Composite unique constraints include tenant_id
- [ ] Cache keys include tenant_id prefix
- [ ] No optional tenant filtering

### Security
- [ ] Tenant mismatch handling specified
- [ ] JWT validation rules documented
- [ ] Permission model includes tenant context
- [ ] Audit logging includes tenant_id
- [ ] Rate limiting per tenant configured

**Critical Check**: Search specification for queries without tenant_id - ❌ FAIL if found

---

## 3. Endpoint Specifications

### For Each Endpoint

#### Basic Specification
- [ ] HTTP method specified (GET/POST/PUT/PATCH/DELETE)
- [ ] URL path documented
- [ ] Summary/description provided
- [ ] Multi-tenant flag (Yes/No)
- [ ] Authentication requirement (Yes/No/Optional)
- [ ] Authorization/permissions specified
- [ ] Lokstra @Handler mapping documented

#### Request Specification
- [ ] Required headers documented
- [ ] Optional headers documented
- [ ] Header validation rules specified
- [ ] Path parameters defined with types
- [ ] Query parameters defined with types
- [ ] Request body schema provided (if applicable)
- [ ] Content-Type specified

#### Request Validation
- [ ] Required fields marked
- [ ] Data types specified for all fields
- [ ] String lengths (min/max) defined
- [ ] Number ranges (min/max) defined
- [ ] Enum values listed
- [ ] Format constraints specified (email, phone, date, etc.)
- [ ] Pattern validation (regex) provided
- [ ] Array size limits defined
- [ ] Custom validation rules documented

#### Response Specification
- [ ] Success status codes defined (200, 201, 204)
- [ ] Success response schema provided
- [ ] Success response example included
- [ ] All error status codes listed (400, 401, 403, 404, 409, 500)
- [ ] Error response schemas provided
- [ ] Error response examples included
- [ ] Error codes documented
- [ ] Response headers documented

#### Error Handling
- [ ] All possible error scenarios covered
- [ ] Error codes follow naming convention
- [ ] Error messages are user-friendly
- [ ] Error details structure defined
- [ ] Field-level errors supported
- [ ] No sensitive data in error messages
- [ ] No cross-tenant information leakage

---

## 4. Data Models & Schemas

### Schema Definition
- [ ] All DTO/model structures defined
- [ ] Field names use consistent casing (camelCase recommended)
- [ ] All fields have type definitions
- [ ] Required vs optional clearly marked
- [ ] Default values specified
- [ ] Nullable fields marked
- [ ] Array/object structures detailed

### Validation Rules
- [ ] String validations (length, pattern, format)
- [ ] Number validations (range, precision)
- [ ] Date/time validations (format, range)
- [ ] Email/URL format validations
- [ ] Enum value validations
- [ ] Cross-field validations (if any)
- [ ] Conditional validations documented

### Data Relationships
- [ ] Foreign key relationships documented
- [ ] Nested objects defined
- [ ] Array items defined
- [ ] Polymorphic types explained
- [ ] tenant_id in all models

---

## 5. Security Specifications

### Authentication
- [ ] Authentication method specified (JWT/API Key/OAuth)
- [ ] Token format documented
- [ ] Token lifetime specified
- [ ] Token refresh mechanism documented
- [ ] Public vs protected endpoints marked
- [ ] Authorization header format specified

### Authorization
- [ ] Permission model documented
- [ ] Role-based access control (RBAC) defined
- [ ] Permission checks per endpoint
- [ ] Admin-only operations marked
- [ ] Tenant-scoped permissions
- [ ] Permission error responses

### JWT Token
- [ ] Algorithm specified (RS256 recommended)
- [ ] Required claims listed
- [ ] Token expiration time defined
- [ ] Refresh token mechanism
- [ ] Token validation rules
- [ ] Claims mapping to user context

### Security Best Practices
- [ ] Password requirements documented (if applicable)
- [ ] Password hashing algorithm (Bcrypt cost 12)
- [ ] Rate limiting specified
- [ ] Account lockout policy (if applicable)
- [ ] CORS configuration documented
- [ ] Input sanitization rules
- [ ] SQL injection prevention mentioned
- [ ] XSS prevention mentioned

---

## 6. Performance & Scalability

### Pagination
- [ ] Pagination strategy specified (cursor/offset)
- [ ] Default page size defined
- [ ] Max page size specified
- [ ] Pagination metadata structure
- [ ] hasNext/hasPrev flags
- [ ] Total count handling (if applicable)

### Filtering & Search
- [ ] Supported filter fields listed
- [ ] Filter operators documented
- [ ] Search query format specified
- [ ] Full-text search strategy (if applicable)
- [ ] Autocomplete endpoint (if applicable)
- [ ] Filter validation rules

### Sorting
- [ ] Sortable fields listed
- [ ] Sort direction format (asc/desc)
- [ ] Multi-field sorting support
- [ ] Default sort order specified
- [ ] Sort validation rules

### Caching
- [ ] Cache strategy specified
- [ ] Cache key format (includes tenant_id)
- [ ] Cache TTL defined
- [ ] Cache invalidation rules
- [ ] ETag support (if applicable)
- [ ] Conditional requests (If-None-Match)

### Rate Limiting
- [ ] Rate limit per endpoint/user/tenant
- [ ] Rate limit window (per minute/hour)
- [ ] Rate limit headers documented
- [ ] Rate limit exceeded response
- [ ] Retry-After header specified

---

## 7. Bulk & Batch Operations

### Bulk Operations
- [ ] Max batch size specified
- [ ] Validation strategy (all or individual)
- [ ] Transaction handling (atomic/partial)
- [ ] Success/failure reporting format
- [ ] Partial failure handling
- [ ] Performance considerations

### Async Processing
- [ ] Job submission endpoint
- [ ] Job ID format
- [ ] Job status endpoint
- [ ] Job result endpoint
- [ ] Progress tracking
- [ ] Estimated duration
- [ ] Notification mechanism

---

## 8. Error Handling & Status Codes

### HTTP Status Codes
- [ ] 200 OK usage documented
- [ ] 201 Created usage documented
- [ ] 204 No Content usage documented
- [ ] 400 Bad Request scenarios
- [ ] 401 Unauthorized scenarios
- [ ] 403 Forbidden scenarios
- [ ] 404 Not Found scenarios
- [ ] 409 Conflict scenarios
- [ ] 422 Unprocessable Entity (if used)
- [ ] 429 Too Many Requests
- [ ] 500 Internal Server Error
- [ ] 503 Service Unavailable

### Error Response Format
- [ ] Standard error structure defined
- [ ] Error code format documented
- [ ] Error message guidelines
- [ ] Field-level error format
- [ ] Multiple errors support
- [ ] Request ID in errors
- [ ] Timestamp in errors

### Error Code Hierarchy
- [ ] Error code naming convention
- [ ] Error code categories defined
- [ ] All error codes documented
- [ ] Error code examples provided
- [ ] No duplicate error codes

---

## 9. Data Formats & Standards

### Date & Time
- [ ] DateTime format specified (ISO 8601)
- [ ] Timezone handling documented
- [ ] Date-only format specified
- [ ] Time-only format specified
- [ ] Duration format specified

### Numbers & Currency
- [ ] Number precision specified
- [ ] Currency code format (ISO 4217)
- [ ] Decimal handling documented
- [ ] Large number format (BigInt)

### Identifiers
- [ ] ID format specified (UUID/ULID/custom)
- [ ] ID generation strategy
- [ ] Resource ID patterns documented
- [ ] Composite key format (if applicable)

### Strings
- [ ] Character encoding (UTF-8)
- [ ] Max string lengths
- [ ] Allowed characters
- [ ] Trim/normalize rules

### Files & Binary Data
- [ ] File upload format (multipart/base64)
- [ ] Max file size
- [ ] Allowed file types
- [ ] File naming conventions
- [ ] Binary data encoding

---

## 10. Integration & Dependencies

### External Services
- [ ] External API dependencies listed
- [ ] Integration patterns documented
- [ ] Fallback behavior specified
- [ ] Timeout settings
- [ ] Retry strategy

### Internal Services
- [ ] Internal service dependencies
- [ ] Service-to-service auth
- [ ] Circuit breaker pattern (if applicable)
- [ ] Service mesh considerations

### Database
- [ ] Database schema referenced
- [ ] Indexes documented
- [ ] Transaction boundaries
- [ ] Concurrency handling
- [ ] Database migrations noted

### Events
- [ ] Events published documented
- [ ] Event payload schemas
- [ ] Event consumers listed
- [ ] Event ordering guarantees
- [ ] Event replay strategy

---

## 11. Testing & Quality

### Test Coverage
- [ ] Functional test cases listed
- [ ] Multi-tenant test scenarios
- [ ] Security test scenarios
- [ ] Performance test criteria
- [ ] Edge cases documented

### Test Data
- [ ] Sample request examples
- [ ] Sample response examples
- [ ] Sample error examples
- [ ] Test tenant IDs
- [ ] Test user credentials

### Performance Criteria
- [ ] Response time targets (p50, p95, p99)
- [ ] Throughput requirements
- [ ] Concurrent user targets
- [ ] Database query limits
- [ ] Memory usage targets

---

## 12. Lokstra Framework Integration

### Handler Annotation
- [ ] @Handler annotation documented
- [ ] Handler name specified
- [ ] URL prefix defined
- [ ] Handler struct defined

### Route Annotation
- [ ] @Route annotation documented
- [ ] HTTP method specified
- [ ] Path parameters format
- [ ] Middlewares listed
- [ ] Handler signature documented

### Service Dependencies
- [ ] @Inject annotations documented
- [ ] Service names specified
- [ ] Service interfaces defined
- [ ] Config injection documented
- [ ] Database injection documented

### Handler Signatures
- [ ] Return type specified
- [ ] Error handling approach
- [ ] Context usage documented
- [ ] DTO parameter types
- [ ] Response helpers usage (ctx.Api.Ok, etc.)

### Configuration
- [ ] Required config keys listed
- [ ] Config injection pattern
- [ ] Environment variables
- [ ] Default values

---

## 13. Versioning & Compatibility

### API Versioning
- [ ] Version scheme documented (v1, v2, etc.)
- [ ] Version location (URL/header)
- [ ] Deprecation policy
- [ ] Migration guide (if applicable)
- [ ] Backward compatibility rules

### Breaking Changes
- [ ] Breaking changes identified
- [ ] Version bump justified
- [ ] Migration path documented
- [ ] Sunset timeline

### Deprecation
- [ ] Deprecated features marked
- [ ] Deprecation warnings documented
- [ ] Sunset header usage
- [ ] Alternative endpoints documented

---

## 14. Documentation Quality

### Completeness
- [ ] All endpoints documented
- [ ] All parameters documented
- [ ] All response codes documented
- [ ] All error codes documented
- [ ] Examples for all endpoints

### Clarity
- [ ] Technical terms explained
- [ ] Ambiguities resolved
- [ ] Consistent terminology
- [ ] No contradictions
- [ ] Clear assumptions

### Examples
- [ ] Request examples for each endpoint
- [ ] Response examples for success cases
- [ ] Response examples for error cases
- [ ] Code snippets for complex operations
- [ ] Common use case scenarios

### Formatting
- [ ] Consistent markdown formatting
- [ ] Proper code blocks with syntax highlighting
- [ ] Tables formatted correctly
- [ ] Links working
- [ ] Proper heading hierarchy

---

## 15. Multi-Tenant Specific Checks

### Critical Security Checks
- [ ] ❌ FAIL: Any query without tenant_id filter
- [ ] ❌ FAIL: Optional tenant filtering
- [ ] ❌ FAIL: tenant_id from request body (should be from JWT)
- [ ] ❌ FAIL: Shared cache keys without tenant prefix
- [ ] ❌ FAIL: Error messages revealing other tenants

### Data Isolation
- [ ] All database queries filter by tenant_id
- [ ] All cache keys include tenant_id
- [ ] All foreign keys include tenant_id
- [ ] All indexes include tenant_id
- [ ] Audit logs include tenant_id

### Testing
- [ ] Cross-tenant isolation test planned
- [ ] Tenant mismatch test planned
- [ ] Performance isolation test planned
- [ ] Security boundary test planned

---

## Approval Checklist

### Before Submitting for Review
- [ ] All sections of this checklist reviewed
- [ ] All ❌ FAIL items resolved
- [ ] All ⚠️ WARNING items addressed or justified
- [ ] Examples tested with real data
- [ ] Specification reviewed by at least one peer

### Review Criteria
- [ ] Specification is complete
- [ ] Specification is consistent
- [ ] Specification follows framework conventions
- [ ] Multi-tenant requirements met
- [ ] Security requirements met
- [ ] Performance requirements met
- [ ] Documentation quality acceptable

### Final Approval
- [ ] Product owner approval
- [ ] Technical lead approval
- [ ] Security review completed
- [ ] Ready for implementation

---

## Common Issues & Solutions

### Issue 1: Incomplete Error Handling
**Problem**: Only 200 and 500 status codes documented  
**Solution**: Document all possible error scenarios with appropriate status codes and error responses

### Issue 2: Missing Multi-Tenant Checks
**Problem**: tenant_id filtering not enforced in all queries  
**Solution**: Add tenant_id to all WHERE clauses and foreign key constraints

### Issue 3: Vague Validation Rules
**Problem**: "Name must be valid" without specifics  
**Solution**: Specify min/max length, allowed characters, patterns

### Issue 4: No Performance Specs
**Problem**: No pagination, filtering, or rate limiting  
**Solution**: Add pagination strategy, filter options, and rate limits

### Issue 5: Security Gaps
**Problem**: No authentication/authorization specified  
**Solution**: Document JWT requirements, permissions, and validation rules

### Issue 6: Missing Examples
**Problem**: Only schema definitions without examples  
**Solution**: Add realistic request/response examples for all endpoints

### Issue 7: Unclear Lokstra Integration
**Problem**: No handler annotations or service dependencies  
**Solution**: Document @Handler, @Route, @Inject patterns for each endpoint

### Issue 8: Inconsistent Naming
**Problem**: Mixed camelCase and snake_case  
**Solution**: Use camelCase for JSON, snake_case for database, consistently

---

## Severity Levels

### ❌ Critical (Must Fix)
- Missing tenant_id filtering
- Security vulnerabilities
- Missing required fields
- Broken examples
- Contradictory information

### ⚠️ Major (Should Fix)
- Missing validation rules
- Incomplete error handling
- No performance specs
- Missing examples
- Poor documentation

### ℹ️ Minor (Nice to Have)
- Additional examples
- Better formatting
- More detailed descriptions
- Extra test scenarios

---

## Review Process

### Step 1: Self-Review (30 min)
- Go through entire checklist
- Mark all items ✅/⚠️/❌
- Fix all ❌ items
- Document ⚠️ justifications

### Step 2: Peer Review (45 min)
- Another developer reviews spec
- Uses this checklist
- Provides feedback
- Identifies gaps

### Step 3: Technical Review (30 min)
- Tech lead reviews
- Validates architecture
- Checks performance
- Approves design

### Step 4: Security Review (30 min)
- Security specialist reviews
- Validates multi-tenant isolation
- Checks authentication/authorization
- Approves security approach

### Step 5: Final Approval (15 min)
- Product owner approval
- Ready for implementation
- Specification locked

**Total Time**: ~2.5 hours for comprehensive review

---

## Quality Score

Calculate quality score based on checklist completion:

```
Score = (Passed Items / Total Items) × 100

Grading:
- 95-100%: Excellent - Ready for implementation
- 85-94%:  Good - Minor improvements needed
- 75-84%:  Fair - Moderate improvements needed
- Below 75%: Poor - Major revisions required
```

**Minimum Passing Score**: 85% with no ❌ FAIL items

---

**File Size**: 13 KB  
**Last Updated**: 2024-01-20  
**Related**: API_SPEC_TEMPLATE.md, MULTI_TENANT_API_PATTERNS.md, AUTH_API_EXAMPLE.md

