# Example 25 - Convention System Integration

## ✅ Updated to Use Service Convention System

Example 25 telah diupdate untuk menggunakan Service Convention System yang baru. Ini mendemonstrasikan bagaimana convention secara otomatis men-generate HTTP routes dari service interfaces.

## What Changed

### Before (Manual Router Registration)

```go
// Manual registration for each service
old_registry.RegisterRouterFunc("user-service", func(app *lokstra.App) error {
    svc := old_registry.GetService[services.UserService]("user-service", nil)
    rt := router.NewFromService(svc, router.DefaultServiceRouterOptions())
    app.AddRouter(rt)
    return nil
})

// Repeat for every service... lots of boilerplate!
```

### After (Convention-Based Registration)

```go
// Configure REST convention once
restOptions := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithConvention("rest")

// Register routers with auto-generated routes
userSvc := services.CreateUserServiceLocal(nil)
userRouter := router.NewFromService(userSvc, restOptions.WithResourceName("user"))
old_registry.RegisterRouter("user-service", userRouter)

// Convention automatically generates:
// GET    /api/v1/users/{id}     -> GetUser
// GET    /api/v1/users          -> ListUsers
// POST   /api/v1/users          -> CreateUser
// PUT    /api/v1/users/{id}     -> UpdateUser
// DELETE /api/v1/users/{id}     -> DeleteUser
```

## Generated Routes

### User Service
- `GET /api/v1/users/{id}` → `GetUser`
- `GET /api/v1/users` → `ListUsers`
- `POST /api/v1/users` → `CreateUser`
- `PUT /api/v1/users/{id}` → `UpdateUser`
- `DELETE /api/v1/users/{id}` → `DeleteUser`

### Auth Service (with custom overrides)
- `POST /api/v1/auth/login` → `Login` (custom path via override)
- `POST /api/v1/auth/logout` → `Logout` (custom path via override)
- `POST /api/v1/auth/validate-token` → `ValidateToken`

### Order Service
- `GET /api/v1/orders/{id}` → `GetOrder`
- `GET /api/v1/orders` → `ListOrders`
- `POST /api/v1/orders` → `CreateOrder`
- `PUT /api/v1/orders/{id}` → `UpdateOrder`

### Cart Service
- `GET /api/v1/carts/{id}` → `GetCart`
- `POST /api/v1/carts` → `CreateCart`
- `POST /api/v1/carts/add-item` → `AddItem`
- `POST /api/v1/carts/remove-item` → `RemoveItem`

### Payment Service
- `POST /api/v1/payments` → `ProcessPayment`
- `GET /api/v1/payments/{id}` → `GetPaymentStatus`

### Invoice Service
- `GET /api/v1/invoices/{id}` → `GetInvoice`
- `POST /api/v1/invoices/generate` → `GenerateInvoice`

## How It Works

### 1. REST Convention Mapping

The REST convention automatically maps method names to HTTP routes:

| Method Pattern | HTTP Method | Path Pattern |
|---------------|-------------|--------------|
| `Get{Resource}` | GET | `/{resources}/{id}` |
| `List{Resource}s` | GET | `/{resources}` |
| `Create{Resource}` | POST | `/{resources}` |
| `Update{Resource}` | PUT | `/{resources}/{id}` |
| Other methods | POST | `/{resources}/{method-name}` |

### 2. Convention Options

```go
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").              // Add prefix to all routes
    WithConvention("rest").             // Use REST convention
    WithResourceName("user")            // Singular resource name
```

### 3. Custom Overrides for Edge Cases

```go
authOptions := restOptions.WithResourceName("auth").
    WithRouteOverride("Login", router.RouteMeta{
        HTTPMethod: "POST",
        Path:       "/login",  // Custom path instead of /auth/login
    })
```

## Benefits

### 1. Less Boilerplate
- **Before**: ~100 lines of manual router registration
- **After**: ~80 lines with clear convention usage
- **Reduction**: 20% less code, much clearer intent

### 2. Consistency
- All services follow REST convention
- Predictable URL patterns
- Easy to understand API structure

### 3. Flexibility
- Override specific routes for special cases (like auth endpoints)
- Mix convention-based and manual routes
- Easy to switch conventions

### 4. Maintainability
- Add new methods → routes auto-generated
- Change convention → all routes update
- Clear separation: service logic vs routing

## Running Example 25

### Monolith Deployment
```bash
# All services in one binary
./25-single-binary-deployment -config config-monolith.yaml

# Routes available:
# http://localhost:8080/api/v1/users
# http://localhost:8080/api/v1/orders
# http://localhost:8080/api/v1/auth/login
# etc...
```

### Multiport Deployment
```bash
# Run different services on different ports
./25-single-binary-deployment -config config-multiport.yaml -server user-server
# user-service on :8080

./25-single-binary-deployment -config config-multiport.yaml -server order-server
# order-service on :8081
```

### Microservices Deployment
```bash
# Each service as separate instance
./25-single-binary-deployment -config config-microservices.yaml -server user-service-pod
./25-single-binary-deployment -config config-microservices.yaml -server order-service-pod
./25-single-binary-deployment -config config-microservices.yaml -server payment-service-pod
```

## Testing with Convention

### Test User Service
```bash
# List users (convention: GET /users)
curl http://localhost:8080/api/v1/users

# Get user (convention: GET /users/{id})
curl http://localhost:8080/api/v1/users/123

# Create user (convention: POST /users)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'
```

### Test Auth Service (with custom overrides)
```bash
# Login (override: POST /auth/login instead of POST /auths)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'

# Validate token (convention: POST /auth/validate-token)
curl -X POST http://localhost:8080/api/v1/auth/validate-token \
  -H "Content-Type: application/json" \
  -d '{"token":"abc123"}'
```

## Key Takeaways

1. **Convention system eliminates boilerplate** while keeping flexibility
2. **REST convention provides sensible defaults** for HTTP APIs
3. **Override system handles edge cases** (like custom auth paths)
4. **All deployment modes benefit** from convention-based routing
5. **Service interfaces drive the API** - change interface, routes update automatically

## Future Enhancements

- Add OpenAPI/Swagger generation from conventions
- Add validation rules via convention
- Add rate limiting configuration per convention
- Add custom conventions for specific domains (e.g., "admin-api" convention)

---

**Status**: ✅ Example 25 successfully updated with convention system
**Build**: ✅ Compiles without errors
**Features**: All deployment modes work (monolith, multiport, microservices)
