# Published Services Design

## Overview

Major architectural improvement for Example 04 to simplify multi-deployment configuration with:
1. **Auto-generated routers** from published services
2. **Auto-resolved remote services** from deployment registry
3. **Simplified configuration** (published services imply local requirement)

---

## Key Design Decisions

### 1. Convention Names

**Decision:** Use **empty string** as default REST convention

```go
// Convention registry behavior:
ConversionRule{
    Convention: "",      // Empty = defaults to "rest"
    Convention: "rest",  // Explicit REST
    Convention: "grpc",  // gRPC
}
```

**Rationale:** 
- Simpler API - no need to remember "standard" vs "rest"
- Clear default behavior
- Explicit when needed

### 2. RouteOverride Priority

**Decision:** **Config overrides Code** (Config has higher priority)

**Scenarios:**

#### Code-only (No config override):
```go
// Code defines base routing
NewUserServiceRemote(baseURL) 
// Uses default REST convention, no overrides
```

#### Config overrides:
```yaml
# Future: When we implement remote-service resolution
remote-service-overrides:
  user-service:
    path-prefix: "/v2"
    custom-routes:
      GetByID: 
        method: "GET"
        path: "/users/{id}"
```

**Rationale:**
- Code: Development defaults
- Config: Deployment-specific flexibility
- Config should be able to override for different environments

### 3. Unix Socket Support

**Decision:** Support unix sockets for local IPC

```go
func buildURL(baseURL, addr string) string {
    if strings.HasPrefix(addr, "unix:") {
        return addr  // e.g., "unix:/tmp/order.sock"
    }
    port := strings.TrimPrefix(addr, ":")
    return baseURL + ":" + port
}
```

**Use Case:** Microservices on same server can communicate via unix socket (faster, more secure)

```yaml
deployments:
  monolith:
    servers:
      api:
        base-url: "http://localhost"
        apps:
          - addr: "unix:/tmp/user-service.sock"  # Local IPC
            published-services: [user-service]
          - addr: "unix:/tmp/order-service.sock"
            required-remote-services: [user-service]  # Via unix socket!
            published-services: [order-service]
```

### 4. Convention Fallback for Non-CRUD Methods

**Decision:** POST to `/actions/{method_name_snake_case}`

```go
// Standard CRUD conventions
"List"     → GET  /resources
"GetByID"  → GET  /resources/{id}
"Create"   → POST /resources
"Update"   → PUT  /resources/{id}
"Delete"   → DELETE /resources/{id}

// Non-CRUD methods → POST /actions/{method_name_snake_case}
"ValidateUser"  → POST /actions/validate_user
"SendEmail"     → POST /actions/send_email
"Login"         → POST /actions/login
"Logout"        → POST /actions/logout
```

**Rationale:**
- `/resources/` for resource-oriented operations (REST)
- `/actions/` for action-oriented operations (RPC-style)
- Clear separation of concerns

### 5. Service Capability Matrix

**Service Types:**

| Service Type | Local Factory | Remote Factory | Can be Published? | Can be Remote? |
|-------------|---------------|----------------|-------------------|----------------|
| Database | ✅ | ❌ | ❌ | ❌ |
| User Service | ✅ | ✅ | ✅ | ✅ |
| Order Service | ✅ | ✅ | ✅ | ✅ |
| Health Service | ✅ | ✅ | ✅ | ✅ |
| External API | ❌ | ✅ | ❌ | ✅ |

**Registration:**
```go
// Local-only service (cannot be published)
RegisterServiceType("database-factory", DatabaseFactory, nil)

// Both local and remote (can be published and consumed remotely)
RegisterServiceType("user-service-factory", UserServiceFactory, UserServiceRemoteFactory)

// Remote-only service (3rd party API)
RegisterServiceType("external-api-factory", nil, ExternalAPIFactory)
```

### 6. Published Services Imply Local Requirement

**Decision:** Services listed in `published-services` are automatically added as local requirements

**Config:**
```yaml
servers:
  order-api:
    base-url: "http://localhost"
    required-services: [database]  # Only non-published services
    required-remote-services: [user-service]
    apps:
      - addr: ":3005"
        published-services: [order-service, health]  # Auto-implies local requirement
```

**Equivalent to:**
```yaml
servers:
  order-api:
    required-services: [database, order-service, health]  # ❌ Redundant!
    required-remote-services: [user-service]
    apps:
      - addr: ":3005"
        published-services: [order-service, health]
```

**Rationale:**
- Less duplication
- Clearer intent: published = local + exposed
- Simpler configuration

---

## Implementation Status

### ✅ Completed:
1. Created `order_service_remote.go` with `proxy.Service`
2. Updated `user_service_remote.go` to use `proxy.Service`
3. Both use empty string convention (default REST)
4. No code-level RouteOverrides (config can override later)

### 🔄 Next Steps:
1. Update schema (add `published-services` to `AppDefMap`)
2. Implement `FindPublishedService()` in GlobalRegistry
3. Update loader to auto-add published services as local requirements
4. Implement remote service auto-resolution
5. Implement auto-router generation from published services
6. Create health service
7. Update Example 04 config
8. Test all deployments

---

## Future Enhancements

### Config-based Remote Service Resolution:
```yaml
# Future syntax for explicit remote service configuration
remote-service-overrides:
  user-service:
    # Auto-resolve from deployment, but override convention
    convention: "grpc"
    path-prefix: "/v2"
    custom-routes:
      GetByID:
        method: "GET"
        path: "/api/users/{id}"
```

### Multi-Publisher Disambiguation:
```yaml
# If service published in multiple places
required-remote-services:
  - user-service@user-api    # Explicit server selection
  - user-service@monolith.api  # deployment.server format
```

---

## Design Philosophy

1. **Convention over Configuration:** Default REST behavior with overrides when needed
2. **Config over Code:** Runtime configuration has higher priority
3. **Simplicity First:** Remove redundancy, make common cases easy
4. **Flexibility When Needed:** Support advanced use cases (unix socket, custom routes, etc.)
5. **Type Safety:** Use generics and interfaces for compile-time checks
6. **Auto-Magic with Escape Hatches:** Auto-generate when possible, allow manual control when needed
