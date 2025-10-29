# Remote Services

This example demonstrates HTTP-based service communication patterns using remote service wrappers and HTTP clients.

## Overview

Remote services allow you to wrap external HTTP APIs as local service interfaces, providing a consistent programming model across local and remote services.

**Topics Covered:**
- Remote service pattern
- HTTP client usage (ClientRouter)
- Type-safe remote calls
- Configuration-driven remote URLs

---

## Remote Service Pattern

### Basic Structure

```go
type RemoteUserService struct {
    client *api_client.ClientRouter
}

func NewRemoteUserService(routerName, pathPrefix string) *RemoteUserService {
    return &RemoteUserService{client: client}
}

func (s *RemoteUserService) GetUser(id int) (*UserResponse, error) {
    // Make HTTP call to remote API
    path := fmt.Sprintf("/api/v1/users/%d", id)
    return api_client.FetchAndCast[*UserResponse](s.client, path,
        api_client.WithMethod("GET"),
    )
}
```

---

## Running the Example

```bash
cd docs/02-deep-dive/02-service/examples/02-remote-services
go run main.go
```

**Test Endpoints:**
```bash
curl http://localhost:3000/users/1
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'
```

---

**Status**: ✅ Complete
