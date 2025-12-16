# Migration Runner Improvements

## ✅ Implemented Improvements

### 1. **Version Conflict Detection**

**Problem:** Multiple developers could create migrations with the same version number but different descriptions, causing silent data loss.

**Solution:** Added strict validation during `Load()`:

```go
// Example conflict scenario:
// Developer A: 001_create_users.up.sql
// Developer B: 001_create_orders.up.sql

// Error output:
❌ Migration conflict detected for version 001:
  Found: 001_create_users.*.sql
  Found: 001_create_orders.*.sql
  → Same version cannot have different descriptions!
  → Please rename one of the migrations to use a different version number.
```

**Benefits:**
- ✅ Detects conflicts immediately
- ✅ Clear error message with resolution steps
- ✅ Prevents silent migration loss
- ✅ Works for both UP and DOWN migrations

---

### 2. **Better Error Messages**

**Before:**
```
Error: no DOWN SQL for migration 2
```

**After:**
```
❌ Cannot rollback migration 002_add_indexes
  → Missing file: 002_add_indexes.down.sql
  → Create the DOWN migration file or manually remove this version from schema_migrations table
```

**Improvements:**
- ✅ Shows version with leading zeros (002 vs 2)
- ✅ Includes migration description
- ✅ Shows exact filename needed
- ✅ Provides actionable solutions
- ✅ Uses emoji for better readability

---

### 3. **Create Command (Auto-versioning)**

**Feature:** Auto-generate migration file pairs with proper versioning.

```bash
# Command
go run main.go --cmd=create --name="create_users_table"

# Output
✅ Created migration files:
   → 001_create_users_table.up.sql
   → 001_create_users_table.down.sql

📝 Next steps:
   1. Edit the migration files with your SQL
   2. Run: go run main.go --cmd=up
```

**How it works:**
1. Scans existing migrations to find highest version
2. Auto-increments to next version (001 → 002 → 003...)
3. Creates both UP and DOWN files with templates
4. Validates migration name format (snake_case)
5. Prevents accidental overwrites

**Benefits:**
- ✅ No manual version numbering
- ✅ No version conflicts (uses next available)
- ✅ Helpful SQL templates included
- ✅ Creates both files atomically
- ✅ Validates naming conventions

---

## Additional Validations

### Duplicate File Detection

```go
// Prevents duplicate UP or DOWN files for same version
❌ Duplicate UP migration file for version 001_create_users
  → Only one .up.sql file allowed per version
```

### Name Format Validation

```bash
# Invalid names are rejected
go run main.go --cmd=create --name="Create Users"

❌ Invalid migration name: 'Create Users'
  → Use snake_case with lowercase letters, numbers, and underscores only
  → Example: create_users_table, add_email_index
```

---

## Migration File Templates

**UP Migration Template:**
```sql
-- Migration: create_users_table
-- Created: auto-generated
-- Version: 001

-- Add your UP migration SQL here
-- Example:
-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     email VARCHAR(255) UNIQUE NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
```

**DOWN Migration Template:**
```sql
-- Migration: create_users_table (ROLLBACK)
-- Created: auto-generated
-- Version: 001

-- Add your DOWN migration SQL here (reverse of UP migration)
-- Example:
-- DROP TABLE IF EXISTS users;
```

---

## Usage Examples

### Typical Workflow

```bash
# 1. Create new migration
go run main.go --cmd=create --name="create_users_table"

# 2. Edit generated files (add your SQL)
# Edit: migrations/001_create_users_table.up.sql
# Edit: migrations/001_create_users_table.down.sql

# 3. Run migration
go run main.go --cmd=up

# 4. Check status
go run main.go --cmd=status

# 5. Rollback if needed
go run main.go --cmd=down
```

### Team Collaboration Scenario

**Developer A (morning):**
```bash
go run main.go --cmd=create --name="create_users_table"
# Creates: 001_create_users_table.*.sql
git add migrations/
git commit -m "Add user table migration"
```

**Developer B (afternoon, before pulling):**
```bash
go run main.go --cmd=create --name="create_orders_table"
# Creates: 001_create_orders_table.*.sql (same version!)
git pull  # Gets Developer A's changes
```

**Developer B (after pull):**
```bash
go run main.go --cmd=up
# Error: Version conflict detected!
# Fix: Rename to version 002
mv migrations/001_create_orders_table.up.sql migrations/002_create_orders_table.up.sql
mv migrations/001_create_orders_table.down.sql migrations/002_create_orders_table.down.sql

go run main.go --cmd=up
# Success!
```

---

## Summary

| Feature | Before | After |
|---------|--------|-------|
| **Conflict Detection** | ❌ Silent overwrite | ✅ Error with clear message |
| **Error Messages** | ❌ Cryptic | ✅ Actionable with solutions |
| **File Creation** | ⚠️ Manual | ✅ Auto-generated with `create` |
| **Version Numbering** | ⚠️ Manual | ✅ Auto-incremented |
| **Duplicate Detection** | ❌ None | ✅ Enforced |
| **Name Validation** | ❌ None | ✅ Snake_case required |
| **Templates** | ❌ None | ✅ Helpful SQL templates |

All improvements maintain backward compatibility with existing migration files! 🎉
