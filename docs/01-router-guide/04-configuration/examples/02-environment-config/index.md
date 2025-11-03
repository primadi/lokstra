# Environment-Based Configuration Example

Demonstrates loading multiple configuration files based on environment.

## What This Demonstrates

- Base configuration with environment variables
- Environment-specific overrides (dev, prod)
- Multi-file configuration merging
- Runtime environment detection

## Files

- `config/base.yaml` - Shared base configuration
- `config/dev.yaml` - Development environment overrides
- `config/prod.yaml` - Production environment overrides
- `main.go` - Application code with environment detection
- `test.http` - HTTP requests for testing (use with VS Code REST Client)

## Run

### Development (default)
```bash
go run main.go
# or explicitly
APP_ENV=dev go run main.go
```

### Production
```bash
APP_ENV=prod go run main.go
```

### With environment variables
```bash
APP_NAME=my-service APP_HOST=0.0.0.0 APP_PORT=9000 APP_ENV=prod go run main.go
```

## Test

### Development
```bash
# Default port is 3000 in dev
curl http://localhost:3000/health
curl http://localhost:3000/info
```

### Production
```bash
# Default port is 8080 in prod
curl http://localhost:8080/health
curl http://localhost:8080/info
```

## Key Features

**Base Configuration (`base.yaml`)**
- Shared routes and handlers
- Environment variable placeholders: `${VAR:default}`
- Works across all environments

**Environment Overrides**
- `dev.yaml` - Sets port 3000, debug mode
- `prod.yaml` - Sets port 8080, production settings
- Merges with base configuration

**Environment Variables**
- `APP_ENV` - Environment name (dev/prod)
- `APP_NAME` - Application name
- `APP_HOST` - Server host
- `APP_PORT` - Server port

## Expected Output

### Development
```
🔧 Starting application in dev environment

🚀 Server starting on http://localhost:3000
📖 Try:
   curl http://localhost:3000/health
   curl http://localhost:3000/info

💡 To change environment:
   APP_ENV=prod go run main.go
```

### Production
```
🔧 Starting application in prod environment

🚀 Server starting on http://localhost:8080
📖 Try:
   curl http://localhost:8080/health
   curl http://localhost:8080/info

💡 To change environment:
   APP_ENV=prod go run main.go
```
