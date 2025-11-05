# Lokstra Multi-App Template

This template demonstrates the **Multi-App pattern** using Lokstra's Server orchestration capabilities. It shows how to run multiple applications on different ports, each with its own routers and middleware, coordinated by a single server instance.

## Architecture Overview

```
Server (lokstra.NewServer)
├── Main App (Port 3000)
│   ├── Health Router (/health, /ready)
│   └── API Router (/api/users, /api/roles)
└── Admin App (Port 3001)
    ├── Health Router (/health, /ready)
    ├── Admin API Router (/admin/users/*, /admin/roles/*)
    └── System Router (/admin/system/*)
```

## Key Concepts

### 1. Multi-App Server
The `lokstra.NewServer` orchestrates multiple apps:
- **Centralized Management**: Single entry point for all applications
- **Coordinated Lifecycle**: All apps start/stop together with graceful shutdown
- **Separate Ports**: Each app runs on its own port for isolation
- **Independent Configuration**: Each app can have different middleware, routes, and settings

### 2. Multiple Routers Per App
Each App can contain multiple routers:
- **Health Router**: Shared across apps for monitoring (`/health`, `/ready`)
- **API Router**: Application-specific routes
- **Admin Router**: Administrative operations
- **System Router**: System-level management

### 3. Port Separation
Different apps on different ports provide:
- **Security Isolation**: Admin endpoints not accessible from public port
- **Traffic Segregation**: Separate analytics and rate limiting
- **Firewall Control**: Different network policies per port
- **Load Balancing**: Route traffic differently based on port

## File Structure

```
03_multi_app/
├── main.go                  # Server setup and entry point
├── test.http                # HTTP requests for testing both apps
├── .gitignore              # Git ignore patterns
├── README.md               # This file
│
├── shared/                  # Shared components across apps
│   ├── health.go           # Health check router factory
│   └── middleware.go       # Custom middleware (logging, etc.)
│
├── mainapp/                 # Main public API (port 3000)
│   ├── app.go              # App setup and configuration
│   ├── router.go           # Router with API routes
│   └── handlers.go         # CRUD handlers for users and roles
│
└── adminapp/                # Admin API (port 3001)
    ├── app.go              # App setup and configuration
    ├── router.go           # Router with admin routes
    └── handlers.go         # Admin-specific handlers (suspend, stats, etc.)
```

## Running the Application

```bash
# From the project root (lokstra-dev2/)
go run ./project_templates/01_router/03_multi_app
```

The server will start both applications:
- **Main App**: http://localhost:3000
- **Admin App**: http://localhost:3001

You'll see startup messages for each app:
```
🚀 Starting application: main-app
   └── Listening on: :3000
🚀 Starting application: admin-app
   └── Listening on: :3001
Server "demo-server" started successfully
Press Ctrl+C to shutdown gracefully...
```

## API Endpoints

### Main App (Port 3000)

#### Health Checks
- `GET /health` - Application health status
- `GET /ready` - Readiness check

#### Users
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create user
- `PUT /api/users/:id` - Update user
- `PATCH /api/users/:id` - Partially update user
- `DELETE /api/users/:id` - Delete user

#### Roles
- `GET /api/roles` - List all roles
- `GET /api/roles/:id` - Get role by ID
- `POST /api/roles` - Create role
- `PUT /api/roles/:id` - Update role
- `PATCH /api/roles/:id` - Partially update role
- `DELETE /api/roles/:id` - Delete role

### Admin App (Port 3001)

#### Health Checks
- `GET /health` - Admin app health status
- `GET /ready` - Admin app readiness check

#### Admin User Management
- `GET /admin/users` - List all users (with admin details)
- `GET /admin/users/:id` - Get user by ID (admin view)
- `POST /admin/users` - Create user
- `PUT /admin/users/:id` - Update user
- `DELETE /admin/users/:id` - Delete user
- `POST /admin/users/:id/suspend` - Suspend user account
- `POST /admin/users/:id/activate` - Activate user account

#### Admin Role Management
- `GET /admin/roles` - List all roles
- `GET /admin/roles/:id` - Get role by ID
- `POST /admin/roles` - Create role
- `PUT /admin/roles/:id` - Update role
- `DELETE /admin/roles/:id` - Delete role
- `DELETE /admin/roles/:id/users/:userId` - Remove role from user

#### System Management
- `GET /admin/system/stats` - Get system statistics (users, roles, CPU, memory)
- `GET /admin/system/config` - Get system configuration
- `POST /admin/system/cache/clear` - Clear application cache

## Testing

Use the included `test.http` file with VS Code REST Client extension or any HTTP client:

```bash
# Test main app health
curl http://localhost:3000/health

# Test admin app health
curl http://localhost:3001/health

# List users from main app
curl http://localhost:3000/api/users

# List users from admin app (with admin details)
curl http://localhost:3001/admin/users

# Get system stats (admin only)
curl http://localhost:3001/admin/system/stats
```

## Features Demonstrated

### 1. Server Orchestration
```go
server := lokstra.NewServer("demo-server", mainApp, adminApp)
```
- Single server manages multiple apps
- Coordinated startup and graceful shutdown
- Central logging and monitoring point

### 2. Shared Components
```go
func setupHealthRouter(appName string) lokstra.Router {
    // Same health router used by both apps
    // Parameterized with app name for identification
}
```
- Reusable routers across apps (in `shared/` folder)
- Consistent health check patterns
- DRY principle applied

### 3. App-Specific Middleware
```go
mainApp.Use(shared.CustomLoggingMiddleware("MAIN"))
adminApp.Use(shared.CustomLoggingMiddleware("ADMIN"))
```
- Different middleware stacks per app
- App identification in logs  
- Independent request processing pipelines

### 4. Graceful Shutdown
```go
server.Shutdown(30 * time.Second)
```
- Coordinated shutdown of all apps
- 30-second timeout for in-flight requests
- Clean resource cleanup

## When to Use Multi-App Pattern

### ✅ Good Use Cases
- **Admin Separation**: Public API and admin panel on different ports
- **API Versioning**: v1 and v2 APIs as separate apps
- **Service Segregation**: Public and internal services
- **Security Boundaries**: Different authentication requirements
- **Traffic Management**: Different rate limits per service

### ❌ When to Use Single App Instead
- Simple applications with one purpose
- All endpoints share same security model
- No need for port-based isolation
- Microservice architecture (separate processes instead)

## Security Considerations

1. **Firewall Rules**: Only expose main app port (3000) publicly, keep admin port (3001) internal
2. **Authentication**: Admin app should require stronger authentication (not shown in template)
3. **Network Segmentation**: Admin app can be on private network/VPN only
4. **Rate Limiting**: Different limits for public vs admin endpoints
5. **Audit Logging**: Enhanced logging for admin operations

## Deployment Strategies

### Development
```bash
# Run from project root
go run ./project_templates/01_router/03_multi_app
```

### Docker
```dockerfile
# Expose both ports
EXPOSE 3000 3001
CMD ["./multi-app-demo"]
```

### Kubernetes
```yaml
# Separate services for each port
apiVersion: v1
kind: Service
metadata:
  name: main-app
spec:
  ports:
  - port: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: admin-app
spec:
  ports:
  - port: 3001
  # Add network policies to restrict access
```

### Reverse Proxy (nginx)
```nginx
# Public traffic to main app
server {
    listen 80;
    server_name api.example.com;
    location / {
        proxy_pass http://localhost:3000;
    }
}

# Internal admin traffic
server {
    listen 80;
    server_name admin.example.com;
    # Restrict by IP
    allow 10.0.0.0/8;
    deny all;
    location / {
        proxy_pass http://localhost:3001;
    }
}
```

## Comparison with Other Patterns

| Feature | Router Only | Single App | Multi-App |
|---------|------------|------------|-----------|
| Complexity | Low | Medium | High |
| Port Isolation | ❌ | ❌ | ✅ |
| Graceful Shutdown | ❌ | ✅ | ✅ |
| Multiple Routers | ❌ | ✅ | ✅ |
| App Separation | ❌ | ❌ | ✅ |
| Coordinated Lifecycle | ❌ | N/A | ✅ |
| Production Ready | ❌ | ✅ | ✅ |

## Next Steps

1. **Add Authentication**: Implement JWT or session-based auth for admin app
2. **Add Database**: Replace mock data with real database
3. **Add Validation**: Enhance input validation
4. **Add Rate Limiting**: Different limits for main vs admin
5. **Add Metrics**: Prometheus metrics per app
6. **Add Tracing**: Distributed tracing across apps

## Learning Path

1. Start with `01_router_only` - Learn basic routing
2. Move to `02_single_app` - Understand App benefits
3. Study this template - Master multi-app orchestration
4. Build your own - Combine patterns for your use case

## Additional Resources

- [Lokstra Documentation](../../docs)
- [Router Guide](../../docs/01-router-guide)
- [Framework Guide](../../docs/02-framework-guide)
- [API Reference](../../docs/03-api-reference)
