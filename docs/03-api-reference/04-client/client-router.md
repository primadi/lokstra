# ClientRouter

> HTTP client with local/remote routing optimization

## Overview

`ClientRouter` is an HTTP client that automatically optimizes routing based on whether the target service is local (same server) or remote. Local calls bypass HTTP completely and use direct router invocation, providing zero-overhead inter-service communication.

## Import Path

```go
import "github.com/primadi/lokstra/common/api_client"
```

---

## ClientRouter Type

**Definition:**
```go
type ClientRouter struct {
    RouterName string        // Router identifier in registry
    ServerName string        // Server identifier in registry
    FullURL    string        // Full URL for remote calls
    IsLocal    bool          // Whether service is on same server
    Router     router.Router // Router instance for local calls
    Timeout    time.Duration // HTTP timeout for remote calls
}
```

**Fields:**
- `RouterName` - Name of the target router (e.g., "user-router")
- `ServerName` - Name of the target server (e.g., "auth-server")
- `FullURL` - Complete URL for remote HTTP calls (e.g., "http://localhost:8080")
- `IsLocal` - If `true`, uses `Router.ServeHTTP`; if `false`, uses HTTP client
- `Router` - Router instance for local optimization
- `Timeout` - HTTP request timeout (default: 30 seconds)

---

## Creating ClientRouter

### From Registry
```go
// Get ClientRouter from registry
client := lokstra_registry.GetClientRouter("user-service")

// Or get with error checking
client, err := lokstra_registry.TryGetClientRouter("user-service")
if err != nil {
    log.Fatal(err)
}
```

### Manual Creation
```go
// Local client
client := &api_client.ClientRouter{
    RouterName: "user-router",
    ServerName: "main-server",
    IsLocal:    true,
    Router:     userRouter,
    Timeout:    30 * time.Second,
}

// Remote client
client := &api_client.ClientRouter{
    RouterName: "user-router",
    ServerName: "user-server",
    FullURL:    "http://user-service:8080",
    IsLocal:    false,
    Timeout:    10 * time.Second,
}
```

---

## HTTP Methods

### GET
Performs HTTP GET request.

**Signature:**
```go
func (c *ClientRouter) GET(
    path string,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
resp, err := client.GET("/users/123", map[string]string{
    "Authorization": "Bearer " + token,
})
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

---

### POST
Performs HTTP POST request with body.

**Signature:**
```go
func (c *ClientRouter) POST(
    path string,
    body any,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
newUser := &User{
    Name:  "John Doe",
    Email: "john@example.com",
}

resp, err := client.POST("/users", newUser, map[string]string{
    "Content-Type": "application/json",
})
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

---

### PUT
Performs HTTP PUT request with body.

**Signature:**
```go
func (c *ClientRouter) PUT(
    path string,
    body any,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
updatedUser := &User{
    Name:  "Jane Doe",
    Email: "jane@example.com",
}

resp, err := client.PUT("/users/123", updatedUser, map[string]string{
    "Content-Type": "application/json",
})
```

---

### PATCH
Performs HTTP PATCH request with body.

**Signature:**
```go
func (c *ClientRouter) PATCH(
    path string,
    body any,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
patch := map[string]any{
    "name": "Updated Name",
}

resp, err := client.PATCH("/users/123", patch, nil)
```

---

### DELETE
Performs HTTP DELETE request.

**Signature:**
```go
func (c *ClientRouter) DELETE(
    path string,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
resp, err := client.DELETE("/users/123", map[string]string{
    "Authorization": "Bearer " + token,
})
```

---

### Method
Performs request with custom HTTP method.

**Signature:**
```go
func (c *ClientRouter) Method(
    method, path string,
    body any,
    headers map[string]string,
) (*http.Response, error)
```

**Example:**
```go
// Custom method
resp, err := client.Method("OPTIONS", "/users", nil, nil)

// HEAD request
resp, err := client.Method("HEAD", "/users/123", nil, nil)
```

---

## Local vs Remote Routing

### Local Routing (Zero Overhead)

When `IsLocal = true`, ClientRouter uses direct router invocation:

```go
client := &api_client.ClientRouter{
    IsLocal: true,
    Router:  userRouter,
}

// This call bypasses HTTP completely
resp, err := client.GET("/users/123", nil)
// Internally: router.ServeHTTP(recorder, req)
```

**Benefits:**
- ✅ No HTTP overhead
- ✅ No network latency
- ✅ No serialization/deserialization
- ✅ Direct memory access
- ✅ Same-process execution

**Use Cases:**
- Inter-service calls on same server
- Monolithic deployment
- Testing with in-memory services

---

### Remote Routing (HTTP)

When `IsLocal = false`, ClientRouter uses standard HTTP client:

```go
client := &api_client.ClientRouter{
    IsLocal: false,
    FullURL: "http://user-service:8080",
    Timeout: 10 * time.Second,
}

// This call uses HTTP client
resp, err := client.GET("/users/123", nil)
// Internally: http.Client.Do(req)
```

**Features:**
- ✅ Standard HTTP protocol
- ✅ Network communication
- ✅ Configurable timeout
- ✅ Load balancing support
- ✅ Service discovery integration

**Use Cases:**
- Microservices architecture
- Distributed deployment
- External API integration

---

## Configuration

### Timeout Configuration

```go
// Default timeout (30 seconds)
client := lokstra_registry.GetClientRouter("user-service")

// Custom timeout
client.Timeout = 5 * time.Second

// Per-service timeout
clients := map[string]*api_client.ClientRouter{
    "fast-service": {
        Timeout: 1 * time.Second,
    },
    "slow-service": {
        Timeout: 60 * time.Second,
    },
}
```

---

### Registry Configuration

**YAML:**
```yaml
clientRouters:
  - routerName: user-router
    serverName: main-server
    isLocal: true
    timeout: 30s

  - routerName: order-router
    serverName: order-server
    fullUrl: http://order-service:8080
    isLocal: false
    timeout: 10s

  - routerName: payment-router
    serverName: payment-server
    fullUrl: http://payment-service:8080
    isLocal: false
    timeout: 60s
```

**Code:**
```go
// Automatically configured from YAML
client := lokstra_registry.GetClientRouter("user-router")
```

---

## Complete Examples

### Local Service Communication
```go
package service

import (
    "github.com/primadi/lokstra/common/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

type OrderService struct {
    userClient *api_client.ClientRouter
}

func NewOrderService() *OrderService {
    return &OrderService{
        // User service is on same server (local optimization)
        userClient: lokstra_registry.GetClientRouter("user-router"),
    }
}

func (s *OrderService) CreateOrder(ctx *request.Context) error {
    // Get user info (local call - zero overhead)
    user, err := api_client.FetchAndCast[*User](s.userClient, 
        fmt.Sprintf("/users/%s", ctx.Params.Get("user_id")))
    if err != nil {
        return ctx.Api.InternalError("Failed to get user")
    }
    
    // Create order
    order := &Order{
        UserID: user.ID,
        Items:  ctx.Body.Get("items"),
    }
    
    return ctx.Api.Created(order)
}
```

---

### Remote Service Communication
```go
package service

import (
    "github.com/primadi/lokstra/common/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

type CheckoutService struct {
    paymentClient *api_client.ClientRouter
}

func NewCheckoutService() *CheckoutService {
    return &CheckoutService{
        // Payment service is remote
        paymentClient: lokstra_registry.GetClientRouter("payment-router"),
    }
}

func (s *CheckoutService) ProcessPayment(ctx *request.Context) error {
    payment := &PaymentRequest{
        Amount:   ctx.Body.Get("amount").(float64),
        Currency: ctx.Body.Get("currency").(string),
        CardID:   ctx.Body.Get("card_id").(string),
    }
    
    // Remote HTTP call with timeout
    result, err := api_client.FetchAndCast[*PaymentResult](
        s.paymentClient, 
        "/payments/process",
        api_client.WithMethod("POST"),
        api_client.WithBody(payment),
        api_client.WithHeaders(map[string]string{
            "Authorization": ctx.Request.Header.Get("Authorization"),
            "X-Request-ID":  ctx.RequestID,
        }),
    )
    
    if err != nil {
        if apiErr, ok := err.(*api_client.ApiError); ok {
            return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
        return ctx.Api.InternalError("Payment processing failed")
    }
    
    return ctx.Api.Ok(result)
}
```

---

### Mixed Local/Remote Architecture
```go
package service

type CompositeService struct {
    // Local services (same server)
    userClient  *api_client.ClientRouter // IsLocal: true
    orderClient *api_client.ClientRouter // IsLocal: true
    
    // Remote services (external)
    paymentClient   *api_client.ClientRouter // IsLocal: false
    shippingClient  *api_client.ClientRouter // IsLocal: false
    analyticsClient *api_client.ClientRouter // IsLocal: false
}

func NewCompositeService() *CompositeService {
    return &CompositeService{
        // Local optimization
        userClient:  lokstra_registry.GetClientRouter("user-router"),
        orderClient: lokstra_registry.GetClientRouter("order-router"),
        
        // Remote HTTP calls
        paymentClient:   lokstra_registry.GetClientRouter("payment-router"),
        shippingClient:  lokstra_registry.GetClientRouter("shipping-router"),
        analyticsClient: lokstra_registry.GetClientRouter("analytics-router"),
    }
}

func (s *CompositeService) PlaceOrder(ctx *request.Context) error {
    // Local calls (zero overhead)
    user, _ := api_client.FetchAndCast[*User](s.userClient, 
        fmt.Sprintf("/users/%s", ctx.Params.Get("user_id")))
    
    // Remote calls (HTTP)
    payment, err := api_client.FetchAndCast[*PaymentResult](
        s.paymentClient, "/payments/process",
        api_client.WithMethod("POST"),
        api_client.WithBody(ctx.Body.Get("payment")),
    )
    if err != nil {
        return handlePaymentError(err)
    }
    
    shipping, err := api_client.FetchAndCast[*ShippingResult](
        s.shippingClient, "/shipping/create",
        api_client.WithMethod("POST"),
        api_client.WithBody(ctx.Body.Get("shipping")),
    )
    if err != nil {
        return handleShippingError(err)
    }
    
    // Local call
    order := &Order{
        UserID:     user.ID,
        PaymentID:  payment.ID,
        ShippingID: shipping.ID,
    }
    created, _ := api_client.FetchAndCast[*Order](s.orderClient, "/orders",
        api_client.WithMethod("POST"),
        api_client.WithBody(order),
    )
    
    // Fire and forget analytics (remote)
    go s.trackOrder(created)
    
    return ctx.Api.Created(created)
}
```

---

### Dynamic Client Selection
```go
type ServiceRouter struct {
    clients map[string]*api_client.ClientRouter
}

func (s *ServiceRouter) GetClient(serviceName string) *api_client.ClientRouter {
    if client, ok := s.clients[serviceName]; ok {
        return client
    }
    return lokstra_registry.GetClientRouter(serviceName)
}

func (s *ServiceRouter) CallService(serviceName, path string, opts ...api_client.FetchOption) (any, error) {
    client := s.GetClient(serviceName)
    return api_client.FetchAndCast[any](client, path, opts...)
}
```

---

### Timeout Management
```go
type TimeoutManager struct {
    clients map[string]*api_client.ClientRouter
}

func NewTimeoutManager() *TimeoutManager {
    return &TimeoutManager{
        clients: map[string]*api_client.ClientRouter{
            "fast": {
                FullURL: "http://fast-service:8080",
                Timeout: 1 * time.Second,
            },
            "normal": {
                FullURL: "http://normal-service:8080",
                Timeout: 10 * time.Second,
            },
            "slow": {
                FullURL: "http://slow-service:8080",
                Timeout: 60 * time.Second,
            },
        },
    }
}

func (tm *TimeoutManager) CallWithTimeout(service, path string) (any, error) {
    client := tm.clients[service]
    return api_client.FetchAndCast[any](client, path)
}
```

---

### Custom Headers Pattern
```go
type AuthenticatedClient struct {
    client *api_client.ClientRouter
    token  string
}

func (ac *AuthenticatedClient) Get(path string) (*http.Response, error) {
    return ac.client.GET(path, map[string]string{
        "Authorization": "Bearer " + ac.token,
    })
}

func (ac *AuthenticatedClient) Post(path string, body any) (*http.Response, error) {
    return ac.client.POST(path, body, map[string]string{
        "Authorization": "Bearer " + ac.token,
        "Content-Type":  "application/json",
    })
}
```

---

### Health Check Pattern
```go
type HealthChecker struct {
    clients map[string]*api_client.ClientRouter
}

func (h *HealthChecker) CheckHealth(serviceName string) error {
    client := h.clients[serviceName]
    
    // Set short timeout for health checks
    originalTimeout := client.Timeout
    client.Timeout = 2 * time.Second
    defer func() { client.Timeout = originalTimeout }()
    
    resp, err := client.GET("/health", nil)
    if err != nil {
        return fmt.Errorf("service unhealthy: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("service unhealthy: status %d", resp.StatusCode)
    }
    
    return nil
}

func (h *HealthChecker) CheckAllServices() map[string]error {
    results := make(map[string]error)
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for name := range h.clients {
        wg.Add(1)
        go func(n string) {
            defer wg.Done()
            err := h.CheckHealth(n)
            mu.Lock()
            results[n] = err
            mu.Unlock()
        }(name)
    }
    
    wg.Wait()
    return results
}
```

---

## Performance Characteristics

### Local Routing Performance
```go
// Benchmark results (from client_helper_bench_test.go)
BenchmarkLocalCall-8    1000000    800 ns/op

// Components:
// - Router lookup:        ~10ns
// - ServeHTTP:           ~700ns
// - Response recording:   ~90ns
```

**Benefits:**
- No HTTP overhead (~50μs)
- No network latency (~1-10ms)
- No JSON serialization (~5-50μs)
- Direct memory access

**Total savings per call:** ~50-100μs (50-100x faster)

---

### Remote Routing Performance
```go
// Typical remote call timing
// - DNS lookup:           ~10-50ms (cached: ~1ms)
// - TCP handshake:        ~1-10ms
// - TLS handshake:        ~10-50ms (if HTTPS)
// - HTTP request:         ~1-5ms
// - Service processing:   varies
// - HTTP response:        ~1-5ms
// - Total:                ~25-120ms
```

---

## Best Practices

### 1. Use Local Routing When Possible
```go
// ✅ Good: Local optimization for same-server services
client := &api_client.ClientRouter{
    IsLocal: true,
    Router:  userRouter,
}

// 🚫 Avoid: HTTP calls for local services
client := &api_client.ClientRouter{
    IsLocal: false,
    FullURL: "http://localhost:8080",
}
```

---

### 2. Set Appropriate Timeouts
```go
// ✅ Good: Service-specific timeouts
fastClient.Timeout = 1 * time.Second
slowClient.Timeout = 60 * time.Second

// 🚫 Avoid: One-size-fits-all timeout
allClients.Timeout = 30 * time.Second
```

---

### 3. Use FetchAndCast for Type Safety
```go
// ✅ Good: Type-safe with FetchAndCast
user, err := api_client.FetchAndCast[*User](client, "/users/123")

// 🚫 Avoid: Manual response parsing
resp, _ := client.GET("/users/123", nil)
var user User
json.NewDecoder(resp.Body).Decode(&user)
```

---

### 4. Configure from Registry
```go
// ✅ Good: Centralized configuration
client := lokstra_registry.GetClientRouter("user-service")

// 🚫 Avoid: Hardcoded URLs
client := &api_client.ClientRouter{
    FullURL: "http://localhost:8080",
}
```

---

### 5. Handle Timeouts Gracefully
```go
// ✅ Good: Check for timeout errors
resp, err := client.GET("/slow-endpoint", nil)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return ctx.Api.GatewayTimeout("Service timeout")
    }
    return ctx.Api.InternalError("Service error")
}

// 🚫 Avoid: Generic error handling
if err != nil {
    return ctx.Api.InternalError("Error")
}
```

---

## See Also

- **[API Client](./api-client)** - FetchAndCast and options
- **[RemoteService](./remote-service)** - Remote service patterns
- **[Router](../01-core-packages/router)** - Router configuration

---

## Related Guides

- **[HTTP Clients](../../04-guides/http-clients/)** - HTTP client patterns
- **[Service Communication](../../04-guides/service-communication/)** - Inter-service patterns
- **[Performance](../../04-guides/performance/)** - Optimization techniques
