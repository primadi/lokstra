# Middleware Usage Example

This example demonstrates comprehensive middleware usage patterns in the Lokstra framework.

## Features Demonstrated

1. **Global Middleware** - Applied to all routes
2. **Route-Specific Middleware** - Applied to individual routes
3. **Group-Level Middleware** - Applied to route groups
4. **Multiple Middleware Chaining** - Multiple middleware on single routes
5. **Middleware Override** - Bypassing global middleware
6. **Context Value Storage** - Storing and retrieving values between middleware

## Middleware Types Shown

- **Logging Middleware** - Request/response logging with timing
- **Request ID Middleware** - Unique request identification
- **Authentication Middleware** - Token-based authentication simulation
- **Rate Limiting Middleware** - Request rate limiting
- **Admin Check Middleware** - Role-based access control
- **Audit Log Middleware** - Action auditing
- **CORS Middleware** - Cross-origin resource sharing
- **JSON Middleware** - Content-type handling

## How to Run

```bash
go run main.go
```

The server will start on port 8080.

## Test Endpoints

### Simple Routes
```bash
# Simple route with global middleware (logging + request_id)
curl http://localhost:8080/ping

# Public route with middleware override (no global middleware)
curl http://localhost:8080/public
```

### Protected Routes
```bash
# Protected route (requires auth + rate limiting)
curl -H "Authorization: Bearer valid-token" http://localhost:8080/protected

# Admin action (requires auth + admin check + audit log)
curl -X POST -H "Authorization: Bearer valid-token" http://localhost:8080/admin/action
```

### API Group Routes
```bash
# API routes with CORS + JSON middleware
curl http://localhost:8080/api/v1/users

# Create user (POST)
curl -X POST -H "Content-Type: application/json" http://localhost:8080/api/v1/users

# Admin stats (nested group with admin check)
curl -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/admin/stats
```

## Expected Behavior

1. **Global middleware** runs on all routes except those using `HandleOverrideMiddleware`
2. **Route-specific middleware** runs after global middleware
3. **Group middleware** applies to all routes in that group
4. **Nested groups** inherit parent group middleware
5. **Context values** are preserved across middleware chain
6. **Headers** are properly set by middleware
7. **Authentication failures** return appropriate error responses

## Key Learning Points

- Middleware execution order: Global → Group → Route-specific
- Context value storage using `context.WithValue`
- Header manipulation with `ctx.WithHeader()`
- Error handling in middleware
- Middleware registration and naming
- Route grouping and nesting
- Middleware override capabilities
