# Module Requirements: Authentication (Auth)
## E-Commerce Order Management System

**Version:** 1.0.0  
**Status:** approved  
**BRD Reference:** BRD v1.0.0 (2026-01-28)  
**Last Updated:** 2026-02-05  
**Module Owner:** Bob Johnson (Tech Lead)  

---

## 1. Module Overview

**Purpose:** Provide secure user authentication and authorization services for customers, staff, and administrators.

**Bounded Context:** User identity management, session handling, role-based access control (RBAC).

**Business Value:**
- Secure customer accounts with industry-standard authentication
- Protect sensitive operations with role-based authorization
- Reduce account takeover incidents by 95% (vs. previous system)
- Support 100,000 concurrent authenticated users

**Dependencies:**
- None (foundational module)

**Dependent Modules:**
- Product Module (requires authentication for admin operations)
- Order Module (requires authentication for order operations)

---

## 2. Functional Requirements

### FR-AUTH-001: User Registration
**BRD Reference:** FR-001  
**Priority:** High  

**User Story:** As a new customer, I want to create an account so that I can place orders.

**Acceptance Criteria:**
- POST `/api/auth/register` endpoint
- Required fields: email, name, password
- Password validation: ≥8 chars, 1 uppercase, 1 lowercase, 1 number
- Email validation: RFC 5322 compliant
- Email verification: Send confirmation link to user email
- Account status: "pending" until email verified
- Default role: "customer"
- Return JWT token on successful registration

**Business Rules:**
- Email must be unique (return 409 Conflict if duplicate)
- Password hashed with bcrypt (cost factor 12)
- JWT token expiry: 24 hours
- Email verification link expires in 1 hour
- Unverified accounts deleted after 7 days

**Input Validation:**
```go
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email,max=100"`
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Password string `json:"password" validate:"required,min=8,max=100"`
}
```

**Success Response (201 Created):**
```json
{
  "data": {
    "user": {
      "id": "usr_abc123",
      "email": "john@example.com",
      "name": "John Doe",
      "role": "customer",
      "email_verified": false,
      "created_at": "2026-02-05T10:30:00Z"
    },
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "error": null,
  "meta": {
    "version": "1.0.0"
  }
}
```

---

### FR-AUTH-002: Email Verification
**BRD Reference:** FR-001  
**Priority:** High  

**User Story:** As a registered user, I want to verify my email so that I can access my account.

**Acceptance Criteria:**
- GET `/api/auth/verify-email?token={token}` endpoint
- Token validation (JWT with email claim)
- Update user status to "active"
- Redirect to success page

**Business Rules:**
- Token expires in 1 hour
- Token can only be used once
- Invalid/expired tokens return 400 Bad Request

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Email verified successfully"
  },
  "error": null
}
```

---

### FR-AUTH-003: User Login
**BRD Reference:** FR-001  
**Priority:** High  

**User Story:** As a registered user, I want to login so that I can access protected features.

**Acceptance Criteria:**
- POST `/api/auth/login` endpoint
- Required fields: email, password
- Password verification with bcrypt.CompareHashAndPassword
- Return JWT token on success
- Account lockout after 5 failed attempts (30 minutes)
- Log login attempts (IP, timestamp, success/failure)

**Business Rules:**
- Email must be verified (status = "active")
- JWT token includes: user_id, role, email
- JWT signed with RS256 (private key)
- Token expiry: 24 hours
- Failed login counter reset on successful login

**Input Validation:**
```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "user": {
      "id": "usr_abc123",
      "email": "john@example.com",
      "name": "John Doe",
      "role": "customer"
    },
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-02-06T10:30:00Z"
  },
  "error": null
}
```

**Error Response (401 Unauthorized):**
```json
{
  "data": null,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": {
      "attempts_remaining": 3
    }
  }
}
```

---

### FR-AUTH-004: User Logout
**BRD Reference:** FR-001  
**Priority:** Medium  

**User Story:** As an authenticated user, I want to logout so that my session is terminated.

**Acceptance Criteria:**
- POST `/api/auth/logout` endpoint
- Requires valid JWT token in Authorization header
- Add token to blacklist (Redis cache)
- Token blacklist TTL = token remaining lifetime

**Business Rules:**
- Blacklisted tokens rejected with 401 Unauthorized
- Token blacklist stored in Redis (expires automatically)

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Logged out successfully"
  },
  "error": null
}
```

---

### FR-AUTH-005: Get Current User Profile
**BRD Reference:** FR-001  
**Priority:** Medium  

**User Story:** As an authenticated user, I want to view my profile so that I can verify my account details.

**Acceptance Criteria:**
- GET `/api/auth/me` endpoint
- Requires valid JWT token
- Return user profile from database

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "usr_abc123",
    "email": "john@example.com",
    "name": "John Doe",
    "role": "customer",
    "email_verified": true,
    "created_at": "2026-02-05T10:30:00Z",
    "updated_at": "2026-02-05T10:35:00Z"
  },
  "error": null
}
```

---

### FR-AUTH-006: Update User Profile
**BRD Reference:** FR-001  
**Priority:** Low  

**User Story:** As an authenticated user, I want to update my profile so that my information is current.

**Acceptance Criteria:**
- PATCH `/api/auth/me` endpoint
- Requires valid JWT token
- Updatable fields: name (email/role NOT updatable)

**Input Validation:**
```go
type UpdateProfileRequest struct {
    Name string `json:"name" validate:"omitempty,min=2,max=50"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "usr_abc123",
    "email": "john@example.com",
    "name": "John Smith",
    "role": "customer",
    "updated_at": "2026-02-06T14:20:00Z"
  },
  "error": null
}
```

---

### FR-AUTH-007: Password Reset Request
**BRD Reference:** FR-001  
**Priority:** Medium  

**User Story:** As a user who forgot my password, I want to request a reset link so that I can regain access.

**Acceptance Criteria:**
- POST `/api/auth/forgot-password` endpoint
- Required field: email
- Send reset link to user email (if email exists)
- Always return success (to prevent email enumeration)

**Business Rules:**
- Reset token expires in 1 hour
- Token can only be used once
- Limit: 3 reset requests per hour per email

**Input Validation:**
```go
type ForgotPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "If email exists, reset link has been sent"
  },
  "error": null
}
```

---

### FR-AUTH-008: Password Reset
**BRD Reference:** FR-001  
**Priority:** Medium  

**User Story:** As a user with a reset token, I want to set a new password so that I can login again.

**Acceptance Criteria:**
- POST `/api/auth/reset-password` endpoint
- Required fields: token, new_password
- Token validation (JWT with user_id claim)
- Password validation (same rules as registration)
- Hash new password and update database
- Invalidate all existing user sessions

**Input Validation:**
```go
type ResetPasswordRequest struct {
    Token       string `json:"token" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Password reset successfully"
  },
  "error": null
}
```

---

### FR-AUTH-009: Role-Based Authorization Middleware
**BRD Reference:** FR-008  
**Priority:** High  

**Purpose:** Protect endpoints based on user roles (customer, staff, admin).

**Roles & Permissions:**
- **customer:** Can access own orders, update own profile
- **staff:** Can manage products, view all orders, update order status
- **admin:** Full access (user management, system settings)

**Implementation:**
```go
// Middleware: Require("staff", "admin")
// JWT claims: {"user_id": "usr_123", "role": "staff"}
// Check: JWT role ∈ allowed roles → proceed
//        JWT role ∉ allowed roles → 403 Forbidden
```

**Error Response (403 Forbidden):**
```json
{
  "data": null,
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions",
    "details": {
      "required_roles": ["staff", "admin"],
      "user_role": "customer"
    }
  }
}
```

---

## 3. Data Models

### User Entity
```go
type User struct {
    ID            string    `json:"id"`              // Primary key: usr_{ulid}
    Email         string    `json:"email"`           // Unique, indexed
    Name          string    `json:"name"`
    PasswordHash  string    `json:"-"`               // Bcrypt hash, never exposed
    Role          string    `json:"role"`            // Enum: customer, staff, admin
    EmailVerified bool      `json:"email_verified"`
    Status        string    `json:"status"`          // Enum: pending, active, locked, deleted
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    DeletedAt     *time.Time `json:"-"`              // Soft delete
}
```

### Login Attempt Entity
```go
type LoginAttempt struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`     // Foreign key to users
    Email      string    `json:"email"`       // For failed attempts (user may not exist)
    IPAddress  string    `json:"ip_address"`
    Success    bool      `json:"success"`
    AttemptedAt time.Time `json:"attempted_at"`
}
```

---

## 4. Business Rules

### BR-AUTH-001: Password Policy
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- Optional: special characters (!@#$%^&*)
- Maximum 100 characters
- Cannot contain user's email or name

### BR-AUTH-002: Account Lockout
- 5 failed login attempts within 30 minutes → lock account
- Lockout duration: 30 minutes
- Counter resets on successful login
- Admin can manually unlock accounts

### BR-AUTH-003: Token Management
- JWT signed with RS256 (2048-bit private key)
- Token expiry: 24 hours (configurable)
- Token claims: user_id, role, email, exp, iat
- Blacklisted tokens stored in Redis (TTL = remaining lifetime)
- Tokens NOT refreshable (must re-login after expiry)

### BR-AUTH-004: Email Verification
- Verification link expires in 1 hour
- Unverified accounts cannot login
- Unverified accounts deleted after 7 days (GDPR compliance)
- Resend verification email: max 3 times per hour

### BR-AUTH-005: Rate Limiting
- Login endpoint: 10 requests/minute per IP
- Register endpoint: 3 requests/hour per IP
- Password reset: 3 requests/hour per email

---

## 5. Integration Points

| Integration      | Purpose                  | Protocol  | Auth Method        |
|------------------|--------------------------|-----------|---------------------|
| SendGrid         | Email notifications      | REST API  | API Key             |
| Redis            | Token blacklist, sessions| Redis CLI | Password            |

---

## 6. Error Codes

| Code                    | HTTP Status | Description                          |
|-------------------------|-------------|--------------------------------------|
| INVALID_CREDENTIALS     | 401         | Email or password incorrect          |
| EMAIL_ALREADY_EXISTS    | 409         | Email already registered             |
| EMAIL_NOT_VERIFIED      | 403         | Email verification required          |
| ACCOUNT_LOCKED          | 423         | Too many failed login attempts       |
| TOKEN_EXPIRED           | 401         | JWT token expired                    |
| TOKEN_INVALID           | 401         | JWT signature invalid                |
| TOKEN_BLACKLISTED       | 401         | Token has been revoked               |
| FORBIDDEN               | 403         | Insufficient permissions             |
| WEAK_PASSWORD           | 400         | Password doesn't meet requirements   |
| RATE_LIMIT_EXCEEDED     | 429         | Too many requests                    |

---

## 7. Performance Requirements

- **Registration:** < 300ms p95 (excluding email send)
- **Login:** < 100ms p95 (cached user lookup)
- **Token Validation:** < 10ms p95 (in-memory JWT verify)
- **Concurrent Users:** 100,000 authenticated sessions
- **Token Blacklist:** Redis lookup < 5ms p95

---

## 8. Security Requirements

- **Password Storage:** Bcrypt with cost factor 12
- **JWT Signing:** RS256 with 2048-bit keys
- **TLS:** Enforce HTTPS (TLS 1.3) for all endpoints
- **CSRF Protection:** Not needed (stateless JWT)
- **XSS Protection:** Sanitize user inputs (name field)
- **SQL Injection:** Use parameterized queries only
- **Audit Logging:** Log all login attempts (success/failure)

---

## 9. Dependencies

**External:**
- SendGrid API (email sending)
- Redis (token blacklist, rate limiting)

**Internal:**
- None (foundational module)

---

## 10. Testing Requirements

### Unit Tests
- Password hashing/verification
- JWT token generation/validation
- Email validation logic
- Rate limiting logic

### Integration Tests
- Complete registration flow (register → email → verify → login)
- Login with valid/invalid credentials
- Account lockout after failed attempts
- Token blacklist (logout → attempt to use token)
- Role-based authorization

### Load Tests
- 10,000 concurrent logins
- 100,000 token validations/second

---

## 11. Acceptance Criteria

- [ ] All functional requirements implemented
- [ ] 80%+ code coverage (unit + integration tests)
- [ ] Security audit passed (OWASP Top 10)
- [ ] Performance benchmarks met (p95 < 100ms for login)
- [ ] API documentation complete (OpenAPI spec)
- [ ] User acceptance testing completed (5 users, 100% success rate)

---

## Document History

| Version | Date       | Author      | Changes                         |
|---------|------------|-------------|---------------------------------|
| 0.1     | 2026-02-03 | Bob Johnson | Initial draft                   |
| 0.2     | 2026-02-04 | Alice Chen  | Added business rules            |
| 1.0.0   | 2026-02-05 | Bob Johnson | Approved after BRD alignment    |

---

**Next Steps:**
1. Generate API specification for Auth module (SKILL 2)
2. Generate database schema for Auth module (SKILL 3)
3. Begin implementation (SKILL 4+)
