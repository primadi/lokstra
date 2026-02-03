---
module: [MODULE_NAME]
version: [VERSION]
status: draft
based_on: Requirements v[VERSION]
---

# [Module Name] - Database Schema

## Schema Information

| Field | Value |
|-------|-------|
| Module | [module_name] |
| Version | [VERSION] |
| Status | draft / approved / implemented |
| Database | PostgreSQL / MySQL / SQLite |
| Schema Name | [schema_name] |
| Last Updated | [DATE] |

---

## 1. Overview

### 1.1 Purpose
[Brief description of the schema purpose]

### 1.2 Database Technology
- **DBMS:** PostgreSQL 14+ / MySQL 8+ / SQLite 3+
- **Character Set:** UTF-8
- **Collation:** [Collation setting]

### 1.3 Schema Naming Convention
- **Tables:** `lowercase_with_underscores` (plural)
- **Columns:** `lowercase_with_underscores`
- **Indexes:** `idx_tablename_columnname`
- **Foreign Keys:** `fk_tablename_columname`
- **Unique Constraints:** `uq_tablename_columnname`

---

## 2. Tables

### 2.1 [table_name_1]

**Description:**
[Detailed description of the table purpose and usage]

**SQL Definition:**
```sql
CREATE TABLE [table_name_1] (
    -- Primary Key
    id VARCHAR(50) PRIMARY KEY,
    
    -- Core Fields
    field1 VARCHAR(100) NOT NULL,
    field2 VARCHAR(255) UNIQUE NOT NULL,
    field3 INTEGER NOT NULL DEFAULT 0,
    field4 DATE NOT NULL,
    field5 VARCHAR(20) NOT NULL,
    
    -- JSON/JSONB Fields (PostgreSQL)
    metadata JSONB,
    settings JSONB,
    
    -- Status/State
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_field5 CHECK (field5 IN ('value1', 'value2', 'value3')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'deleted')),
    CONSTRAINT chk_field3_positive CHECK (field3 >= 0),
    CONSTRAINT chk_field4_past CHECK (field4 <= CURRENT_DATE)
);

-- Indexes
CREATE UNIQUE INDEX idx_[table]_field2 ON [table_name_1](field2);
CREATE INDEX idx_[table]_status ON [table_name_1](status);
CREATE INDEX idx_[table]_created ON [table_name_1](created_at DESC);
CREATE INDEX idx_[table]_field1 ON [table_name_1](field1);

-- Partial Index (only non-null values)
CREATE INDEX idx_[table]_field2_active ON [table_name_1](field2) 
  WHERE deleted_at IS NULL;

-- Composite Index
CREATE INDEX idx_[table]_status_created ON [table_name_1](status, created_at DESC);

-- Comments
COMMENT ON TABLE [table_name_1] IS 'Primary entity table for [description]';
COMMENT ON COLUMN [table_name_1].id IS 'Unique identifier (UUID or custom format)';
COMMENT ON COLUMN [table_name_1].field2 IS 'Unique identifier for business logic';
COMMENT ON COLUMN [table_name_1].metadata IS 'Additional metadata (JSONB format)';
COMMENT ON COLUMN [table_name_1].deleted_at IS 'Soft delete timestamp';
```

**Column Details:**

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | VARCHAR(50) | NO | - | Primary key, unique identifier |
| field1 | VARCHAR(100) | NO | - | [Description] |
| field2 | VARCHAR(255) | NO | - | [Description], unique |
| field3 | INTEGER | NO | 0 | [Description], must be >= 0 |
| field4 | DATE | NO | - | [Description], must be past date |
| field5 | VARCHAR(20) | NO | - | [Description], enum: value1\|value2\|value3 |
| metadata | JSONB | YES | NULL | Additional metadata (flexible structure) |
| settings | JSONB | YES | NULL | Entity-specific settings |
| status | VARCHAR(20) | NO | 'active' | Current status: active\|inactive\|deleted |
| created_at | TIMESTAMP | NO | NOW() | Record creation timestamp |
| updated_at | TIMESTAMP | NO | NOW() | Last modification timestamp |
| deleted_at | TIMESTAMP | YES | NULL | Soft delete timestamp (NULL = active) |

**Constraints:**
- **Primary Key:** `id`
- **Unique:** `field2` (business identifier)
- **Check:** `field5 IN ('value1', 'value2', 'value3')`
- **Check:** `status IN ('active', 'inactive', 'deleted')`
- **Check:** `field3 >= 0` (non-negative)
- **Check:** `field4 <= CURRENT_DATE` (past date only)

**Indexes:**
- **Primary:** `id` (clustered)
- **Unique:** `field2` (fast lookup by business ID)
- **Regular:** `status` (filter by status)
- **Regular:** `created_at DESC` (sorting by creation date)
- **Partial:** `field2 WHERE deleted_at IS NULL` (only active records)
- **Composite:** `(status, created_at DESC)` (filter + sort)

**Sample Data:**
```sql
INSERT INTO [table_name_1] (id, field1, field2, field3, field4, field5, status)
VALUES 
  ('ID-001', 'Example 1', 'BIZ-001', 100, '2026-01-15', 'value1', 'active'),
  ('ID-002', 'Example 2', 'BIZ-002', 200, '2026-01-20', 'value2', 'active');
```

---

### 2.2 [table_name_2]

**Description:**
[Related entity table - child of table_name_1]

**SQL Definition:**
```sql
CREATE TABLE [table_name_2] (
    -- Primary Key
    id SERIAL PRIMARY KEY,
    
    -- Foreign Key
    parent_id VARCHAR(50) NOT NULL REFERENCES [table_name_1](id) ON DELETE CASCADE,
    
    -- Core Fields
    name VARCHAR(100) NOT NULL,
    description TEXT,
    value DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    type VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_type CHECK (type IN ('type1', 'type2', 'type3')),
    CONSTRAINT chk_value_positive CHECK (value >= 0)
);

-- Indexes
CREATE INDEX idx_[table2]_parent ON [table_name_2](parent_id);
CREATE INDEX idx_[table2]_type ON [table_name_2](type);
CREATE INDEX idx_[table2]_active ON [table_name_2](is_active);

-- Composite Index for common query
CREATE INDEX idx_[table2]_parent_type ON [table_name_2](parent_id, type);

-- Comments
COMMENT ON TABLE [table_name_2] IS 'Child entities related to [table_name_1]';
COMMENT ON COLUMN [table_name_2].parent_id IS 'Reference to parent entity (CASCADE delete)';
```

**Column Details:**

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | SERIAL | NO | AUTO | Auto-increment primary key |
| parent_id | VARCHAR(50) | NO | - | Foreign key to [table_name_1] |
| name | VARCHAR(100) | NO | - | Entity name |
| description | TEXT | YES | NULL | Optional description |
| value | DECIMAL(10,2) | NO | 0.00 | Numeric value (must be >= 0) |
| type | VARCHAR(50) | NO | - | Entity type: type1\|type2\|type3 |
| is_active | BOOLEAN | NO | true | Active status flag |
| created_at | TIMESTAMP | NO | NOW() | Creation timestamp |
| updated_at | TIMESTAMP | NO | NOW() | Last update timestamp |

**Relationships:**
- **Belongs To:** [table_name_1] (N:1)
  - Foreign Key: `parent_id` → `[table_name_1].id`
  - On Delete: CASCADE (delete children when parent deleted)
  - On Update: NO ACTION

**Sample Data:**
```sql
INSERT INTO [table_name_2] (parent_id, name, description, value, type)
VALUES 
  ('ID-001', 'Child 1', 'Description 1', 10.50, 'type1'),
  ('ID-001', 'Child 2', 'Description 2', 20.75, 'type2'),
  ('ID-002', 'Child 3', 'Description 3', 30.00, 'type1');
```

---

### 2.3 [junction_table] (Many-to-Many)

**Description:**
Junction table for many-to-many relationship between [table_name_1] and [other_table]

**SQL Definition:**
```sql
CREATE TABLE [junction_table] (
    -- Composite Primary Key
    entity1_id VARCHAR(50) NOT NULL REFERENCES [table_name_1](id) ON DELETE CASCADE,
    entity2_id VARCHAR(50) NOT NULL REFERENCES [other_table](id) ON DELETE CASCADE,
    
    -- Additional Fields (optional)
    role VARCHAR(50),
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by VARCHAR(50),
    
    -- Primary Key
    PRIMARY KEY (entity1_id, entity2_id)
);

-- Indexes
CREATE INDEX idx_junction_entity1 ON [junction_table](entity1_id);
CREATE INDEX idx_junction_entity2 ON [junction_table](entity2_id);

-- Comments
COMMENT ON TABLE [junction_table] IS 'Many-to-many relationship between [table1] and [table2]';
```

---

## 3. Relationships

### 3.1 Entity Relationship Diagram

```
[table_name_1] (1) ──────< (*) [table_name_2]
      │
      │ (*)
      │
      └──────< (*) [junction_table] ──────> (*) [other_table]
```

### 3.2 Relationship Details

#### Relationship 1: [table_name_1] → [table_name_2]
- **Type:** One-to-Many (1:N)
- **Parent:** [table_name_1]
- **Child:** [table_name_2]
- **Foreign Key:** `[table_name_2].parent_id` → `[table_name_1].id`
- **On Delete:** CASCADE (delete all children)
- **On Update:** NO ACTION
- **Description:** One parent can have multiple children

#### Relationship 2: [table_name_1] ↔ [other_table]
- **Type:** Many-to-Many (M:N)
- **Junction Table:** [junction_table]
- **Foreign Keys:** 
  - `entity1_id` → `[table_name_1].id` (CASCADE)
  - `entity2_id` → `[other_table].id` (CASCADE)
- **Description:** Many-to-many relationship with additional attributes

---

## 4. Indexes

### 4.1 Index Strategy

**Primary Indexes:**
- All tables have PRIMARY KEY on `id` column

**Unique Indexes:**
- Business identifiers (e.g., `field2`)
- Natural keys

**Regular Indexes:**
- Foreign keys (for JOIN performance)
- Status/filter columns (e.g., `status`, `is_active`)
- Sort columns (e.g., `created_at`)

**Composite Indexes:**
- Common filter + sort combinations
- Foreign key + type/status

**Partial Indexes:**
- Index only active records: `WHERE deleted_at IS NULL`
- Index only specific values: `WHERE status = 'active'`

### 4.2 Index Maintenance

```sql
-- Analyze query performance
EXPLAIN ANALYZE
SELECT * FROM [table_name_1]
WHERE status = 'active' AND created_at > '2026-01-01'
ORDER BY created_at DESC
LIMIT 20;

-- Rebuild indexes (if needed)
REINDEX TABLE [table_name_1];

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE schemaname = '[schema_name]'
ORDER BY idx_scan ASC;
```

---

## 5. Data Types

### 5.1 Standard Data Types

| Type | PostgreSQL | MySQL | Usage |
|------|-----------|-------|-------|
| Primary Key (UUID) | VARCHAR(50) | VARCHAR(50) | Unique identifiers |
| Primary Key (Auto) | SERIAL / BIGSERIAL | INT AUTO_INCREMENT | Auto-increment IDs |
| Short Text | VARCHAR(n) | VARCHAR(n) | Names, codes (max 255) |
| Long Text | TEXT | TEXT | Descriptions, content |
| Integer | INTEGER / BIGINT | INT / BIGINT | Counts, quantities |
| Decimal | DECIMAL(p,s) | DECIMAL(p,s) | Money, precise numbers |
| Boolean | BOOLEAN | TINYINT(1) | Yes/No flags |
| Date | DATE | DATE | Date only |
| Timestamp | TIMESTAMP | DATETIME | Date + time |
| JSON | JSONB | JSON | Flexible data |

### 5.2 Custom Types (PostgreSQL)

```sql
-- Enum Type
CREATE TYPE [status_type] AS ENUM ('active', 'inactive', 'deleted');

-- Composite Type
CREATE TYPE [address_type] AS (
    street TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10)
);

-- Use in table
CREATE TABLE example (
    id VARCHAR(50) PRIMARY KEY,
    status [status_type] NOT NULL,
    address [address_type]
);
```

---

## 6. Constraints

### 6.1 Constraint Types

**Primary Key:**
```sql
PRIMARY KEY (id)
PRIMARY KEY (entity1_id, entity2_id)  -- Composite
```

**Foreign Key:**
```sql
FOREIGN KEY (parent_id) REFERENCES [table](id) ON DELETE CASCADE
```

**Unique:**
```sql
UNIQUE (email)
CONSTRAINT uq_table_email UNIQUE (email)
```

**Check:**
```sql
CHECK (age >= 0 AND age <= 150)
CHECK (status IN ('active', 'inactive'))
CHECK (start_date <= end_date)
```

**Not Null:**
```sql
field1 VARCHAR(100) NOT NULL
```

**Default:**
```sql
status VARCHAR(20) DEFAULT 'active'
created_at TIMESTAMP DEFAULT NOW()
```

---

## 7. Triggers & Functions

### 7.1 Update Timestamp Trigger

**Purpose:** Automatically update `updated_at` timestamp on row modification

```sql
-- Function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger
CREATE TRIGGER trg_[table]_updated_at
    BEFORE UPDATE ON [table_name_1]
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### 7.2 Audit Log Trigger (Optional)

**Purpose:** Track all changes to critical tables

```sql
-- Audit table
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id VARCHAR(50) NOT NULL,
    action VARCHAR(10) NOT NULL,  -- INSERT, UPDATE, DELETE
    old_data JSONB,
    new_data JSONB,
    changed_by VARCHAR(50),
    changed_at TIMESTAMP DEFAULT NOW()
);

-- Audit function
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_log (table_name, record_id, action, old_data)
        VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', row_to_json(OLD));
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_log (table_name, record_id, action, old_data, new_data)
        VALUES (TG_TABLE_NAME, OLD.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_log (table_name, record_id, action, new_data)
        VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', row_to_json(NEW));
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to table
CREATE TRIGGER trg_[table]_audit
    AFTER INSERT OR UPDATE OR DELETE ON [table_name_1]
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();
```

---

## 8. Migrations

### 8.1 Migration Files

**Up Migration:** `migrations/[module]/001_create_tables.up.sql`
```sql
-- Migration: Initial schema for [module]
-- Version: 1.0

BEGIN;

CREATE TABLE [table_name_1] (
    -- ... (as defined above)
);

CREATE TABLE [table_name_2] (
    -- ... (as defined above)
);

-- Indexes
CREATE UNIQUE INDEX idx_[table]_field2 ON [table_name_1](field2);
-- ... (all indexes)

-- Comments
COMMENT ON TABLE [table_name_1] IS '...';
-- ... (all comments)

COMMIT;
```

**Down Migration:** `migrations/[module]/001_create_tables.down.sql`
```sql
-- Rollback: Drop tables for [module]

BEGIN;

DROP TABLE IF EXISTS [table_name_2] CASCADE;
DROP TABLE IF EXISTS [table_name_1] CASCADE;

COMMIT;
```

### 8.2 Migration Naming Convention
- `001_create_initial_schema.up.sql`
- `002_add_field_to_table.up.sql`
- `003_create_index_on_field.up.sql`
- `004_add_audit_trigger.up.sql`

---

## 9. Data Integrity

### 9.1 Soft Delete Pattern

All primary entities use soft delete:
```sql
deleted_at TIMESTAMP  -- NULL = active, NOT NULL = deleted
```

**Query active records:**
```sql
SELECT * FROM [table] WHERE deleted_at IS NULL;
```

**Soft delete:**
```sql
UPDATE [table] SET deleted_at = NOW() WHERE id = 'xxx';
```

**Restore:**
```sql
UPDATE [table] SET deleted_at = NULL WHERE id = 'xxx';
```

### 9.2 Referential Integrity

**Cascade Delete:**
```sql
-- Child records deleted when parent deleted
FOREIGN KEY (parent_id) REFERENCES [parent](id) ON DELETE CASCADE
```

**Restrict Delete:**
```sql
-- Cannot delete parent if children exist
FOREIGN KEY (parent_id) REFERENCES [parent](id) ON DELETE RESTRICT
```

**Set Null:**
```sql
-- Set foreign key to NULL when parent deleted
FOREIGN KEY (parent_id) REFERENCES [parent](id) ON DELETE SET NULL
```

---

## 10. Performance Considerations

### 10.1 Query Optimization

**Use Indexes:**
```sql
-- Good: Uses index
SELECT * FROM [table] WHERE status = 'active';

-- Bad: Function on indexed column (no index used)
SELECT * FROM [table] WHERE UPPER(name) = 'JOHN';

-- Good: Use functional index
CREATE INDEX idx_table_name_upper ON [table](UPPER(name));
```

**Pagination:**
```sql
-- Use LIMIT/OFFSET
SELECT * FROM [table]
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;

-- Better: Use cursor-based pagination
SELECT * FROM [table]
WHERE created_at < '2026-01-27 10:00:00'
ORDER BY created_at DESC
LIMIT 20;
```

### 10.2 Data Volume Estimates

| Table | Initial | Year 1 | Year 5 |
|-------|---------|--------|--------|
| [table_name_1] | 1,000 | 10,000 | 50,000 |
| [table_name_2] | 5,000 | 50,000 | 250,000 |

### 10.3 Partitioning (for large tables)

```sql
-- Partition by date range
CREATE TABLE [table] (
    id VARCHAR(50),
    created_at TIMESTAMP NOT NULL,
    ...
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE [table]_2026 PARTITION OF [table]
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
```

---

## 11. Security

### 11.1 Access Control

```sql
-- Read-only user
CREATE USER readonly_user WITH PASSWORD 'xxx';
GRANT SELECT ON ALL TABLES IN SCHEMA [schema] TO readonly_user;

-- Application user (read/write)
CREATE USER app_user WITH PASSWORD 'xxx';
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA [schema] TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA [schema] TO app_user;
```

### 11.2 Sensitive Data

**Encrypted Columns:**
- Passwords: Hashed with bcrypt/argon2
- API Keys: Encrypted
- PII: Consider encryption at application layer

**Row-Level Security (PostgreSQL):**
```sql
-- Enable RLS
ALTER TABLE [table] ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see their own records
CREATE POLICY user_isolation ON [table]
    FOR SELECT
    USING (user_id = current_setting('app.current_user_id')::VARCHAR);
```

---

## 12. Backup & Recovery

### 12.1 Backup Strategy
- **Full Backup:** Daily at 2 AM
- **Incremental:** Every 6 hours
- **Transaction Log:** Continuous
- **Retention:** 30 days

### 12.2 Backup Commands

**PostgreSQL:**
```bash
# Full backup
pg_dump -h localhost -U postgres -d [database] -F c -f backup_$(date +%Y%m%d).dump

# Restore
pg_restore -h localhost -U postgres -d [database] -c backup_20260127.dump
```

**MySQL:**
```bash
# Backup
mysqldump -u root -p [database] > backup_$(date +%Y%m%d).sql

# Restore
mysql -u root -p [database] < backup_20260127.sql
```

---

## 13. Maintenance

### 13.1 Regular Maintenance Tasks

**PostgreSQL:**
```sql
-- Analyze statistics
ANALYZE [table];

-- Vacuum (reclaim storage)
VACUUM [table];

-- Reindex
REINDEX TABLE [table];
```

### 13.2 Monitoring Queries

**Table Size:**
```sql
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = '[schema]'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Index Usage:**
```sql
SELECT 
    tablename,
    indexname,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
WHERE schemaname = '[schema]'
ORDER BY idx_scan ASC;
```

---

## 14. Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| [VERSION] | [DATE] | [AUTHOR] | [CHANGES] |

---

**End of Schema Document**
