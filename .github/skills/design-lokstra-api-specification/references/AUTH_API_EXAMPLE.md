# Auth API Example - Multi-Tenant Authentication

**Purpose**: Real-world authentication API specification demonstrating multi-tenant patterns, JWT token handling, and Lokstra framework integration.

**Context**: This API manages user authentication for clinic applications. Each clinic is a separate tenant with isolated user data.

---

## API Overview

**Base URL**: `/api/v1/auth`

**Module**: auth  
**Domain**: User authentication and authorization  
**Multi-Tenant Strategy**: Header-based tenant identification with JWT claims

---

## Common Headers

### Request Headers
```yaml
X-Tenant-ID:
  description: Tenant identifier for the clinic
  type: string
  required: true (except for public endpoints)
  example: "clinic_001"
  validation: Must match pattern ^[a-z0-9_]+$

Authorization:
  description: JWT bearer token for authenticated requests
  type: string
  required: false (only for protected endpoints)
  example: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
  validation: Must start with "Bearer "
```

### Response Headers
```yaml
X-Request-ID:
  description: Unique request identifier for tracing
  type: string
  example: "req_abc123def456"

X-RateLimit-Limit:
  description: Maximum requests per window
  type: integer
  example: 100

X-RateLimit-Remaining:
  description: Remaining requests in current window
  type: integer
  example: 95
```

---

## Endpoints

### 1. POST /register

**Summary**: Register new user in tenant  
**Multi-Tenant**: Yes (tenant_id required)  
**Authentication**: No  
**Handler Mapping**: `@Route "POST /register"`

#### Request

**Headers**:
```yaml
X-Tenant-ID: clinic_001
Content-Type: application/json
```

**Body**:
```json
{
  "email": "doctor@clinic.com",
  "password": "SecurePass123!",
  "fullName": "Dr. John Doe",
  "role": "doctor",
  "metadata": {
    "licenseNumber": "MD12345",
    "specialization": "General Practitioner"
  }
}
```

**Schema**:
```yaml
RegisterRequest:
  type: object
  required: [email, password, fullName, role]
  properties:
    email:
      type: string
      format: email
      minLength: 5
      maxLength: 255
      example: "doctor@clinic.com"
    password:
      type: string
      minLength: 8
      maxLength: 72
      pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]+"
      description: Must contain uppercase, lowercase, digit, special char
      example: "SecurePass123!"
    fullName:
      type: string
      minLength: 2
      maxLength: 255
      example: "Dr. John Doe"
    role:
      type: string
      enum: [admin, doctor, nurse, receptionist]
      example: "doctor"
    metadata:
      type: object
      description: Role-specific additional information
      additionalProperties: true
```

**Validation Rules**:
- Email must be unique within tenant
- Password must be at least 8 characters
- Password must contain: 1 uppercase, 1 lowercase, 1 digit, 1 special character
- Role must be valid enum value
- Metadata is optional but validated if provided

#### Response

**Success (201 Created)**:
```json
{
  "status": "success",
  "message": "User registered successfully",
  "data": {
    "userId": "usr_abc123",
    "email": "doctor@clinic.com",
    "fullName": "Dr. John Doe",
    "role": "doctor",
    "tenantId": "clinic_001",
    "createdAt": "2024-01-15T10:30:00Z",
    "emailVerified": false
  }
}
```

**Error (400 Bad Request)**:
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": [
      {
        "field": "password",
        "message": "Password must contain at least one uppercase letter",
        "code": "INVALID_PASSWORD_FORMAT"
      }
    ]
  },
  "requestId": "req_abc123"
}
```

**Error (409 Conflict)**:
```json
{
  "status": "error",
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email already registered in this tenant",
    "field": "email"
  },
  "requestId": "req_abc123"
}
```

**Error Codes**:
- `VALIDATION_ERROR`: Invalid input data
- `EMAIL_EXISTS`: Email already registered
- `TENANT_NOT_FOUND`: Invalid tenant ID
- `WEAK_PASSWORD`: Password doesn't meet requirements

---

### 2. POST /login

**Summary**: Authenticate user and get JWT token  
**Multi-Tenant**: Yes (tenant_id required)  
**Authentication**: No  
**Handler Mapping**: `@Route "POST /login"`

#### Request

**Headers**:
```yaml
X-Tenant-ID: clinic_001
Content-Type: application/json
```

**Body**:
```json
{
  "email": "doctor@clinic.com",
  "password": "SecurePass123!",
  "deviceInfo": {
    "userAgent": "Mozilla/5.0...",
    "ipAddress": "192.168.1.100"
  }
}
```

**Schema**:
```yaml
LoginRequest:
  type: object
  required: [email, password]
  properties:
    email:
      type: string
      format: email
      example: "doctor@clinic.com"
    password:
      type: string
      example: "SecurePass123!"
    deviceInfo:
      type: object
      description: Optional device information for audit log
      properties:
        userAgent:
          type: string
        ipAddress:
          type: string
          format: ipv4
```

#### Response

**Success (200 OK)**:
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 3600,
    "tokenType": "Bearer",
    "user": {
      "userId": "usr_abc123",
      "email": "doctor@clinic.com",
      "fullName": "Dr. John Doe",
      "role": "doctor",
      "tenantId": "clinic_001",
      "permissions": ["patient:read", "patient:write", "appointment:manage"]
    }
  }
}
```

**JWT Token Claims**:
```json
{
  "sub": "usr_abc123",
  "email": "doctor@clinic.com",
  "tenant_id": "clinic_001",
  "role": "doctor",
  "permissions": ["patient:read", "patient:write", "appointment:manage"],
  "iat": 1705315800,
  "exp": 1705319400,
  "iss": "lokstra-auth"
}
```

**Error (401 Unauthorized)**:
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  },
  "requestId": "req_abc123"
}
```

**Error (403 Forbidden)**:
```json
{
  "status": "error",
  "error": {
    "code": "ACCOUNT_LOCKED",
    "message": "Account locked due to multiple failed login attempts",
    "retryAfter": 900
  },
  "requestId": "req_abc123"
}
```

**Error Codes**:
- `INVALID_CREDENTIALS`: Wrong email or password
- `ACCOUNT_LOCKED`: Too many failed attempts
- `ACCOUNT_DISABLED`: User account disabled
- `EMAIL_NOT_VERIFIED`: Email verification required

---

### 3. POST /refresh

**Summary**: Refresh access token using refresh token  
**Multi-Tenant**: Yes (extracted from refresh token)  
**Authentication**: Refresh token required  
**Handler Mapping**: `@Route "POST /refresh"`

#### Request

**Headers**:
```yaml
Content-Type: application/json
```

**Body**:
```json
{
  "refreshToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Schema**:
```yaml
RefreshRequest:
  type: object
  required: [refreshToken]
  properties:
    refreshToken:
      type: string
      description: Valid refresh token from login
      example: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Response

**Success (200 OK)**:
```json
{
  "status": "success",
  "message": "Token refreshed successfully",
  "data": {
    "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 3600,
    "tokenType": "Bearer"
  }
}
```

**Error (401 Unauthorized)**:
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_REFRESH_TOKEN",
    "message": "Refresh token is invalid or expired"
  },
  "requestId": "req_abc123"
}
```

---

### 4. POST /logout

**Summary**: Invalidate user session  
**Multi-Tenant**: Yes (from JWT)  
**Authentication**: Yes (JWT required)  
**Handler Mapping**: `@Route "POST /logout"`

#### Request

**Headers**:
```yaml
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Body**:
```json
{
  "allDevices": false
}
```

**Schema**:
```yaml
LogoutRequest:
  type: object
  properties:
    allDevices:
      type: boolean
      default: false
      description: If true, logout from all devices
```

#### Response

**Success (200 OK)**:
```json
{
  "status": "success",
  "message": "Logged out successfully"
}
```

---

### 5. GET /me

**Summary**: Get current user profile  
**Multi-Tenant**: Yes (from JWT)  
**Authentication**: Yes (JWT required)  
**Handler Mapping**: `@Route "GET /me"`

#### Request

**Headers**:
```yaml
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Query Parameters**: None

#### Response

**Success (200 OK)**:
```json
{
  "status": "success",
  "data": {
    "userId": "usr_abc123",
    "email": "doctor@clinic.com",
    "fullName": "Dr. John Doe",
    "role": "doctor",
    "tenantId": "clinic_001",
    "emailVerified": true,
    "createdAt": "2024-01-15T10:30:00Z",
    "lastLoginAt": "2024-01-20T08:15:00Z",
    "metadata": {
      "licenseNumber": "MD12345",
      "specialization": "General Practitioner"
    },
    "permissions": ["patient:read", "patient:write", "appointment:manage"]
  }
}
```

**Error (401 Unauthorized)**:
```json
{
  "status": "error",
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Access token has expired"
  },
  "requestId": "req_abc123"
}
```

---

### 6. POST /change-password

**Summary**: Change user password  
**Multi-Tenant**: Yes (from JWT)  
**Authentication**: Yes (JWT required)  
**Handler Mapping**: `@Route "POST /change-password"`

#### Request

**Headers**:
```yaml
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Body**:
```json
{
  "currentPassword": "OldPass123!",
  "newPassword": "NewSecurePass456!",
  "logoutAllDevices": true
}
```

**Schema**:
```yaml
ChangePasswordRequest:
  type: object
  required: [currentPassword, newPassword]
  properties:
    currentPassword:
      type: string
      example: "OldPass123!"
    newPassword:
      type: string
      minLength: 8
      maxLength: 72
      pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]+"
      example: "NewSecurePass456!"
    logoutAllDevices:
      type: boolean
      default: false
      description: Force logout from all other devices
```

#### Response

**Success (200 OK)**:
```json
{
  "status": "success",
  "message": "Password changed successfully"
}
```

**Error (400 Bad Request)**:
```json
{
  "status": "error",
  "error": {
    "code": "SAME_PASSWORD",
    "message": "New password must be different from current password"
  },
  "requestId": "req_abc123"
}
```

**Error (401 Unauthorized)**:
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_CURRENT_PASSWORD",
    "message": "Current password is incorrect"
  },
  "requestId": "req_abc123"
}
```

---

## Security Specifications

### Password Requirements
- **Minimum length**: 8 characters
- **Maximum length**: 72 characters (Bcrypt limit)
- **Complexity**: Must contain:
  - At least 1 uppercase letter (A-Z)
  - At least 1 lowercase letter (a-z)
  - At least 1 digit (0-9)
  - At least 1 special character (@$!%*?&)
- **Hashing**: Bcrypt with cost factor 12
- **Storage**: Never store plain text passwords

### JWT Token Specifications
- **Algorithm**: RS256 (RSA with SHA-256)
- **Access Token TTL**: 1 hour (3600 seconds)
- **Refresh Token TTL**: 30 days (2592000 seconds)
- **Required Claims**:
  - `sub`: User ID
  - `email`: User email
  - `tenant_id`: Tenant identifier
  - `role`: User role
  - `permissions`: Array of permission strings
  - `iat`: Issued at timestamp
  - `exp`: Expiration timestamp
  - `iss`: Issuer (lokstra-auth)

### Rate Limiting
- **Login endpoint**: 5 requests per 15 minutes per IP
- **Register endpoint**: 3 requests per hour per IP
- **Other endpoints**: 100 requests per minute per user

### Account Lockout
- **Failed attempts threshold**: 5 consecutive failures
- **Lockout duration**: 15 minutes
- **Reset condition**: Successful login or password reset

---

## Multi-Tenant Implementation

### Tenant Isolation
```yaml
Database Level:
  - All user tables include tenant_id column
  - Composite unique index: (tenant_id, email)
  - Row-level security enabled
  - Foreign keys include tenant_id

API Level:
  - X-Tenant-ID header required for all non-public endpoints
  - JWT token includes tenant_id claim
  - All queries filter by tenant_id
  - Cross-tenant access blocked

Validation:
  - Validate X-Tenant-ID on every request
  - Verify JWT tenant_id matches X-Tenant-ID
  - Return 403 if tenant mismatch detected
```

### Data Isolation Example
```sql
-- User registration query with tenant isolation
INSERT INTO auth.users (id, tenant_id, email, password_hash, full_name, role, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
WHERE NOT EXISTS (
  SELECT 1 FROM auth.users 
  WHERE tenant_id = $2 AND email = $3
);

-- User lookup query with tenant isolation
SELECT * FROM auth.users 
WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL;
```

---

## Error Response Format

### Standard Error Structure
```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "field": "fieldName",
    "details": [
      {
        "field": "password",
        "message": "Password is too weak",
        "code": "WEAK_PASSWORD"
      }
    ]
  },
  "requestId": "req_abc123"
}
```

### HTTP Status Code Mapping
- **200 OK**: Successful operation (login, logout, profile fetch)
- **201 Created**: Resource created (user registration)
- **400 Bad Request**: Validation error, invalid input
- **401 Unauthorized**: Missing/invalid/expired token
- **403 Forbidden**: Account locked, insufficient permissions
- **404 Not Found**: User not found
- **409 Conflict**: Email already exists
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server-side error

### Common Error Codes
```yaml
Authentication:
  - INVALID_CREDENTIALS: Wrong email/password
  - TOKEN_EXPIRED: JWT token expired
  - TOKEN_INVALID: Malformed or invalid token
  - REFRESH_TOKEN_INVALID: Invalid refresh token

Authorization:
  - ACCOUNT_LOCKED: Too many failed attempts
  - ACCOUNT_DISABLED: User account disabled
  - EMAIL_NOT_VERIFIED: Email verification required
  - INSUFFICIENT_PERMISSIONS: Missing required permissions

Validation:
  - VALIDATION_ERROR: Generic validation error
  - WEAK_PASSWORD: Password doesn't meet requirements
  - EMAIL_INVALID: Invalid email format
  - REQUIRED_FIELD: Required field missing

Conflict:
  - EMAIL_EXISTS: Email already registered
  - SAME_PASSWORD: New password same as old

Tenant:
  - TENANT_NOT_FOUND: Invalid tenant ID
  - TENANT_MISMATCH: Token tenant != header tenant
  - TENANT_DISABLED: Tenant account disabled
```

---

## Lokstra Handler Implementation Example

```go
package auth

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	authService *AuthService  // @Inject "auth-service"
}

// @Route "POST /register"
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) error {
	// Tenant ID extracted from X-Tenant-ID header
	tenantID := ctx.Request.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return ctx.Api.BadRequest("X-Tenant-ID header is required")
	}

	// Register user with tenant isolation
	user, err := h.authService.RegisterUser(ctx.Context(), tenantID, params)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return ctx.Api.Conflict("Email already registered in this tenant")
		}
		return ctx.Api.InternalServerError("Failed to register user")
	}

	return ctx.Api.Created(map[string]interface{}{
		"userId":        user.ID,
		"email":         user.Email,
		"fullName":      user.FullName,
		"role":          user.Role,
		"tenantId":      user.TenantID,
		"createdAt":     user.CreatedAt,
		"emailVerified": user.EmailVerified,
	})
}

// @Route "POST /login"
func (h *AuthHandler) Login(ctx *request.Context, params *LoginRequest) error {
	tenantID := ctx.Request.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return ctx.Api.BadRequest("X-Tenant-ID header is required")
	}

	// Authenticate and generate tokens
	authResult, err := h.authService.Authenticate(ctx.Context(), tenantID, params)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return ctx.Api.Unauthorized("Invalid email or password")
		}
		if errors.Is(err, ErrAccountLocked) {
			return ctx.Api.Forbidden("Account locked due to multiple failed attempts")
		}
		return ctx.Api.InternalServerError("Login failed")
	}

	return ctx.Api.Ok(authResult)
}

// @Route "GET /me"
func (h *AuthHandler) GetProfile(ctx *request.Context) error {
	// User ID and tenant ID extracted from JWT token
	userID := ctx.Auth.UserID()
	tenantID := ctx.Auth.TenantID()

	profile, err := h.authService.GetUserProfile(ctx.Context(), tenantID, userID)
	if err != nil {
		return ctx.Api.NotFound("User not found")
	}

	return ctx.Api.Ok(profile)
}
```

---

## Testing Checklist

### Functional Tests
- [ ] Register with valid data returns 201
- [ ] Register with duplicate email returns 409
- [ ] Register with weak password returns 400
- [ ] Login with valid credentials returns 200 with tokens
- [ ] Login with invalid credentials returns 401
- [ ] Login after 5 failed attempts returns 403
- [ ] Refresh token returns new access token
- [ ] Expired refresh token returns 401
- [ ] Logout invalidates tokens
- [ ] Get profile with valid token returns 200
- [ ] Get profile with expired token returns 401
- [ ] Change password with correct current password succeeds

### Multi-Tenant Tests
- [ ] User in tenant A cannot access tenant B data
- [ ] Login in tenant A doesn't work in tenant B
- [ ] JWT token includes correct tenant_id claim
- [ ] X-Tenant-ID header validation works
- [ ] Token tenant_id mismatch returns 403
- [ ] Same email can register in different tenants

### Security Tests
- [ ] Passwords are bcrypt hashed with cost 12
- [ ] JWT tokens use RS256 algorithm
- [ ] Rate limiting works on login endpoint
- [ ] Account lockout after 5 failed attempts
- [ ] CORS headers properly configured
- [ ] SQL injection attempts are blocked
- [ ] XSS attempts are sanitized

### Performance Tests
- [ ] Login response time < 200ms (p95)
- [ ] Register response time < 300ms (p95)
- [ ] Profile fetch response time < 100ms (p95)
- [ ] System handles 100 concurrent logins

---

## Integration Points

### Dependencies
```yaml
Services:
  - email-service: Send verification emails
  - audit-service: Log authentication events
  - redis-cache: Store session data and rate limits

Databases:
  - PostgreSQL: User and session storage
  - Redis: Token blacklist and rate limiting

External:
  - SMTP server: Email delivery
  - Monitoring: Prometheus metrics
```

### Events Published
```yaml
user.registered:
  payload: { userId, tenantId, email, role }
  consumers: [email-service, audit-service]

user.logged_in:
  payload: { userId, tenantId, ipAddress, deviceInfo }
  consumers: [audit-service, analytics-service]

user.password_changed:
  payload: { userId, tenantId, timestamp }
  consumers: [email-service, audit-service]
```

---

## Versioning Strategy

### API Version
- **Current**: v1
- **Deprecation Policy**: 6 months notice
- **Breaking Changes**: Require major version bump

### Backward Compatibility
- New optional fields can be added without version bump
- New endpoints can be added without version bump
- Changing required fields requires new version
- Removing endpoints requires new version

---

**File Size**: 19 KB  
**Last Updated**: 2024-01-20  
**Related**: API_SPEC_TEMPLATE.md, MULTI_TENANT_API_PATTERNS.md
