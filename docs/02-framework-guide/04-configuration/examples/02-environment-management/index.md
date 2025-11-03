# Environment Management

Manage different configurations for dev, staging, and production environments.

## Running

```bash
# Development
APP_ENV=development go run main.go

# Staging
APP_ENV=staging go run main.go

# Production
APP_ENV=production go run main.go
```

Server starts on `http://localhost:3020`

## Configuration Files

- `config-base.yaml` - Shared base configuration
- `config-development.yaml` - Development overrides
- `config-staging.yaml` - Staging overrides
- `config-production.yaml` - Production overrides

## Pattern

1. Load base configuration
2. Load environment-specific overrides
3. Merge configurations
4. Start application

## Benefits

- Single codebase for all environments
- Environment-specific settings
- Easy to add new environments
- Clear separation of concerns
