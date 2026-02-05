# Lokstra CLI

Command-line tool for creating new Lokstra projects from templates.

## Installation

### Install from source

```bash
go install github.com/primadi/lokstra/cmd/lokstra@latest
```

### Or build locally

```bash
# From lokstra project root
go install ./cmd/lokstra

```

## Usage

### Create a new project (interactive)

```bash
lokstra new myproject
```

This will show an interactive menu to select a template.

### Create a new project with specific template

```bash
lokstra new myproject -template 01_router/01_router_only
lokstra new myapp -template 02_app_framework/01_medium_system
lokstra new enterprise-app -template 02_app_framework/03_enterprise_router_service
```

### Generate code from @Handler, @Service annotation (for app framework)

```bash
# Generate code in current directory
lokstra autogen

# Generate code in specific folder
lokstra autogen ./myproject
lokstra autogen c:\path\to\project
```

This command is equivalent to running `go run . --generate-only` and is useful for:
- Enterprise Router Service templates that use annotations
- Regenerating routers after changing annotations
- CI/CD pipelines for code generation

### Auto code generation from @Handler, @Service annotation

When running in development mode (starting project inside VSCode or running `go run .`), the following files are automatically generated:

1. **`zz_lokstra_imports.go`** - Generated at the project root
   - Contains automatic imports for all services with `@Handler` or `@Service` annotations
   - Ensures all annotated services are registered with the framework
   - Automatically updated when new annotations are detected

2. **Per-module generated files** - Created in each folder containing `@Handler` or `@Service` annotations:
   - **`zz_cache.lokstra.json`** - Cache file containing metadata about annotations
     - Tracks annotation changes to determine when regeneration is needed
     - Contains parsed annotation data for faster subsequent builds
   - **`zz_generated.lokstra.go`** - Generated code for service registration and routing
     - Contains service factory functions
     - Contains route registration code
     - Should not be manually edited (regenerated automatically)

**Note:** All `zz_*.go` and `zz_*.json` files are auto-generated. Do not modify them manually as they will be overwritten on the next build.

### Generate code (alias)

```bash
# Alias for autogen command
lokstra generate

# Generate code in specific folder
lokstra generate ./myproject
lokstra generate c:\path\to\project
```

`lokstra generate` is an alias for `lokstra autogen` - both commands do the same thing.

### Update AI skills and templates

```bash
# Update skills in current project
lokstra update-skills

# Update skills in specific project
lokstra update-skills ./myproject
lokstra update-skills c:\path\to\project

# Use different branch
lokstra update-skills -branch main
```

This command updates the following files in your project:
- `.github/skills/` - All AI agent skill files
- `.github/copilot-instructions.md` - Copilot configuration
- `docs/templates/` - Document templates (BRD, API Spec, etc.)

**When to use:**
- Update skills to latest version from Lokstra framework
- Get new AI capabilities added to the framework
- Sync your project with latest best practices
- After framework updates

**Note:** Existing files are backed up to `.github/skills.backup/` before updating.

### Database migrations

```bash
# Create new migration
lokstra migration create create_users_table

# Create migration in specific directory
lokstra migration create create_users_table -dir migrations/auth

# Run pending migrations
lokstra migration up

# Run migrations for specific database
lokstra migration up -db replica-db

# Rollback last migration
lokstra migration down

# Rollback multiple migrations
lokstra migration down -steps 3

# Show migration status
lokstra migration status

# Show current version
lokstra migration version
```

**Migration Flags:**
- `-dir <path>` - Migrations directory (default: `migrations`)
- `-db <name>` - Database pool name from config.yaml (used when `migration.yaml` is missing or doesn't specify `dbpool-name`; default: `db_main`)
- `-steps <n>` - Number of migrations to rollback (default: 1)
- `-config <file>` - Config file path (default: `config.yaml`)

**Database Pool Resolution:**

The database pool name is resolved in this order:
1. If `{dir}/migration.yaml` exists and has `dbpool-name` → use that
2. Otherwise → use `-db <name>`

**Two Migration Strategies:**

1. **Single Database (All Migrations in One Folder)**
   ```
   migrations/
   ├── 001_create_users.up.sql
   ├── 001_create_users.down.sql
   ├── 002_create_products.up.sql
   └── 002_create_products.down.sql
   ```
   
   Use: `lokstra migration up`

2. **Multi-Database (Migrations Per Module/Database)**
   ```
   migrations/
   ├── 01_main-db/
   │   ├── migration.yaml          # Required
   │   ├── 001_create_users.up.sql
   │   └── 001_create_users.down.sql
   ├── 02_tenant-db/
   │   ├── migration.yaml          # Required
   │   ├── 001_create_tenants.up.sql
   │   └── 001_create_tenants.down.sql
   └── 03_ledger-db/
       ├── migration.yaml          # Required
       ├── 001_create_accounts.up.sql
       └── 001_create_accounts.down.sql
   ```
   
   **migration.yaml example:**
   ```yaml
   dbpool-name: main-db      # From config.yaml service-definitions
   schema-table: schema_migrations
   enabled: true
   description: Main application database
   ```
   
   Use: `lokstra migration up -dir migrations/01_main-db`

**Notes:**
- Each subfolder with `migration.yaml` is treated as separate database
- Subfolders without `migration.yaml` are ignored
- Use numeric prefixes (01_, 02_) for execution order
- Migration files format: `{version}_{name}.{up|down}.sql`

**migration.yaml fields:**
- `dbpool-name` (recommended) - Database pool name from config.yaml `service-definitions`
- `schema-table` (optional) - Migration tracking table (default: `schema_migrations`)
- `enabled` (optional) - If `false`, `up`/`down` will be skipped
- `description` (optional) - Documentation only

### Use different branch

```bash
lokstra new myproject -template 01_router/01_router_only -branch main
```

### Show version

```bash
lokstra version
```

### Show help

```bash
lokstra help
```

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `new` | Create new project from template | `lokstra new myapp` |
| `autogen` | Generate code from annotations | `lokstra autogen` |
| `generate` | Alias for autogen | `lokstra generate` |
| `update-skills` | Update AI skills and templates | `lokstra update-skills` |
| `migration` | Manage database migrations | `lokstra migration up` |
| `version` | Show CLI version | `lokstra version` |
| `help` | Show help information | `lokstra help` |

## Available Templates

### Router Patterns

1. **01_router/01_router_only**
   - Pure router with CRUD operations
   - Best for: Learning Lokstra routing basics

2. **01_router/02_single_app**
   - App wrapper with graceful shutdown
   - Best for: Single application servers

3. **01_router/03_multi_app**
   - Multiple apps on different ports
   - Best for: Admin/API separation

### Framework Patterns

4. **02_app_framework/01_medium_system**
   - Flat domain-driven structure
   - Best for: 2-10 entities, single team

5. **02_app_framework/02_enterprise_modular**
   - DDD with bounded contexts
   - Best for: 10+ entities, multiple teams

6. **02_app_framework/03_enterprise_router_service**
   - DDD with annotation-based router service
   - Best for: Enterprise scale applications

### AI-Driven Development

7. **03_ai_driven/01_starter (Recommended)**
   - Design-first development with AI agent skills
   - Generate code from specifications
   - Best for: Production applications, all scales

## What the CLI does

1. ✅ Downloads template directly from GitHub (branch: dev2)
2. ✅ Copies all template files to your project directory
3. ✅ Copies AI agent skills to `.github/skills/` (for AI-driven templates)
4. ✅ Automatically fixes all import paths
5. ✅ Runs `go mod init <project-name>`
6. ✅ Runs `go mod tidy` to download dependencies
7. ✅ Code generation for annotation-based templates (via `autogen` command)
8. ✅ Updates AI skills and templates (via `update-skills` command)

## After Creating a Project

```bash
cd myproject
go run .
```

The server will start and show you the available endpoints.

## Example

```bash
# Create a medium system project
lokstra new blog-api -template 02_app_framework/01_medium_system

# Output:
# 🚀 Creating new Lokstra project: blog-api
#
# 📦 Selected template: 02_app_framework/01_medium_system
# 🌿 Branch: dev2
#
# 📥 Downloading template from GitHub...
# 📋 Copying template files...
# 🔧 Fixing imports...
# 📦 Initializing Go module...
# 🧹 Running go mod tidy...
#
# ✅ Project created successfully!
#
# Next steps:
#   cd blog-api
#   go run .

cd blog-api
go run .
```

## Requirements

- Go 1.23 or higher
- Internet connection (to download templates from GitHub)

## Troubleshooting

### Error: directory already exists

The CLI won't overwrite existing directories. Choose a different project name or remove the existing directory.

### Error: failed to download template

Check your internet connection and verify that the branch name is correct (default: dev2).

### Error: template not found

Verify the template path is correct. Use `lokstra new myproject` without the `-template` flag to see all available templates interactively.

## Development

### Building

```bash
cd cmd/lokstra
go build -o lokstra.exe .
```

### Testing locally

```bash
# Test creating a project
./lokstra.exe new test-project

# Clean up
rm -rf test-project
```

## How It Works

1. **Download**: Downloads the entire Lokstra repository as a tar.gz archive from GitHub
2. **Extract**: Extracts the specific template folder from the archive
3. **Copy**: Copies all files to your new project directory
4. **Fix Imports**: Automatically replaces all imports like:
   - From: `github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain`
   - To: `myproject/domain`
5. **Initialize**: Runs `go mod init` with your project name
6. **Tidy**: Runs `go mod tidy` to download all dependencies

## License

Part of the Lokstra framework. See LICENSE file in project root.
