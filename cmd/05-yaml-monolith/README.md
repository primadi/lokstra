# YAML Configuration Monolith Example

**Purpose:** Configure Lokstra via YAML for a monolithic server.

## Features
- Server, services, middleware defined in YAML
- Multiple apps in one server
- Env interpolation supported

## Run
```
export DUMMY_INITIAL=5   # optional
go run main.go
curl http://localhost:8080/blog/
```
