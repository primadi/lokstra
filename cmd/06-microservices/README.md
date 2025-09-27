# Microservices Deployment Example

**Purpose:** Configure multiple servers (microservices).

## Features
- Multiple servers in YAML
- Independent apps/services
- Isolation and parallel startup

## Run
```
go run main.go
curl http://localhost:8080/user/hello
curl http://localhost:8081/order/status
```
