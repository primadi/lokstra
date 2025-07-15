# Service Development Guide

This guide explains how to create custom services in Lokstra.

## Service Interface

All services must implement the basic `Service` interface:

```go
type Service interface {
    InstanceName() string
    GetConfig(key string) any
}
```

## Service Module Interface

To make your service configurable and registerable, implement the `ServiceModule` interface:

```go
type ServiceModule interface {
    Name() string
    Factory(config any) (Service, error)
    Meta() *ServiceMeta
}
```

## Creating a Custom Service

### 1. Define Your Service

```go
package myservice

import "lokstra/common/iface"

type MyService struct {
    instanceName string
    config       map[string]any
    // Add your service-specific fields
}

func (s *MyService) InstanceName() string {
    return s.instanceName
}

func (s *MyService) GetConfig(key string) any {
    return s.config[key]
}

// Add your service-specific methods
func (s *MyService) DoSomething() error {
    // Implementation
    return nil
}
```

### 2. Create Service Factory

```go
func newMyService(instanceName string, config map[string]any) (*MyService, error) {
    // Validate configuration
    requiredField, ok := config["required_field"].(string)
    if !ok {
        return nil, fmt.Errorf("myservice requires 'required_field' in config")
    }

    return &MyService{
        instanceName: instanceName,
        config:       config,
        // Initialize your fields
    }, nil
}

func ServiceFactory(config any) (iface.Service, error) {
    configMap, ok := config.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("myservice requires configuration as map")
    }

    instanceName := "myservice"
    if name, ok := configMap["instance_name"].(string); ok {
        instanceName = name
    }

    return newMyService(instanceName, configMap)
}
```

### 3. Implement Service Module

```go
type ServiceModule struct{}

func (s *ServiceModule) Name() string {
    return "myapp.myservice"
}

func (s *ServiceModule) Factory(config any) (iface.Service, error) {
    return ServiceFactory(config)
}

func (s *ServiceModule) Meta() *iface.ServiceMeta {
    return &iface.ServiceMeta{
        Description: "My custom service for doing something",
        Tags:        []string{"custom", "business"},
    }
}

func GetModule() iface.ServiceModule {
    return &ServiceModule{}
}
```

### 4. Registration Helper

```go
type Registration struct{}

func (r *Registration) RegisterService(ctx module.RegistrationContext) {
    ctx.RegisterServiceFactory("myservice", ServiceFactory)
}
```

## Using Your Service

### In Code

```go
func main() {
    ctx := lokstra.NewGlobalContext()
    
    // Register your service module
    ctx.RegisterServiceModule(myservice.GetModule())
    
    app := lokstra.NewApp(ctx, "my-app", ":8080")
    
    app.GET("/test", func(ctx *lokstra.Context) error {
        // Get your service
        service, err := ctx.GetService("myservice")
        if err != nil {
            return ctx.ErrorInternal("Service not available")
        }
        
        // Cast to your service type
        myService := service.(*myservice.MyService)
        
        // Use your service
        err = myService.DoSomething()
        if err != nil {
            return ctx.ErrorInternal("Service error")
        }
        
        return ctx.Ok("Success")
    })
    
    app.Start()
}
```

### In YAML Configuration

```yaml
services:
  - type: myapp.myservice
    name: my-custom-service
    config:
      required_field: "some value"
      optional_field: 42
```

## Best Practices

### Configuration Handling

1. **Validate Required Fields**: Always check for required configuration fields
2. **Provide Defaults**: Set sensible defaults for optional fields
3. **Type Safety**: Use type assertions with proper error handling
4. **Environment Variables**: Support environment variable overrides

```go
func parseConfig(config map[string]any) (*MyConfig, error) {
    cfg := &MyConfig{
        // Set defaults
        Timeout: 30 * time.Second,
        Retries: 3,
    }
    
    // Required fields
    if host, ok := config["host"].(string); ok {
        cfg.Host = host
    } else {
        return nil, fmt.Errorf("host is required")
    }
    
    // Optional fields with type checking
    if timeout, ok := config["timeout"].(int); ok {
        cfg.Timeout = time.Duration(timeout) * time.Second
    }
    
    return cfg, nil
}
```

### Error Handling

1. **Initialization Errors**: Return clear errors during service creation
2. **Runtime Errors**: Handle errors gracefully in service methods
3. **Context Cancellation**: Respect context cancellation in long-running operations

```go
func (s *MyService) ProcessData(ctx context.Context, data []byte) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process data
        return nil
    }
}
```

### Resource Management

1. **Cleanup**: Implement cleanup methods for resources
2. **Connection Pooling**: Use connection pools for external services
3. **Graceful Shutdown**: Support graceful shutdown

```go
type MyService struct {
    // ... other fields
    client *http.Client
    done   chan struct{}
}

func (s *MyService) Close() error {
    close(s.done)
    // Cleanup resources
    return nil
}
```

### Testing

1. **Unit Tests**: Test service logic independently
2. **Integration Tests**: Test service with real dependencies
3. **Mock Dependencies**: Use interfaces for testability

```go
type ExternalAPI interface {
    Call(data []byte) error
}

type MyService struct {
    api ExternalAPI // Interface for testing
}

// In tests
func TestMyService(t *testing.T) {
    mockAPI := &MockExternalAPI{}
    service := &MyService{api: mockAPI}
    // Test service
}
```

## Service Lifecycle

Services in Lokstra follow this lifecycle:

1. **Registration**: Service module is registered with GlobalContext
2. **Configuration**: Service factory receives configuration
3. **Creation**: Service instance is created and validated
4. **Usage**: Service is accessed via `ctx.GetService()`
5. **Cleanup**: Service resources are cleaned up on shutdown

## Examples

See the following examples for reference:
- [Redis Service](../services/redis/) - External service integration
- [Email Service](../services/email/) - SMTP service implementation
- [Health Service](../services/health/) - Health check service
- [Metrics Service](../services/metrics/) - Prometheus metrics
