# Monolith to Microservices

Deploy the same codebase as either a monolith or microservices using different configurations.

## Running

### Monolith Deployment
```bash
go run main.go monolith
```
Access at `http://localhost:3010`

### Microservices Deployment
```bash
go run main.go microservices
```
- User Service: `http://localhost:3011`
- Order Service: `http://localhost:3012`

## Key Concepts

### Same Code, Different Configs

**config-monolith.yaml**
- Single server on port 3010
- All services in one process
- Single router with all routes

**config-microservices.yaml**
- Multiple servers (3011, 3012)
- Services separated by domain
- Each service has its own router

## Benefits

**Monolith:**
- Simple deployment
- Easy development
- Lower latency (no network calls)

**Microservices:**
- Independent scaling
- Technology flexibility
- Team autonomy

## Migration Strategy

1. Start with monolith (faster development)
2. Split config by service boundaries
3. Deploy as microservices when needed
4. No code changes required!
