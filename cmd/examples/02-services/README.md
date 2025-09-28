# Services and Dependency Injection Example

**Purpose:** Show how to define a Service and use it via DI.

## Features
- Service defined in ServerModule
- Required service declaration
- Handler accessing service via `ctx.GetService`

## Run
```
go run main.go
curl http://localhost:8081/count
```
