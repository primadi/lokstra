# Example 02 - Multi-File Configuration

Demonstrates splitting configuration across multiple files and merging them.

## What's Demonstrated

- ✅ Multi-file configuration
- ✅ Base + environment-specific configs
- ✅ Config merging strategy
- ✅ Environment-specific overrides
- ✅ Shared vs deployment-specific definitions

## File Structure

```
02-multi-file/
├── main.go
└── config/
    ├── base.yaml        # Shared configuration
    ├── dev.yaml         # Development environment
    └── production.yaml  # Production environment
```

## Configuration Strategy

### 1. Base Configuration (base.yaml)
Contains shared configuration across all environments:
- Global configs
- Common middleware definitions
- Service definitions
- Router definitions

### 2. Environment-Specific (dev.yaml, production.yaml)
Contains environment-specific configuration:
- Deployment definitions
- Config overrides
- Environment-specific middleware
- Handler configurations

## Merge Behavior

```go
lokstra.RunFromConfig(
    "config/base.yaml",      // Loaded first
    "config/dev.yaml",       // Overrides base
)
```

**Merge rules:**
- **Maps**: Deep merged (nested keys preserved)
- **Arrays**: Replaced (not merged)
- **Primitives**: Replaced

**Example:**

**base.yaml:**
```yaml
configs:
  app:
    name: "MyApp"
    version: "1.0.0"
  database:
    host: "localhost"
```

**dev.yaml:**
```yaml
configs:
  app:
    debug: true  # Added to app
  database:
    host: "dev-db"  # Overrides host
```

**Result:**
```yaml
configs:
  app:
    name: "MyApp"      # From base
    version: "1.0.0"   # From base
    debug: true        # From dev
  database:
    host: "dev-db"     # Overridden by dev
```

## Development vs Production

### Development (dev.yaml)

**Focus:**
- Debug logging
- Local database
- Dev tools mounted
- Detailed error messages

```yaml
deployments:
  development:
    config-overrides:
      app:
        debug: true
        log-level: "debug"
```

### Production (production.yaml)

**Focus:**
- Minimal logging
- Environment variables for secrets
- Rate limiting
- Reverse proxies
- Production SPAs

```yaml
deployments:
  production:
    config-overrides:
      app:
        debug: false
        log-level: "info"
      database:
        host: "${DB_HOST}"
```

## Loading Strategy

### Option 1: Build-time selection
```bash
# Development
go build -tags dev
./app  # Uses dev.yaml

# Production
go build -tags prod
./app  # Uses production.yaml
```

### Option 2: Runtime selection
```go
env := os.Getenv("ENV")
if env == "" {
    env = "dev"
}

configFiles := []string{
    "config/base.yaml",
    fmt.Sprintf("config/%s.yaml", env),
}

lokstra.RunFromConfig(configFiles...)
```

### Option 3: Load all from folder
```go
// Loads all .yaml files in order
lokstra.RunFromConfigFolder("config")
```

## Best Practices

### 1. Organize by Concern
```
config/
├── base.yaml           # Shared
├── services.yaml       # Service definitions
├── middleware.yaml     # Middleware definitions
└── deployments/
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

### 2. Use Environment Variables for Secrets
```yaml
# ❌ Don't hardcode secrets
database:
  password: "secret123"

# ✅ Use environment variables
database:
  password: "${DB_PASSWORD}"
```

### 3. Override Only What's Different
```yaml
# base.yaml - all shared config
# dev.yaml - only development overrides
# prod.yaml - only production overrides
```

### 4. Validate Before Deploy
```bash
# Check which config is loaded
go run main.go --print-config

# Validate YAML syntax
yamllint config/*.yaml
```

## Run

```bash
# Development
go run main.go  # Uses base.yaml + dev.yaml

# Production (change main.go to load production.yaml)
ENV=production go run main.go
```

## Summary

Multi-file configuration allows you to:
- ✅ Share common config across environments
- ✅ Override only what's different per environment
- ✅ Keep secrets in environment variables
- ✅ Organize config by concern
- ✅ Version control environment configs separately
