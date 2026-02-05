# PostgreSQL Migration Patterns for Lokstra

**Purpose**: Best practices for creating, organizing, and executing database migrations in Lokstra projects with multi-tenant architecture.

**Context**: Migrations must be safe, reversible, and maintain data integrity across all tenants. This guide covers patterns for common migration scenarios.

---

## Table of Contents

1. [Migration File Structure](#migration-file-structure)
2. [Naming Conventions](#naming-conventions)
3. [Common Migration Patterns](#common-migration-patterns)
4. [Multi-Tenant Considerations](#multi-tenant-considerations)
5. [Zero-Downtime Migrations](#zero-downtime-migrations)
6. [Data Migrations](#data-migrations)
7. [Rollback Strategies](#rollback-strategies)

---

## Migration File Structure

### Directory Organization

```
migrations/
├── auth/
│   ├── 001_create_tenants.up.sql
│   ├── 001_create_tenants.down.sql
│   ├── 002_create_users.up.sql
│   ├── 002_create_users.down.sql
│   ├── 003_create_roles.up.sql
│   ├── 003_create_roles.down.sql
│   └── ...
├── patient/
│   ├── 001_create_patients.up.sql
│   ├── 001_create_patients.down.sql
│   ├── 002_add_emergency_contact.up.sql
│   ├── 002_add_emergency_contact.down.sql
│   └── ...
└── shared/
    ├── 000_create_extensions.up.sql
    ├── 000_create_extensions.down.sql
    └── ...
```

### Migration Template

```sql
-- Migration: [Description of what this migration does]
-- Module: [module-name]
-- Version: [version]
-- Date: [YYYY-MM-DD]
-- Author: [author]

-- UP Migration
BEGIN;

-- Your changes here
CREATE TABLE ...;
CREATE INDEX ...;

COMMIT;
```

```sql
-- DOWN Migration (Rollback)
BEGIN;

-- Reverse the changes
DROP TABLE IF EXISTS ... CASCADE;

COMMIT;
```

---

## Naming Conventions

### File Naming Format

```
{sequence}_{action}_{resource}.{direction}.sql

Examples:
001_create_users.up.sql
001_create_users.down.sql
002_add_email_verification.up.sql
002_add_email_verification.down.sql
003_alter_users_add_phone.up.sql
003_alter_users_add_phone.down.sql
```

### Sequence Numbers

- Start at 001, increment by 1
- Zero-pad to 3 digits (001, 002, 003, ...)
- Use 000 for foundation migrations (extensions, schemas)
- Never reuse sequence numbers

### Action Verbs

- `create_` - Create new table/schema/extension
- `alter_` - Modify existing table
- `add_` - Add column/index/constraint
- `remove_` - Remove column/index/constraint
- `rename_` - Rename table/column
- `drop_` - Delete table/index
- `backfill_` - Data migration

---

## Common Migration Patterns

### Pattern 1: Create Table

**UP Migration**:
```sql
-- Migration: Create patients table
-- Module: patient
-- Version: 1.0.0
-- Date: 2026-01-20

BEGIN;

CREATE TABLE IF NOT EXISTS patient.patients (
    -- Primary Key
    id TEXT PRIMARY KEY,
    
    -- Multi-Tenant
    tenant_id TEXT NOT NULL,
    
    -- Business Fields
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    date_of_birth DATE NOT NULL,
    gender VARCHAR(20),
    
    -- Address
    address TEXT,
    city VARCHAR(100),
    postal_code VARCHAR(20),
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT chk_patients_status CHECK (status IN ('active', 'inactive', 'archived')),
    CONSTRAINT chk_patients_gender CHECK (gender IN ('male', 'female', 'other')),
    CONSTRAINT uq_patients_tenant_email UNIQUE(tenant_id, email)
);

-- Indexes
CREATE INDEX idx_patients_tenant ON patient.patients(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_patients_tenant_email ON patient.patients(tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_patients_status ON patient.patients(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_patients_created_at ON patient.patients(tenant_id, created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER trg_patients_updated_at
    BEFORE UPDATE ON patient.patients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE patient.patients IS 'Patient records with multi-tenant isolation';
COMMENT ON COLUMN patient.patients.tenant_id IS 'Clinic/organization identifier';
COMMENT ON COLUMN patient.patients.deleted_at IS 'Soft delete timestamp (NULL = active)';

COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Drop patients table
BEGIN;

DROP TRIGGER IF EXISTS trg_patients_updated_at ON patient.patients;
DROP TABLE IF EXISTS patient.patients CASCADE;

COMMIT;
```

### Pattern 2: Add Column

**UP Migration**:
```sql
-- Migration: Add emergency contact to patients
-- Module: patient
-- Date: 2026-01-25

BEGIN;

-- Add column (nullable initially for existing rows)
ALTER TABLE patient.patients 
ADD COLUMN emergency_contact_name VARCHAR(255),
ADD COLUMN emergency_contact_phone VARCHAR(50),
ADD COLUMN emergency_contact_relationship VARCHAR(50);

-- Add constraint
ALTER TABLE patient.patients
ADD CONSTRAINT chk_patients_emergency_relationship 
    CHECK (emergency_contact_relationship IN ('spouse', 'parent', 'sibling', 'child', 'friend', 'other'));

-- Add index if needed
CREATE INDEX idx_patients_emergency_phone 
    ON patient.patients(tenant_id, emergency_contact_phone) 
    WHERE emergency_contact_phone IS NOT NULL;

-- Comments
COMMENT ON COLUMN patient.patients.emergency_contact_name IS 'Emergency contact full name';
COMMENT ON COLUMN patient.patients.emergency_contact_phone IS 'Emergency contact phone number';

COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Remove emergency contact from patients
BEGIN;

DROP INDEX IF EXISTS patient.idx_patients_emergency_phone;

ALTER TABLE patient.patients
DROP CONSTRAINT IF EXISTS chk_patients_emergency_relationship,
DROP COLUMN IF EXISTS emergency_contact_name,
DROP COLUMN IF EXISTS emergency_contact_phone,
DROP COLUMN IF EXISTS emergency_contact_relationship;

COMMIT;
```

### Pattern 3: Add Index (Non-Blocking)

**UP Migration**:
```sql
-- Migration: Add index for patient search by name
-- Module: patient
-- Date: 2026-02-01

BEGIN;

-- CREATE INDEX CONCURRENTLY cannot run inside a transaction
COMMIT;

-- Run outside transaction for zero-downtime
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_patients_tenant_name 
    ON patient.patients(tenant_id, LOWER(full_name)) 
    WHERE deleted_at IS NULL;

BEGIN;
COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Remove patient name index
BEGIN;

DROP INDEX CONCURRENTLY IF EXISTS patient.idx_patients_tenant_name;

COMMIT;
```

### Pattern 4: Rename Column

**UP Migration**:
```sql
-- Migration: Rename 'name' to 'full_name' in patients table
-- Module: patient
-- Date: 2026-02-05

BEGIN;

-- Rename column
ALTER TABLE patient.patients 
RENAME COLUMN name TO full_name;

-- Update comment
COMMENT ON COLUMN patient.patients.full_name IS 'Patient full name';

-- Update any affected indexes (if needed)
DROP INDEX IF EXISTS patient.idx_patients_tenant_name;
CREATE INDEX idx_patients_tenant_fullname 
    ON patient.patients(tenant_id, LOWER(full_name)) 
    WHERE deleted_at IS NULL;

COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Rename 'full_name' back to 'name'
BEGIN;

ALTER TABLE patient.patients 
RENAME COLUMN full_name TO name;

DROP INDEX IF EXISTS patient.idx_patients_tenant_fullname;
CREATE INDEX idx_patients_tenant_name 
    ON patient.patients(tenant_id, LOWER(name)) 
    WHERE deleted_at IS NULL;

COMMIT;
```

### Pattern 5: Add Foreign Key with tenant_id

**UP Migration**:
```sql
-- Migration: Add foreign key from appointments to patients
-- Module: appointment
-- Date: 2026-02-10

BEGIN;

-- First, ensure composite primary key exists on parent table
ALTER TABLE patient.patients 
DROP CONSTRAINT IF EXISTS patients_pkey,
ADD PRIMARY KEY (id),
ADD CONSTRAINT uq_patients_tenant_id UNIQUE(tenant_id, id);

-- Add foreign key to child table
ALTER TABLE appointment.appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patient.patients(tenant_id, id)
    ON DELETE RESTRICT
    ON UPDATE CASCADE;

-- Add index for foreign key lookups
CREATE INDEX idx_appointments_tenant_patient 
    ON appointment.appointments(tenant_id, patient_id);

COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Remove foreign key from appointments to patients
BEGIN;

DROP INDEX IF EXISTS appointment.idx_appointments_tenant_patient;

ALTER TABLE appointment.appointments
DROP CONSTRAINT IF EXISTS fk_appointments_patient;

COMMIT;
```

### Pattern 6: Change Column Type

**UP Migration**:
```sql
-- Migration: Change phone column from VARCHAR(20) to VARCHAR(50)
-- Module: patient
-- Date: 2026-02-15

BEGIN;

-- Change column type
ALTER TABLE patient.patients 
ALTER COLUMN phone TYPE VARCHAR(50);

-- If data transformation needed:
-- ALTER TABLE patient.patients 
-- ALTER COLUMN phone TYPE VARCHAR(50) USING phone::VARCHAR(50);

COMMIT;
```

**DOWN Migration**:
```sql
-- Migration: Revert phone column to VARCHAR(20)
BEGIN;

-- Warning: May truncate data if values exceed 20 characters
ALTER TABLE patient.patients 
ALTER COLUMN phone TYPE VARCHAR(20);

COMMIT;
```

---

## Multi-Tenant Considerations

### Rule 1: Always Include tenant_id

```sql
-- ✅ CORRECT: New table with tenant_id
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← Always include
    patient_id TEXT NOT NULL,
    -- ...
);

-- ❌ WRONG: Missing tenant_id
CREATE TABLE appointments (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL
);
```

### Rule 2: Composite Foreign Keys

```sql
-- ✅ CORRECT: Composite FK with tenant_id
ALTER TABLE appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patients(tenant_id, id);

-- ❌ WRONG: FK without tenant_id
ALTER TABLE appointments
ADD CONSTRAINT fk_appointments_patient 
    FOREIGN KEY (patient_id) 
    REFERENCES patients(id);
```

### Rule 3: Indexes with tenant_id First

```sql
-- ✅ CORRECT: tenant_id first in index
CREATE INDEX idx_appointments_tenant_date 
    ON appointments(tenant_id, appointment_date);

-- ❌ WRONG: tenant_id not first
CREATE INDEX idx_appointments_date_tenant 
    ON appointments(appointment_date, tenant_id);
```

### Rule 4: Unique Constraints with tenant_id

```sql
-- ✅ CORRECT: Unique per tenant
CREATE UNIQUE INDEX uq_patients_tenant_email 
    ON patients(tenant_id, email) 
    WHERE deleted_at IS NULL;

-- ❌ WRONG: Unique globally (prevents email reuse across tenants)
CREATE UNIQUE INDEX uq_patients_email 
    ON patients(email);
```

---

## Zero-Downtime Migrations

### Pattern 1: Add Column with Default

```sql
-- Phase 1: Add nullable column
BEGIN;
ALTER TABLE patients ADD COLUMN priority VARCHAR(20);
COMMIT;

-- Phase 2: Backfill data (in batches)
DO $$
DECLARE
    batch_size INTEGER := 1000;
    offset_val INTEGER := 0;
    rows_updated INTEGER;
BEGIN
    LOOP
        UPDATE patients
        SET priority = 'normal'
        WHERE priority IS NULL
          AND id IN (
              SELECT id FROM patients 
              WHERE priority IS NULL 
              LIMIT batch_size
          );
        
        GET DIAGNOSTICS rows_updated = ROW_COUNT;
        EXIT WHEN rows_updated = 0;
        
        RAISE NOTICE 'Updated % rows', rows_updated;
        PERFORM pg_sleep(0.1);  -- Avoid locking issues
    END LOOP;
END $$;

-- Phase 3: Make NOT NULL
BEGIN;
ALTER TABLE patients ALTER COLUMN priority SET NOT NULL;
ALTER TABLE patients ALTER COLUMN priority SET DEFAULT 'normal';
COMMIT;
```

### Pattern 2: Rename Column (3-Phase)

```sql
-- Phase 1: Add new column
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);

-- Phase 2: Dual-write period (app writes to both columns)
-- Deploy app that writes to both 'name' and 'full_name'

-- Phase 3: Backfill existing data
UPDATE users SET full_name = name WHERE full_name IS NULL;

-- Phase 4: Make new column NOT NULL
ALTER TABLE users ALTER COLUMN full_name SET NOT NULL;

-- Phase 5: Drop old column (after app only reads from full_name)
ALTER TABLE users DROP COLUMN name;
```

### Pattern 3: Add Index (Non-Blocking)

```sql
-- Use CONCURRENTLY to avoid blocking reads/writes
CREATE INDEX CONCURRENTLY idx_patients_tenant_phone 
    ON patients(tenant_id, phone) 
    WHERE deleted_at IS NULL;

-- Check if index is valid
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'patients' 
  AND indexname = 'idx_patients_tenant_phone';
```

---

## Data Migrations

### Pattern 1: Batch Updates

```sql
-- Migration: Set default value for existing records
-- Module: patient
-- Date: 2026-03-01

DO $$
DECLARE
    batch_size INTEGER := 1000;
    total_updated INTEGER := 0;
    rows_updated INTEGER;
BEGIN
    LOOP
        -- Update in batches
        WITH batch AS (
            SELECT id
            FROM patient.patients
            WHERE status IS NULL
            LIMIT batch_size
        )
        UPDATE patient.patients p
        SET status = 'active', updated_at = NOW()
        FROM batch
        WHERE p.id = batch.id;
        
        GET DIAGNOSTICS rows_updated = ROW_COUNT;
        EXIT WHEN rows_updated = 0;
        
        total_updated := total_updated + rows_updated;
        RAISE NOTICE 'Updated % rows (total: %)', rows_updated, total_updated;
        
        -- Small delay to avoid lock contention
        PERFORM pg_sleep(0.1);
    END LOOP;
    
    RAISE NOTICE 'Migration complete. Total rows updated: %', total_updated;
END $$;
```

### Pattern 2: Data Transformation

```sql
-- Migration: Split full_name into first_name and last_name
-- Module: patient
-- Date: 2026-03-05

BEGIN;

-- Add new columns
ALTER TABLE patient.patients 
ADD COLUMN first_name VARCHAR(255),
ADD COLUMN last_name VARCHAR(255);

-- Transform data
UPDATE patient.patients
SET 
    first_name = SPLIT_PART(full_name, ' ', 1),
    last_name = CASE 
        WHEN POSITION(' ' IN full_name) > 0 
        THEN SUBSTRING(full_name FROM POSITION(' ' IN full_name) + 1)
        ELSE ''
    END
WHERE full_name IS NOT NULL;

-- Make NOT NULL after backfill
ALTER TABLE patient.patients 
ALTER COLUMN first_name SET NOT NULL,
ALTER COLUMN last_name SET NOT NULL;

COMMIT;
```

### Pattern 3: Populate from Another Table

```sql
-- Migration: Copy email from users to patients
-- Module: patient
-- Date: 2026-03-10

BEGIN;

-- Add column
ALTER TABLE patient.patients ADD COLUMN email VARCHAR(255);

-- Populate from users table
UPDATE patient.patients p
SET email = u.email
FROM auth.users u
WHERE p.tenant_id = u.tenant_id 
  AND p.user_id = u.id;

-- Make NOT NULL (or leave nullable if some patients have no user)
-- ALTER TABLE patient.patients ALTER COLUMN email SET NOT NULL;

COMMIT;
```

---

## Rollback Strategies

### Strategy 1: Immediate Rollback

```sql
-- If migration fails, PostgreSQL automatically rolls back
-- due to BEGIN/COMMIT transaction block

BEGIN;

CREATE TABLE new_table (...);
-- Error occurs here
CREATE INDEX idx_something ON non_existent_table(...);

COMMIT;  -- Never reached, entire transaction rolled back
```

### Strategy 2: Manual Rollback

```bash
# Apply migration
psql -U postgres -d mydb -f migrations/auth/005_add_two_factor.up.sql

# If issues detected, rollback
psql -U postgres -d mydb -f migrations/auth/005_add_two_factor.down.sql
```

### Strategy 3: Point-in-Time Recovery (PITR)

```bash
# Before migration, create manual backup
pg_dump -U postgres -d mydb -F c -f backup_before_migration.dump

# If migration causes issues, restore
pg_restore -U postgres -d mydb -c backup_before_migration.dump
```

### Strategy 4: Shadow Table Pattern

```sql
-- Phase 1: Create new table structure
CREATE TABLE patients_new (
    -- New schema with improvements
);

-- Phase 2: Copy data
INSERT INTO patients_new SELECT * FROM patients;

-- Phase 3: Swap tables (atomic)
BEGIN;
ALTER TABLE patients RENAME TO patients_old;
ALTER TABLE patients_new RENAME TO patients;
COMMIT;

-- Phase 4: If all OK, drop old table
-- If issues, swap back
BEGIN;
ALTER TABLE patients RENAME TO patients_new;
ALTER TABLE patients_old RENAME TO patients;
COMMIT;
```

---

## Migration Testing Checklist

### Before Applying Migration

- [ ] Test on local development database
- [ ] Test on staging environment with production-like data
- [ ] Verify UP migration executes without errors
- [ ] Verify DOWN migration properly reverts changes
- [ ] Check for syntax errors
- [ ] Verify foreign key constraints work correctly
- [ ] Test with sample tenant data
- [ ] Measure execution time on large dataset
- [ ] Check for lock contention issues
- [ ] Verify indexes are used (EXPLAIN ANALYZE)

### During Migration

- [ ] Monitor database connections
- [ ] Watch for long-running queries
- [ ] Check replication lag (if applicable)
- [ ] Monitor disk space usage
- [ ] Watch application error logs

### After Migration

- [ ] Verify data integrity (row counts, constraints)
- [ ] Test application functionality
- [ ] Check query performance
- [ ] Verify indexes are valid
- [ ] Update schema documentation
- [ ] Tag migration in version control

---

## Common Pitfalls

### ❌ Pitfall 1: Long-Running Migrations Without Batching

```sql
-- WRONG: Updates all rows in one transaction (locks table)
UPDATE patients SET status = 'active' WHERE status IS NULL;
```

**Fix**: Use batched updates

```sql
-- CORRECT: Process in batches
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

### ❌ Pitfall 2: Adding NOT NULL Without Default

```sql
-- WRONG: Fails if any existing rows have NULL
ALTER TABLE patients ADD COLUMN email VARCHAR(255) NOT NULL;
```

**Fix**: Add column as nullable, backfill, then set NOT NULL

```sql
-- CORRECT: 3-phase approach
ALTER TABLE patients ADD COLUMN email VARCHAR(255);
UPDATE patients SET email = 'unknown@example.com' WHERE email IS NULL;
ALTER TABLE patients ALTER COLUMN email SET NOT NULL;
```

### ❌ Pitfall 3: Creating Indexes Inside Transaction

```sql
-- WRONG: Blocks other transactions
BEGIN;
CREATE INDEX idx_patients_name ON patients(name);
COMMIT;
```

**Fix**: Use CONCURRENTLY outside transaction

```sql
-- CORRECT: Non-blocking
CREATE INDEX CONCURRENTLY idx_patients_name ON patients(name);
```

### ❌ Pitfall 4: Missing tenant_id in Foreign Keys

```sql
-- WRONG: FK without tenant_id (allows cross-tenant references)
ALTER TABLE appointments 
ADD CONSTRAINT fk_patient 
    FOREIGN KEY (patient_id) REFERENCES patients(id);
```

**Fix**: Composite FK with tenant_id

```sql
-- CORRECT: Composite FK prevents cross-tenant references
ALTER TABLE appointments 
ADD CONSTRAINT fk_patient 
    FOREIGN KEY (tenant_id, patient_id) 
    REFERENCES patients(tenant_id, id);
```

---

## Lokstra Migration Runner

### Usage Example

```go
// main.go
package main

import (
    "github.com/primadi/lokstra/lokstra_init"
)

func main() {
    app := lokstra_init.BootstrapAndRun(
        lokstra_init.WithMigrations("./migrations"),
        lokstra_init.WithAutoMigrate(true),
    )
    
    app.Start()
}
```

### Manual Migration

```bash
# Run all pending migrations
go run . migrate up

# Rollback last migration
go run . migrate down

# Check migration status
go run . migrate status

# Create new migration
go run . migrate create add_user_avatar
```

---

## Best Practices Summary

### DO
✅ Use transactions (BEGIN/COMMIT) for atomic changes  
✅ Include both UP and DOWN migrations  
✅ Test migrations on staging before production  
✅ Use batched updates for large datasets  
✅ Add indexes CONCURRENTLY in production  
✅ Include tenant_id in all tables  
✅ Use composite foreign keys with tenant_id  
✅ Add comments to tables and columns  
✅ Monitor migration execution time  
✅ Keep migrations small and focused

### DON'T
❌ Run migrations without backup  
❌ Skip writing DOWN migrations  
❌ Make breaking changes without deprecation period  
❌ Add NOT NULL column without default value  
❌ Create blocking indexes in production  
❌ Update millions of rows in single transaction  
❌ Forget to test rollback procedure  
❌ Mix DDL and DML in same migration  
❌ Hardcode tenant-specific data  
❌ Reuse migration sequence numbers

---

**File Size**: 17 KB  
**Last Updated**: 2024-01-20  
**Related**: MULTI_TENANT_SCHEMA_PATTERNS.md, AUTH_SCHEMA_EXAMPLE.md
