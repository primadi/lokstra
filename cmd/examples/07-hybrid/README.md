# Hybrid Architecture Example

**Purpose:** Combine monolith + microservice using a gateway and reverse proxy.

## Features
- Gateway app with multiple modules
- Reverse proxy using `HandleRaw`
- Unified interface for client

## Run
```
go run main.go
curl http://localhost:8082/user/hello
curl http://localhost:8082/order/status
```
