---
name: lokstra-schema-design
description: Generate PostgreSQL database schemas for Lokstra modules. Creates tables, indexes, constraints, triggers, and migration files. Use after module requirements and API specs are approved to design data layer.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  database: postgresql
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---

# Lokstra Schema Design

## When to use this skill

Use this skill when:
- Module requirements and API specs are approved
- Need database table definitions
- Creating migration files
- Defining indexes and constraints for performance

## How to generate database schema

Save to: `docs/modules/{module-name}/SCHEMA.md`

### Key Elements

1. **Table Definitions** - Columns, types, constraints
2. **Indexes** - For query performance
3. **Foreign Keys** - Relationships between tables
4. **Triggers** - Auto-update timestamps
5. **Migration Files** - UP/DOWN SQL

### Table Definition Format

```sql
CREATE TABLE {table_name} (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Business fields
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Foreign keys
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_{table}_name_length CHECK (LENGTH(name) >= 3),
    CONSTRAINT uq_{table}_email UNIQUE(email)
);

-- Indexes
CREATE INDEX idx_{table}_user_id ON {table_name}(user_id);
CREATE INDEX idx_{table}_created_at ON {table_name}(created_at DESC);

-- Triggers
CREATE TRIGGER trg_{table}_updated_at
    BEFORE UPDATE ON {table_name}
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Migration File Format

Create: `migrations/{module}/001_create_{table}.sql`

```sql
-- Migration: Create {table} table
-- Module: {module-name}
-- Version: 1.0.0
-- Date: 2026-01-28

-- UP
CREATE TABLE IF NOT EXISTS {table_name} (
    -- table definition
);

-- Indexes
CREATE INDEX...;

-- Triggers
CREATE TRIGGER...;

-- DOWN
DROP TABLE IF EXISTS {table_name} CASCADE;
```

## Resources

- **Template:** [references/SCHEMA_TEMPLATE.md](references/SCHEMA_TEMPLATE.md)
- **PostgreSQL Docs:** https://www.postgresql.org/docs/
- **Best Practices:** Use UUID for IDs, TIMESTAMPTZ for dates
