# Module Requirements Validation Checklist

Comprehensive checklist untuk memastikan module requirements quality sebelum proceeding ke API spec dan schema design.

---

## Section 1: Module Definition & Scope

### Module Overview ✅
- [ ] Module name is clear and follows naming convention (lowercase, hyphen-separated)
- [ ] Purpose stated in one sentence
- [ ] Bounded context clearly defined (owns X, does NOT own Y)
- [ ] Multi-tenant strategy documented (if applicable)

**Validation Rule:** All 4 items checked = Clear Module Definition

**Red Flags:**
```
❌ Module name: "UserStuff" → Too vague
✅ Module name: "user-profile" → Clear

❌ Purpose: "Handles users" → Too broad
✅ Purpose: "Manages user profile information and preferences" → Specific

❌ No bounded context defined → Risk of scope creep
✅ "Owns user profiles, NOT authentication (auth module)" → Clear boundary
```

---

### Dependencies ✅
- [ ] All module dependencies listed
- [ ] Each dependency has business justification (WHY needed)
- [ ] No circular dependencies
- [ ] Dependencies on external systems documented
- [ ] Tenant context dependency explicit (for multi-tenant)

**Validation Rule:** All 5 items checked = Clean Dependencies

**Circular Dependency Check:**
```
Run this check:
1. List all dependencies for module A
2. For each dependency B, list its dependencies
3. If B (or B's dependencies) include A → CIRCULAR!

Example:
Module: user-profile
├─ Depends on: auth ✅
├─ Depends on: notification ✅
└─ auth depends on: tenant ✅
    └─ tenant does NOT depend on user-profile ✅
Result: NO CIRCULAR DEPENDENCIES ✅
```

---

## Section 2: Functional Requirements

### Requirement Quality ✅
- [ ] Each FR has unique ID (FR-{MODULE}-001 format)
- [ ] Each FR has BRD reference traceability
- [ ] Each FR has priority (P0/P1/P2)
- [ ] Each FR has user story ("As a X, I want Y, so Z")
- [ ] Acceptance criteria specific and testable (not vague)
- [ ] Business rules documented
- [ ] API endpoint specified (method, path, auth)
- [ ] Total FRs reasonable for module scope (5-15 typical)

**Validation Rule:** At least 7 items checked per FR

**Common Issues:**
```
❌ FR-001: User Management → No module prefix
✅ FR-AUTH-001: User Login → Clear module ownership

❌ "System should be secure" → Vague
✅ "Password must be hashed with bcrypt cost 12" → Measurable

❌ No API endpoint → Implementation unclear
✅ "POST /api/auth/login, JWT required" → Clear

❌ 25 FRs in one module → Too large, split module
✅ 8 FRs → Manageable scope
```

---

### Multi-Tenant Awareness ✅
For multi-tenant systems only:

- [ ] Each FR considers tenant isolation
- [ ] Tenant_id included in data access
- [ ] Cross-tenant access explicitly prevented
- [ ] Tenant limits documented (e.g., user limits per plan)
- [ ] Super admin vs tenant admin roles clear

**Validation Rule:** All 5 items checked for multi-tenant modules

**Example Checks:**
```
FR-PATIENT-001: Create Patient
❌ "User can create patient" → No tenant context
✅ "Tenant admin can create patient within their tenant" → Tenant-aware

FR-AUTH-002: User Login
❌ "Login with email + password" → No tenant isolation
✅ "Login with email + password + tenant_id" → Tenant-aware
```

---

## Section 3: Domain Model

### Entity Definition ✅
- [ ] All entities have attributes defined
- [ ] Primary key specified (typically UUID)
- [ ] Tenant_id included for multi-tenant (if applicable)
- [ ] Timestamps included (created_at, updated_at)
- [ ] Soft delete timestamp if needed (deleted_at)
- [ ] Relationships documented (1:1, 1:N, N:M)
- [ ] Constraints documented (unique, not null, foreign keys)
- [ ] Indexes identified for performance

**Validation Rule:** At least 6 items checked per entity

**Entity Checklist Example:**
```
Entity: User
✅ id: UUID (primary key)
✅ tenant_id: UUID (foreign key, indexed) 
✅ email: String (unique within tenant)
✅ created_at: Timestamp
✅ updated_at: Timestamp
✅ deleted_at: Timestamp (nullable, soft delete)
✅ Relationships: Belongs to Tenant (N:1)
✅ Index: (tenant_id, email) composite unique
```

---

### Data Model Validation ✅
- [ ] All entity names singular (User, not Users)
- [ ] Attributes follow naming convention (snake_case)
- [ ] No redundant data (normalized, or documented denormalization)
- [ ] Value objects identified
- [ ] Enums defined with all values
- [ ] No cross-module entity ownership (DDD bounded context)

**Validation Rule:** All 6 items checked

**Red Flags:**
```
❌ Entity: Users → Plural
✅ Entity: User → Singular

❌ Attribute: userName → camelCase
✅ Attribute: user_name → snake_case

❌ Patient entity in both patient & visit modules → Ownership unclear
✅ Patient owned by patient module, visit references it → Clear boundary
```

---

## Section 4: Use Cases

### Use Case Completeness ✅
- [ ] Each UC has unique ID (UC-{MODULE}-001)
- [ ] Actor identified (user role)
- [ ] Goal stated clearly
- [ ] Preconditions listed
- [ ] Main flow documented (step-by-step)
- [ ] Alternative flows documented
- [ ] Postconditions stated
- [ ] Tenant context in multi-tenant UCs

**Validation Rule:** All 8 items checked per use case

**Use Case Quality Check:**
```
UC-AUTH-001: User Login

✅ Actor: Clinic Staff
✅ Goal: Authenticate and access system
✅ Preconditions: User registered, tenant active
✅ Main Flow: 8 steps documented
✅ Alt Flow: Invalid credentials, locked account
✅ Postconditions: User authenticated, session created
✅ Tenant context: User logs into specific tenant
```

---

## Section 5: Validation Rules

### Validation Completeness ✅
- [ ] All input fields have validation rules
- [ ] Rules are specific (not "valid input")
- [ ] Error messages user-friendly
- [ ] Validation format: field → rules → message
- [ ] Multi-tenant validation included (tenant_id checks)

**Validation Rule:** All 5 items checked

**Good vs Bad Validation:**
```
❌ email | Required | "Error"
   → Vague message

✅ email | Required, Email format, Max 100 | "Valid email required (max 100 chars)"
   → Specific rules + clear message

❌ password | Strong | "Invalid"
   → Undefined "strong"

✅ password | Min 8, 1 uppercase, 1 number, 1 special | "Password must be at least 8 characters with 1 uppercase, 1 number, 1 special"
   → Measurable rules
```

---

## Section 6: Error Handling

### Error Code Quality ✅
- [ ] Each error has unique code ({MODULE}_001 format)
- [ ] HTTP status appropriate (400, 401, 403, 404, 500)
- [ ] Description technical (for logs)
- [ ] User message non-technical (for UI)
- [ ] Multi-tenant errors included (tenant inactive, limits)

**Validation Rule:** All 5 items checked per error

**Error Definition Example:**
```
✅ AUTH_001 | 401 | Invalid credentials | "Email or password incorrect"
   → Code + Status + Tech description + User message

✅ AUTH_008 | 403 | Tenant user limit | "User limit reached. Upgrade plan"
   → Multi-tenant specific error
```

---

## Section 7: Security Requirements

### Security Completeness ✅
- [ ] Authentication method specified
- [ ] Authorization model documented (RBAC, ABAC)
- [ ] Data protection strategy (encryption at rest/transit)
- [ ] Rate limiting defined
- [ ] Tenant isolation security documented (multi-tenant)
- [ ] Password policy specified (if auth module)

**Validation Rule:** At least 5 items checked

**Security Checklist:**
```
Module: auth
✅ Authentication: JWT with RS256
✅ Authorization: RBAC with tenant-scoped roles
✅ Data protection: Bcrypt password hash (cost 12)
✅ Rate limiting: 5 login attempts per 15 min
✅ Tenant isolation: Row-level tenant_id filter
✅ Password policy: Min 8 chars, 1 upper, 1 number, 1 special
```

---

## Section 8: Performance Requirements

### Performance Measurability ✅
- [ ] All performance targets have units (ms, seconds)
- [ ] Percentile specified (p50, p95, p99)
- [ ] Concurrent user capacity defined
- [ ] Database query performance specified
- [ ] Multi-tenant load considered (total across tenants)

**Validation Rule:** All 5 items checked

**Measurable vs Vague:**
```
❌ "Fast response time" → No number
✅ "API response < 500ms (p95)" → Measurable

❌ "Handle many users" → Undefined
✅ "Support 10,000 concurrent users" → Specific

❌ "Queries should be quick" → Vague
✅ "Database queries < 30ms (p99)" → Measurable
```

---

## Section 9: Integration Points

### Integration Documentation ✅
- [ ] All dependencies listed (this module depends on)
- [ ] All consumers listed (other modules depend on this)
- [ ] Data exchanged documented
- [ ] Tenant context propagation clear (multi-tenant)
- [ ] No circular dependencies verified

**Validation Rule:** All 5 items checked

**Integration Table Format:**
```
✅ Dependencies (This module depends on):
| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| auth   | Token validation | Token → User context (id, tenant_id, role) |
| tenant | Validate tenant | Tenant ID → Tenant info (status, plan) |

✅ Provides To (Other modules depend on this):
| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| visit  | Patient lookup | Patient ID → Patient details |
```

---

## Section 10: Multi-Tenant Specific

### Tenant Isolation Validation ✅
For multi-tenant modules only:

- [ ] Tenant_id in all entities
- [ ] All queries include tenant_id filter
- [ ] Cross-tenant access prevention documented
- [ ] Tenant-specific configuration documented
- [ ] Super admin cross-tenant access documented
- [ ] Tenant limits enforced (user count, storage, etc)
- [ ] Tenant context in JWT claims

**Validation Rule:** All 7 items checked

**Multi-Tenant Red Flags:**
```
❌ SELECT * FROM patients WHERE id = ? 
   → No tenant_id filter (data leak risk!)

✅ SELECT * FROM patients WHERE id = ? AND tenant_id = ?
   → Tenant isolation enforced

❌ No tenant limits documented
   → Free tier abuse risk

✅ "Free tier: 5 users, Premium: unlimited"
   → Limits clear
```

---

## Section 11: Testing Requirements

### Test Coverage ✅
- [ ] Unit tests identified
- [ ] Integration tests identified
- [ ] Test scenarios with input/output examples
- [ ] Multi-tenant cross-access prevention tests (if applicable)
- [ ] Load tests for performance targets

**Validation Rule:** At least 4 items checked

**Test Scenario Quality:**
```
❌ "Test login" → Too vague

✅ Test Scenario 1: Successful Login
Input: email="user@test.com", password="Pass123!", tenant_id="uuid"
Expected: HTTP 200, access_token returned, token contains tenant_id
   → Specific and testable

✅ Test Scenario 2: Cross-Tenant Access Prevention (Multi-Tenant)
Input: User from Tenant A, Request data from Tenant B
Expected: HTTP 403 Forbidden, "Access denied"
   → Multi-tenant security test
```

---

## Section 12: Module-Level Acceptance Criteria

### Acceptance Criteria Quality ✅
- [ ] All FRs implemented checkbox
- [ ] All tests passing with coverage target (e.g., > 80%)
- [ ] API documentation complete
- [ ] Database migrations working
- [ ] Performance targets met (with actual numbers)
- [ ] Multi-tenant isolation verified (if applicable)

**Validation Rule:** At least 5 items checked

---

## Overall Module Quality Score

### Scoring System

| Category | Weight | Max Points |
|----------|--------|------------|
| Module Definition | 10% | 10 |
| Functional Requirements | 25% | 25 |
| Domain Model | 15% | 15 |
| Use Cases | 10% | 10 |
| Validation & Errors | 10% | 10 |
| Security & Performance | 15% | 15 |
| Integration Points | 10% | 10 |
| Multi-Tenant (if applicable) | 5% | 5 |

**Total:** 100 points

**Quality Gates:**
- 🟢 **90-100 points:** Excellent - Ready for API spec
- 🟡 **70-89 points:** Good - Minor improvements needed
- 🟠 **50-69 points:** Fair - Significant gaps, revise
- 🔴 **< 50 points:** Poor - Major issues, restart

---

## Quick Validation Commands (For AI Agents)

### Pre-Flight Checklist

Before proceeding to API specification:

```
✅ MUST HAVE (Blockers):
- [ ] BRD reference exists and approved
- [ ] No circular dependencies
- [ ] All FRs have acceptance criteria
- [ ] Domain entities defined
- [ ] Tenant isolation strategy (multi-tenant)

✅ SHOULD HAVE (Important):
- [ ] Use cases documented
- [ ] Validation rules specified
- [ ] Error handling complete
- [ ] Performance targets measurable
- [ ] Integration points documented

✅ NICE TO HAVE (Quality):
- [ ] Test scenarios with examples
- [ ] Security requirements detailed
- [ ] Migration plan documented
```

---

## Common Validation Failures

### Top 10 Issues Found in Reviews

1. **No tenant_id in multi-tenant entities** (🔴 Critical)
2. **Circular dependencies between modules** (🔴 Critical)
3. **Vague acceptance criteria** ("system works well")
4. **No BRD traceability** (can't verify against requirements)
5. **Missing validation rules** (allows bad data)
6. **No error handling documented** (poor UX)
7. **Performance targets unmeasurable** ("fast", "quick")
8. **Cross-module entity ownership unclear** (breaks bounded context)
9. **Missing integration points** (hard to implement later)
10. **No test scenarios** (impossible to verify)

---

## Automated Validation Script (Concept)

```yaml
# validation-rules.yaml
rules:
  - name: "Tenant ID in entities"
    check: "All entities in multi-tenant modules have tenant_id field"
    severity: critical
    
  - name: "FR naming convention"
    pattern: "FR-{MODULE}-\\d{3}"
    severity: high
    
  - name: "Acceptance criteria count"
    min: 3
    severity: medium
    
  - name: "Performance targets have units"
    pattern: "< \\d+(ms|s|GB)"
    severity: high
```

**Usage:** Agent can auto-validate before proceeding to next phase.

---

**End of Validation Checklist**
