# Core Router Unit Tests

Unit tests for the `core/router` package in the Lokstra framework.

## Test Files

### 1. `basic_test.go`
Basic tests for compilation verification and interface implementation.

### 2. `router_test.go` 
Tests for the Router interface and its method completeness.

### 3. `router_impl_test.go`
Comprehensive tests for RouterImpl implementation:
- **Factory Functions**: `NewRouter`, `NewRouterWithEngine`, `NewListener`, `NewListenerWithEngine`
- **Basic Methods**: `Prefix`, `WithPrefix`, `GetMeta`, `GetMiddleware`
- **HTTP Methods**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- **Route Handling**: `Handle`, `HandleOverrideMiddleware`
- **Middleware**: `Use`, `WithOverrideMiddleware`, `OverrideMiddleware`, `LockMiddleware`
- **Grouping**: `Group`, `GroupBlock`
- **Mounting**: `MountStatic`, `MountSPA`, `MountReverseProxy`, `MountRpcService`
- **Utilities**: `RecurseAllHandler`, `DumpRoutes`, `ServeHTTP`, `FastHttpHandler`

### 4. `group_impl_test.go`
Tests for GroupImpl implementation:
- **Basic Methods**: `Prefix`, `WithPrefix`, `GetMeta`, `GetMiddleware` 
- **HTTP Methods**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- **Route Handling**: `Handle`, `HandleOverrideMiddleware`
- **Nested Groups**: Groups within groups, middleware inheritance
- **Middleware**: `Use`, `WithOverrideMiddleware`, `OverrideMiddleware`, `LockMiddleware`
- **Mounting**: Static files, SPA, reverse proxy
- **Panic Scenarios**: Methods that should panic

### 5. `integration_test.go`
Integration tests and edge cases:
- **RPC Service Mounting**: String service, Service interface, RpcServiceMeta
- **Handler Types**: HandlerFunc, string, HandlerMeta
- **Static Mounting**: Static files, SPA, reverse proxy
- **Prefix Cleaning**: Various prefix combinations

## Mock Objects

Tests use mock implementations:

### MockRegistrationContext
Mock for `registration.RegistrationContext` interface providing:
- Service creation dan retrieval
- Handler registration
- Middleware registration
- Module registration

### MockRouterEngine  
Mock for `serviceapi.RouterEngine` interface providing:
- HTTP method handling
- Static file serving
- SPA serving
- Reverse proxy

### MockHttpListener
Mock for `serviceapi.HttpListener` interface providing:
- HTTP server lifecycle
- Request handling

### MockRpcService
Mock for `service.Service` interface for testing RPC services.

## Running Tests

```bash
# Run all tests in the router package
go test ./core/router

# Run tests with verbose output
go test -v ./core/router

# Run specific test
go test -v ./core/router -run TestRouterImpl_HTTPMethods

# Run tests with coverage
go test -v -cover ./core/router
```

## Coverage Areas

Tests cover:

1. **Interface Compliance**: Ensuring implementations meet interface contracts
2. **Method Functionality**: All public methods function correctly
3. **Error Handling**: Panic and error scenarios
4. **State Management**: Middleware locking, route registration
5. **Integration**: Component interactions
6. **Edge Cases**: Empty values, nil pointers, invalid inputs

## Notes

- Tests use mock objects to avoid dependencies on concrete implementations
- Tests focus on behavior and contracts rather than implementation details
- Tests are designed to be easy to maintain and extend
- All test cases have clear descriptions and specific assertions
