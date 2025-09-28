# Basic Server Example

**Purpose:** Demonstrate a minimal Lokstra server using the Server abstraction, with multiple apps, router chaining, and graceful shutdown.

## Features
- Server creation with `server.New`
- Multiple app support
- Router chaining
- Listener configuration (address, type)
- Direct route definition
- Group route support
- Middleware usage
- Graceful shutdown on signal

## Run
```
go run main.go
curl http://localhost:8080/ping
```
