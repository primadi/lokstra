# Basic YAML Configuration Example

Demonstrates loading configuration from a single YAML file.

## What This Demonstrates

- Loading YAML configuration
- Registering handlers
- Applying configuration to create server
- Starting server from config

## Files

- `config.yaml` - YAML configuration with routers and server
- `main.go` - Application code

## Run

```bash
go run main.go
```

## Test

```bash
# Health check
curl http://localhost:8080/health

# Version
curl http://localhost:8080/version
```

## Expected Output

```
Server starting on :8080
Server 'web-server' starting with 1 app(s):
Starting [api-app] with 1 router(s) on address :8080
  GET /health
  GET /version
Press CTRL+C to stop the server...
```
