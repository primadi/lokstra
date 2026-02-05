# Module Requirements: Authentication (Auth)
## Clinic Management System - Multi-Tenant

**Version:** 1.0.0  
**Status:** Approved  
**BRD Reference:** Clinic BRD v1.0  
**Last Updated:** 2026-02-01  
**Module Owner:** Security Team

---

## 1. Module Overview

**Module Name:** auth  
**Purpose:** Handles user authentication, authorization, session management, and role-based access control (RBAC) for multi-tenant clinic system.

**Bounded Context:**
- Owns: User credentials, sessions, tokens, roles, permissions, tenant associations
- Does NOT own: User profiles (belongs to `user-profile` module), audit logs (belongs to `audit` module)

**Multi-Tenant Strategy:**
- **Tenant Isolation:** Each tenant has separate user database with `tenant_id` partition
- **Shared Services:** Single auth service handles all tenants with tenant context
- **Cross-Tenant:** Super admin can switch tenants; regular users cannot

**Dependencies:**
- **tenant** module - Validates tenant existence and status before authentication
- **user-profile** module - Fetches additional user data after successful authentication

**Dependent Modules:**
- All modules depend on auth for token validation
- **api-gateway** - Validates JWT tokens on every request
- **audit** - Logs authentication events

---

## 2. Functional Requirements

### FR-AUTH-001: User Registration (Multi-Tenant)
**BRD Reference:** FR-001  
**Priority:** P0 (Must Have)  
**User Story:** As a clinic administrator, I want to register new users within my tenant, so that staff can access the system.

**Acceptance Criteria:**
- User must belong to exactly one tenant (except super admins)
- Email must be unique within tenant scope (same email allowed across tenants)
- Username must be globally unique (across all tenants)
- Password must meet complexity requirements (min 8 chars, 1 uppercase, 1 number, 1 special)
- User gets default role "staff" unless specified otherwise
- Tenant admin can only create users for their own tenant

**Business Rules:**
- Tenant must be active to register users
- Free tier tenants limited to 5 users max
- Premium tier unlimited users
- Super admin role can only be assigned by system admin (not tenant admin)

**API Endpoint:**
```
POST /api/auth/register
Headers: X-Tenant-ID (required), Authorization (admin token)
Body: {
  "email": "doctor@clinic.com",
  "username": "dr_john",
  "password": "SecureP@ss123",
  "role": "doctor",
  "tenant_id": "uuid" // must match X-Tenant-ID header
}
Response: 201 Created
{
  "user_id": "uuid",
  "tenant_id": "uuid",
  "email": "doctor@clinic.com",
  "role": "doctor",
  "created_at": "2026-02-01T10:00:00Z"
}
```

---

### FR-AUTH-002: User Login (Tenant-Aware)
**BRD Reference:** FR-002  
**Priority:** P0 (Must Have)  
**User Story:** As a clinic staff member, I want to login with my credentials, so that I can access my tenant's data.

**Acceptance Criteria:**
- Login with email/username + password + tenant identifier
- System validates credentials within tenant scope
- Returns JWT token with tenant_id claim
- Token expires after 8 hours (configurable per tenant)
- Failed login attempts tracked (max 5 attempts per 15 min)
- Account locked after 5 failed attempts

**Business Rules:**
- Inactive tenant users cannot login
- Suspended users cannot login
- User can only login to their assigned tenant
- Super admin can specify tenant during login (tenant switcher)

**API Endpoint:**
```
POST /api/auth/login
Body: {
  "email": "doctor@clinic.com",
  "password": "SecureP@ss123",
  "tenant_id": "uuid" // required for regular users
}
Response: 200 OK
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 28800,
  "user": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "doctor@clinic.com",
    "role": "doctor",
    "permissions": ["read:patients", "write:prescriptions"]
  }
}
```

---

### FR-AUTH-003: Token Validation (Multi-Tenant)
**BRD Reference:** FR-003  
**Priority:** P0 (Must Have)  
**User Story:** As a backend service, I want to validate JWT tokens, so that I can ensure requests are authenticated and authorized.

**Acceptance Criteria:**
- Validate JWT signature with public key
- Check token expiration
- Extract tenant_id from token claims
- Verify user still active in tenant
- Return user permissions for authorization

**Business Rules:**
- Token must contain valid tenant_id claim
- User must still be active in that tenant
- Tenant must be active
- Expired tokens rejected (HTTP 401)
- Invalid tenant rejected (HTTP 403)

**API Endpoint:**
```
POST /api/auth/validate
Headers: Authorization: Bearer {token}
Response: 200 OK
{
  "valid": true,
  "user_id": "uuid",
  "tenant_id": "uuid",
  "role": "doctor",
  "permissions": ["read:patients", "write:prescriptions"],
  "expires_at": "2026-02-01T18:00:00Z"
}
```

---

### FR-AUTH-004: Role-Based Access Control (RBAC)
**BRD Reference:** FR-004  
**Priority:** P0 (Must Have)  
**User Story:** As a system admin, I want to define roles and permissions, so that users have appropriate access levels.

**Acceptance Criteria:**
- Predefined roles: super_admin, tenant_admin, doctor, nurse, front_office, pharmacist, viewer
- Each role has specific permissions
- Permissions follow resource:action pattern (e.g., "patients:read", "prescriptions:write")
- Tenant admin can assign roles within their tenant
- Super admin can assign any role globally

**Business Rules:**
- Roles are tenant-scoped (except super_admin)
- User can have only one role per tenant
- Super admin has all permissions across all tenants
- Tenant admin has all permissions within their tenant only

**Predefined Roles & Permissions:**

| Role | Permissions | Scope |
|------|-------------|-------|
| super_admin | All (`*:*`) | Global (all tenants) |
| tenant_admin | All within tenant | Single tenant |
| doctor | patients:*, prescriptions:*, visits:*, lab_results:read | Single tenant |
| nurse | patients:read, visits:*, vital_signs:* | Single tenant |
| front_office | patients:*, appointments:*, queue:* | Single tenant |
| pharmacist | prescriptions:read, inventory:*, dispensing:* | Single tenant |
| viewer | *:read (read-only all resources) | Single tenant |

---

### FR-AUTH-005: Tenant Switching (Super Admin)
**BRD Reference:** FR-005  
**Priority:** P1 (Should Have)  
**User Story:** As a super admin, I want to switch between tenants, so that I can manage multiple clinics.

**Acceptance Criteria:**
- Only super_admin role can switch tenants
- Token updated with new tenant_id
- Previous tenant context cleared
- Audit log records tenant switch

**API Endpoint:**
```
POST /api/auth/switch-tenant
Headers: Authorization: Bearer {super_admin_token}
Body: {
  "tenant_id": "uuid"
}
Response: 200 OK
{
  "access_token": "eyJhbGc...", // new token with updated tenant_id
  "tenant": {
    "id": "uuid",
    "name": "Clinic XYZ",
    "status": "active"
  }
}
```

---

### FR-AUTH-006: Password Reset (Tenant-Scoped)
**BRD Reference:** FR-006  
**Priority:** P1 (Should Have)  
**User Story:** As a user, I want to reset my forgotten password, so that I can regain access to my account.

**Acceptance Criteria:**
- User requests reset via email + tenant_id
- System sends reset token valid for 1 hour
- Reset link includes tenant context
- User sets new password meeting complexity rules
- Old password immediately invalidated

**Business Rules:**
- Reset token single-use only
- Email must exist within specified tenant
- Reset emails throttled (max 3 per hour per user)

---

## 3. Domain Model

### Entities

#### User
**Attributes:**
- `id`: UUID - Primary key
- `tenant_id`: UUID - Tenant association (indexed)
- `email`: String - Email (unique within tenant)
- `username`: String - Username (globally unique, indexed)
- `password_hash`: String - Bcrypt hashed password
- `role`: Enum - User role (see RBAC)
- `status`: Enum - [active, inactive, suspended]
- `failed_login_attempts`: Integer - Failed login counter
- `last_login_at`: Timestamp - Last successful login
- `password_changed_at`: Timestamp - Last password change
- `created_at`: Timestamp
- `updated_at`: Timestamp
- `deleted_at`: Timestamp (soft delete)

**Relationships:**
- Belongs to: Tenant (N:1)
- Has many: Sessions (1:N)
- Has many: AuditLogs (1:N)

**Indexes:**
- Composite: (tenant_id, email) - Unique
- Single: username - Unique
- Single: tenant_id - For tenant filtering

**Business Rules:**
- Email unique within tenant (same email OK across tenants)
- Username globally unique
- Password must be hashed with bcrypt (cost 12)

---

#### Session
**Attributes:**
- `id`: UUID - Primary key
- `user_id`: UUID - Foreign key to User
- `tenant_id`: UUID - Tenant context
- `access_token`: String - JWT token (indexed)
- `refresh_token`: String - Refresh token (indexed)
- `ip_address`: String - Client IP
- `user_agent`: String - Browser/client info
- `expires_at`: Timestamp - Token expiration
- `created_at`: Timestamp
- `revoked_at`: Timestamp (nullable) - Manual revocation

**Relationships:**
- Belongs to: User (N:1)
- Belongs to: Tenant (N:1)

**Indexes:**
- Single: access_token - For fast lookup
- Single: user_id - For user session list
- Composite: (tenant_id, expires_at) - Cleanup queries

---

#### Role
**Attributes:**
- `id`: UUID - Primary key
- `name`: String - Role name (e.g., "doctor")
- `description`: String - Role description
- `permissions`: JSON - Array of permission strings
- `is_system_role`: Boolean - System-defined vs custom
- `tenant_id`: UUID (nullable) - Null for global roles

**Relationships:**
- Belongs to: Tenant (optional, N:1)

**Predefined System Roles:**
```json
[
  {
    "name": "super_admin",
    "permissions": ["*:*"],
    "is_system_role": true,
    "tenant_id": null
  },
  {
    "name": "tenant_admin",
    "permissions": ["*:*"],
    "is_system_role": true,
    "tenant_id": "uuid"
  },
  {
    "name": "doctor",
    "permissions": [
      "patients:*",
      "prescriptions:*",
      "visits:*",
      "lab_results:read"
    ],
    "is_system_role": true
  }
]
```

---

## 4. Use Cases

### UC-AUTH-001: User Login Flow (Multi-Tenant)
**Actor:** Clinic Staff  
**Goal:** Authenticate and access tenant-specific data

**Preconditions:**
- User registered in system
- Tenant is active
- User is active in tenant

**Main Flow:**
1. User opens login page
2. User enters email, password, and selects tenant (from dropdown)
3. System validates tenant is active
4. System finds user by email within tenant scope
5. System verifies password hash
6. System checks user status (must be active)
7. System generates JWT with claims: user_id, tenant_id, role, permissions
8. System creates session record
9. System returns access_token and refresh_token
10. User redirected to dashboard with tenant context

**Alternative Flows:**
- **Alt-1 (Invalid Credentials):**
  - System increments failed_login_attempts
  - If attempts >= 5: Lock account for 15 minutes
  - Return HTTP 401 with error message
- **Alt-2 (Inactive Tenant):**
  - Return HTTP 403 "Tenant is not active"
- **Alt-3 (Suspended User):**
  - Return HTTP 403 "Account suspended. Contact administrator"

**Postconditions:**
- User authenticated with tenant context
- Session created in database
- Audit log recorded

---

### UC-AUTH-002: Token Validation in API Gateway
**Actor:** API Gateway  
**Goal:** Validate incoming request token

**Main Flow:**
1. Client sends request with Authorization header
2. Gateway extracts JWT from header
3. Gateway calls auth service to validate token
4. Auth service verifies JWT signature
5. Auth service checks expiration
6. Auth service extracts tenant_id from claims
7. Auth service verifies user still active in tenant
8. Auth service returns validation result + user context
9. Gateway attaches user context to request
10. Gateway forwards request to downstream service

**Alternative Flows:**
- **Alt-1 (Expired Token):**
  - Return HTTP 401 "Token expired"
  - Client should refresh token
- **Alt-2 (Invalid Tenant):**
  - Return HTTP 403 "Invalid tenant context"
- **Alt-3 (User Deactivated):**
  - Return HTTP 403 "User account no longer active"

---

## 5. Data Validation Rules

| Field | Rules | Error Message |
|-------|-------|---------------|
| email | Required, Email format, Max 100 | "Valid email required" |
| username | Required, Alphanumeric + underscore, Min 3, Max 50 | "Username must be 3-50 characters (alphanumeric)" |
| password | Required, Min 8, 1 uppercase, 1 number, 1 special char | "Password must be at least 8 characters with 1 uppercase, 1 number, 1 special character" |
| tenant_id | Required (except super_admin), Valid UUID, Tenant must exist | "Valid tenant required" |
| role | Required, Must be valid role name | "Invalid role" |

---

## 6. Error Handling

| Code | HTTP Status | Description | User Message |
|------|-------------|-------------|--------------|
| AUTH_001 | 401 | Invalid credentials | "Email or password incorrect" |
| AUTH_002 | 401 | Token expired | "Session expired. Please login again" |
| AUTH_003 | 403 | Account suspended | "Account suspended. Contact administrator" |
| AUTH_004 | 403 | Tenant inactive | "Tenant account is not active" |
| AUTH_005 | 429 | Too many login attempts | "Too many failed attempts. Try again in 15 minutes" |
| AUTH_006 | 400 | Password complexity failed | "Password does not meet requirements" |
| AUTH_007 | 403 | Insufficient permissions | "You don't have permission to perform this action" |
| AUTH_008 | 403 | Tenant limit reached | "User limit reached for this plan. Upgrade to add more users" |

---

## 7. Security Requirements

**Authentication:** 
- Password hashing: Bcrypt with cost factor 12
- JWT signing: RS256 (asymmetric key)
- Token expiration: 8 hours (configurable per tenant)
- Refresh token: 30 days

**Authorization:**
- Role-based access control (RBAC)
- Permission format: `resource:action` (e.g., "patients:read")
- Tenant isolation enforced at data layer

**Data Protection:**
- Passwords never stored in plain text
- JWT tokens contain minimal claims (no sensitive data)
- Refresh tokens rotated on use
- Sessions revocable

**Rate Limiting:**
- Login endpoint: 5 attempts per 15 minutes per IP
- Password reset: 3 requests per hour per user
- Token validation: 1000 requests per minute per token

---

## 8. Performance Requirements

- Login operation: < 500ms (p95)
- Token validation: < 50ms (p95)
- Token generation: < 100ms (p95)
- Database queries: < 30ms (p99)
- Support 10,000 concurrent users across all tenants
- Session cleanup: Automated daily job for expired sessions

---

## 9. Integration Points

### Dependencies (This module depends on:)

| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| tenant | Validate tenant status | Tenant ID → Tenant info (status, plan) |
| user-profile | Fetch user details after login | User ID → Profile data (name, avatar) |

### Provides To (Other modules depend on this:)

| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| api-gateway | Token validation | Token → User context (user_id, tenant_id, permissions) |
| All modules | Authorization checks | User ID → Role & permissions |
| audit | Authentication events | Login/logout events → Audit logs |

---

## 10. Multi-Tenant Considerations

**Data Isolation:**
- All queries MUST include `tenant_id` filter (except super_admin)
- Database indexes optimized for tenant-based queries
- Row-level security policies in PostgreSQL (optional)

**Tenant Context Propagation:**
- JWT token includes `tenant_id` claim
- Every API request validated against token's tenant_id
- Prevent cross-tenant data access

**Tenant-Specific Configuration:**
- Token expiration per tenant
- Password policy per tenant
- MFA requirements per tenant
- Session limits per tenant

**Super Admin Access:**
- Can switch tenant context
- Cross-tenant operations logged
- Separate audit trail for super admin actions

---

## 11. Testing Requirements

### Unit Tests
- [ ] Password hashing and verification
- [ ] JWT token generation and validation
- [ ] Permission checking logic
- [ ] Tenant isolation validation
- [ ] Failed login attempt tracking

### Integration Tests
- [ ] Full login flow with database
- [ ] Token validation with tenant checks
- [ ] Cross-tenant access prevention
- [ ] Super admin tenant switching
- [ ] Session management (create, validate, revoke)

### Load Tests
- [ ] 10,000 concurrent logins
- [ ] 100,000 token validations per second
- [ ] Tenant isolation under load

### Test Scenarios

**Scenario 1: Successful Multi-Tenant Login**
```
Input: 
  - email: "doctor@clinic.com"
  - password: "SecureP@ss123"
  - tenant_id: "tenant-uuid-123"
Expected Output:
  - HTTP 200
  - access_token with tenant_id claim
  - user.tenant_id matches request tenant_id
```

**Scenario 2: Cross-Tenant Access Prevention**
```
Input:
  - User from Tenant A tries to access Tenant B data
  - Token contains tenant_id: "tenant-a"
  - Request URL: /api/patients?tenant_id=tenant-b
Expected Output:
  - HTTP 403 Forbidden
  - Error: "Access denied to this tenant"
```

---

## 12. Migration Requirements

### Data Migration
- **Source:** Legacy authentication system (if applicable)
- **Volume:** Estimated 500 users across 10 tenants
- **Mapping:**
  - Legacy user ID → New UUID
  - Legacy clinic ID → tenant_id
  - Legacy role → New RBAC role

### Migration Steps
1. Export users from legacy system
2. Map legacy roles to new RBAC roles
3. Assign users to correct tenant_id
4. Generate secure password hashes (force reset on first login)
5. Validate data integrity (unique constraints)
6. Test login with migrated users

---

## 13. Acceptance Criteria (Module-Level)

- [ ] Users can register within tenant scope
- [ ] Users can login with tenant awareness
- [ ] JWT tokens include tenant_id claim
- [ ] Tokens validated with tenant checks
- [ ] Cross-tenant access prevented (tested)
- [ ] RBAC enforced with tenant context
- [ ] Super admin can switch tenants
- [ ] Failed login attempts tracked and locked
- [ ] Password complexity enforced
- [ ] Sessions revocable
- [ ] All tests passing (> 85% coverage)
- [ ] API documentation complete
- [ ] Database migrations working
- [ ] Performance targets met (< 500ms login, < 50ms validation)

---

## 14. Glossary

| Term | Definition |
|------|------------|
| Tenant | An organization (clinic) using the system with isolated data |
| RBAC | Role-Based Access Control - Permission model based on user roles |
| JWT | JSON Web Token - Standard for secure authentication tokens |
| Bounded Context | DDD concept - Module owns specific domain concepts |
| Multi-Tenant | Single application instance serving multiple isolated customers |
| Tenant Isolation | Ensuring data from one tenant cannot be accessed by another |

---

## 15. References

- Clinic Management BRD v1.0
- Lokstra Framework - Multi-Tenant Guide
- OAuth 2.0 & JWT Best Practices (RFC 7519)
- OWASP Authentication Cheat Sheet

---

**End of Requirements Document**
