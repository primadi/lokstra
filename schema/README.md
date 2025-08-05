# Lokstra YAML Configuration Schema

This directory contains the JSON Schema for Lokstra framework YAML configuration files.

## Usage

### VS Code with YAML Language Server

Add the following to your VS Code `settings.json`:

```json
{
  "yaml.schemas": {
    "./schema/lokstra.json": [
      "**/configs/**/*.yaml",
      "**/config/**/*.yaml",
      "lokstra.yaml",
      "server.yaml",
      "**/lokstra/**/*.yaml"
    ],
    "./schema/group-include.json": [
      "**/*group*.yaml",
      "**/*include*.yaml",
      "**/group-*.yaml",
      "**/include-*.yaml"
    ]
  },
  "yaml.format.enable": true,
  "yaml.validate": true,
  "yaml.completion": true,
  "yaml.hover": true
}
```

### Direct Schema Reference

Add to the top of your YAML files:

```yaml
# yaml-language-server: $schema=./schema/lokstra.json

server:
  name: my-server
  # IntelliSense will provide auto-completion here
```

### Other Editors

Most modern editors with YAML language server support can use this schema:

- **JetBrains IDEs** (IntelliJ, WebStorm, etc.): Add schema mapping in Settings → JSON Schema Mappings
- **Vim/Neovim** with coc-yaml or vim-lsp
- **Emacs** with lsp-mode
- **Sublime Text** with LSP package

## Schema Features

The schema provides:

✅ **Auto-completion** for all configuration keys  
✅ **Validation** with error highlighting  
✅ **Documentation** via hover tooltips  
✅ **Type checking** for values  
✅ **Enum validation** for restricted values (e.g., HTTP methods, log levels)

## Configuration Structure

The schema supports the complete Lokstra configuration structure:

### Server Configuration
- Server name and global settings
- Log level configuration

### Apps Configuration  
- HTTP listener types (default, fasthttp, http3, secure)
- Router engine types
- Address binding (TCP ports, Unix sockets)
- Route definitions with middleware
- Static file serving and SPA mounting
- Reverse proxy configuration

### Services Configuration
- Service type definitions
- Configuration parameters
- Dependency management

### Modules Configuration
- External module loading
- Plugin configurations
- Permission systems

### Middleware Configuration
- Built-in middleware (cors, logger, recovery, etc.)
- Custom middleware with configuration
- Per-route and per-group middleware

## Examples

See the `cmd/examples/01_basic_overview/03_with_yaml_config/configs/` directory for complete working examples with schema validation.

## Group Include Files

Lokstra supports modular configuration through the `load_from` feature in groups. This allows you to split large configurations into smaller, reusable files.

### Group Include Schema

The `schema/group-include.json` provides validation for files loaded via `load_from`:

```yaml
# yaml-language-server: $schema=./schema/group-include.json

# This file can be loaded via load_from in a group configuration
routes:
  - method: GET
    path: /users
    handler: user.list
  - method: POST
    path: /users
    handler: user.create

groups:
  - prefix: /admin
    routes:
      - method: GET
        path: /stats
        handler: admin.stats
    middleware:
      - name: admin_auth

mount_static:
  - prefix: /uploads
    folder: ./uploads
```

### Usage in Main Configuration

```yaml
apps:
  - name: api-app
    address: ":8080"
    groups:
      - prefix: /api/v1
        load_from:
          - api-routes.yaml
          - admin-routes.yaml
        middleware:
          - name: api_auth
```

### Restrictions

Group include files have specific restrictions to maintain configuration clarity:

❌ **`prefix`** - Not allowed at root level (define groups inside instead)  
❌ **`middleware`** - Not allowed at root level (define groups inside instead)  
❌ **`override_middleware`** - Not allowed at root level  

✅ **`routes`** - Allowed at root level  
✅ **`groups`** - Allowed at root level with their own prefixes  
✅ **`mount_*`** - All mount types allowed at root level  

## Contributing

When adding new configuration options to Lokstra:

1. Update the Go structs in `core/config/types.go`
2. Update this JSON schema in `schema/lokstra.json`
3. Add examples to demonstrate the new configuration
4. Test with VS Code YAML language server

## Schema Validation

The schema follows JSON Schema Draft 7 specification and can be validated using any JSON Schema validator.
