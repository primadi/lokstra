# Multi-Tenant API Design Patterns

**Purpose**: Comprehensive patterns for designing APIs in multi-tenant systems with strict data isolation, security, and scalability.

**Context**: Multi-tenant systems require careful API design to prevent data leakage, ensure performance, and maintain security across tenant boundaries.

---

## Table of Contents

1. [Tenant Identification Patterns](#tenant-identification-patterns)
2. [Data Isolation Patterns](#data-isolation-patterns)
3. [Security Patterns](#security-patterns)
4. [Error Handling Patterns](#error-handling-patterns)
5. [Performance Patterns](#performance-patterns)
6. [Testing Patterns](#testing-patterns)

---

## Tenant Identification Patterns

### Pattern 1: Header-Based Identification (Recommended)

**When to Use**: Most common pattern, works well with JWT authentication

```yaml
Request:
  headers:
    X-Tenant-ID: clinic_001
    Authorization: Bearer <jwt_token>

Validation:
  1. Extract X-Tenant-ID from header
  2. Extract tenant_id from JWT claims
  3. Verify both match
  4. Return 403 if mismatch

Advantages:
  - Clean URL structure
  - Works with all HTTP methods
  - Easy to validate
  - Standard practice

Disadvantages:
  - Requires header on every request
  - Can be forgotten by clients
```

**Implementation**:
```go
// Middleware to validate tenant context
func ValidateTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerTenantID := r.Header.Get("X-Tenant-ID")
		if headerTenantID == "" {
			http.Error(w, "X-Tenant-ID header required", http.StatusBadRequest)
			return
		}

		// Extract from JWT
		claims := r.Context().Value("jwt_claims").(JWTClaims)
		tokenTenantID := claims.TenantID

		if headerTenantID != tokenTenantID {
			http.Error(w, "Tenant mismatch", http.StatusForbidden)
			return
		}

		// Add to context for handlers
		ctx := context.WithValue(r.Context(), "tenant_id", headerTenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### Pattern 2: URL Path-Based Identification

**When to Use**: Public-facing APIs, tenant-specific subdomains not available

```yaml
URL Structure:
  /api/v1/tenants/{tenantId}/patients
  /api/v1/tenants/{tenantId}/appointments

Validation:
  1. Extract tenantId from URL path
  2. Compare with JWT tenant_id claim
  3. Return 403 if mismatch

Advantages:
  - Self-documenting URLs
  - No additional headers needed
  - Easy to test with curl

Disadvantages:
  - Longer URLs
  - Tenant ID exposed in logs
  - More complex routing
```

**Example**:
```yaml
GET /api/v1/tenants/clinic_001/patients/123

vs.

GET /api/v1/patients/123
Headers: X-Tenant-ID: clinic_001
```

### Pattern 3: Subdomain-Based Identification

**When to Use**: SaaS applications with custom domains

```yaml
URL Structure:
  https://clinic001.yourdomain.com/api/v1/patients
  https://clinic002.yourdomain.com/api/v1/patients

Tenant Resolution:
  1. Extract subdomain from Host header
  2. Map subdomain to tenant_id
  3. Validate against JWT tenant_id

Advantages:
  - Clean API URLs
  - Natural tenant separation
  - Good for branding

Disadvantages:
  - Requires DNS configuration
  - SSL certificate complexity
  - Subdomain mapping overhead
```

---

## Data Isolation Patterns

### Pattern 1: Query-Level Filtering (Mandatory)

**Description**: Always include tenant_id in WHERE clauses

```sql
-- ✅ CORRECT: With tenant filtering
SELECT * FROM patients 
WHERE tenant_id = $1 AND patient_id = $2 AND deleted_at IS NULL;

-- ❌ WRONG: Missing tenant filter (security vulnerability!)
SELECT * FROM patients 
WHERE patient_id = $1 AND deleted_at IS NULL;
```

**Lokstra Implementation**:
```go
// Repository with tenant isolation
// @Service "patient-repo"
type PatientRepository struct {
	db *sql.DB  // @Inject "database"
}

func (r *PatientRepository) GetByID(ctx context.Context, tenantID, patientID string) (*Patient, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, created_at, updated_at
		FROM patients
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`
	
	var patient Patient
	err := r.db.QueryRowContext(ctx, query, tenantID, patientID).Scan(
		&patient.ID, &patient.TenantID, &patient.Name, 
		&patient.Email, &patient.Phone, &patient.CreatedAt, &patient.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, ErrPatientNotFound
	}
	return &patient, err
}
```

### Pattern 2: Database-Level Row Security

**Description**: Use PostgreSQL Row-Level Security (RLS) as defense-in-depth

```sql
-- Enable RLS on table
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON patients
    USING (tenant_id = current_setting('app.current_tenant_id')::text);

-- Set tenant context at connection level
SET app.current_tenant_id = 'clinic_001';
```

**Advantages**:
- Automatic enforcement
- Defense against SQL injection
- Catches developer mistakes

**Disadvantages**:
- Performance overhead
- Requires session variables
- Complex debugging

### Pattern 3: Composite Foreign Keys

**Description**: Include tenant_id in foreign key relationships

```sql
-- Parent table
CREATE TABLE clinics (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    UNIQUE(tenant_id, id)
);

-- Child table with composite FK
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    clinic_id TEXT NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (tenant_id, clinic_id) 
        REFERENCES clinics(tenant_id, id),
    UNIQUE(tenant_id, id)
);

-- This prevents cross-tenant references!
-- INSERT will fail if tenant_id doesn't match
```

### Pattern 4: Tenant-Specific Indexes

**Description**: Create indexes that include tenant_id for performance

```sql
-- Composite index for tenant queries
CREATE INDEX idx_patients_tenant_email 
    ON patients(tenant_id, email) 
    WHERE deleted_at IS NULL;

-- Covering index for common queries
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date, patient_id, doctor_id)
    INCLUDE (status, notes);
```

---

## Security Patterns

### Pattern 1: JWT Token Claims Validation

**Required Claims**:
```json
{
  "sub": "user_id",
  "tenant_id": "clinic_001",
  "role": "doctor",
  "permissions": ["patient:read", "patient:write"],
  "iat": 1705315800,
  "exp": 1705319400
}
```

**Validation Rules**:
```go
func ValidateJWTClaims(token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)
	
	// 1. Verify tenant_id exists
	tenantID, ok := claims["tenant_id"].(string)
	if !ok || tenantID == "" {
		return ErrMissingTenantClaim
	}
	
	// 2. Verify role exists
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return ErrMissingRoleClaim
	}
	
	// 3. Verify expiration
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp {
		return ErrTokenExpired
	}
	
	// 4. Verify issuer
	if claims["iss"] != "lokstra-auth" {
		return ErrInvalidIssuer
	}
	
	return nil
}
```

### Pattern 2: Permission-Based Authorization

**Structure**:
```yaml
Permissions Format: "resource:action"

Examples:
  - patient:read
  - patient:write
  - patient:delete
  - appointment:manage
  - report:export

Role Mappings:
  admin:
    - *:*
  doctor:
    - patient:read
    - patient:write
    - appointment:manage
    - prescription:create
  nurse:
    - patient:read
    - appointment:read
  receptionist:
    - patient:read
    - appointment:manage
```

**Implementation**:
```go
// Permission checker middleware
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("jwt_claims").(JWTClaims)
			
			// Check if user has permission
			if !hasPermission(claims.Permissions, permission) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Usage in handler
// @Route "DELETE /patients/:id", middlewares=["auth", "permission:patient:delete"]
```

### Pattern 3: Rate Limiting Per Tenant

**Strategy**: Prevent one tenant from affecting others

```yaml
Rate Limit Scopes:
  1. Per Tenant: Prevent single tenant from overloading
  2. Per User: Prevent single user abuse
  3. Per IP: Prevent DDoS attacks

Limits:
  Normal Operations:
    - 1000 requests/minute per tenant
    - 100 requests/minute per user
    
  Resource-Intensive:
    - 10 report exports/hour per tenant
    - 5 bulk operations/hour per user
    
  Authentication:
    - 5 login attempts/15min per IP
    - 3 registration/hour per IP
```

**Redis Implementation**:
```go
func CheckRateLimit(tenantID, userID string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", tenantID, userID)
	
	// Increment counter
	count, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	
	// Set expiration on first request
	if count == 1 {
		redisClient.Expire(ctx, key, window)
	}
	
	// Check limit
	return count <= int64(limit), nil
}
```

### Pattern 4: Audit Logging

**What to Log**:
```yaml
Authentication Events:
  - Login success/failure
  - Logout
  - Token refresh
  - Password change

Data Access:
  - Read sensitive data (patient records)
  - Create/Update/Delete operations
  - Bulk exports
  - Permission changes

Security Events:
  - Failed authorization attempts
  - Tenant mismatch attempts
  - Rate limit violations
  - Suspicious patterns
```

**Log Format**:
```json
{
  "timestamp": "2024-01-20T10:30:00Z",
  "event_type": "data_access",
  "tenant_id": "clinic_001",
  "user_id": "usr_abc123",
  "resource": "patient",
  "resource_id": "pat_xyz789",
  "action": "read",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "request_id": "req_abc123",
  "status": "success"
}
```

---

## Error Handling Patterns

### Pattern 1: Tenant-Specific Error Responses

**Never leak cross-tenant information in errors!**

```yaml
# ❌ BAD: Reveals tenant existence
{
  "error": "Patient not found in tenant clinic_002, did you mean clinic_001?"
}

# ✅ GOOD: Generic error
{
  "status": "error",
  "error": {
    "code": "PATIENT_NOT_FOUND",
    "message": "Patient not found"
  }
}
```

### Pattern 2: Error Code Hierarchy

```yaml
Format: DOMAIN_CATEGORY_DETAIL

Examples:
  AUTH_VALIDATION_EMAIL_INVALID
  PATIENT_AUTHORIZATION_INSUFFICIENT_PERMISSION
  APPOINTMENT_CONFLICT_TIME_OVERLAP
  TENANT_VALIDATION_MISMATCH

Hierarchy:
  - DOMAIN: Module/resource (AUTH, PATIENT, APPOINTMENT)
  - CATEGORY: Error type (VALIDATION, AUTHORIZATION, CONFLICT)
  - DETAIL: Specific issue (EMAIL_INVALID, INSUFFICIENT_PERMISSION)
```

### Pattern 3: Tenant Context in Error Logs

**Log Format**:
```json
{
  "level": "error",
  "timestamp": "2024-01-20T10:30:00Z",
  "message": "Failed to fetch patient",
  "tenant_id": "clinic_001",
  "user_id": "usr_abc123",
  "request_id": "req_abc123",
  "error": {
    "type": "database_error",
    "detail": "connection timeout",
    "query": "SELECT * FROM patients WHERE..."
  },
  "stack_trace": "..."
}
```

**Never log in response to user**:
- SQL queries with tenant IDs
- Stack traces
- Internal error details
- Other tenants' data

---

## Performance Patterns

### Pattern 1: Tenant-Aware Caching

**Strategy**: Cache per tenant to prevent cross-contamination

```yaml
Cache Key Format: "tenant:{tenant_id}:resource:{resource_type}:{id}"

Examples:
  - tenant:clinic_001:patient:pat_123
  - tenant:clinic_001:config:app_settings
  - tenant:clinic_002:patient:pat_456

Invalidation Rules:
  - Invalidate on UPDATE/DELETE
  - TTL: 5-15 minutes for volatile data
  - Invalidate all tenant keys on tenant config change
```

**Redis Implementation**:
```go
func GetPatientCached(ctx context.Context, tenantID, patientID string) (*Patient, error) {
	// Check cache
	cacheKey := fmt.Sprintf("tenant:%s:patient:%s", tenantID, patientID)
	cached, err := redisClient.Get(ctx, cacheKey).Result()
	
	if err == nil {
		var patient Patient
		json.Unmarshal([]byte(cached), &patient)
		return &patient, nil
	}
	
	// Fetch from DB
	patient, err := repo.GetPatient(ctx, tenantID, patientID)
	if err != nil {
		return nil, err
	}
	
	// Cache for 10 minutes
	json, _ := json.Marshal(patient)
	redisClient.Set(ctx, cacheKey, json, 10*time.Minute)
	
	return patient, nil
}
```

### Pattern 2: Query Optimization for Multi-Tenancy

**Indexing Strategy**:
```sql
-- Always index tenant_id as first column in composite indexes
CREATE INDEX idx_patients_tenant_name 
    ON patients(tenant_id, name);

-- Index for common tenant queries
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date)
    WHERE cancelled_at IS NULL;

-- Partial index for active records only
CREATE INDEX idx_users_tenant_email_active 
    ON users(tenant_id, email)
    WHERE deleted_at IS NULL;
```

**Query Pattern**:
```sql
-- ✅ GOOD: Uses tenant index
EXPLAIN SELECT * FROM patients 
WHERE tenant_id = 'clinic_001' AND name LIKE 'John%'
LIMIT 20;

-- Index Scan using idx_patients_tenant_name

-- ❌ BAD: Full table scan
EXPLAIN SELECT * FROM patients 
WHERE name LIKE 'John%' AND tenant_id = 'clinic_001'
LIMIT 20;

-- Sequential Scan (slow!)
```

### Pattern 3: Pagination for Large Datasets

**Cursor-Based Pagination** (recommended for multi-tenant):
```yaml
Request:
  GET /api/v1/patients?limit=20&cursor=eyJpZCI6InBhdF8xMjMiLCJ0ZW5hbnRfaWQiOiJjbGluaWNfMDAxIn0=

Response:
  {
    "data": [...],
    "pagination": {
      "hasNext": true,
      "nextCursor": "eyJpZCI6InBhdF8yMDAiLCJ0ZW5hbnRfaWQiOiJjbGluaWNfMDAxIn0=",
      "limit": 20
    }
  }

Cursor Format (Base64):
  {
    "id": "pat_200",
    "tenant_id": "clinic_001"
  }

Query:
  SELECT * FROM patients
  WHERE tenant_id = $1 
    AND id > $2  -- from decoded cursor
  ORDER BY id
  LIMIT $3;
```

**Advantages**:
- Consistent results during pagination
- No missing/duplicate records
- Better performance than offset
- Works well with tenant isolation

### Pattern 4: Bulk Operations

**Pattern**: Process in tenant-specific batches

```go
// ✅ GOOD: Batch by tenant
func BulkUpdatePatients(ctx context.Context, tenantID string, updates []PatientUpdate) error {
	// All updates are for same tenant
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	
	for _, update := range updates {
		_, err := tx.Exec(
			"UPDATE patients SET name = $1 WHERE tenant_id = $2 AND id = $3",
			update.Name, tenantID, update.ID,
		)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

// ❌ BAD: Mixed tenants in one transaction
func BulkUpdatePatientsCrossTenant(updates []PatientUpdate) error {
	// Security risk! Updates might cross tenant boundaries
	for _, update := range updates {
		db.Exec("UPDATE patients SET name = $1 WHERE id = $2", 
			update.Name, update.ID)
	}
}
```

---

## Testing Patterns

### Pattern 1: Cross-Tenant Isolation Tests

**Test Cases**:
```go
func TestCrossTenantIsolation(t *testing.T) {
	// Setup: Create patients in two tenants
	patientTenant1 := createPatient(t, "clinic_001", "John Doe")
	patientTenant2 := createPatient(t, "clinic_002", "Jane Doe")
	
	// Test: Try to access tenant2 patient with tenant1 credentials
	jwtTenant1 := generateJWT(t, "clinic_001", "user_123")
	
	resp := apiRequest(t, "GET", "/api/v1/patients/" + patientTenant2.ID, 
		map[string]string{
			"X-Tenant-ID": "clinic_001",
			"Authorization": "Bearer " + jwtTenant1,
		})
	
	// Assert: Should return 404 or 403, never the data
	assert.Equal(t, 404, resp.StatusCode)
	assert.NotContains(t, resp.Body, "Jane Doe")
}
```

### Pattern 2: Tenant Mismatch Tests

```go
func TestTenantMismatch(t *testing.T) {
	// JWT has tenant_id = clinic_001
	jwt := generateJWT(t, "clinic_001", "user_123")
	
	// But header says clinic_002
	resp := apiRequest(t, "GET", "/api/v1/patients", 
		map[string]string{
			"X-Tenant-ID": "clinic_002",
			"Authorization": "Bearer " + jwt,
		})
	
	// Should return 403 Forbidden
	assert.Equal(t, 403, resp.StatusCode)
	assert.Contains(t, resp.Body, "tenant_mismatch")
}
```

### Pattern 3: Performance Tests Per Tenant

```go
func TestTenantPerformanceIsolation(t *testing.T) {
	// Setup: Create heavy load on tenant1
	go simulateHeavyLoad(t, "clinic_001", 1000) // 1000 req/sec
	
	// Test: Measure tenant2 response times
	tenant2Times := measureResponseTimes(t, "clinic_002", 100)
	
	// Assert: Tenant2 not affected by tenant1 load
	p95 := percentile(tenant2Times, 0.95)
	assert.Less(t, p95, 200*time.Millisecond, 
		"Tenant2 response time degraded due to tenant1 load")
}
```

---

## Common Anti-Patterns

### ❌ Anti-Pattern 1: Optional Tenant Filtering

```go
// WRONG: Tenant filtering is optional
func GetPatients(ctx context.Context, tenantID string) ([]Patient, error) {
	query := "SELECT * FROM patients WHERE 1=1"
	if tenantID != "" {  // ← BUG! Should always be required
		query += " AND tenant_id = $1"
	}
	// ...
}
```

**Fix**: Always require tenant_id
```go
func GetPatients(ctx context.Context, tenantID string) ([]Patient, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	query := "SELECT * FROM patients WHERE tenant_id = $1"
	// ...
}
```

### ❌ Anti-Pattern 2: Trusting Client-Provided Tenant ID

```go
// WRONG: Using tenant ID from request body
func CreatePatient(ctx context.Context, req *CreatePatientRequest) error {
	patient := &Patient{
		TenantID: req.TenantID,  // ← BUG! Client can spoof this
		Name: req.Name,
	}
	return repo.Save(ctx, patient)
}
```

**Fix**: Always extract tenant from JWT
```go
func CreatePatient(ctx context.Context, req *CreatePatientRequest) error {
	// Get tenant from authenticated context
	tenantID := ctx.Value("tenant_id").(string)
	
	patient := &Patient{
		TenantID: tenantID,  // ✅ From JWT, not from client
		Name: req.Name,
	}
	return repo.Save(ctx, patient)
}
```

### ❌ Anti-Pattern 3: Shared Cache Keys

```go
// WRONG: Cache key doesn't include tenant
cacheKey := fmt.Sprintf("patient:%s", patientID)  // ← BUG!
```

**Fix**: Always include tenant in cache keys
```go
cacheKey := fmt.Sprintf("tenant:%s:patient:%s", tenantID, patientID)
```

---

## Checklist for API Review

### Tenant Isolation
- [ ] All queries include tenant_id filter
- [ ] Foreign keys include tenant_id
- [ ] Indexes include tenant_id as first column
- [ ] Cache keys include tenant_id
- [ ] No optional tenant filtering

### Security
- [ ] X-Tenant-ID header validated
- [ ] JWT tenant_id claim validated
- [ ] Header and token tenant_id match verified
- [ ] Permissions checked for operations
- [ ] Rate limiting per tenant configured

### Error Handling
- [ ] Errors don't leak cross-tenant information
- [ ] Generic messages for not found errors
- [ ] Audit logs include tenant context
- [ ] Error codes follow hierarchy

### Performance
- [ ] Tenant-specific indexes created
- [ ] Caching strategy per tenant
- [ ] Pagination implemented
- [ ] Bulk operations tenant-scoped

### Testing
- [ ] Cross-tenant isolation tests added
- [ ] Tenant mismatch tests added
- [ ] Performance isolation tests added
- [ ] Security tests cover tenant boundaries

---

**File Size**: 18 KB  
**Last Updated**: 2024-01-20  
**Related**: AUTH_API_EXAMPLE.md, API_VALIDATION_CHECKLIST.md
