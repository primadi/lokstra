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
# Lokstra Configuration Schema

This directory contains JSON schema definitions for validating Lokstra configuration files.

## Files

- `lokstra.json` - Main JSON schema file that defines the structure and validation rules for Lokstra YAML configuration files
- `group-include.json` - Group include configuration schema

## Schema Features

The `lokstra.json` schema provides comprehensive validation for:

### Service Configuration (`ServiceConfig`)

The schema supports conditional validation based on service type. Each service type has its own specific configuration properties:

#### PostgreSQL Database Pool (`lokstra.db_pool.pg`)
- **Connection Options**: DSN string or individual parameters (host, port, database, username, password)
- **Pool Settings**: min_connections, max_connections, max_idle_time, max_lifetime
- **SSL Configuration**: sslmode/ssl_mode with enum validation
- **Advanced Options**: tenant_mode, default_schema, connection_timeout, query_timeout

#### Redis (`lokstra.redis`)
- **Connection**: addr, host, port, password, db
- **Connection Pool**: pool_size, min_idle_conns, max_conn_age
- **Timeouts**: pool_timeout, idle_timeout, read_timeout, write_timeout

#### Logger (`lokstra.logger`)
- **Basic Settings**: level (debug/info/warn/error/fatal/panic), format (json/text/console), output (stdout/stderr/file)
- **File Rotation**: file_path, max_size, max_backups, max_age, compress
- **Advanced Options**: caller, stacktrace

#### Metrics (`lokstra.metrics`)
- **Collection**: enabled, endpoint, namespace, subsystem
- **Configuration**: buckets (histogram), labels, collect_interval

#### Internationalization (`lokstra.i18n`)
- **Language Settings**: default_language, supported_languages, fallback_language
- **File Configuration**: messages_path, format (json/yaml/toml)
- **Options**: auto_reload, case_sensitive

#### JWT Authentication (`lokstra.jwt_auth`)
- **Token Configuration**: secret, algorithm, expires_in, refresh_expires_in
- **Claims**: issuer, audience
- **Key Files**: public_key_path, private_key_path (for RSA/ECDSA)
- **HTTP Integration**: token_header, token_prefix, skip_paths

#### HTTP Listener (`lokstra.http_listener`)
- **Basic Settings**: addr, host, port
- **Timeouts**: read_timeout, write_timeout, idle_timeout, shutdown_timeout
- **Advanced**: max_header_bytes
- **TLS Configuration**: enabled, cert_file, key_file, min_version, max_version, cipher_suites
- **CORS Support**: allowed_origins, allowed_methods, allowed_headers, exposed_headers, allow_credentials, max_age

### Additional Configuration Types

- `ModuleConfig` - External module loading and configuration
- `MountStaticConfig` - Static file serving
- `MountSPAConfig` - Single Page Application serving
- `MountReverseProxyConfig` - Reverse proxy configuration
- `MountRpcServiceConfig` - RPC service mounting

## Usage

### IDE Integration

Most modern IDEs with YAML support can use this schema for:
- **Autocompletion**: Get suggestions for configuration properties
- **Validation**: Real-time validation while editing YAML files
- **Documentation**: Hover tooltips showing property descriptions

To enable schema validation in your IDE, add this to the top of your YAML configuration file:

```yaml
# yaml-language-server: $schema=../../../schema/lokstra.json
```

### Manual Validation

You can validate configuration files programmatically using JSON schema validation libraries.

## Examples

### PostgreSQL Service Configuration

```yaml
services:
  - name: "main_db"
    type: "lokstra.db_pool.pg"
    config:
      host: "localhost"
      port: 5432
      database: "myapp"
      username: "postgres"
      password: "secret"
      max_connections: 20
      min_connections: 2
      sslmode: "require"
      tenant_mode: true
```

### HTTP Listener with TLS

```yaml
services:
  - name: "web_server"
    type: "lokstra.http_listener"
    config:
      port: 8443
      read_timeout: "10s"
      write_timeout: "10s"
      tls:
        enabled: true
        cert_file: "./certs/server.crt"
        key_file: "./certs/server.key"
        min_version: "1.2"
      cors:
        enabled: true
        allowed_origins: ["https://app.example.com"]
        allowed_methods: ["GET", "POST", "PUT", "DELETE"]
        allow_credentials: true
```

### JWT Authentication

```yaml
services:
  - name: "jwt_auth"
    type: "lokstra.jwt_auth"
    config:
      secret: "your-secret-key"
      algorithm: "HS256"
      expires_in: "24h"
      issuer: "myapp"
      skip_paths: ["/login", "/register", "/health"]
```

## Schema Validation Rules

### Duration Format
Many timeout and interval fields use Go duration format:
- Pattern: `^[0-9]+(ns|us|µs|ms|s|m|h)$`
- Examples: `"5s"`, `"30m"`, `"1h"`, `"2h30m"`

### Required Fields
The schema uses conditional validation with `anyOf` to allow different configuration styles:
- PostgreSQL: Either `dsn` OR (`host` + `database` + `username`)
- JWT Auth: Either `secret` OR (`public_key_path` + `private_key_path`)

### Enums
Many fields have restricted values:
- Log levels: `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- SSL modes: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`
- JWT algorithms: `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `ES256`, `ES384`, `ES512`

## Contributing

When adding new service types or modifying existing ones:

1. Add the service type to the `allOf` array in `ServiceConfig`
2. Define the conditional schema with `if/then` structure
3. Include comprehensive property descriptions and examples
4. Add appropriate validation rules (patterns, enums, ranges)
5. Test the schema with real configuration files
6. Update this documentation with examples

## Schema Versioning

The schema follows semantic versioning principles:
- Major version: Breaking changes to existing service configurations
- Minor version: New service types or optional properties
- Patch version: Bug fixes, improved descriptions, or additional examples

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
