---
name: lokstra-schema-design
description: Generate PostgreSQL database schemas for Lokstra modules. Creates tables, indexes, constraints, triggers, and migration files. Use after module requirements and API specs are approved to design data layer.
phase: design
order: 5
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
  database: postgresql
---

# Lokstra Schema Design

**Purpose**: Generate production-ready PostgreSQL database schemas for Lokstra modules with multi-tenant architecture, enforcing data integrity, security, and performance best practices.

---

## Table of Contents

1. [When to Use This Skill](#when-to-use-this-skill)
2. [Prerequisites](#prerequisites)
3. [Step-by-Step Process](#step-by-step-process)
4. [Multi-Tenant Requirements](#multi-tenant-requirements)
5. [Table Design Standards](#table-design-standards)
6. [Indexing Strategy](#indexing-strategy)
7. [Security & Access Control](#security--access-control)
8. [Migration Workflow](#migration-workflow)
9. [Quality Checklist](#quality-checklist)
10. [Examples](#examples)
11. [Reference Files](#reference-files)

---

## When to Use This Skill

### Trigger Conditions

Use this skill when:

- ✅ **Module requirements are approved** - Business logic is finalized
- ✅ **API specifications are complete** - Endpoints and data models defined
- ✅ **Ready to design data layer** - Need PostgreSQL table definitions
- ✅ **Creating migrations** - Need UP/DOWN SQL files
- ✅ **Defining relationships** - Need foreign keys and constraints

### Prerequisites

Before starting schema design, ensure you have:

1. **Approved Module Requirements** (`docs/modules/{module-name}/REQUIREMENTS.md`)
   - Functional requirements with acceptance criteria
   - Use cases and business rules
   - Data retention policies

2. **Approved API Specifications** (`docs/modules/{module-name}/API_SPEC.md`)
   - Request/response DTOs
   - Validation rules
   - Endpoint definitions

3. **Multi-Tenant Architecture Decision**
   - Shared schema (recommended for 100-10K tenants)
   - Separate schemas (for 10-1K tenants)
   - Separate databases (for 1-100 large tenants)

4. **Database Environment**
   - PostgreSQL 14+ installed
   - Migration tool configured (golang-migrate, Lokstra migration runner)

---

## Step-by-Step Process

### Step 1: Analyze Requirements (30 minutes)

**Goal**: Extract data entities and relationships from requirements

**Actions**:
1. Read module requirements document
2. Identify core entities (nouns in use cases)
3. Map entity relationships (one-to-many, many-to-many)
4. List required queries and access patterns
5. Identify data volume estimates

**Output**: Entity-Relationship diagram (text or tool)

**Example**:
```
Module: Patient Management
Entities:
- Patients (10K-100K records per tenant)
- Appointments (100K-1M records per tenant)
- Medical Records (500K-5M records per tenant)

Relationships:
- Patient 1:N Appointments
- Patient 1:N Medical Records
- Appointment N:1 Doctor (from User module)
```

---

### Step 2: Design Table Structures (60 minutes)

**Goal**: Define tables with all columns, types, and constraints

**Actions**:
1. Create table for each entity
2. Add `tenant_id TEXT NOT NULL` to every table (multi-tenant)
3. Choose ID strategy: ULID (recommended) or UUID v7
4. Define business columns with appropriate types
5. Add audit columns: `created_at`, `updated_at`, `deleted_at`
6. Add CHECK constraints for enum-like fields
7. Add unique constraints (scoped to tenant)

**Output**: SQL CREATE TABLE statements

**Template**:
```sql
-- Table: {module}.{entity_plural}
-- Purpose: {Brief description}
-- Estimated rows: {volume estimate}
-- Multi-tenant: Yes (tenant_id required on all queries)

CREATE TABLE {module}.{entity_plural} (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- ULID: Sortable, URL-safe, 26 chars
    
    -- Multi-Tenant (MANDATORY)
    tenant_id TEXT NOT NULL,
    
    -- Business Columns
    {column_name} {TYPE} {NOT NULL} {DEFAULT},
    
    -- Foreign Keys (composite with tenant_id)
    {reference_id} TEXT NOT NULL,
    
    -- Audit Columns
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- Soft delete (NULL = active)
    
    -- Check Constraints
    CONSTRAINT chk_{table}_{column} CHECK ({column} IN ('value1', 'value2')),
    
    -- Unique Constraints (scoped to tenant)
    CONSTRAINT uq_{table}_tenant_{column} UNIQUE(tenant_id, {column})
);
```

**See**: [AUTH_SCHEMA_EXAMPLE.md](references/AUTH_SCHEMA_EXAMPLE.md) for complete 10-table example

---

### Step 3: Define Indexes (30 minutes)

**Goal**: Optimize query performance for common access patterns

**Actions**:
1. Add index on `tenant_id` for every table
2. Create composite indexes with `tenant_id` as first column
3. Add partial indexes to exclude soft-deleted records
4. Create GIN indexes for JSONB columns
5. Add covering indexes for hot queries

**Index Priority**:
- **Critical**: `tenant_id` (every table)
- **High**: Foreign keys, status columns, date ranges
- **Medium**: Full-text search, JSONB paths
- **Low**: Rarely queried columns

**Template**:
```sql
-- Indexes for {table_name}

-- 1. Tenant isolation (partial index excludes deleted)
CREATE INDEX idx_{table}_tenant 
    ON {module}.{table}(tenant_id) 
    WHERE deleted_at IS NULL;

-- 2. Composite index for common query (tenant_id FIRST)
CREATE INDEX idx_{table}_tenant_{column} 
    ON {module}.{table}(tenant_id, {column} DESC) 
    WHERE deleted_at IS NULL;

-- 3. Foreign key lookup
CREATE INDEX idx_{table}_tenant_{fk} 
    ON {module}.{table}(tenant_id, {fk_column});

-- 4. GIN index for JSONB
CREATE INDEX idx_{table}_metadata_gin 
    ON {module}.{table} USING GIN(metadata_jsonb);

-- 5. Covering index (avoid table lookup)
CREATE INDEX idx_{table}_tenant_{col}_covering 
    ON {module}.{table}(tenant_id, {column}) 
    INCLUDE (id, status, created_at);
```

**See**: [MULTI_TENANT_SCHEMA_PATTERNS.md](references/MULTI_TENANT_SCHEMA_PATTERNS.md) - Section on Indexing Strategies

---

### Step 4: Add Security (30 minutes)

**Goal**: Enforce tenant isolation at database level

**Actions**:
1. Enable Row-Level Security (RLS) on all tenant tables
2. Create RLS policies for SELECT, INSERT, UPDATE, DELETE
3. Define composite foreign keys with `tenant_id`
4. Add comments for sensitive columns (PII, passwords)

**Template**:
```sql
-- Row-Level Security for {table_name}

-- 1. Enable RLS
ALTER TABLE {module}.{table} ENABLE ROW LEVEL SECURITY;

-- 2. SELECT policy (filter by tenant)
CREATE POLICY tenant_isolation_select ON {module}.{table}
    FOR SELECT
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 3. INSERT policy (enforce tenant)
CREATE POLICY tenant_isolation_insert ON {module}.{table}
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 4. UPDATE policy (own tenant only)
CREATE POLICY tenant_isolation_update ON {module}.{table}
    FOR UPDATE
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 5. DELETE policy (own tenant only)
CREATE POLICY tenant_isolation_delete ON {module}.{table}
    FOR DELETE
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- Composite Foreign Key (prevents cross-tenant references)
ALTER TABLE {module}.{child_table}
ADD CONSTRAINT fk_{child}_tenant_{parent} 
    FOREIGN KEY (tenant_id, {parent_id}) 
    REFERENCES {module}.{parent_table}(tenant_id, id)
    ON DELETE RESTRICT
    ON UPDATE CASCADE;
```

**See**: [MULTI_TENANT_SCHEMA_PATTERNS.md](references/MULTI_TENANT_SCHEMA_PATTERNS.md) - Section on Row-Level Security

---

### Step 5: Add Triggers & Functions (15 minutes)

**Goal**: Automate repetitive tasks (timestamps, audit logs)

**Actions**:
1. Create `update_updated_at_column()` function (once per database)
2. Attach trigger to each table for auto-updating `updated_at`
3. Create audit trigger for sensitive tables (optional)

**Template**:
```sql
-- Function: Auto-update updated_at column (create once)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Attach to each table
CREATE TRIGGER trg_{table}_updated_at
    BEFORE UPDATE ON {module}.{table}
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Optional: Audit trigger for sensitive tables
CREATE OR REPLACE FUNCTION audit_{table}_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO {module}.audit_logs (
            tenant_id, table_name, record_id, action, 
            old_values, new_values, changed_by
        ) VALUES (
            OLD.tenant_id, TG_TABLE_NAME, OLD.id, 'UPDATE',
            row_to_json(OLD), row_to_json(NEW),
            current_setting('app.user_id', true)::TEXT
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_{table}_audit
    AFTER UPDATE ON {module}.{table}
    FOR EACH ROW
    EXECUTE FUNCTION audit_{table}_changes();
```

**See**: [AUTH_SCHEMA_EXAMPLE.md](references/AUTH_SCHEMA_EXAMPLE.md) - Section on Triggers

---

### Step 6: Create Migration Files (30 minutes)

**Goal**: Version-controlled schema changes with rollback capability

**Actions**:
1. Create `migrations/{module}/` directory
2. Generate sequential UP migration files
3. Generate corresponding DOWN migration files
4. Include all tables, indexes, RLS policies, triggers
5. Test UP/DOWN on clean database

**Naming Convention**:
```
migrations/{module}/
├── 001_create_{table1}.up.sql
├── 001_create_{table1}.down.sql
├── 002_create_{table2}.up.sql
├── 002_create_{table2}.down.sql
└── ...
```

**UP Migration Template**:
```sql
-- Migration: Create {table} table
-- Module: {module-name}
-- Version: 1.0.0
-- Date: 2026-01-20
-- Author: {your-name}

BEGIN;

-- Create table
CREATE TABLE IF NOT EXISTS {module}.{table} (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    -- ... columns
);

-- Create indexes
CREATE INDEX idx_{table}_tenant ON {module}.{table}(tenant_id);
-- ... more indexes

-- Enable RLS
ALTER TABLE {module}.{table} ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY tenant_isolation ON {module}.{table}
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- Attach triggers
CREATE TRIGGER trg_{table}_updated_at
    BEFORE UPDATE ON {module}.{table}
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE {module}.{table} IS '{Description}';
COMMENT ON COLUMN {module}.{table}.tenant_id IS 'Multi-tenant isolation';

COMMIT;
```

**DOWN Migration Template**:
```sql
-- Migration: Rollback {table} table creation
-- Module: {module-name}

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_{table}_updated_at ON {module}.{table};

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation ON {module}.{table};

-- Drop table (CASCADE removes dependent objects)
DROP TABLE IF EXISTS {module}.{table} CASCADE;

COMMIT;
```

**See**: [MIGRATION_PATTERNS.md](references/MIGRATION_PATTERNS.md) for comprehensive migration guide

---

## Multi-Tenant Requirements

### Mandatory Elements

Every schema MUST include:

#### 1. tenant_id Column (CRITICAL)

```sql
-- ✅ CORRECT: All business tables
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← MANDATORY
    full_name VARCHAR(255) NOT NULL
);

-- ✅ Exception: Global lookup tables
CREATE TABLE countries (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
COMMENT ON TABLE countries IS 'Global lookup table (no tenant_id)';

-- ❌ WRONG: Missing tenant_id
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL  -- ← Missing tenant_id
);
```

#### 2. Composite Foreign Keys (CRITICAL)

```sql
-- ✅ CORRECT: FK includes tenant_id
ALTER TABLE appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patients(tenant_id, id);

-- ❌ WRONG: FK without tenant_id
ALTER TABLE appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (patient_id)  -- ← Missing tenant_id
    REFERENCES patients(id);
```

#### 3. Scoped Unique Constraints (CRITICAL)

```sql
-- ✅ CORRECT: Unique per tenant
CREATE UNIQUE INDEX uq_patients_tenant_email 
    ON patients(tenant_id, email) 
    WHERE deleted_at IS NULL;

-- ❌ WRONG: Unique globally
CREATE UNIQUE INDEX uq_patients_email 
    ON patients(email);  -- ← Prevents email reuse across tenants
```

#### 4. Row-Level Security (CRITICAL)

```sql
-- ✅ CORRECT: RLS enforces tenant isolation
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON patients
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- ❌ WRONG: No RLS (relies on app filtering only)
-- Missing: ALTER TABLE patients ENABLE ROW LEVEL SECURITY;
```

**See**: [MULTI_TENANT_SCHEMA_PATTERNS.md](references/MULTI_TENANT_SCHEMA_PATTERNS.md) for complete patterns

---

## Table Design Standards

### Data Type Guidelines

| Use Case | Type | Example |
|----------|------|---------|
| IDs (Primary/Foreign) | `TEXT` | ULID: `01ARZ3NDEKTSV4RRFFQ69G5FAV` |
| Tenant ID | `TEXT NOT NULL` | Same as primary ID |
| Timestamps | `TIMESTAMPTZ` | `2026-01-20 10:30:00+00` |
| Money/Decimal | `NUMERIC(10,2)` | `1234.56` |
| Boolean | `BOOLEAN` | `true/false` |
| Enum-like | `VARCHAR(20)` + CHECK | `status VARCHAR(20) CHECK (status IN (...))` |
| JSON Data | `JSONB` | `{"key": "value"}` |
| Full Text | `TEXT` + tsvector | For search indexing |
| Short Text | `VARCHAR(n)` | Names, emails (n=50-255) |
| Long Text | `TEXT` | Descriptions, notes |

### Column Naming

- Use `snake_case` (lowercase with underscores)
- Avoid abbreviations: `full_name` not `fname`
- Boolean prefix: `is_active`, `has_profile`
- Timestamp suffix: `created_at`, `deleted_at`
- Foreign key suffix: `user_id`, `patient_id`

### Required Columns

Every table MUST have:

```sql
CREATE TABLE {table} (
    -- Primary Key
    id TEXT PRIMARY KEY,
    
    -- Multi-Tenant
    tenant_id TEXT NOT NULL,
    
    -- Business columns...
    
    -- Audit Columns (REQUIRED)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- Soft delete
);
```

---

## Indexing Strategy

### Index Types

| Type | Use Case | Syntax |
|------|----------|--------|
| B-tree (default) | Equality, range queries | `CREATE INDEX idx_name ON table(col)` |
| Partial | Filter subset | `CREATE INDEX ... WHERE deleted_at IS NULL` |
| Composite | Multi-column queries | `CREATE INDEX ... ON table(col1, col2)` |
| Covering | Avoid table lookup | `CREATE INDEX ... INCLUDE (col3, col4)` |
| GIN | JSONB, arrays, full-text | `CREATE INDEX ... USING GIN(jsonb_col)` |
| UNIQUE | Enforce uniqueness | `CREATE UNIQUE INDEX ... ON table(col)` |

### Index Priorities

**Always Create**:
1. Primary key (automatic)
2. `tenant_id` (every table)
3. Foreign keys
4. Unique constraints

**High Priority**:
- Status columns used in WHERE
- Date ranges (created_at, updated_at)
- Frequently joined columns

**Medium Priority**:
- JSONB columns (GIN)
- Full-text search (GIN + tsvector)
- Covering indexes for hot queries

### Index Best Practices

```sql
-- ✅ CORRECT: tenant_id first, partial index
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date DESC) 
    WHERE deleted_at IS NULL;

-- ✅ CORRECT: Covering index (avoids table lookup)
CREATE INDEX idx_patients_tenant_email_covering 
    ON patients(tenant_id, LOWER(email)) 
    INCLUDE (id, full_name, status);

-- ✅ CORRECT: GIN for JSONB
CREATE INDEX idx_patients_metadata_gin 
    ON patients USING GIN(metadata);

-- ❌ WRONG: tenant_id not first
CREATE INDEX idx_appointments_date_tenant 
    ON appointments(appointment_date, tenant_id);  -- ← Wrong order

-- ❌ WRONG: No partial index (includes deleted)
CREATE INDEX idx_patients_tenant 
    ON patients(tenant_id);  -- ← Should add WHERE deleted_at IS NULL
```

**See**: [MULTI_TENANT_SCHEMA_PATTERNS.md](references/MULTI_TENANT_SCHEMA_PATTERNS.md) - Indexing Strategies section

---

## Security & Access Control

### Row-Level Security (RLS)

**Purpose**: Database-level enforcement of tenant isolation (defense-in-depth)

**When to use**:
- All tables with `tenant_id` column
- Prevents accidental cross-tenant data access
- Complements application-level filtering

**Setup**:
```sql
-- 1. Enable RLS on table
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;

-- 2. Create policy for SELECT
CREATE POLICY tenant_isolation_select ON patients
    FOR SELECT
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 3. Create policy for INSERT
CREATE POLICY tenant_isolation_insert ON patients
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 4. Create policy for UPDATE
CREATE POLICY tenant_isolation_update ON patients
    FOR UPDATE
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);

-- 5. Create policy for DELETE (or disable if soft delete only)
CREATE POLICY tenant_isolation_delete ON patients
    FOR DELETE
    USING (tenant_id = current_setting('app.tenant_id', true)::TEXT);
```

**Application Code** (Go):
```go
func (r *PatientRepository) List(ctx *request.Context) ([]*Patient, error) {
    // Set tenant_id in session variable
    _, err := r.db.Exec("SET LOCAL app.tenant_id = $1", ctx.TenantID)
    if err != nil {
        return nil, err
    }
    
    // Query automatically filtered by RLS policy
    rows, err := r.db.Query(`
        SELECT id, tenant_id, full_name, email, created_at 
        FROM patients 
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
    `)
    // RLS ensures: AND tenant_id = current_setting('app.tenant_id')::TEXT
    
    // ... parse results
}
```

### Sensitive Data Handling

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,  -- PII: Email
    password_hash TEXT NOT NULL,  -- Bcrypt hash (cost 12)
    phone VARCHAR(50),  -- PII: Phone
    ssn_encrypted TEXT,  -- PII: Encrypted with AES-256
    -- ...
);

COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash with cost factor 12 (NEVER store plaintext)';
COMMENT ON COLUMN users.email IS 'PII: User email address (GDPR applicable)';
COMMENT ON COLUMN users.ssn_encrypted IS 'PII: Social Security Number (encrypted with AES-256, key in KMS)';
```

---

## Migration Workflow

### Development Workflow

```bash
# 1. Create new migration files
lokstra migrate create add_emergency_contact

# 2. Edit UP migration
# migrations/patient/003_add_emergency_contact.up.sql

# 3. Edit DOWN migration
# migrations/patient/003_add_emergency_contact.down.sql

# 4. Test UP migration
go run . migrate up

# 5. Test DOWN migration (rollback)
go run . migrate down

# 6. Test UP again (ensure idempotent)
go run . migrate up

# 7. Commit to version control
git add migrations/patient/003_*.sql
git commit -m "feat(patient): add emergency contact fields"
```

### Production Deployment

```bash
# 1. Backup database
pg_dump -U postgres -d prod_db -F c -f backup_$(date +%Y%m%d).dump

# 2. Check migration status
go run . migrate status

# 3. Apply migrations
go run . migrate up

# 4. Verify data integrity
psql -U postgres -d prod_db -c "SELECT COUNT(*) FROM patients WHERE tenant_id IS NULL;"

# 5. If issues, rollback
go run . migrate down
# Then restore from backup if needed
```

### Zero-Downtime Patterns

**Adding Column**:
```sql
-- Phase 1: Add nullable column
ALTER TABLE patients ADD COLUMN priority VARCHAR(20);

-- Phase 2: Backfill (in batches, off-peak hours)
-- (Use DO $$ loop with batching - see MIGRATION_PATTERNS.md)

-- Phase 3: Make NOT NULL
ALTER TABLE patients ALTER COLUMN priority SET NOT NULL;
ALTER TABLE patients ALTER COLUMN priority SET DEFAULT 'normal';
```

**Adding Index (Non-Blocking)**:
```sql
-- Use CONCURRENTLY (runs outside transaction, no table lock)
CREATE INDEX CONCURRENTLY idx_patients_priority 
    ON patients(tenant_id, priority) 
    WHERE deleted_at IS NULL;
```

**See**: [MIGRATION_PATTERNS.md](references/MIGRATION_PATTERNS.md) for complete migration patterns

---

## Quality Checklist

Use this checklist before finalizing schema. Minimum passing score: **85/100 points**.

### Critical Requirements (Must have all)

- [ ] **[10 pts]** All tables have `tenant_id TEXT NOT NULL`
- [ ] **[10 pts]** Composite foreign keys include `tenant_id`
- [ ] **[10 pts]** RLS enabled on all tenant tables
- [ ] **[5 pts]** Audit columns: `created_at`, `updated_at`

### Multi-Tenancy (30 pts total)

- [ ] **[5 pts]** Unique constraints scoped to tenant
- [ ] **[5 pts]** Indexes have `tenant_id` as first column
- [ ] **[5 pts]** Global tables documented as exceptions

### Data Types (15 pts)

- [ ] **[5 pts]** IDs are `TEXT` (ULID/UUID v7)
- [ ] **[5 pts]** Timestamps are `TIMESTAMPTZ`
- [ ] **[5 pts]** CHECK constraints for enum-like fields

### Indexing (20 pts)

- [ ] **[5 pts]** Index on `tenant_id` for every table
- [ ] **[5 pts]** Composite indexes (tenant_id first)
- [ ] **[5 pts]** Partial indexes exclude soft-deleted
- [ ] **[5 pts]** GIN indexes for JSONB columns

### Security (15 pts)

- [ ] **[5 pts]** RLS policies (SELECT, INSERT, UPDATE, DELETE)
- [ ] **[3 pts]** Composite FKs prevent cross-tenant refs
- [ ] **[2 pts]** Password hashing documented

### Audit (10 pts)

- [ ] **[5 pts]** `updated_at` trigger attached
- [ ] **[5 pts]** Soft delete (`deleted_at`) implemented

### Performance (10 pts)

- [ ] **[5 pts]** Partitioning for large tables (>10M rows)
- [ ] **[5 pts]** JSONB used appropriately

**Total Score**: ___ / 100

**Passing**: 85+ points (all critical requirements + 50+ other)

**See**: [SCHEMA_VALIDATION_CHECKLIST.md](references/SCHEMA_VALIDATION_CHECKLIST.md) for detailed validation

---

## Examples

### Complete Auth Module Schema

See [AUTH_SCHEMA_EXAMPLE.md](references/AUTH_SCHEMA_EXAMPLE.md) for a production-ready example with:
- 10 tables (tenants, users, roles, permissions, role_permissions, user_roles, refresh_tokens, password_reset_tokens, login_attempts, audit_logs)
- 57 indexes (tenant, composite, partial, GIN)
- RLS policies on all tables
- Triggers for auto-update timestamps and audit logging
- Composite foreign keys
- Partitioning examples
- Sample queries

### Key Patterns from Example

**1. Tenant Scoped Queries**:
```sql
-- Always filter by tenant_id
SELECT id, full_name, email, status
FROM patients
WHERE tenant_id = $1  -- ← MANDATORY
  AND deleted_at IS NULL
ORDER BY created_at DESC;
```

**2. Composite FK Enforcement**:
```sql
-- Parent table must have unique (tenant_id, id)
ALTER TABLE patients 
ADD CONSTRAINT uq_patients_tenant_id UNIQUE(tenant_id, id);

-- Child table references both columns
ALTER TABLE appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patients(tenant_id, id);
```

**3. Soft Delete Pattern**:
```sql
-- Mark as deleted (never DELETE)
UPDATE patients 
SET deleted_at = NOW() 
WHERE tenant_id = $1 AND id = $2;

-- Query active records only
SELECT * FROM patients 
WHERE tenant_id = $1 AND deleted_at IS NULL;
```

---

## Reference Files

### Comprehensive Guides

1. **[AUTH_SCHEMA_EXAMPLE.md](references/AUTH_SCHEMA_EXAMPLE.md)** (28 KB)
   - Complete 10-table auth module schema
   - All multi-tenant patterns demonstrated
   - RLS policies, triggers, indexes
   - Sample queries and data volume estimates

2. **[MULTI_TENANT_SCHEMA_PATTERNS.md](references/MULTI_TENANT_SCHEMA_PATTERNS.md)** (22 KB)
   - 3 multi-tenancy approaches (shared schema recommended)
   - Tenant isolation patterns
   - Composite foreign key patterns
   - 8 indexing strategies
   - RLS implementation with app code
   - Partitioning strategies
   - Anti-patterns to avoid

3. **[MIGRATION_PATTERNS.md](references/MIGRATION_PATTERNS.md)** (17 KB)
   - UP/DOWN migration templates
   - 6 common migration scenarios
   - Zero-downtime migration techniques
   - Batched data updates
   - Rollback strategies
   - Testing checklist

4. **[SCHEMA_VALIDATION_CHECKLIST.md](references/SCHEMA_VALIDATION_CHECKLIST.md)** (18 KB)
   - 100-point quality scoring system
   - 16 validation criteria with SQL queries
   - Minimum 85/100 to pass
   - Common issues and fixes
   - 2.5-hour review process

### Templates

5. **[SCHEMA_TEMPLATE.md](references/SCHEMA_TEMPLATE.md)** (748 lines)
   - Generic schema documentation template
   - Use as starting point for new modules

---

## Output Format

Save schema design to: `docs/modules/{module-name}/SCHEMA.md`

### Document Structure

```markdown
# {Module Name} - Database Schema

**Module**: {module-name}
**Version**: 1.0.0
**Database**: PostgreSQL 14+
**Multi-Tenant**: Yes (Shared Schema)
**Date**: 2026-01-20

## Overview

{Brief description of module's data model}

## Tables

### 1. {table_name_plural}

**Purpose**: {What this table stores}
**Estimated Rows**: {volume estimate per tenant}
**Partitioning**: {Yes/No - strategy if applicable}

#### Table Definition

```sql
CREATE TABLE {module}.{table_plural} (
    -- ... full SQL definition
);
```

#### Indexes

```sql
-- List all indexes
CREATE INDEX ...;
```

#### Row-Level Security

```sql
ALTER TABLE ... ENABLE ROW LEVEL SECURITY;
CREATE POLICY ...;
```

#### Triggers

```sql
CREATE TRIGGER ...;
```

#### Sample Queries

```sql
-- Common query 1
SELECT ...;

-- Common query 2
SELECT ...;
```

## Relationships

{ER diagram or text description}

## Data Volume Estimates

| Table | Year 1 | Year 3 | Year 5 |
|-------|--------|--------|--------|
| {table1} | 10K | 50K | 100K |
| {table2} | 100K | 500K | 1M |

## Migration Files

- `migrations/{module}/001_create_{table1}.up.sql`
- `migrations/{module}/001_create_{table1}.down.sql`
- ...

## Validation Checklist

- [x] All tables have tenant_id
- [x] Composite foreign keys
- [x] RLS enabled
- ...

**Score**: 92/100 (A - Very Good)
```

---

## Best Practices Summary

### DO

✅ Include `tenant_id` in ALL business tables  
✅ Use composite foreign keys with `tenant_id`  
✅ Scope unique constraints to tenant  
✅ Enable RLS on all tenant tables  
✅ Use `TIMESTAMPTZ` (not TIMESTAMP)  
✅ Use `TEXT` for IDs (ULID/UUID v7)  
✅ Add partial indexes: `WHERE deleted_at IS NULL`  
✅ Put `tenant_id` first in composite indexes  
✅ Attach `updated_at` trigger to all tables  
✅ Soft delete with `deleted_at` column  
✅ Add CHECK constraints for enum-like fields  
✅ Test migrations on staging first  
✅ Write DOWN migrations for rollback  
✅ Comment sensitive columns (PII, passwords)  
✅ Partition tables with >10M rows

### DON'T

❌ Skip `tenant_id` in any business table  
❌ Create foreign keys without `tenant_id`  
❌ Use global unique constraints  
❌ Rely only on app-level tenant filtering  
❌ Use `TIMESTAMP` without timezone  
❌ Use auto-increment integers for IDs  
❌ Create indexes without tenant_id first  
❌ Hard delete business data  
❌ Store passwords in plaintext  
❌ Mix structured data with JSONB  
❌ Create blocking indexes in production  
❌ Update millions of rows in one transaction  
❌ Deploy migrations without backup  
❌ Forget to test rollback procedure  
❌ Reuse migration sequence numbers

---

## Common Pitfalls

### Pitfall 1: Missing tenant_id in Queries

```sql
-- ❌ WRONG: Missing tenant filter (returns ALL tenants' data)
SELECT * FROM patients WHERE email = 'john@example.com';

-- ✅ CORRECT: Always filter by tenant
SELECT * FROM patients 
WHERE tenant_id = $1 AND email = 'john@example.com';
```

### Pitfall 2: Non-Composite Foreign Keys

```sql
-- ❌ WRONG: Allows cross-tenant references
CREATE TABLE appointments (
    patient_id TEXT REFERENCES patients(id)  -- ← No tenant_id
);

-- ✅ CORRECT: Composite FK prevents cross-tenant
CREATE TABLE appointments (
    tenant_id TEXT NOT NULL,
    patient_id TEXT NOT NULL,
    FOREIGN KEY (tenant_id, patient_id) 
        REFERENCES patients(tenant_id, id)
);
```

### Pitfall 3: Long-Running Migrations

```sql
-- ❌ WRONG: Updates all rows in one transaction (locks table)
UPDATE patients SET status = 'active' WHERE status IS NULL;

-- ✅ CORRECT: Batch updates with delays
DO $$
DECLARE batch_size INTEGER := 1000;
BEGIN
    LOOP
        UPDATE patients SET status = 'active' 
        WHERE status IS NULL 
        AND id IN (SELECT id FROM patients WHERE status IS NULL LIMIT batch_size);
        EXIT WHEN NOT FOUND;
        PERFORM pg_sleep(0.1);
    END LOOP;
END $$;
```

---

## PostgreSQL-Specific Features

### Extensions

```sql
-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Full-text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- JSONB operators
CREATE EXTENSION IF NOT EXISTS "btree_gin";
```

### Advanced Types

```sql
-- Enum type (use CHECK constraint instead for flexibility)
CREATE TYPE appointment_status AS ENUM ('scheduled', 'confirmed', 'completed');

-- Range type
CREATE TABLE bookings (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    time_range TSTZRANGE NOT NULL,
    EXCLUDE USING GIST (tenant_id WITH =, time_range WITH &&)
);
```

### Partitioning

```sql
-- Range partitioning by date
CREATE TABLE audit_logs (
    id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- List partitioning by tenant (for very large tenants)
CREATE TABLE transactions (
    id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY LIST (tenant_id);

CREATE TABLE transactions_tenant_abc123 PARTITION OF transactions
    FOR VALUES IN ('abc123');
```

---

## Troubleshooting

### Issue: Migration Fails on Foreign Key

**Error**: `ERROR: insert or update on table violates foreign key constraint`

**Cause**: Child table has records referencing non-existent parent records

**Fix**:
```sql
-- Find orphaned records
SELECT DISTINCT a.patient_id
FROM appointments a
LEFT JOIN patients p ON a.tenant_id = p.tenant_id AND a.patient_id = p.id
WHERE p.id IS NULL;

-- Remove or fix orphaned records before adding FK
DELETE FROM appointments WHERE patient_id IN (SELECT ...);
-- OR
UPDATE appointments SET patient_id = NULL WHERE patient_id IN (SELECT ...);

-- Then add FK
ALTER TABLE appointments ADD CONSTRAINT fk_appointments_patient ...;
```

### Issue: RLS Policy Not Working

**Error**: Queries return no rows even though data exists

**Cause**: `app.tenant_id` session variable not set

**Fix**:
```go
// Set session variable before querying
_, err := db.Exec("SET LOCAL app.tenant_id = $1", ctx.TenantID)
if err != nil {
    return nil, err
}

// Now RLS policy can access current_setting('app.tenant_id')
rows, err := db.Query("SELECT * FROM patients WHERE deleted_at IS NULL")
```

### Issue: Slow Queries on Large Tables

**Error**: Queries take >1 second on tables with >1M rows

**Cause**: Missing indexes or wrong index column order

**Fix**:
```sql
-- Check query plan
EXPLAIN ANALYZE 
SELECT * FROM appointments 
WHERE tenant_id = 'abc123' 
  AND appointment_date >= '2026-01-01';

-- If "Seq Scan" appears, add index
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date DESC) 
    WHERE deleted_at IS NULL;

-- Verify index is used
EXPLAIN ANALYZE SELECT ...;
-- Should show "Index Scan using idx_appointments_tenant_date"
```

---

**Total Length**: ~650 lines  
**Estimated Reading Time**: 30 minutes  
**Last Updated**: 2024-01-20  
**Skill Version**: 2.0.0
