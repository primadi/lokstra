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
cd cmd/lokstra
go build -o lokstra.exe .

# Add to PATH or run directly
./lokstra.exe
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

### Generate code (for Enterprise Router Service templates)

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

## What the CLI does

1. ✅ Downloads template directly from GitHub (branch: dev2)
2. ✅ Copies all template files to your project directory
3. ✅ Automatically fixes all import paths
4. ✅ Runs `go mod init <project-name>`
5. ✅ Runs `go mod tidy` to download dependencies
6. ✅ Code generation for annotation-based templates (via `autogen` command)

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
