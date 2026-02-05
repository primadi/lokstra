# Multi-Tenant Database Schema Patterns

**Purpose**: Comprehensive patterns for designing PostgreSQL schemas in multi-tenant systems with strict data isolation, performance, and scalability.

**Context**: Multi-tenant database design requires careful consideration of isolation, performance, security, and cost. This guide covers proven patterns and anti-patterns.

---

## Table of Contents

1. [Multi-Tenancy Approaches](#multi-tenancy-approaches)
2. [Tenant Isolation Patterns](#tenant-isolation-patterns)
3. [Composite Foreign Key Pattern](#composite-foreign-key-pattern)
4. [Indexing Strategies](#indexing-strategies)
5. [Row-Level Security (RLS)](#row-level-security-rls)
6. [Partitioning for Scale](#partitioning-for-scale)
7. [Data Types & Constraints](#data-types--constraints)
8. [Audit & Soft Delete](#audit--soft-delete)

---

## Multi-Tenancy Approaches

### Approach 1: Shared Schema with tenant_id (Recommended)

**Description**: Single database, single schema, all tenants share tables with tenant_id column

```sql
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← Tenant discriminator
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    -- ... other fields
    
    CONSTRAINT uq_patients_tenant_email UNIQUE(tenant_id, email)
);

CREATE INDEX idx_patients_tenant ON patients(tenant_id);
```

**Advantages**:
- ✅ Cost-effective (shared resources)
- ✅ Easy to scale horizontally
- ✅ Simple backup/restore
- ✅ Cross-tenant analytics possible
- ✅ Lower infrastructure complexity

**Disadvantages**:
- ❌ Requires strict app-level filtering
- ❌ Risk of data leakage if bugs
- ❌ Hard limits on tenant size
- ❌ Noisy neighbor problem

**When to Use**:
- SaaS with many small tenants (100-10,000+)
- Similar workload patterns across tenants
- Cost optimization priority
- Standard features for all tenants

**Security Requirements**:
- MUST filter by tenant_id in ALL queries
- MUST use RLS (Row-Level Security) as defense-in-depth
- MUST include tenant_id in foreign keys
- MUST audit all data access

### Approach 2: Shared Database, Separate Schemas

**Description**: Single database, one schema per tenant

```sql
-- Tenant 1
CREATE SCHEMA tenant_clinic_001;
CREATE TABLE tenant_clinic_001.patients ( ... );

-- Tenant 2
CREATE SCHEMA tenant_clinic_002;
CREATE TABLE tenant_clinic_002.patients ( ... );
```

**Advantages**:
- ✅ Better isolation than shared schema
- ✅ Easier to restore single tenant
- ✅ Can customize schema per tenant
- ✅ Tenant-specific optimizations

**Disadvantages**:
- ❌ Schema proliferation (PostgreSQL limit: ~9,900 schemas)
- ❌ Complex migrations (must update all schemas)
- ❌ Cross-tenant queries difficult
- ❌ More complex monitoring

**When to Use**:
- Medium number of tenants (10-1,000)
- Custom features per tenant
- Regulatory isolation requirements
- Tenant-specific data retention

### Approach 3: Separate Databases

**Description**: One database per tenant

```sql
-- Create databases
CREATE DATABASE tenant_clinic_001;
CREATE DATABASE tenant_clinic_002;
```

**Advantages**:
- ✅ Complete isolation
- ✅ Easy to backup/restore individual tenant
- ✅ Tenant-specific tuning possible
- ✅ No risk of cross-tenant data leakage

**Disadvantages**:
- ❌ High infrastructure cost
- ❌ Complex migrations (update all databases)
- ❌ Connection pool exhaustion
- ❌ No cross-tenant analytics
- ❌ Database limit (PostgreSQL default: 100 connections per DB)

**When to Use**:
- Few large tenants (1-100)
- Enterprise customers
- Strict compliance requirements (HIPAA, GDPR)
- Different SLAs per tenant

---

## Tenant Isolation Patterns

### Pattern 1: tenant_id Column (Mandatory)

**Rule**: ALL tables MUST have tenant_id column

```sql
-- ✅ CORRECT
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← REQUIRED
    patient_id TEXT NOT NULL,
    doctor_id TEXT NOT NULL,
    -- ... fields
);

-- ❌ WRONG: Missing tenant_id
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL,
    doctor_id TEXT NOT NULL
);
```

**Exceptions** (tables without tenant_id):
- Global lookup tables (countries, currencies, timezones)
- System configuration tables
- Permission definitions (reused across tenants)

### Pattern 2: Query-Level Filtering (Always)

**Rule**: EVERY query MUST filter by tenant_id

```sql
-- ✅ CORRECT: With tenant filter
SELECT * FROM patients 
WHERE tenant_id = $1 AND id = $2;

-- ✅ CORRECT: Even for INSERT
INSERT INTO patients (id, tenant_id, name, email)
VALUES ($1, $2, $3, $4);

-- ✅ CORRECT: Even for UPDATE
UPDATE patients 
SET name = $1, updated_at = NOW()
WHERE tenant_id = $2 AND id = $3;

-- ✅ CORRECT: Even for DELETE
DELETE FROM patients 
WHERE tenant_id = $1 AND id = $2;

-- ❌ WRONG: Missing tenant filter (SECURITY VULNERABILITY!)
SELECT * FROM patients WHERE id = $1;
UPDATE patients SET name = $1 WHERE id = $2;
DELETE FROM patients WHERE id = $1;
```

### Pattern 3: Unique Constraints with tenant_id

**Rule**: Include tenant_id in all unique constraints

```sql
-- ✅ CORRECT: tenant_id + email unique
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
    -- ...
    
    CONSTRAINT uq_users_tenant_email UNIQUE(tenant_id, email)
);

-- This allows same email in different tenants:
-- tenant_001: john@example.com ✅
-- tenant_002: john@example.com ✅

-- ❌ WRONG: Email unique across all tenants
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE  -- ← Bug: prevents email reuse
);
```

---

## Composite Foreign Key Pattern

### Pattern: Include tenant_id in Foreign Keys

**Rule**: All foreign keys MUST include tenant_id for data integrity

```sql
-- Parent table
CREATE TABLE clinics (
    id TEXT,
    tenant_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    -- ...
    
    PRIMARY KEY (tenant_id, id),  -- ← Composite PK
    CONSTRAINT uq_clinics_tenant_id UNIQUE(tenant_id, id)
);

-- Child table with composite FK
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    clinic_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    -- ...
    
    -- Composite FK prevents cross-tenant references
    CONSTRAINT fk_patients_clinic 
        FOREIGN KEY (tenant_id, clinic_id) 
        REFERENCES clinics(tenant_id, id)
        ON DELETE RESTRICT
);

-- This INSERT will FAIL if tenant_id doesn't match:
INSERT INTO patients (id, tenant_id, clinic_id, name)
VALUES ('pat_123', 'tenant_002', 'clinic_from_tenant_001', 'John');
-- ERROR: foreign key constraint violation
```

**Benefits**:
- ✅ Prevents cross-tenant data corruption
- ✅ Database-enforced integrity
- ✅ Catches app bugs early

**Implementation for Self-Referencing Tables**:

```sql
CREATE TABLE comments (
    id TEXT,
    tenant_id TEXT NOT NULL,
    parent_comment_id TEXT,
    content TEXT NOT NULL,
    -- ...
    
    PRIMARY KEY (tenant_id, id),
    
    -- Self-referencing composite FK
    CONSTRAINT fk_comments_parent 
        FOREIGN KEY (tenant_id, parent_comment_id) 
        REFERENCES comments(tenant_id, id)
        ON DELETE CASCADE
);
```

---

## Indexing Strategies

### Rule 1: tenant_id MUST be First Column in Composite Indexes

```sql
-- ✅ CORRECT: tenant_id first
CREATE INDEX idx_patients_tenant_name 
    ON patients(tenant_id, name);

CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date, status);

-- ❌ WRONG: tenant_id not first
CREATE INDEX idx_patients_name_tenant 
    ON patients(name, tenant_id);  -- ← Won't be used efficiently
```

**Why**: PostgreSQL uses leftmost prefix matching. Index is only used if tenant_id is in WHERE clause.

### Rule 2: Partial Indexes for Active Records

```sql
-- Exclude soft-deleted records from index
CREATE INDEX idx_patients_tenant_email_active 
    ON patients(tenant_id, email)
    WHERE deleted_at IS NULL;

-- Only index active appointments
CREATE INDEX idx_appointments_tenant_date_active 
    ON appointments(tenant_id, appointment_date)
    WHERE cancelled_at IS NULL AND deleted_at IS NULL;
```

**Benefits**:
- Smaller index size
- Faster queries
- Reduced maintenance cost

### Rule 3: Covering Indexes for Common Queries

```sql
-- Covering index (includes all columns needed)
CREATE INDEX idx_patients_tenant_email_covering 
    ON patients(tenant_id, email)
    INCLUDE (id, full_name, phone, status);

-- Query can use index-only scan (faster)
SELECT id, full_name, phone, status
FROM patients
WHERE tenant_id = 'tenant_001' AND email = 'john@example.com';
```

### Index Patterns Summary

```sql
-- 1. Tenant isolation (always needed)
CREATE INDEX idx_table_tenant ON table_name(tenant_id);

-- 2. Tenant + filter column
CREATE INDEX idx_table_tenant_status ON table_name(tenant_id, status);

-- 3. Tenant + sort column
CREATE INDEX idx_table_tenant_created ON table_name(tenant_id, created_at DESC);

-- 4. Tenant + filter + sort (composite)
CREATE INDEX idx_table_tenant_status_created 
    ON table_name(tenant_id, status, created_at DESC);

-- 5. Unique constraint with tenant
CREATE UNIQUE INDEX idx_table_tenant_email 
    ON table_name(tenant_id, email) WHERE deleted_at IS NULL;

-- 6. Foreign key lookup
CREATE INDEX idx_table_tenant_fk 
    ON table_name(tenant_id, foreign_key_id);

-- 7. JSONB search (GIN index)
CREATE INDEX idx_table_metadata 
    ON table_name USING GIN(metadata);

-- 8. Full-text search
CREATE INDEX idx_table_search 
    ON table_name USING GIN(to_tsvector('english', name || ' ' || description));
```

---

## Row-Level Security (RLS)

### Pattern: Enable RLS for Defense-in-Depth

**Purpose**: Even if app forgets tenant_id filter, database enforces isolation

```sql
-- Step 1: Enable RLS on table
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

-- Step 2: Create policy
CREATE POLICY tenant_isolation_policy ON patients
    FOR ALL  -- Applies to SELECT, INSERT, UPDATE, DELETE
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Step 3: Set tenant context in application
BEGIN;
SET LOCAL app.current_tenant_id = 'tenant_clinic_001';

-- Now all queries are automatically filtered
SELECT * FROM patients WHERE id = 'pat_123';
-- Automatically becomes:
-- SELECT * FROM patients WHERE id = 'pat_123' AND tenant_id = 'tenant_clinic_001'

COMMIT;
```

**Go Implementation**:

```go
func (r *PatientRepo) GetByID(ctx context.Context, tenantID, patientID string) (*Patient, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()
    
    // Set tenant context
    _, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
    if err != nil {
        return nil, err
    }
    
    // Query (RLS automatically filters by tenant_id)
    var patient Patient
    err = tx.QueryRowContext(ctx, 
        "SELECT id, tenant_id, name, email FROM patients WHERE id = $1", 
        patientID,
    ).Scan(&patient.ID, &patient.TenantID, &patient.Name, &patient.Email)
    
    if err != nil {
        return nil, err
    }
    
    return &patient, tx.Commit()
}
```

**Advanced Policies**:

```sql
-- Read-only policy for specific role
CREATE POLICY readonly_user_policy ON patients
    FOR SELECT
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)::TEXT
        AND current_setting('app.user_role', true) = 'readonly'
    );

-- Owner-only policy
CREATE POLICY owner_only_policy ON medical_records
    FOR ALL
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)::TEXT
        AND created_by = current_setting('app.current_user_id', true)::TEXT
    );
```

**Pros**:
- ✅ Defense against app bugs
- ✅ Additional security layer
- ✅ Audit compliance

**Cons**:
- ❌ Performance overhead (~5-10%)
- ❌ Complex debugging
- ❌ Requires session variable management

---

## Partitioning for Scale

### Pattern 1: Range Partitioning by Date

**Use Case**: Time-series data (audit logs, login attempts)

```sql
-- Parent table
CREATE TABLE audit_logs (
    id TEXT,
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    -- ... other fields
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE audit_logs_2026_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE audit_logs_2026_03 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Indexes on each partition
CREATE INDEX idx_audit_logs_2026_01_tenant 
    ON audit_logs_2026_01(tenant_id, created_at DESC);
```

**Benefits**:
- Query only scans relevant partition
- Easy to drop old data (DROP TABLE audit_logs_2025_01)
- Vacuum/analyze only affected partitions

**Automatic Partition Creation** (pg_partman extension):

```sql
CREATE EXTENSION pg_partman;

SELECT partman.create_parent(
    'public.audit_logs',
    'created_at',
    'native',
    'monthly'
);
```

### Pattern 2: List Partitioning by tenant_id

**Use Case**: Large tenants with dedicated hardware

```sql
-- Parent table
CREATE TABLE patient_records (
    id TEXT,
    tenant_id TEXT NOT NULL,
    -- ... fields
) PARTITION BY LIST (tenant_id);

-- Dedicated partition for large tenant
CREATE TABLE patient_records_tenant_enterprise_001 PARTITION OF patient_records
    FOR VALUES IN ('tenant_enterprise_001');

-- Default partition for smaller tenants
CREATE TABLE patient_records_default PARTITION OF patient_records
    DEFAULT;
```

**When to Use**:
- 1-2 large tenants (80% of data)
- Tenant-specific performance tuning
- Isolation for enterprise customers

---

## Data Types & Constraints

### ID Generation Patterns

**Pattern 1: ULID (Recommended)**

```sql
-- Install pgulid extension
CREATE EXTENSION pgulid;

CREATE TABLE patients (
    id TEXT PRIMARY KEY DEFAULT generate_ulid()::TEXT,
    -- ...
);

-- Example ULID: 01HQT1K8N9M5QFKBPJWX7YZ123
-- Sortable, URL-safe, globally unique
```

**Pattern 2: UUID v7 (Time-ordered)**

```sql
CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- ...
);
```

**Pattern 3: Custom Prefix Format**

```sql
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    -- ...
);

-- Application generates: pat_01HQT1K8N9M5QFKBPJWX7YZ123
-- Prefix helps identify entity type
```

### Timestamps

**Always use TIMESTAMPTZ**:

```sql
CREATE TABLE table_name (
    -- ✅ CORRECT: Timezone-aware
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- ❌ WRONG: No timezone
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### JSONB for Flexible Data

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    
    -- Structured fields
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    
    -- Flexible metadata
    metadata JSONB DEFAULT '{}',  -- {licenseNumber, specialization, etc}
    preferences JSONB DEFAULT '{}'  -- {theme, language, notifications}
);

-- GIN index for JSONB queries
CREATE INDEX idx_users_metadata ON users USING GIN(metadata);

-- Query specific JSON key
SELECT * FROM users 
WHERE tenant_id = 'tenant_001' 
  AND metadata->>'licenseNumber' = 'MD12345';
```

### Check Constraints

```sql
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    appointment_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    
    -- Enum validation
    CONSTRAINT chk_appointments_status 
        CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled')),
    
    -- Date validation
    CONSTRAINT chk_appointments_future_date 
        CHECK (appointment_date >= CURRENT_DATE),
    
    -- Time range validation
    CONSTRAINT chk_appointments_time_range 
        CHECK (end_time > start_time),
    
    -- Duration validation (max 4 hours)
    CONSTRAINT chk_appointments_duration 
        CHECK (end_time - start_time <= INTERVAL '4 hours')
);
```

---

## Audit & Soft Delete

### Pattern 1: Audit Columns

**Standard columns for all tables**:

```sql
CREATE TABLE table_name (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    
    -- Business fields
    -- ...
    
    -- Audit columns
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,  -- User ID who created
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT,  -- User ID who last updated
    deleted_at TIMESTAMPTZ,  -- Soft delete timestamp
    deleted_by TEXT  -- User ID who deleted
);

-- Auto-update updated_at trigger
CREATE TRIGGER trg_table_updated_at
    BEFORE UPDATE ON table_name
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Pattern 2: Soft Delete

```sql
-- Soft delete (set deleted_at)
UPDATE patients 
SET deleted_at = NOW(), deleted_by = 'usr_123'
WHERE tenant_id = 'tenant_001' AND id = 'pat_456';

-- Queries exclude soft-deleted by default
SELECT * FROM patients 
WHERE tenant_id = 'tenant_001' 
  AND deleted_at IS NULL;

-- Indexes exclude soft-deleted
CREATE INDEX idx_patients_tenant_email 
    ON patients(tenant_id, email)
    WHERE deleted_at IS NULL;

-- Restore soft-deleted record
UPDATE patients 
SET deleted_at = NULL, deleted_by = NULL
WHERE tenant_id = 'tenant_001' AND id = 'pat_456';

-- Hard delete (permanent)
DELETE FROM patients 
WHERE tenant_id = 'tenant_001' 
  AND id = 'pat_456' 
  AND deleted_at IS NOT NULL 
  AND deleted_at < NOW() - INTERVAL '90 days';
```

### Pattern 3: Audit Trail Table

```sql
CREATE TABLE audit_trail (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    
    -- What changed
    table_name VARCHAR(100) NOT NULL,
    record_id TEXT NOT NULL,
    action VARCHAR(20) NOT NULL,  -- INSERT, UPDATE, DELETE
    
    -- Who changed
    user_id TEXT NOT NULL,
    
    -- When
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- What was changed
    old_values JSONB,  -- Before update
    new_values JSONB,  -- After update
    changed_fields TEXT[],  -- Array of field names
    
    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id TEXT
);

-- Trigger function
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_trail (id, tenant_id, table_name, record_id, action, 
                              user_id, old_values, new_values)
    VALUES (
        'audit_' || gen_random_uuid()::TEXT,
        NEW.tenant_id,
        TG_TABLE_NAME,
        NEW.id,
        TG_OP,
        current_setting('app.current_user_id', true),
        row_to_json(OLD),
        row_to_json(NEW)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to table
CREATE TRIGGER trg_patients_audit
    AFTER INSERT OR UPDATE OR DELETE ON patients
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();
```

---

## Anti-Patterns

### ❌ Anti-Pattern 1: Optional tenant_id Filtering

```go
// WRONG: tenant_id filtering optional
func GetPatients(ctx context.Context, tenantID string) ([]Patient, error) {
    query := "SELECT * FROM patients WHERE 1=1"
    if tenantID != "" {  // ← BUG: Should ALWAYS be required
        query += " AND tenant_id = $1"
    }
    // ...
}
```

**Fix**: Always require tenant_id

```go
func GetPatients(ctx context.Context, tenantID string) ([]Patient, error) {
    if tenantID == "" {
        return nil, errors.New("tenant_id is required")
    }
    query := "SELECT * FROM patients WHERE tenant_id = $1"
    // ...
}
```

### ❌ Anti-Pattern 2: tenant_id from Request Body

```go
// WRONG: Using tenant_id from client request
func CreatePatient(req *CreatePatientRequest) error {
    patient := &Patient{
        TenantID: req.TenantID,  // ← SECURITY BUG: Client can spoof this
        Name: req.Name,
    }
    return repo.Save(patient)
}
```

**Fix**: Extract tenant_id from JWT

```go
func CreatePatient(ctx *request.Context, req *CreatePatientRequest) error {
    tenantID := ctx.Auth.TenantID()  // ← From authenticated JWT
    patient := &Patient{
        TenantID: tenantID,
        Name: req.Name,
    }
    return repo.Save(patient)
}
```

### ❌ Anti-Pattern 3: Shared Cache Keys

```go
// WRONG: Cache key without tenant_id
cacheKey := fmt.Sprintf("patient:%s", patientID)
```

**Fix**: Include tenant_id in cache key

```go
cacheKey := fmt.Sprintf("tenant:%s:patient:%s", tenantID, patientID)
```

---

## Migration Best Practices

### 1. Always Add tenant_id to New Tables

```sql
-- Migration UP
CREATE TABLE new_table (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← REQUIRED
    -- ... fields
);

CREATE INDEX idx_new_table_tenant ON new_table(tenant_id);
```

### 2. Adding tenant_id to Existing Table

```sql
-- Step 1: Add column (nullable initially)
ALTER TABLE existing_table 
ADD COLUMN tenant_id TEXT;

-- Step 2: Backfill data (may need manual intervention)
UPDATE existing_table 
SET tenant_id = (SELECT tenant_id FROM users WHERE users.id = existing_table.user_id);

-- Step 3: Make NOT NULL
ALTER TABLE existing_table 
ALTER COLUMN tenant_id SET NOT NULL;

-- Step 4: Add index
CREATE INDEX idx_existing_table_tenant ON existing_table(tenant_id);

-- Step 5: Update unique constraints
ALTER TABLE existing_table 
DROP CONSTRAINT IF EXISTS uq_existing_table_email;

ALTER TABLE existing_table 
ADD CONSTRAINT uq_existing_table_tenant_email UNIQUE(tenant_id, email);
```

---

## Checklist

### Design Phase
- [ ] All tables include tenant_id (except global lookup tables)
- [ ] All foreign keys are composite (include tenant_id)
- [ ] All unique constraints include tenant_id
- [ ] All indexes have tenant_id as first column
- [ ] Soft delete columns (deleted_at) included
- [ ] Audit columns (created_at, updated_at, created_by) included
- [ ] Check constraints for enum fields
- [ ] TIMESTAMPTZ (not TIMESTAMP) for all timestamps
- [ ] JSONB for flexible/extensible data

### Security Phase
- [ ] RLS policies defined for all tenant tables
- [ ] Composite foreign keys prevent cross-tenant references
- [ ] No queries without tenant_id filter
- [ ] Audit trail for sensitive operations
- [ ] Password columns use TEXT (for bcrypt hash)
- [ ] API tokens hashed (SHA-256), never plain text

### Performance Phase
- [ ] Indexes for all foreign keys
- [ ] Partial indexes for active records (WHERE deleted_at IS NULL)
- [ ] Covering indexes for common queries
- [ ] GIN indexes for JSONB columns
- [ ] Partitioning strategy for large tables (>10M rows)
- [ ] Cleanup jobs for expired/old data

### Migration Phase
- [ ] UP migration adds table/columns
- [ ] DOWN migration reverts changes
- [ ] Data backfill for existing records
- [ ] Index creation (CONCURRENTLY in production)
- [ ] No breaking changes (backwards compatible)

---

**File Size**: 22 KB  
**Last Updated**: 2024-01-20  
**Related**: AUTH_SCHEMA_EXAMPLE.md, SCHEMA_VALIDATION_CHECKLIST.md
