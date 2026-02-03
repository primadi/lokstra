# PostgreSQL Schema Validation Checklist for Lokstra

**Purpose**: Quality assurance checklist to validate database schemas before implementation in Lokstra projects with multi-tenant architecture.

**Usage**: Review each generated schema against this checklist. Minimum passing score: **85/100 points**.

---

## Table of Contents

1. [Multi-Tenancy Requirements](#multi-tenancy-requirements) (30 points)
2. [Data Types & Constraints](#data-types--constraints) (15 points)
3. [Indexing Strategy](#indexing-strategy) (20 points)
4. [Security & Access Control](#security--access-control) (15 points)
5. [Audit & Compliance](#audit--compliance) (10 points)
6. [Performance Optimization](#performance-optimization) (10 points)
7. [Scoring Guidelines](#scoring-guidelines)

---

## Multi-Tenancy Requirements (30 points)

### 1. Tenant Isolation (10 points)

**Requirement**: Every business table MUST include `tenant_id` column

- [ ] **[5 pts]** All tables have `tenant_id TEXT NOT NULL` column
- [ ] **[5 pts]** Global lookup tables (countries, timezones) explicitly documented as exceptions

**Validation Query**:
```sql
-- Find tables missing tenant_id
SELECT 
    schemaname, 
    tablename
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
  AND tablename NOT IN ('schema_migrations', 'countries', 'timezones')
  AND NOT EXISTS (
      SELECT 1 
      FROM information_schema.columns c
      WHERE c.table_schema = pg_tables.schemaname
        AND c.table_name = pg_tables.tablename
        AND c.column_name = 'tenant_id'
  );
```

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    -- ...
);

-- ❌ WRONG: Missing tenant_id
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL
);
```

---

### 2. Composite Foreign Keys (10 points)

**Requirement**: All foreign keys MUST include `tenant_id` for cross-tenant protection

- [ ] **[5 pts]** All foreign keys are composite: `(tenant_id, reference_id)`
- [ ] **[3 pts]** Parent tables have unique constraint: `UNIQUE(tenant_id, id)`
- [ ] **[2 pts]** ON DELETE/ON UPDATE policies explicitly defined

**Validation Query**:
```sql
-- Find foreign keys missing tenant_id
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND NOT EXISTS (
      SELECT 1 
      FROM information_schema.key_column_usage kcu2
      WHERE kcu2.constraint_name = tc.constraint_name
        AND kcu2.column_name = 'tenant_id'
  );
```

**Example**:
```sql
-- ✅ CORRECT: Composite FK with tenant_id
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    patient_id TEXT NOT NULL,
    CONSTRAINT fk_appointments_patient 
        FOREIGN KEY (tenant_id, patient_id) 
        REFERENCES patients(tenant_id, id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);

-- ❌ WRONG: FK without tenant_id
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    patient_id TEXT NOT NULL,
    CONSTRAINT fk_appointments_patient 
        FOREIGN KEY (patient_id) 
        REFERENCES patients(id)  -- ← Missing tenant_id
);
```

---

### 3. Unique Constraints with tenant_id (10 points)

**Requirement**: All unique constraints MUST scope to tenant

- [ ] **[5 pts]** Unique constraints include `tenant_id` as first column
- [ ] **[3 pts]** Soft-deleted records excluded: `WHERE deleted_at IS NULL`
- [ ] **[2 pts]** No global unique constraints (except system tables)

**Validation Query**:
```sql
-- Find unique constraints without tenant_id
SELECT
    tc.constraint_name,
    tc.table_name,
    array_agg(kcu.column_name ORDER BY kcu.ordinal_position) AS columns
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type = 'UNIQUE'
  AND tc.table_schema = 'public'
GROUP BY tc.constraint_name, tc.table_name
HAVING NOT ('tenant_id' = ANY(array_agg(kcu.column_name)));
```

**Example**:
```sql
-- ✅ CORRECT: Scoped to tenant
CREATE UNIQUE INDEX uq_patients_tenant_email 
    ON patients(tenant_id, email) 
    WHERE deleted_at IS NULL;

-- ❌ WRONG: Global unique (prevents email reuse)
CREATE UNIQUE INDEX uq_patients_email 
    ON patients(email);  -- ← Wrong: email unique globally
```

---

## Data Types & Constraints (15 points)

### 4. ID Generation (5 points)

**Requirement**: Use ULID or UUID v7 for sortable IDs

- [ ] **[3 pts]** ID columns are `TEXT` (not INTEGER/BIGINT)
- [ ] **[2 pts]** Documentation specifies ID generation strategy (ULID/UUID v7)

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE users (
    id TEXT PRIMARY KEY,  -- ULID or UUID v7
    tenant_id TEXT NOT NULL,
    -- ...
);

-- ❌ WRONG: Auto-incrementing integer
CREATE TABLE users (
    id SERIAL PRIMARY KEY,  -- ← Wrong: not globally unique
    tenant_id TEXT NOT NULL
);
```

---

### 5. Timestamp Handling (5 points)

**Requirement**: Use timezone-aware timestamps

- [ ] **[3 pts]** All timestamp columns use `TIMESTAMPTZ` (not TIMESTAMP)
- [ ] **[2 pts]** Default values use `NOW()` or `CURRENT_TIMESTAMP`

**Validation Query**:
```sql
-- Find TIMESTAMP columns (should be TIMESTAMPTZ)
SELECT 
    table_schema,
    table_name,
    column_name,
    data_type
FROM information_schema.columns
WHERE data_type = 'timestamp without time zone'
  AND table_schema NOT IN ('pg_catalog', 'information_schema');
```

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    appointment_date TIMESTAMPTZ NOT NULL,  -- ← Timezone-aware
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ❌ WRONG: No timezone
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    appointment_date TIMESTAMP NOT NULL,  -- ← Wrong: no timezone
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

### 6. Check Constraints (5 points)

**Requirement**: Enforce data integrity at database level

- [ ] **[3 pts]** Enum-like columns have CHECK constraints
- [ ] **[2 pts]** Business rules enforced via CHECK constraints (where applicable)

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    duration_minutes INTEGER NOT NULL,
    
    -- Check constraints
    CONSTRAINT chk_appointments_status 
        CHECK (status IN ('scheduled', 'confirmed', 'completed', 'cancelled', 'no_show')),
    CONSTRAINT chk_appointments_duration 
        CHECK (duration_minutes BETWEEN 5 AND 480)
);

-- ❌ WRONG: No constraints
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    status VARCHAR(20),  -- ← No constraint, any value allowed
    duration_minutes INTEGER  -- ← No constraint, negative values allowed
);
```

---

## Indexing Strategy (20 points)

### 7. Primary Indexes (5 points)

**Requirement**: Optimize for multi-tenant queries

- [ ] **[3 pts]** All tables have primary key index
- [ ] **[2 pts]** Composite primary key consideration documented (if needed)

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    -- ...
    CONSTRAINT uq_patients_tenant_id UNIQUE(tenant_id, id)  -- For composite FKs
);
```

---

### 8. Tenant Indexes (5 points)

**Requirement**: Efficient tenant-scoped queries

- [ ] **[3 pts]** Index on `tenant_id` for every table
- [ ] **[2 pts]** Partial index excludes soft-deleted: `WHERE deleted_at IS NULL`

**Validation Query**:
```sql
-- Find tables with tenant_id but no index
SELECT 
    t.schemaname,
    t.tablename
FROM pg_tables t
WHERE EXISTS (
    SELECT 1 FROM information_schema.columns c
    WHERE c.table_schema = t.schemaname
      AND c.table_name = t.tablename
      AND c.column_name = 'tenant_id'
)
AND NOT EXISTS (
    SELECT 1 FROM pg_indexes i
    WHERE i.schemaname = t.schemaname
      AND i.tablename = t.tablename
      AND i.indexdef LIKE '%tenant_id%'
);
```

**Example**:
```sql
-- ✅ CORRECT
CREATE INDEX idx_patients_tenant 
    ON patients(tenant_id) 
    WHERE deleted_at IS NULL;

-- ❌ WRONG: No index on tenant_id
-- Missing: CREATE INDEX idx_patients_tenant ON patients(tenant_id);
```

---

### 9. Composite Indexes (5 points)

**Requirement**: Optimize common query patterns

- [ ] **[3 pts]** Composite indexes have `tenant_id` as first column
- [ ] **[2 pts]** Covering indexes for hot queries (SELECT without table lookup)

**Example**:
```sql
-- ✅ CORRECT: tenant_id first
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date DESC);

-- Covering index for: SELECT id, patient_id, status WHERE tenant_id = ? AND appointment_date = ?
CREATE INDEX idx_appointments_tenant_date_covering 
    ON appointments(tenant_id, appointment_date) 
    INCLUDE (id, patient_id, status);

-- ❌ WRONG: tenant_id not first
CREATE INDEX idx_appointments_date_tenant 
    ON appointments(appointment_date, tenant_id);  -- ← tenant_id should be first
```

---

### 10. Specialized Indexes (5 points)

**Requirement**: Use appropriate index types

- [ ] **[3 pts]** GIN indexes for JSONB columns
- [ ] **[2 pts]** Text search indexes (GIN with tsvector) if needed

**Example**:
```sql
-- ✅ CORRECT: GIN for JSONB
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    metadata JSONB
);

CREATE INDEX idx_patients_metadata_gin 
    ON patients USING GIN(metadata);

-- Full-text search
CREATE INDEX idx_patients_fulltext 
    ON patients USING GIN(to_tsvector('english', full_name));
```

---

## Security & Access Control (15 points)

### 11. Row-Level Security (10 points)

**Requirement**: Database-enforced tenant isolation

- [ ] **[5 pts]** RLS enabled on all tenant tables: `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
- [ ] **[3 pts]** RLS policies defined for SELECT, INSERT, UPDATE, DELETE
- [ ] **[2 pts]** RLS uses `current_setting('app.tenant_id')` for filtering

**Validation Query**:
```sql
-- Find tables with tenant_id but RLS not enabled
SELECT 
    schemaname,
    tablename
FROM pg_tables t
WHERE EXISTS (
    SELECT 1 FROM information_schema.columns c
    WHERE c.table_schema = t.schemaname
      AND c.table_name = t.tablename
      AND c.column_name = 'tenant_id'
)
AND NOT EXISTS (
    SELECT 1 FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = t.schemaname
      AND c.relname = t.tablename
      AND c.relrowsecurity = true
);
```

**Example**:
```sql
-- ✅ CORRECT: RLS enabled with policies
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON patients
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

CREATE POLICY tenant_insert_policy ON patients
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- ❌ WRONG: No RLS
-- Missing: ALTER TABLE patients ENABLE ROW LEVEL SECURITY;
```

---

### 12. Sensitive Data (5 points)

**Requirement**: Protect sensitive information

- [ ] **[3 pts]** Passwords hashed (never plaintext) - documented in comments
- [ ] **[2 pts]** PII columns documented (for GDPR/compliance)

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,  -- PII: Email address
    password_hash TEXT NOT NULL,  -- Bcrypt hash (cost 12)
    phone VARCHAR(50),  -- PII: Phone number
    -- ...
);

COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash with cost factor 12 (never store plaintext)';
COMMENT ON COLUMN users.email IS 'PII: User email address (GDPR applicable)';
```

---

## Audit & Compliance (10 points)

### 13. Audit Columns (5 points)

**Requirement**: Track data lifecycle

- [ ] **[2 pts]** `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- [ ] **[2 pts]** `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- [ ] **[1 pt]** Trigger to auto-update `updated_at` on UPDATE

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger function (create once)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger
CREATE TRIGGER trg_patients_updated_at
    BEFORE UPDATE ON patients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

### 14. Soft Delete (5 points)

**Requirement**: Never hard-delete business data

- [ ] **[3 pts]** `deleted_at TIMESTAMPTZ` column for soft delete
- [ ] **[2 pts]** Indexes exclude deleted records: `WHERE deleted_at IS NULL`

**Example**:
```sql
-- ✅ CORRECT
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMPTZ,  -- NULL = active, NOT NULL = deleted
    -- ...
);

-- Index only active records
CREATE INDEX idx_patients_tenant_active 
    ON patients(tenant_id) 
    WHERE deleted_at IS NULL;

-- ❌ WRONG: Hard delete only
-- Missing: deleted_at column
```

---

## Performance Optimization (10 points)

### 15. Partitioning (5 points)

**Requirement**: Scale large tables efficiently

- [ ] **[3 pts]** Tables with >10M rows use partitioning (RANGE or LIST)
- [ ] **[2 pts]** Partitioning strategy documented in comments

**Example**:
```sql
-- ✅ CORRECT: Partition large audit log
CREATE TABLE audit_logs (
    id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- ...
) PARTITION BY RANGE (created_at);

-- Create partitions (monthly)
CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE audit_logs_2026_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

COMMENT ON TABLE audit_logs IS 'Partitioned by month for performance (10M+ rows expected)';
```

---

### 16. JSONB Usage (5 points)

**Requirement**: Use JSONB appropriately

- [ ] **[3 pts]** JSONB columns have GIN indexes
- [ ] **[2 pts]** Structured data uses proper columns (not JSONB)

**Example**:
```sql
-- ✅ CORRECT: Structured + metadata
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,  -- ← Structured column (not JSONB)
    email VARCHAR(255) NOT NULL,      -- ← Structured column
    metadata JSONB,  -- ← Optional/flexible data only
    -- ...
);

CREATE INDEX idx_patients_metadata_gin 
    ON patients USING GIN(metadata);

-- ❌ WRONG: Everything in JSONB
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    data JSONB  -- ← Wrong: full_name, email should be columns
);
```

---

## Scoring Guidelines

### Score Calculation

| Category | Max Points | Weight |
|----------|-----------|--------|
| Multi-Tenancy Requirements | 30 | Critical |
| Data Types & Constraints | 15 | High |
| Indexing Strategy | 20 | High |
| Security & Access Control | 15 | Critical |
| Audit & Compliance | 10 | Medium |
| Performance Optimization | 10 | Medium |
| **TOTAL** | **100** | |

### Passing Criteria

- **Minimum Score**: 85/100
- **Critical Requirements** (must have all):
  - ✅ All tables have `tenant_id` (10 pts)
  - ✅ Composite foreign keys with `tenant_id` (10 pts)
  - ✅ RLS enabled on tenant tables (10 pts)
  - ✅ Audit columns (created_at, updated_at) (5 pts)

### Grade Levels

| Score | Grade | Status |
|-------|-------|--------|
| 95-100 | A+ | Excellent - Production Ready |
| 90-94 | A | Very Good - Minor improvements |
| 85-89 | B+ | Good - Meets minimum requirements |
| 80-84 | B | Acceptable - Significant improvements needed |
| 75-79 | C | Poor - Major issues, revise |
| <75 | F | Fail - Does not meet standards |

---

## Validation Process

### Step 1: Automated Checks (30 minutes)

Run SQL validation queries to check:
- Missing `tenant_id` columns
- Foreign keys without `tenant_id`
- Unique constraints without `tenant_id`
- Tables without RLS enabled
- TIMESTAMP columns (should be TIMESTAMPTZ)
- Missing indexes on `tenant_id`

### Step 2: Manual Review (45 minutes)

Review schema for:
- Business logic correctness
- Appropriate data types
- Check constraints coverage
- Index efficiency for common queries
- Partitioning strategy for large tables
- Documentation completeness

### Step 3: Performance Testing (60 minutes)

Test with realistic data:
- Load 100K+ rows with multiple tenants
- Run common queries with EXPLAIN ANALYZE
- Verify tenant isolation (no cross-tenant leaks)
- Check index usage
- Measure query response times

### Step 4: Security Audit (30 minutes)

Verify:
- RLS policies work correctly
- Composite FKs prevent cross-tenant references
- No sensitive data in plaintext
- Audit trail captures all changes

---

## Common Issues & Fixes

### Issue 1: Low Score on Multi-Tenancy (< 20/30)

**Symptoms**:
- Missing `tenant_id` in some tables
- Foreign keys without `tenant_id`
- Global unique constraints

**Fix**:
```sql
-- Add tenant_id to all tables
ALTER TABLE appointments ADD COLUMN tenant_id TEXT NOT NULL;

-- Fix foreign keys
ALTER TABLE appointments DROP CONSTRAINT fk_patient;
ALTER TABLE appointments ADD CONSTRAINT fk_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patients(tenant_id, id);

-- Fix unique constraints
DROP INDEX uq_patients_email;
CREATE UNIQUE INDEX uq_patients_tenant_email 
    ON patients(tenant_id, email) 
    WHERE deleted_at IS NULL;
```

### Issue 2: Poor Indexing (< 15/20)

**Symptoms**:
- Missing index on `tenant_id`
- Composite indexes with wrong column order
- No partial indexes for soft delete

**Fix**:
```sql
-- Add tenant index
CREATE INDEX idx_patients_tenant 
    ON patients(tenant_id) 
    WHERE deleted_at IS NULL;

-- Fix composite index order
DROP INDEX idx_appointments_date_tenant;
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date DESC);
```

### Issue 3: Missing Security (< 10/15)

**Symptoms**:
- RLS not enabled
- No RLS policies
- Missing password hashing documentation

**Fix**:
```sql
-- Enable RLS
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY tenant_isolation_policy ON patients
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- Add documentation
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash with cost factor 12';
```

---

## Checklist Summary

Use this quick reference during schema review:

**Multi-Tenancy** (30 pts)
- [ ] tenant_id in all tables (10 pts)
- [ ] Composite foreign keys (10 pts)
- [ ] Scoped unique constraints (10 pts)

**Data Types** (15 pts)
- [ ] TEXT for IDs (5 pts)
- [ ] TIMESTAMPTZ for timestamps (5 pts)
- [ ] CHECK constraints for enums (5 pts)

**Indexes** (20 pts)
- [ ] Primary keys (5 pts)
- [ ] Tenant indexes with partial (5 pts)
- [ ] Composite with tenant_id first (5 pts)
- [ ] GIN for JSONB (5 pts)

**Security** (15 pts)
- [ ] RLS enabled + policies (10 pts)
- [ ] Password hashing documented (5 pts)

**Audit** (10 pts)
- [ ] created_at, updated_at, trigger (5 pts)
- [ ] deleted_at soft delete (5 pts)

**Performance** (10 pts)
- [ ] Partitioning for large tables (5 pts)
- [ ] JSONB used appropriately (5 pts)

---

**Minimum Passing Score**: 85/100  
**Critical Items**: Must have tenant_id, composite FKs, RLS, audit columns  
**Review Time**: ~2.5 hours (automated + manual + testing + audit)

---

**File Size**: 18 KB  
**Last Updated**: 2024-01-20  
**Related**: MULTI_TENANT_SCHEMA_PATTERNS.md, AUTH_SCHEMA_EXAMPLE.md, MIGRATION_PATTERNS.md
