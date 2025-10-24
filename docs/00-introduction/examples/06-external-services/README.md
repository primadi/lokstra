# Example 06 - External Services Integration

This example demonstrates how to **integrate external APIs** (like payment gateways, email services, SMS providers) as Lokstra services using `proxy.Service` for convention-based remote calls.

## 📋 What You'll Learn

- ✅ Wrapping third-party APIs as Lokstra services
- ✅ Using `proxy.Service` for remote HTTP calls
- ✅ **Custom route overrides in code** (not config!) using `ServiceMeta`
- ✅ `external-service-definitions` with auto-wrapper creation
- ✅ Business services depending on external services
- ✅ Error handling when external service fails
- ✅ **Flexible metadata**: Works for both local and remote services
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
│  │  - Refund()    → POST /orders/{id}/refund (custom)   │  │
│  │                                                       │  │
│  │  📋 ServiceMeta: route override for Refund()         │  │
│  └─────────────────┬────────────────────────────────────┘  │
│                    │ depends on                            │
│                    ▼                                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PaymentServiceRemote (proxy.Service)                │  │
│  │  - CreatePayment()  → POST /payments                 │  │
│  │  - GetPayment()     → GET /payments/{id}             │  │
│  │  - Refund()         → POST /payments/{id}/refund     │  │
│  │                                                       │  │
│  │  📋 ServiceMeta: route overrides for all methods     │  │
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
├── README.md                         # This file
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

Define external services in `config.yaml` with **auto-wrapper creation**:

```yaml
external-service-definitions:
  payment-gateway:
    url: "http://localhost:9000"
    type: payment-service-remote-factory  # Auto-creates service wrapper!
```

**New DX improvement:** Framework automatically creates service definition from `external-service-definitions` when `type` is specified. No need to duplicate in `service-definitions`!

### 2. Remote Service Wrapper with ServiceMeta

Create a service that wraps the external API using `ServiceMeta` (works for local & remote):

```go
type PaymentServiceRemote struct {
    service.ServiceMetaAdapter  // Renamed from RemoteServiceMetaAdapter
}

func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
    return &PaymentServiceRemote{
        ServiceMetaAdapter: service.ServiceMetaAdapter{
            Resource:     "payment",
            Plural:       "payments",
            Convention:   "rest",
            ProxyService: proxyService,
            // Route overrides IN CODE (not config!)
            Override: autogen.RouteOverride{
                Custom: map[string]autogen.Route{
                    // All methods need overrides because names don't match REST convention
                    "CreatePayment": {Method: "POST", Path: "/payments"},
                    "GetPayment":    {Method: "GET", Path: "/payments/{id}"},
                    "Refund":        {Method: "POST", Path: "/payments/{id}/refund"},
                },
            },
        },
    }
}
```

**Key points:**
- Uses `ServiceMetaAdapter` (renamed, works for both local & remote)
- `Convention: "rest"` enables auto-routing for standard names
- `Override.Custom` allows custom routes for **non-standard method names**
- Method names like `CreatePayment` don't match REST convention (should be `Create`), so we override them

### 3. Remote Factory Pattern

The framework injects `proxy.Service` via `config["remote"]`:

```go
func PaymentServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return NewPaymentServiceRemote(
        service.CastProxyService(config["remote"]),
    )
}
```

Register the factory with **nil local factory** (remote-only):

```go
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil,                                    // Local factory = nil
    service.PaymentServiceRemoteFactory,    // Remote factory
)
```

### 4. Custom Route Overrides in Code

**New best practice:** Route overrides are now in **code**, not config!

For external services with non-standard method names:

```go
Override: autogen.RouteOverride{
    Custom: map[string]autogen.Route{
        "CreatePayment": {Method: "POST", Path: "/payments"},
        "GetPayment":    {Method: "GET", Path: "/payments/{id}"},
        "Refund":        {Method: "POST", Path: "/payments/{id}/refund"},
    },
},
```

For local services with custom actions:

```go
// In OrderService
func (s *OrderService) GetRouteOverride() autogen.RouteOverride {
    return autogen.RouteOverride{
        Custom: map[string]autogen.Route{
            "Refund": {Method: "POST", Path: "/orders/{id}/refund"},
        },
    }
}
```

**Why in code?**
- ✅ Type-safe and discoverable
- ✅ Co-located with service implementation
- ✅ No duplication between code and config
- ✅ Easier refactoring

**When to use overrides:**
- Method name doesn't match REST convention (`CreatePayment` vs `Create`)
- Custom actions (`POST /orders/{id}/refund` vs standard `PUT /orders/{id}`)
- Non-standard HTTP methods

### 5. Business Service Using External Service

```go
type OrderService struct {
    Payment *service.Cached[*PaymentServiceRemote]
}

// Implement ServiceMeta for route overrides
func (s *OrderService) GetResourceName() (string, string) {
    return "order", "orders"
}

func (s *OrderService) GetConventionName() string {
    return "rest"
}

func (s *OrderService) GetRouteOverride() autogen.RouteOverride {
    return autogen.RouteOverride{
        Custom: map[string]autogen.Route{
            "Refund": {Method: "POST", Path: "/orders/{id}/refund"},
        },
    }
}

// Method names match REST convention (Create, Get, not CreateOrder, GetOrder)
func (s *OrderService) Create(p *OrderCreateParams) (*Order, error) {
    // Create order first
    order := &Order{...}
    
    // Process payment via external gateway
    payment, err := s.Payment.MustGet().CreatePayment(&CreatePaymentParams{
        Amount: p.TotalAmount,
        Currency: p.Currency,
    })
    
    if err != nil {
        order.Status = "failed"
        return nil, fmt.Errorf("payment failed: %w", err)
    }
    
    order.PaymentID = payment.ID
    order.Status = "paid"
    return order, nil
}
```

**Key changes:**
- ✅ Implement `ServiceMeta` interface for metadata
- ✅ Method names: `Create`, `Get`, `Refund` (match REST convention)
- ✅ Custom route override for `Refund` in code

## 🎯 Service Configuration

In `config.yaml`:

```yaml
# Define external API and auto-create wrapper
external-service-definitions:
  payment-gateway:
    url: "http://localhost:9000"
    type: payment-service-remote-factory  # Auto-creates service definition!

# Define local business service
service-definitions:
  order-service:
    type: order-service-factory
    depends-on:
      - payment-gateway  # Direct reference to external service

deployments:
  app:
    servers:
      api-server:
        base-url: "http://localhost"
        addr: ":3000"
        
        # External payment service (remote only)
        required-remote-services:
          - payment-gateway  # Framework resolves URL automatically
        
        # Auto-generates router with metadata from OrderService
        published-services:
          - order-service
```

**Important DX improvements:**
- ✅ `external-service-definitions` with `type` auto-creates service wrapper
- ✅ No duplication in `service-definitions` for external services
- ✅ `published-services` auto-generates router with metadata from code
- ✅ Route overrides in code, not config!

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

The mock gateway simulates a real payment provider:

```go
// In-memory payment storage
var payments = make(map[string]*Payment)

// Create payment
router.POST("/payments", func(ctx *Context) error {
    var req CreatePaymentRequest
    if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid request"})
    }
    
    payment := &Payment{
        ID:       fmt.Sprintf("pay_%d", paymentID),
        Amount:   req.Amount,
        Currency: req.Currency,
        Status:   "completed",
    }
    
    payments[payment.ID] = payment
    return ctx.JSON(200, payment)
})
```

**Endpoints:**
- `POST /payments` - Create payment
- `GET /payments/{id}` - Get payment status
- `POST /payments/{id}/refund` - Refund payment

## 🎓 Learning Points

### 1. External Service Integration Pattern

```
External API → Service Wrapper (proxy.Service) → Business Service
```

This pattern:
- Isolates external API details
- Provides typed interface
- Enables testing with mocks
- Centralizes error handling

### 2. Convention-Based Routing

`proxy.Service` auto-generates routes for **standard REST method names**:
- `Create()` → `POST /payments`
- `Get(id)` → `GET /payments/{id}`
- `Update(id)` → `PUT /payments/{id}`
- `Delete(id)` → `DELETE /payments/{id}`

**Non-standard names need overrides:**
- `CreatePayment()` → Override: `POST /payments`
- `GetPayment(id)` → Override: `GET /payments/{id}`
- `Refund(id)` → Override: `POST /payments/{id}/refund`

### 3. Custom Routes for Non-Standard APIs

Not all APIs follow REST conventions. Use `Override.Custom` for:
- Custom actions: `POST /payments/{id}/refund` (not `PUT /payments/{id}`)
- Non-standard method names: `CreatePayment` vs `Create`
- Special operations: `POST /users/{id}/reset-password`

**Best practice:** If possible, use standard REST method names (`Create`, `Get`, `Update`, `Delete`) to avoid needing overrides. Use overrides only when necessary (custom actions, external API constraints).

### 4. Error Handling

When external service fails:

```go
payment, err := s.Payment.MustGet().CreatePayment(...)
if err != nil {
    // Mark order as failed
    order.Status = "failed"
    return nil, fmt.Errorf("payment failed: %w", err)
}
```

Always handle external failures gracefully!

## 🔄 Next Steps

1. ✅ **Example 06** - External Services (You are here)
2. 📖 **Example 07** - Remote Router (`proxy.Router` for quick API access)
3. 📖 **Example 08** - Testing with Mock Services

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

- [Architecture - Service Categories](../../architecture.md#service-categories)
- [Architecture - Proxy Patterns](../../architecture.md#proxy-patterns)
- [Remote Services Guide](../../../01-essentials/02-service/README.md)
- [Configuration Guide](../../../01-essentials/03-configuration/README.md)

---

**💡 Key Takeaway:** Use `proxy.Service` to wrap external APIs as typed Lokstra services with convention-based routing and custom overrides. For simpler one-off calls, use `proxy.Router` (Example 07).
