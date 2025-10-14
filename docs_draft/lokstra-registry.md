
# Lokstra Registry

Lokstra Registry is a package for Dependency Injection and registry management in Lokstra applications, especially for routers, service factories, services, and middleware. 

This registry centralizes instance management and supports lazy service creation (created only when needed).

## Main Features

- Register and retrieve routers
- Register and retrieve service factories
- Register and retrieve services (direct and lazy)
- Registration override option (AllowOverride)

---

## Router

### RegisterRouter
```go
func RegisterRouter(name string, r router.Router, opts ...RegisterOption)
```
Registers a router with a given name. If the name is already registered and `AllowOverride` is not enabled, it will panic.

### GetRouter
```go
func GetRouter(name string) router.Router
```
Retrieves a router by name. Returns `nil` if not found.

---

## Service Factory

### RegisterServiceFactory
```go
func RegisterServiceFactory(serviceType string, factory ServiceFactory, opts ...RegisterOption)
```
Registers a factory function for a given service type. If the type is already registered and `AllowOverride` is not enabled, it will panic.

### GetServiceFactory
```go
func GetServiceFactory(serviceType string) ServiceFactory
```
Retrieves a factory by type. Returns `nil` if not found.

---

## Service

### RegisterService
```go
func RegisterService(svcName string, svcInstance any, opts ...RegisterOption)
```
Registers a service instance with a given name. If the name is already registered and `AllowOverride` is not enabled, it will panic.

### RegisterLazyService
```go
func RegisterLazyService(svcName string, svcType string, config map[string]any, opts ...RegisterOption)
```
Registers a lazy service configuration. The actual instance will be created when first requested. If the name is already registered and `AllowOverride` is not enabled, it will panic.


### GetService
```go
func GetService[T comparable](name string, current T) T
```
Retrieves a service from the registry. The `current` parameter acts as a cache:
- If `current` is already set (non-nil), it is returned immediately and the registry is not accessed again.
- If `current` is nil/zero, the function will look up the registry (or create from lazy config if available).
- If not found or type mismatch, it will panic.

This mechanism ensures that service lookup and initialization only happen once, and subsequent calls simply reuse the cached value in the same variable.

### TryGetService
```go
func TryGetService[T comparable](svcName string, current T) (T, bool)
```
Similar to `GetService`, but does not panic if not found or type mismatch. The `current` parameter also acts as a cache: if already set, it is returned immediately. If not, the registry is checked and lazy config is used if available.

### NewService
```go
func NewService[T any](svcName, svcType string, config map[string]any, opts ...RegisterOption) T
```
Creates a new service using a registered factory, registers it, and returns it. If the factory is not found or creation fails, returns the zero value of T.

---

## RegisterOption & AllowOverride

All registration functions (RegisterRouter, RegisterService, RegisterServiceFactory, RegisterLazyService, NewService) accept an optional `RegisterOption` parameter. This option controls registration behavior, such as whether to override an existing entry.

### AllowOverride
```go
func AllowOverride(enable bool) RegisterOption
```
If `AllowOverride(true)` is provided, registration will override the previous entry with the same name/type. If not provided (default), registration with the same name/type will panic.

#### Example Usage
```go
// Register service without override (will panic if already exists)
RegisterService("db", dbInstance)

// Register service with override
RegisterService("db", dbInstance, AllowOverride(true))
```

---

## Best Practice

- Use unique names for each router and service.
- For lazy services, use RegisterLazyService so resources are only used when truly needed.
- Always use AllowOverride carefully, only when you intend to replace an old entry.
- For dependency injection, use TryGetService to avoid panics if not registered yet.

---

## Simple Example

```go
// Register a factory
RegisterServiceFactory("mytype", func(cfg map[string]any) any {
	return &MyService{Param: cfg["param"].(string)}
})

// Register a lazy service
RegisterLazyService("svcA", "mytype", map[string]any{"param": "value"})

// Retrieve service (will be created from factory if not already present)
var svc *MyService
svc = GetService("svcA", svc)
```

---

## Error Handling

- Registering the same name/type without AllowOverride will panic.
- Retrieving a service with the wrong type will panic.
- TryGetService is safer for non-panicking checks.

---

## FAQ

**Q: What is the difference between RegisterService and RegisterLazyService?**
A: RegisterService immediately registers an instance, while RegisterLazyService only stores the config and creates the instance when first requested.

**Q: When should I use AllowOverride?**
A: Use AllowOverride(true) if you want to replace an old entry, e.g., for hot reload or re-configure.

**Q: How do I do dependency injection between services?**
A: Register all dependencies in the registry, then retrieve them using TryGetService or GetService as needed.
