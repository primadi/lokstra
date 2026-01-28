# Database Schema: [Module Name]
## [Project Name]

**Version:** 1.0.0  
**Status:** draft  
**Requirements Reference:** [Module] Requirements v[version]  
**Last Updated:** [Date]  
**Owner:** [Name/Team]  

---

## 1. Overview

**Database:** PostgreSQL 15+

**Schema:** `public` (or `[schema_name]`)

**Character Set:** UTF-8

**Collation:** en_US.UTF-8

---

## 2. Tables

### Table: [table_name]

**Purpose:** [Description of what this table stores]

**Estimated Size:** [X] records (Year 1), [Y] records (Year 3)

---

#### Schema

```sql
CREATE TABLE [table_name] (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Business fields
    field1 VARCHAR(100) NOT NULL,
    field2 TEXT,
    field3 DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    field4 INTEGER NOT NULL,
    
    -- Foreign keys
    related_id UUID NOT NULL REFERENCES related_table(id) ON DELETE RESTRICT,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- Soft delete (optional)
    deleted_at TIMESTAMPTZ NULL,
    
    -- Constraints
    CONSTRAINT chk_[table]_field3_positive CHECK (field3 >= 0),
    CONSTRAINT chk_[table]_field1_length CHECK (LENGTH(field1) >= 3),
    CONSTRAINT uq_[table]_field1 UNIQUE(field1)
);
```

---

#### Column Details

| Column       | Type         | Nullable | Default        | Description                     |
|--------------|--------------|----------|----------------|---------------------------------|
| `id`         | UUID         | No       | gen_random()   | Primary key                     |
| `field1`     | VARCHAR(100) | No       | -              | [Description]                   |
| `field2`     | TEXT         | Yes      | NULL           | [Description]                   |
| `field3`     | DECIMAL(10,2)| No       | 0.00           | [Description]                   |
| `field4`     | INTEGER      | No       | -              | [Description]                   |
| `related_id` | UUID         | No       | -              | FK to related_table             |
| `created_at` | TIMESTAMPTZ  | No       | NOW()          | Creation timestamp              |
| `updated_at` | TIMESTAMPTZ  | No       | NOW()          | Last update timestamp           |
| `deleted_at` | TIMESTAMPTZ  | Yes      | NULL           | Soft delete timestamp           |

---

#### Indexes

```sql
-- Primary key index (automatic)
-- pk_[table_name] on (id)

-- Foreign key indexes (for query performance)
CREATE INDEX idx_[table]_related_id ON [table_name](related_id);

-- Query optimization indexes
CREATE INDEX idx_[table]_field1 ON [table_name](field1);
CREATE INDEX idx_[table]_created_at ON [table_name](created_at DESC);

-- Composite indexes for common queries
CREATE INDEX idx_[table]_field1_created ON [table_name](field1, created_at DESC);

-- Partial indexes for active records (if using soft delete)
CREATE INDEX idx_[table]_active ON [table_name](id) WHERE deleted_at IS NULL;

-- Full-text search index (if needed)
CREATE INDEX idx_[table]_search ON [table_name] USING gin(to_tsvector('english', field1 || ' ' || COALESCE(field2, '')));
```

---

#### Constraints

**Primary Key:**
- `pk_[table_name]`: Primary key on `id`

**Foreign Keys:**
- `fk_[table]_related`: References `related_table(id)` ON DELETE RESTRICT

**Unique Constraints:**
- `uq_[table]_field1`: Unique constraint on `field1`

**Check Constraints:**
- `chk_[table]_field3_positive`: Ensures `field3 >= 0`
- `chk_[table]_field1_length`: Ensures `LENGTH(field1) >= 3`

---

#### Triggers

```sql
-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_[table]_updated_at
    BEFORE UPDATE ON [table_name]
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## 3. Relationships

### Entity Relationship Diagram

```
[table_name]
    |
    |-- belongs to --> related_table
    |
    |-- has many --> dependent_table
```

### Relationship Details

| From Table   | Relationship | To Table     | FK Column    | On Delete   |
|--------------|--------------|--------------|--------------|-------------|
| [table_name] | Many-to-One  | related_table| related_id   | RESTRICT    |
| dependent_tbl| Many-to-One  | [table_name] | [table]_id   | CASCADE     |

---

## 4. Views

### View: [view_name]

**Purpose:** [Description]

```sql
CREATE OR REPLACE VIEW [view_name] AS
SELECT 
    t1.id,
    t1.field1,
    t1.field2,
    t2.related_field,
    t1.created_at
FROM [table_name] t1
INNER JOIN related_table t2 ON t1.related_id = t2.id
WHERE t1.deleted_at IS NULL;
```

---

## 5. Stored Procedures

### Procedure: [procedure_name]

**Purpose:** [Description]

**Parameters:**
- `p_param1` (UUID): [Description]
- `p_param2` (VARCHAR): [Description]

**Returns:** [Return type/description]

```sql
CREATE OR REPLACE FUNCTION [procedure_name](
    p_param1 UUID,
    p_param2 VARCHAR
)
RETURNS TABLE (
    id UUID,
    field1 VARCHAR,
    field2 TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT t.id, t.field1, t.field2
    FROM [table_name] t
    WHERE t.related_id = p_param1
      AND t.field1 ILIKE '%' || p_param2 || '%'
      AND t.deleted_at IS NULL
    ORDER BY t.created_at DESC;
END;
$$ LANGUAGE plpgsql;
```

---

## 6. Sequences

```sql
-- Only if not using UUID
CREATE SEQUENCE [sequence_name]
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
```

---

## 7. Data Types Reference

| Application Type | PostgreSQL Type | Notes                          |
|------------------|-----------------|--------------------------------|
| ID               | UUID            | Use gen_random_uuid()          |
| String (short)   | VARCHAR(n)      | For constrained strings        |
| String (long)    | TEXT            | For unlimited text             |
| Integer          | INTEGER         | 4 bytes, -2B to +2B            |
| Big Integer      | BIGINT          | 8 bytes                        |
| Decimal          | DECIMAL(p,s)    | For money, precise numbers     |
| Boolean          | BOOLEAN         | TRUE/FALSE/NULL                |
| Date             | DATE            | Date only                      |
| DateTime         | TIMESTAMPTZ     | Timestamp with timezone        |
| JSON             | JSONB           | Binary JSON (prefer over JSON) |
| Array            | ARRAY           | PostgreSQL array type          |

---

## 8. Indexes Strategy

### Index Types

1. **B-Tree (Default)**
   - Use for: equality, range queries
   - Columns: IDs, dates, numbers, strings

2. **GIN (Generalized Inverted Index)**
   - Use for: full-text search, JSONB, arrays
   - Columns: JSONB fields, text search

3. **BRIN (Block Range Index)**
   - Use for: large tables with natural ordering
   - Columns: timestamps, sequential IDs

### Indexing Guidelines

- Index foreign keys (for JOIN performance)
- Index frequently filtered columns
- Index columns used in ORDER BY
- Use composite indexes for multi-column queries
- Use partial indexes for subset queries
- Monitor and remove unused indexes

---

## 9. Migration Files

### Migration: 001_create_[table_name]_table.sql

```sql
-- Migration: Create [table_name] table
-- Module: [module_name]
-- Version: 1.0.0
-- Date: [Date]

-- UP
CREATE TABLE IF NOT EXISTS [table_name] (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field1 VARCHAR(100) NOT NULL,
    field2 TEXT,
    related_id UUID NOT NULL REFERENCES related_table(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_[table]_field1_length CHECK (LENGTH(field1) >= 3)
);

-- Indexes
CREATE INDEX idx_[table]_related_id ON [table_name](related_id);
CREATE INDEX idx_[table]_created_at ON [table_name](created_at DESC);

-- Triggers
CREATE TRIGGER trg_[table]_updated_at
    BEFORE UPDATE ON [table_name]
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- DOWN
DROP TABLE IF EXISTS [table_name] CASCADE;
```

---

## 10. Performance Considerations

### Query Optimization

**Slow Query Example:**
```sql
-- Avoid: Full table scan
SELECT * FROM [table_name] WHERE field1 LIKE '%keyword%';
```

**Optimized Query:**
```sql
-- Better: Use index
SELECT * FROM [table_name] 
WHERE field1 ILIKE 'keyword%'  -- Uses index
LIMIT 100;

-- Best: Use full-text search index
SELECT * FROM [table_name]
WHERE to_tsvector('english', field1) @@ to_tsquery('keyword')
LIMIT 100;
```

### Connection Pooling

- **Min Connections:** 5
- **Max Connections:** 50
- **Idle Timeout:** 30 seconds
- **Max Lifetime:** 1 hour

### Partitioning Strategy

For tables with > 100M records, consider partitioning by:
- Date range (e.g., monthly partitions)
- Hash (e.g., by user_id)

---

## 11. Backup & Recovery

### Backup Strategy

- **Full Backup:** Daily at 2 AM UTC
- **Incremental Backup:** Every 6 hours
- **Retention:** 30 days
- **Testing:** Monthly restore test

### Point-in-Time Recovery

- **WAL Archiving:** Enabled
- **Recovery Window:** 7 days

---

## 12. Security

### Permissions

```sql
-- Create role for application
CREATE ROLE [app_role] WITH LOGIN PASSWORD '[secure_password]';

-- Grant permissions
GRANT CONNECT ON DATABASE [database_name] TO [app_role];
GRANT USAGE ON SCHEMA public TO [app_role];
GRANT SELECT, INSERT, UPDATE, DELETE ON [table_name] TO [app_role];
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO [app_role];
```

### Row-Level Security (if needed)

```sql
-- Enable RLS
ALTER TABLE [table_name] ENABLE ROW LEVEL SECURITY;

-- Create policy
CREATE POLICY [policy_name] ON [table_name]
    FOR ALL
    USING (created_by = current_setting('app.user_id')::UUID);
```

---

## Appendix

### A. Glossary

- **UUID**: Universally Unique Identifier
- **TIMESTAMPTZ**: Timestamp with timezone
- **JSONB**: Binary JSON format (more efficient than JSON)
- **GIN**: Generalized Inverted Index
- **WAL**: Write-Ahead Log

### B. Change Log

| Version | Date   | Author | Changes                |
|---------|--------|--------|------------------------|
| 1.0.0   | [Date] | [Name] | Initial schema design  |

### C. References

- Module Requirements: [Link]
- API Specification: [Link]
- PostgreSQL Documentation: https://www.postgresql.org/docs/
