# Lokstra Framework Integration Guide for API Specifications

**Purpose**: Comprehensive guide for mapping API specifications to Lokstra Framework implementation using annotations, handler signatures, and code generation.

**Context**: This guide bridges the gap between API design and Lokstra code, showing exactly how API specs translate into handlers, services, and configurations.

---

## Table of Contents

1. [Overview](#overview)
2. [Annotation Mapping](#annotation-mapping)
3. [Handler Signatures](#handler-signatures)
4. [Request/Response Mapping](#requestresponse-mapping)
5. [Dependency Injection](#dependency-injection)
6. [Configuration Integration](#configuration-integration)
7. [Code Generation](#code-generation)

---

## Overview

### API Specification → Lokstra Implementation Flow

```
API Spec
    ↓
@Handler Annotation (struct + prefix)
    ↓
@Route Annotations (per endpoint)
    ↓
Handler Methods (29+ signatures supported)
    ↓
@Inject Dependencies (services, config)
    ↓
lokstra autogen . (code generation)
    ↓
Running Application
```

### Key Principles

1. **Annotations Drive Everything**: API specs map directly to @Handler and @Route annotations
2. **Type Safety**: DTO structs with validation tags match API schemas
3. **Auto-Validation**: Lokstra validates request params automatically
4. **Flexible Signatures**: 29+ handler signatures support different use cases
5. **Zero Boilerplate**: Framework generates routing, validation, serialization

---

## Annotation Mapping

### @Handler: API Base Path

**API Spec**:
```yaml
Base URL: /api/v1/auth
Module: auth
Endpoints:
  - POST /register
  - POST /login
  - GET /me
```

**Lokstra Code**:
```go
package auth

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	authService *AuthService  // @Inject "auth-service"
	emailService *EmailService  // @Inject "email-service"
}
```

**Mapping Rules**:
- `Base URL` → `prefix` parameter
- `Module name` → `name` parameter (use kebab-case)
- Handler struct name = PascalCase + "Handler"

### @Route: Individual Endpoints

**API Spec**:
```yaml
Endpoint: POST /api/v1/auth/register
Method: POST
Path: /register
Authentication: No
Middlewares: rate-limit
```

**Lokstra Code**:
```go
// @Route "POST /register", middlewares=["rate-limit"]
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) error {
	// Implementation
}
```

**Mapping Rules**:
- `Method + Path` → First parameter of @Route
- `Middlewares` → `middlewares` parameter (array)
- Method name = PascalCase verb (Register, Login, GetProfile, etc.)
- Parameters depend on signature pattern (see Handler Signatures section)

---

## Handler Signatures

Lokstra supports **29+ handler signatures** with automatic parameter binding.

### Pattern 1: Context + DTO → Error

**API Spec**:
```yaml
POST /register
Request Body:
  email: string (required)
  password: string (required)
  fullName: string (required)
Response: 201 Created with user data
```

**Lokstra Code**:
```go
// Request DTO
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"fullName" validate:"required,min=2,max=255"`
}

// @Route "POST /register"
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) error {
	tenantID := ctx.Request.Header.Get("X-Tenant-ID")
	
	user, err := h.authService.RegisterUser(ctx.Context(), tenantID, params)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return ctx.Api.Conflict("Email already registered")
		}
		return ctx.Api.InternalServerError("Registration failed")
	}
	
	return ctx.Api.Created(user)
}
```

**When to Use**:
- POST/PUT/PATCH with request body
- Need full request context
- Complex validation or error handling
- Multi-tenant operations

### Pattern 2: Context + Path Param + DTO → Error

**API Spec**:
```yaml
PATCH /patients/{id}
Path Params:
  id: string (required)
Request Body:
  name: string (optional)
  phone: string (optional)
Response: 200 OK with updated patient
```

**Lokstra Code**:
```go
type UpdatePatientRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,phone"`
}

// @Route "PATCH /patients/:id"
func (h *PatientHandler) UpdatePatient(
	ctx *request.Context, 
	id string, 
	params *UpdatePatientRequest,
) error {
	tenantID := ctx.Auth.TenantID()
	
	patient, err := h.patientService.UpdatePatient(ctx.Context(), tenantID, id, params)
	if err != nil {
		return ctx.Api.NotFound("Patient not found")
	}
	
	return ctx.Api.Ok(patient)
}
```

**When to Use**:
- Update operations with resource ID
- Path parameters + request body
- Partial updates (PATCH)

### Pattern 3: Context + Path Param → Error

**API Spec**:
```yaml
GET /patients/{id}
Path Params:
  id: string (required)
Response: 200 OK with patient data
```

**Lokstra Code**:
```go
// @Route "GET /patients/:id"
func (h *PatientHandler) GetPatient(ctx *request.Context, id string) error {
	tenantID := ctx.Auth.TenantID()
	
	patient, err := h.patientService.GetPatient(ctx.Context(), tenantID, id)
	if err != nil {
		return ctx.Api.NotFound("Patient not found")
	}
	
	return ctx.Api.Ok(patient)
}
```

**When to Use**:
- GET by ID
- DELETE operations
- No request body

### Pattern 4: Context Only → Error

**API Spec**:
```yaml
GET /me
Authentication: Required (JWT)
Response: 200 OK with current user profile
```

**Lokstra Code**:
```go
// @Route "GET /me"
func (h *AuthHandler) GetProfile(ctx *request.Context) error {
	userID := ctx.Auth.UserID()
	tenantID := ctx.Auth.TenantID()
	
	profile, err := h.authService.GetUserProfile(ctx.Context(), tenantID, userID)
	if err != nil {
		return ctx.Api.NotFound("User not found")
	}
	
	return ctx.Api.Ok(profile)
}
```

**When to Use**:
- Authenticated endpoints using JWT claims
- No path params or request body
- Current user operations

### Pattern 5: No Context → (Data, Error)

**API Spec**:
```yaml
GET /health
Authentication: No
Response: 200 OK with health status
```

**Lokstra Code**:
```go
// @Route "GET /health"
func (h *HealthHandler) HealthCheck() (map[string]string, error) {
	return map[string]string{
		"status": "ok",
		"version": "1.0.0",
	}, nil
}
```

**When to Use**:
- Simple public endpoints
- Health checks, version info
- No authentication needed

### Pattern 6: Query Parameters

**API Spec**:
```yaml
GET /patients?status=active&page=2&limit=20
Query Params:
  status: string (optional)
  page: integer (optional, default: 1)
  limit: integer (optional, default: 20)
```

**Lokstra Code**:
```go
type ListPatientsQuery struct {
	Status string `query:"status" validate:"omitempty,oneof=active inactive"`
	Page   int    `query:"page" validate:"omitempty,min=1" default:"1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100" default:"20"`
}

// @Route "GET /patients"
func (h *PatientHandler) ListPatients(ctx *request.Context, query *ListPatientsQuery) error {
	tenantID := ctx.Auth.TenantID()
	
	patients, pagination, err := h.patientService.ListPatients(
		ctx.Context(), tenantID, query,
	)
	if err != nil {
		return ctx.Api.InternalServerError("Failed to list patients")
	}
	
	return ctx.Api.Ok(map[string]interface{}{
		"data": patients,
		"pagination": pagination,
	})
}
```

**Tag Mapping**:
- `query:"paramName"` → Query parameter binding
- `validate:"..."` → Validation rules
- `default:"value"` → Default value if not provided

---

## Request/Response Mapping

### Request DTO Structure

**API Spec Schema**:
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
    password:
      type: string
      minLength: 8
      maxLength: 72
      pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).*$"
    fullName:
      type: string
      minLength: 2
      maxLength: 255
    role:
      type: string
      enum: [admin, doctor, nurse, receptionist]
    metadata:
      type: object
      additionalProperties: true
```

**Lokstra DTO**:
```go
type RegisterRequest struct {
	Email    string                 `json:"email" validate:"required,email,min=5,max=255"`
	Password string                 `json:"password" validate:"required,min=8,max=72,password_strength"`
	FullName string                 `json:"fullName" validate:"required,min=2,max=255"`
	Role     string                 `json:"role" validate:"required,oneof=admin doctor nurse receptionist"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

**Tag Mapping**:

| API Spec | Lokstra Validation Tag | Example |
|----------|------------------------|---------|
| required: true | `validate:"required"` | Required field |
| minLength: 8 | `validate:"min=8"` | Min string length |
| maxLength: 72 | `validate:"max=72"` | Max string length |
| format: email | `validate:"email"` | Email format |
| format: phone | `validate:"phone"` | Phone number |
| format: uuid | `validate:"uuid"` | UUID format |
| pattern: "^[A-Z]" | `validate:"regexp=^[A-Z]"` | Regex pattern |
| enum: [a, b] | `validate:"oneof=a b"` | Enum values |
| minimum: 0 | `validate:"min=0"` | Min number |
| maximum: 100 | `validate:"max=100"` | Max number |
| type: array | `validate:"dive"` | Validate array items |

### Response Helpers

**API Spec**:
```yaml
Success Response (200 OK):
  {
    "status": "success",
    "data": {...}
  }

Error Response (400 Bad Request):
  {
    "status": "error",
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Invalid input"
    }
  }
```

**Lokstra Response Helpers**:
```go
// Success responses
ctx.Api.Ok(data)                    // 200 OK
ctx.Api.Created(data)               // 201 Created
ctx.Api.NoContent()                 // 204 No Content

// Error responses
ctx.Api.BadRequest("message")       // 400 Bad Request
ctx.Api.Unauthorized("message")     // 401 Unauthorized
ctx.Api.Forbidden("message")        // 403 Forbidden
ctx.Api.NotFound("message")         // 404 Not Found
ctx.Api.Conflict("message")         // 409 Conflict
ctx.Api.InternalServerError("msg")  // 500 Internal Server Error

// Custom response
ctx.Api.Response(statusCode, data)  // Custom status code
```

**Standard Response Format** (Auto-generated):
```json
{
  "status": "success",
  "message": "Operation successful",
  "data": {
    // Your response data
  }
}
```

---

## Dependency Injection

### Service Injection

**API Spec**:
```yaml
Dependencies:
  - auth-service: User authentication logic
  - email-service: Send verification emails
  - redis-cache: Token blacklist
```

**Lokstra Code**:
```go
// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	authService  *AuthService   // @Inject "auth-service"
	emailService *EmailService  // @Inject "email-service"
	cache        redis.Client   // @Inject "redis"
}
```

**Injection Patterns**:

#### Pattern 1: Direct Service Injection
```go
authService *AuthService  // @Inject "auth-service"
```

#### Pattern 2: Config Value Injection
```go
jwtSecret string  // @Inject "cfg:jwt.secret"
jwtExpiry int     // @Inject "cfg:jwt.expiry"
```

#### Pattern 3: Config-Based Service Injection
```go
// In config.yaml:
# repository:
#   impl: "patient-repo-pg"

patientRepo repository.PatientRepo  // @Inject "@repository.impl"
```

#### Pattern 4: Database Injection
```go
db *sql.DB  // @Inject "database"
```

### Service Registration

**Service Implementation**:
```go
package auth

import "database/sql"

// @Service "auth-service"
type AuthService struct {
	db    *sql.DB          // @Inject "database"
	cache redis.Client     // @Inject "redis"
	cfg   *AuthConfig      // @Inject "cfg:auth"
}

func NewAuthService(db *sql.DB, cache redis.Client, cfg *AuthConfig) *AuthService {
	return &AuthService{
		db: db,
		cache: cache,
		cfg: cfg,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, tenantID string, req *RegisterRequest) (*User, error) {
	// Implementation
}
```

**Factory Function Pattern**:
```go
// Auto-generated by lokstra
func CreateAuthService(registry service.Registry) (*AuthService, error) {
	db := registry.MustGet("database").(*sql.DB)
	cache := registry.MustGet("redis").(redis.Client)
	cfg := registry.MustGetConfig("auth").(*AuthConfig)
	
	return NewAuthService(db, cache, cfg), nil
}
```

---

## Configuration Integration

### API Spec → Configuration

**API Spec**:
```yaml
Configuration:
  jwt:
    algorithm: RS256
    accessTokenExpiry: 3600
    refreshTokenExpiry: 2592000
    publicKeyPath: /keys/public.pem
    privateKeyPath: /keys/private.pem
  
  security:
    bcryptCost: 12
    maxLoginAttempts: 5
    lockoutDuration: 900
  
  rateLimit:
    enabled: true
    requestsPerMinute: 100
```

**Lokstra Config (configs/auth.yaml)**:
```yaml
jwt:
  algorithm: RS256
  accessTokenExpiry: 3600
  refreshTokenExpiry: 2592000
  publicKeyPath: ${JWT_PUBLIC_KEY_PATH:/keys/public.pem}
  privateKeyPath: ${JWT_PRIVATE_KEY_PATH:/keys/private.pem}

security:
  bcryptCost: ${BCRYPT_COST:12}
  maxLoginAttempts: ${MAX_LOGIN_ATTEMPTS:5}
  lockoutDuration: ${LOCKOUT_DURATION:900}

rateLimit:
  enabled: ${RATE_LIMIT_ENABLED:true}
  requestsPerMinute: ${RATE_LIMIT_RPM:100}
```

**Config Struct**:
```go
type AuthConfig struct {
	JWT      JWTConfig      `yaml:"jwt"`
	Security SecurityConfig `yaml:"security"`
	RateLimit RateLimitConfig `yaml:"rateLimit"`
}

type JWTConfig struct {
	Algorithm           string `yaml:"algorithm"`
	AccessTokenExpiry   int    `yaml:"accessTokenExpiry"`
	RefreshTokenExpiry  int    `yaml:"refreshTokenExpiry"`
	PublicKeyPath       string `yaml:"publicKeyPath"`
	PrivateKeyPath      string `yaml:"privateKeyPath"`
}

type SecurityConfig struct {
	BcryptCost       int `yaml:"bcryptCost"`
	MaxLoginAttempts int `yaml:"maxLoginAttempts"`
	LockoutDuration  int `yaml:"lockoutDuration"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requestsPerMinute"`
}
```

**Injection in Handler**:
```go
// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	authService *AuthService   // @Inject "auth-service"
	jwtExpiry   int            // @Inject "cfg:jwt.accessTokenExpiry"
	rateLimitRPM int           // @Inject "cfg:rateLimit.requestsPerMinute"
}
```

---

## Code Generation

### Process Flow

```
1. Write Handler with Annotations
   ↓
2. Run: lokstra autogen .
   ↓
3. Generated Code:
   - lokstra_registry.go (service registry)
   - Route mappings
   - Validation rules
   - Dependency wiring
   ↓
4. go run .
```

### Example: Full Handler → Generated Code

**Source (handler.go)**:
```go
package auth

import "github.com/primadi/lokstra/core/request"

// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	authService *AuthService  // @Inject "auth-service"
}

// @Route "POST /register"
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) error {
	// Implementation
}

// @Route "POST /login"
func (h *AuthHandler) Login(ctx *request.Context, params *LoginRequest) error {
	// Implementation
}

// @Route "GET /me"
func (h *AuthHandler) GetProfile(ctx *request.Context) error {
	// Implementation
}
```

**Generated (lokstra_registry.go)**:
```go
package lokstra_registry

import (
	"github.com/primadi/lokstra/core/route"
	"yourapp/auth"
)

func init() {
	// Register handler
	route.RegisterHandler("auth-handler", func(registry service.Registry) interface{} {
		authService := registry.MustGet("auth-service").(*auth.AuthService)
		
		return &auth.AuthHandler{
			authService: authService,
		}
	})
	
	// Register routes
	route.RegisterRoute("POST", "/api/v1/auth/register", "auth-handler", "Register")
	route.RegisterRoute("POST", "/api/v1/auth/login", "auth-handler", "Login")
	route.RegisterRoute("GET", "/api/v1/auth/me", "auth-handler", "GetProfile")
}
```

### Auto-Generated Features

1. **Route Mapping**: HTTP method + path → handler method
2. **Dependency Resolution**: @Inject → service lookup and injection
3. **Parameter Binding**: Request → DTO struct
4. **Validation**: validate tags → automatic validation
5. **Response Serialization**: Return value → JSON response
6. **Error Handling**: error return → appropriate HTTP status

---

## Multi-Tenant Integration

### Tenant Context Extraction

**API Spec**:
```yaml
Multi-Tenant: Yes
Tenant Identification: X-Tenant-ID header + JWT tenant_id claim
Validation: Both must match
```

**Lokstra Implementation**:
```go
// @Route "POST /patients"
func (h *PatientHandler) CreatePatient(ctx *request.Context, params *CreatePatientRequest) error {
	// Option 1: From header (validated by middleware)
	tenantID := ctx.Request.Header.Get("X-Tenant-ID")
	
	// Option 2: From JWT claims (recommended)
	tenantID := ctx.Auth.TenantID()
	
	// Option 3: Custom context value
	tenantID := ctx.Value("tenant_id").(string)
	
	// Always pass tenantID to service layer
	patient, err := h.patientService.CreatePatient(ctx.Context(), tenantID, params)
	// ...
}
```

### Tenant Validation Middleware

```go
package middleware

import "github.com/primadi/lokstra/core/request"

// @Middleware "tenant-validation"
func TenantValidationMiddleware() func(*request.Context) error {
	return func(ctx *request.Context) error {
		// Get from header
		headerTenantID := ctx.Request.Header.Get("X-Tenant-ID")
		if headerTenantID == "" {
			return ctx.Api.BadRequest("X-Tenant-ID header required")
		}
		
		// Get from JWT
		claims := ctx.Auth.Claims()
		tokenTenantID, ok := claims["tenant_id"].(string)
		if !ok {
			return ctx.Api.Unauthorized("Invalid token: missing tenant_id")
		}
		
		// Validate match
		if headerTenantID != tokenTenantID {
			return ctx.Api.Forbidden("Tenant mismatch")
		}
		
		// Add to context for handlers
		ctx.Set("tenant_id", headerTenantID)
		return nil
	}
}
```

**Usage in Handler**:
```go
// @Route "POST /patients", middlewares=["auth", "tenant-validation"]
func (h *PatientHandler) CreatePatient(ctx *request.Context, params *CreatePatientRequest) error {
	tenantID := ctx.Value("tenant_id").(string)
	// ...
}
```

---

## Complete Example: Patient API

### API Specification
```yaml
Module: patient
Base URL: /api/v1/patients
Multi-Tenant: Yes

Endpoints:
  - POST /patients (Create patient)
  - GET /patients (List patients with pagination)
  - GET /patients/{id} (Get patient by ID)
  - PATCH /patients/{id} (Update patient)
  - DELETE /patients/{id} (Delete patient)

Dependencies:
  - patient-service: Business logic
  - database: PostgreSQL connection
```

### Lokstra Implementation

**handler.go**:
```go
package patient

import "github.com/primadi/lokstra/core/request"

// @Handler name="patient-handler", prefix="/api/v1/patients"
type PatientHandler struct {
	patientService *PatientService  // @Inject "patient-service"
}

// @Route "POST /", middlewares=["auth", "tenant-validation", "permission:patient:create"]
func (h *PatientHandler) CreatePatient(ctx *request.Context, params *CreatePatientRequest) error {
	tenantID := ctx.Auth.TenantID()
	
	patient, err := h.patientService.CreatePatient(ctx.Context(), tenantID, params)
	if err != nil {
		return ctx.Api.InternalServerError("Failed to create patient")
	}
	
	return ctx.Api.Created(patient)
}

// @Route "GET /", middlewares=["auth", "tenant-validation", "permission:patient:read"]
func (h *PatientHandler) ListPatients(ctx *request.Context, query *ListPatientsQuery) error {
	tenantID := ctx.Auth.TenantID()
	
	patients, pagination, err := h.patientService.ListPatients(ctx.Context(), tenantID, query)
	if err != nil {
		return ctx.Api.InternalServerError("Failed to list patients")
	}
	
	return ctx.Api.Ok(map[string]interface{}{
		"data": patients,
		"pagination": pagination,
	})
}

// @Route "GET /:id", middlewares=["auth", "tenant-validation", "permission:patient:read"]
func (h *PatientHandler) GetPatient(ctx *request.Context, id string) error {
	tenantID := ctx.Auth.TenantID()
	
	patient, err := h.patientService.GetPatient(ctx.Context(), tenantID, id)
	if err != nil {
		return ctx.Api.NotFound("Patient not found")
	}
	
	return ctx.Api.Ok(patient)
}

// @Route "PATCH /:id", middlewares=["auth", "tenant-validation", "permission:patient:update"]
func (h *PatientHandler) UpdatePatient(ctx *request.Context, id string, params *UpdatePatientRequest) error {
	tenantID := ctx.Auth.TenantID()
	
	patient, err := h.patientService.UpdatePatient(ctx.Context(), tenantID, id, params)
	if err != nil {
		return ctx.Api.NotFound("Patient not found")
	}
	
	return ctx.Api.Ok(patient)
}

// @Route "DELETE /:id", middlewares=["auth", "tenant-validation", "permission:patient:delete"]
func (h *PatientHandler) DeletePatient(ctx *request.Context, id string) error {
	tenantID := ctx.Auth.TenantID()
	
	err := h.patientService.DeletePatient(ctx.Context(), tenantID, id)
	if err != nil {
		return ctx.Api.NotFound("Patient not found")
	}
	
	return ctx.Api.NoContent()
}
```

**dto.go**:
```go
package patient

type CreatePatientRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,phone"`
	Address  string `json:"address" validate:"omitempty,max=500"`
	DateOfBirth string `json:"dateOfBirth" validate:"required,datetime=2006-01-02"`
}

type UpdatePatientRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,phone"`
	Address  *string `json:"address,omitempty" validate:"omitempty,max=500"`
}

type ListPatientsQuery struct {
	Page   int    `query:"page" validate:"omitempty,min=1" default:"1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100" default:"20"`
	Status string `query:"status" validate:"omitempty,oneof=active inactive"`
	Search string `query:"q" validate:"omitempty,min=2"`
}
```

**service.go**:
```go
package patient

import (
	"context"
	"database/sql"
)

// @Service "patient-service"
type PatientService struct {
	db *sql.DB  // @Inject "database"
}

func NewPatientService(db *sql.DB) *PatientService {
	return &PatientService{db: db}
}

func (s *PatientService) CreatePatient(ctx context.Context, tenantID string, req *CreatePatientRequest) (*Patient, error) {
	query := `
		INSERT INTO patients (id, tenant_id, name, email, phone, address, date_of_birth, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, tenant_id, name, email, phone, address, date_of_birth, created_at, updated_at
	`
	
	id := generateID()
	var patient Patient
	err := s.db.QueryRowContext(ctx, query, 
		id, tenantID, req.Name, req.Email, req.Phone, req.Address, req.DateOfBirth,
	).Scan(
		&patient.ID, &patient.TenantID, &patient.Name, &patient.Email, 
		&patient.Phone, &patient.Address, &patient.DateOfBirth, 
		&patient.CreatedAt, &patient.UpdatedAt,
	)
	
	return &patient, err
}

func (s *PatientService) GetPatient(ctx context.Context, tenantID, id string) (*Patient, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, address, date_of_birth, created_at, updated_at
		FROM patients
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`
	
	var patient Patient
	err := s.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&patient.ID, &patient.TenantID, &patient.Name, &patient.Email, 
		&patient.Phone, &patient.Address, &patient.DateOfBirth, 
		&patient.CreatedAt, &patient.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, ErrPatientNotFound
	}
	
	return &patient, err
}

// ... other methods
```

---

## Quick Reference

### Annotation Cheat Sheet

```go
// Handler definition
// @Handler name="handler-name", prefix="/api/v1/path"
type MyHandler struct {
	service *MyService  // @Inject "my-service"
	config  *MyConfig   // @Inject "cfg:my.config"
}

// Route definition
// @Route "METHOD /path/:param", middlewares=["mw1", "mw2"]
func (h *MyHandler) MethodName(ctx *request.Context, param string, dto *RequestDTO) error {
	return ctx.Api.Ok(data)
}

// Service definition
// @Service "service-name"
type MyService struct {
	db *sql.DB  // @Inject "database"
}
```

### Common Validation Tags

```go
validate:"required"              // Must be provided
validate:"required,email"        // Required email
validate:"required,min=8,max=72" // Required, length 8-72
validate:"omitempty,email"       // Optional email
validate:"oneof=a b c"          // Must be a, b, or c
validate:"min=0,max=100"        // Number range
validate:"phone"                 // Phone number
validate:"uuid"                  // UUID format
validate:"datetime=2006-01-02"   // Date format
```

### Response Helper Quick Reference

```go
ctx.Api.Ok(data)                   // 200 OK
ctx.Api.Created(data)              // 201 Created
ctx.Api.NoContent()                // 204 No Content
ctx.Api.BadRequest("msg")          // 400
ctx.Api.Unauthorized("msg")        // 401
ctx.Api.Forbidden("msg")           // 403
ctx.Api.NotFound("msg")            // 404
ctx.Api.Conflict("msg")            // 409
ctx.Api.InternalServerError("msg") // 500
```

---

**File Size**: 18 KB  
**Last Updated**: 2024-01-20  
**Related**: AUTH_API_EXAMPLE.md, API_SPEC_TEMPLATE.md, MULTI_TENANT_API_PATTERNS.md
