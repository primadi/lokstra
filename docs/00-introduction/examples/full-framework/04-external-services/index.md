# Example 06 - External Services Integration

This example demonstrates how to **integrate external APIs** (like payment gateways, email services, SMS providers) as Lokstra services using `proxy.Service` for convention-based remote calls.

## 📋 What You'll Learn

- ✅ Wrapping third-party APIs as Lokstra services
- ✅ Using `proxy.Service` for remote HTTP calls
- ✅ **Route overrides in `RegisterServiceType`** (not in config!)
- ✅ `external-service-definitions` with URL and factory type
- ✅ Business services depending on external services
- ✅ Error handling when external service fails
- ✅ **Clean service code** without metadata embedding
- ✅ Difference between `proxy.Service` vs `proxy.Router` (see Example 07)

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────┐
│                      Main App (:3000)                      │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  OrderService (Business Logic)                       │  │
│  │  - Create()    → POST /orders                        │  │
│  │  - Get()       → GET /orders/{id}                    │  │
│  │  - Refund()    → POST /orders/{id}/refund            │  │
│  └─────────────────┬────────────────────────────────────┘  │
│                    │ depends on                            │
│                    ▼                                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PaymentServiceRemote (proxy.Service)                │  │
│  │  - CreatePayment()  → POST /payments                 │  │
│  │  - GetPayment()     → GET /payments/{id}             │  │
│  │  - Refund()         → POST /payments/{id}/refund     │  │
│  └─────────────────┬────────────────────────────────────┘  │
│                    │ HTTP calls                            │
└────────────────────┼───────────────────────────────────────┘
                     ▼
     ┌───────────────────────────────────────────────┐
     │   Mock Payment Gateway (:9000)                │
     │   (Simulates Stripe, PayPal, etc.)            │
     │                                               │
     │   POST   /payments                            │
     │   GET    /payments/{id}                       │
     │   POST   /payments/{id}/refund                │
     └───────────────────────────────────────────────┘
```

**Key:** All route overrides defined in `RegisterServiceType` in `main.go`!

## 🚀 How to Run

### Step 1: Start Mock Payment Gateway

```bash
cd mock-payment-gateway
go run main.go
```

This starts the mock payment gateway on `http://localhost:9000`. It simulates an external payment provider like Stripe or PayPal.

### Step 2: Start Main Application

```bash
# From the example root directory
go run main.go
```

This starts the main application on `http://localhost:3000`.

### Step 3: Test with HTTP Requests

Use the `test.http` file or curl:

```bash
# Create order (processes payment via external gateway)
curl -X POST http://localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": ["Laptop", "Mouse", "Keyboard"],
    "total_amount": 1299.99,
    "currency": "USD"
  }'

# Get order
curl http://localhost:3000/orders/order_1

# Refund order (via external gateway)
curl -X POST http://localhost:3000/orders/order_1/refund
```

## 📂 Project Structure

```
06-external-services/
├── main.go                           # Main application entry point
├── config.yaml                       # Configuration with external service definitions
├── test.http                         # HTTP test scenarios
├── index                         # This file
│
├── mock-payment-gateway/
│   └── main.go                       # Mock external payment API
│
└── service/
    ├── payment_service_remote.go     # Proxy to external payment gateway
    └── order_service.go              # Business logic using external payment
```

## 🔑 Key Concepts

### 1. External Service Definition

Define external services in `config.yaml` with URL and factory type:

```yaml
external-service-definitions:
  payment-gateway:
    url: "http://localhost:9000"
    type: payment-service-remote-factory
```

**What it does:**
- Declares external API location
- Specifies factory type for creating wrapper
- Framework creates proxy.Service automatically with this URL

### 2. Remote Service Wrapper

Create a clean service wrapper without embedded metadata:

```go
// PaymentServiceRemote wraps external payment API
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
    return &PaymentServiceRemote{
        proxyService: proxyService,
    }
}

// Method names can be non-standard (routes defined in RegisterServiceType)
func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}

func (s *PaymentServiceRemote) GetPayment(p *GetPaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "GetPayment", p)
}

func (s *PaymentServiceRemote) Refund(p *RefundParams) (*RefundResponse, error) {
    return proxy.CallWithData[*RefundResponse](s.proxyService, "Refund", p)
}
```

**Key points:**
- ✅ Simple struct with `proxyService` field
- ✅ NO metadata interfaces (no ServiceMeta!)
- ✅ Method names can be non-standard (CreatePayment, GetPayment)
- ✅ Routes defined separately in `RegisterServiceType`

### 3. Service Registration with Metadata

Register in `main.go` with all metadata and route overrides:

```go
// Register remote-only service (nil local factory)
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil,                                    // No local implementation
    svc.PaymentServiceRemoteFactory,        // Remote factory
    deploy.WithResource("payment", "payments"),
    deploy.WithConvention("rest"),
    // Route overrides for non-standard method names
    deploy.WithRouteOverride("CreatePayment", "POST /payments"),
    deploy.WithRouteOverride("GetPayment", "GET /payments/{id}"),
    deploy.WithRouteOverride("Refund", "POST /payments/{id}/refund"),
)

// Register local business service with custom action
lokstra_registry.RegisterServiceType(
    "order-service-factory",
    svc.OrderServiceFactory, nil,
    deploy.WithResource("order", "orders"),
    deploy.WithConvention("rest"),
    // Custom action route
    deploy.WithRouteOverride("Refund", "POST /orders/{id}/refund"),
)
```

**Why route overrides?**
- `CreatePayment`, `GetPayment` ≠ standard REST names (`Create`, `Get`)
- `Refund` is custom action (not standard REST)
- Allows matching external API exactly as-is

### 4. Remote Factory Implementation

Framework injects `proxy.Service` via `config["remote"]`:

```go
func PaymentServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return NewPaymentServiceRemote(
        service.CastProxyService(config["remote"]),
    )
}
```

**What happens:**
1. Framework reads `external-service-definitions.payment-gateway.url`
2. Creates `proxy.Service` with URL = `"http://localhost:9000"`
3. Passes it via `config["remote"]` to factory
4. Factory wraps it in `PaymentServiceRemote`

### 5. Business Service Using External Service

Clean service code with standard REST method names:

```go
type OrderService struct {
    Payment *service.Cached[*PaymentServiceRemote]
}

func OrderServiceFactory(deps map[string]any, config map[string]any) any {
    return &OrderService{
        Payment: service.Cast[*PaymentServiceRemote](deps["payment-gateway"]),
    }
}

// Standard REST method names (Create, Get, not CreateOrder, GetOrder)
func (s *OrderService) Create(p *OrderCreateParams) (*Order, error) {
    // Create order
    order := &Order{
        ID:     fmt.Sprintf("order_%d", orderID),
        Status: "pending",
        ...
    }
    
    // Process payment via external gateway
    payment, err := s.Payment.MustGet().CreatePayment(&CreatePaymentParams{
        Amount:      p.TotalAmount,
        Currency:    p.Currency,
        Description: fmt.Sprintf("Payment for order %s", order.ID),
    })
    
    if err != nil {
        order.Status = "failed"
        return nil, fmt.Errorf("payment failed: %w", err)
    }
    
    order.PaymentID = payment.ID
    order.Status = "paid"
    return order, nil
}

func (s *OrderService) Get(p *OrderGetParams) (*Order, error) {
    // Retrieve order by ID
}

func (s *OrderService) Refund(p *OrderRefundParams) (*Order, error) {
    // Process refund via external gateway
    _, err := s.Payment.MustGet().Refund(&RefundParams{
        ID: order.PaymentID,
    })
    
    if err != nil {
        return nil, fmt.Errorf("refund failed: %w", err)
    }
    
    order.Status = "refunded"
    return order, nil
}
```

**Key points:**
- ✅ Clean service struct (no metadata interfaces!)
- ✅ Standard REST method names: `Create`, `Get`, `Refund`
- ✅ Only `Refund` needs route override (custom action)
- ✅ Depends on external service via `deps["payment-gateway"]`

## 🎯 Service Configuration

In `config.yaml`:

```yaml
# Define external API
external-service-definitions:
  payment-gateway:
    url: "http://localhost:9000"
    type: payment-service-remote-factory

# Define local business service
service-definitions:
  order-service:
    type: order-service-factory
    depends-on:
      - payment-gateway  # Reference external service

deployments:
  app:
    servers:
      api-server:
        base-url: "http://localhost"
        addr: ":3000"
        published-services:
          - order-service
        # Framework auto-detects payment-gateway dependency
```

**How it works:**
1. Framework reads `order-service` dependencies
2. Finds `payment-gateway` in `external-service-definitions`
3. Creates `proxy.Service` with URL from config
4. Calls `PaymentServiceRemoteFactory` with proxy
5. Injects into `OrderService` via `deps["payment-gateway"]`

## 🔄 Request Flow

1. **Client** → `POST /orders` to main app (:3000)
2. **OrderService** → Validate request, create order
3. **OrderService** → Call `Payment.MustGet().CreatePayment()`
4. **PaymentServiceRemote** → HTTP call to `:9000/payments`
5. **Mock Gateway** → Process payment, return payment ID
6. **OrderService** → Update order with payment ID, status = "paid"
7. **Client** ← Return order with payment details

## 📊 Comparison: proxy.Service vs proxy.Router

| Feature | proxy.Service (This Example) | proxy.Router (Example 07) |
|---------|------------------------------|---------------------------|
| **Use Case** | Structured external services | Quick API access |
| **Convention** | ✅ REST/JSON-RPC auto-routing | ❌ Manual paths |
| **Type Safety** | ✅ Typed methods | ❌ Generic calls |
| **Overrides** | ✅ Custom route overrides | N/A |
| **Service Wrapper** | ✅ Required | ❌ Not needed |
| **Best For** | Payment, Email, SMS APIs | Weather, Maps, Ad-hoc APIs |

**When to use proxy.Service:**
- External API has multiple related endpoints
- You want typed methods and reusability
- Need service dependency injection
- Example: Stripe, SendGrid, Twilio

**When to use proxy.Router:**
- One-off API calls
- Quick integration without wrapper
- No need for service abstraction
- Example: Weather API, Currency converter

## 🧪 Mock Payment Gateway

The mock gateway simulates a real payment provider **using Lokstra framework**:

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    "github.com/primadi/lokstra"
)

// In-memory storage
var (
    payments   = make(map[string]*Payment)
    paymentsMu sync.RWMutex
    nextID     = 1
)

// Handlers using Lokstra's handler form variations
func createPayment(req *CreatePaymentRequest) (*Payment, error) {
    if req.Currency == "" {
        req.Currency = "USD"
    }
    
    paymentsMu.Lock()
    id := fmt.Sprintf("pay_%d", nextID)
    nextID++
    
    payment := &Payment{
        ID:          id,
        Amount:      req.Amount,
        Currency:    req.Currency,
        Status:      "completed",
        Description: req.Description,
        CreatedAt:   time.Now(),
    }
    payments[id] = payment
    paymentsMu.Unlock()
    
    log.Printf("✅ Payment created: %s - $%.2f %s", id, req.Amount, req.Currency)
    return payment, nil
}

func getPayment(req *GetPaymentRequest) (*Payment, error) {
    paymentsMu.RLock()
    payment, exists := payments[req.ID]
    paymentsMu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("payment not found: %s", req.ID)
    }
    
    return payment, nil
}

func refundPayment(req *RefundRequest) (*RefundResponse, error) {
    paymentsMu.Lock()
    defer paymentsMu.Unlock()
    
    payment, exists := payments[req.ID]
    if !exists {
        return nil, fmt.Errorf("payment not found: %s", req.ID)
    }
    
    if payment.Status != "completed" {
        return nil, fmt.Errorf("only completed payments can be refunded")
    }
    
    now := time.Now()
    payment.Status = "refunded"
    payment.RefundedAt = &now
    
    log.Printf("💸 Payment refunded: %s", req.ID)
    
    return &RefundResponse{
        PaymentID:  req.ID,
        RefundedAt: now,
        Status:     "refunded",
        Message:    fmt.Sprintf("Payment %s has been refunded", req.ID),
    }, nil
}

func main() {
    // Create router with Lokstra
    r := lokstra.NewRouter("payment-api")
    
    // Register routes
    r.POST("/payments", createPayment)
    r.GET("/payments/{id}", getPayment)
    r.POST("/payments/{id}/refund", refundPayment)
    
    // Start server
    app := lokstra.NewApp("payment-gateway", ":9000", r)
    if err := app.Run(30 * time.Second); err != nil {
        log.Fatalf("Failed to run app: %v", err)
    }
}
```

**Key points:**
- ✅ Built with Lokstra (not standard http package)
- ✅ Demonstrates Lokstra's handler form flexibility
- ✅ Uses struct parameters with validation tags
- ✅ In-memory storage with sync.RWMutex
- ✅ Instant success (status = "completed")
- ✅ Simple refund logic

**Endpoints:**
- `POST /payments` - Create payment
- `GET /payments/{id}` - Get payment status
- `POST /payments/{id}/refund` - Refund payment

## 🎓 Learning Points

### 1. External Service Integration Pattern

```
External API → proxy.Service → Service Wrapper → Business Service
```

This pattern:
- Isolates external API details
- Provides typed interface
- Enables testing with mocks
- Centralizes error handling

### 2. Route Overrides for Non-Standard APIs

Use `deploy.WithRouteOverride()` when:
- Method names don't match REST (`CreatePayment` vs `Create`)
- Custom actions needed (`POST /orders/{id}/refund`)
- External API has specific requirements

**Standard REST methods (no override needed):**
- `Create()` → `POST /resource`
- `Get()` → `GET /resource/{id}`
- `Update()` → `PUT /resource/{id}`
- `Delete()` → `DELETE /resource/{id}`
- `List()` → `GET /resource`

**Non-standard (override required):**
- `CreatePayment()` → needs `POST /payments`
- `Refund()` → needs `POST /payments/{id}/refund`

### 3. Clean Separation of Concerns

- **Service code**: Pure logic, no metadata
- **Registration**: Metadata + route overrides in `main.go`
- **Config**: Deployment topology only

This makes services:
- Easier to test (no framework coupling)
- Simpler to understand (one responsibility)
- More maintainable (metadata in one place)

### 4. Error Handling

When external service fails:

```go
payment, err := s.Payment.MustGet().CreatePayment(...)
if err != nil {
    order.Status = "failed"
    return nil, fmt.Errorf("payment failed: %w", err)
}
```

Always handle external failures gracefully and update your domain state accordingly!

## 🔄 Next Steps

1. ✅ **Example 06** - External Services (You are here)
2. 📖 **Example 07** - Remote Router (`proxy.Router` for quick API access)

## 🎯 Real-World Examples

This pattern works for any external API:

**Payment Gateways:**
- Stripe: `stripe-service-remote` → `https://api.stripe.com`
- PayPal: `paypal-service-remote` → `https://api.paypal.com`

**Communication:**
- SendGrid: `email-service-remote` → `https://api.sendgrid.com`
- Twilio: `sms-service-remote` → `https://api.twilio.com`

**Storage:**
- AWS S3: `s3-service-remote` → `https://s3.amazonaws.com`
- Cloudinary: `cdn-service-remote` → `https://api.cloudinary.com`

All follow the same pattern: define external service → create wrapper → use in business services!

## 📚 Related Documentation

- [Architecture - Service Categories](../../architecture#service-categories)
- [Architecture - Proxy Patterns](../../architecture#proxy-patterns)
- [Remote Services Guide](../../../01-essentials/02-service)
- [Configuration Guide](../../../01-essentials/03-configuration)

---

**💡 Key Takeaway:** Use `proxy.Service` to wrap external APIs as typed Lokstra services with convention-based routing and custom overrides. For simpler one-off calls, use `proxy.Router` (Example 07).
