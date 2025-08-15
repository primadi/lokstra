# Module Configuration Implementation

This document explains the implementation of module configuration features in `core/config/new_server.go`.

## Overview

The `startModulesFromConfig` function has been enhanced to support all fields defined in `ModuleConfig`:

- ✅ **RequiredServices** - Check if required services exist before loading the module
- ✅ **CreateServices** - Create service instances from module configuration  
- ✅ **RegisterServiceFactories** - Call plugin methods to register service factories
- ✅ **RegisterHandlers** - Call plugin methods to register handlers
- ✅ **RegisterMiddleware** - Call plugin methods to register middleware

## Implementation Details

### 1. Required Services Validation

Before loading a module, the system checks if all required services are available:

```go
// Check required services
for _, serviceName := range mod.RequiredServices {
    if _, err := regCtx.GetService(serviceName); err != nil {
        return fmt.Errorf("module %s requires service %s which is not available: %w", mod.Name, serviceName, err)
    }
}
```

### 2. Service Creation

Services defined in the module configuration are created automatically:

```go
// Create services defined in the module
for _, serviceConfig := range mod.CreateServices {
    if err := createServiceFromConfig(regCtx, &serviceConfig); err != nil {
        return fmt.Errorf("module %s failed to create service %s: %w", mod.Name, serviceConfig.Name, err)
    }
}
```

### 3. Plugin Method Calls

For modules with plugins, specific methods can be called to register factories, handlers, and middleware:

```go
// Register service factories from the module
if mod.Path != "" && len(mod.RegisterServiceFactories) > 0 {
    if err := callModuleMethods(mod.Path, mod.RegisterServiceFactories, regCtx, "service factory"); err != nil {
        return fmt.Errorf("module %s failed to register service factories: %w", mod.Name, err)
    }
}
```

## Helper Functions

### `createServiceFromConfig`

Creates a service instance from `ServiceConfig`:

```go
func createServiceFromConfig(regCtx iface.RegistrationContext, serviceConfig *ServiceConfig) error {
    _, err := regCtx.CreateService(serviceConfig.Type, serviceConfig.Name, serviceConfig.Config)
    return err
}
```

### `callModuleMethods`

Calls specified methods from a plugin module:

```go
func callModuleMethods(pluginPath string, methodNames []string, regCtx iface.RegistrationContext, methodType string) error {
    // Open plugin, lookup methods, call with proper signature validation
}
```

## Expected Plugin Method Signatures

All plugin methods must have the signature:
```go
func MethodName(regCtx iface.RegistrationContext) error
```

This applies to:
- Service factory registration methods
- Handler registration methods  
- Middleware registration methods

## Configuration Example

```yaml
modules:
  - name: auth-module
    path: ./plugins/auth.so
    entry: GetAuthModule
    
    required_services:
      - "logger.default"
      - "main-db"
    
    create_services:
      - name: "jwt-validator"
        type: "lokstra.jwt_auth"
        config:
          secret_key: "${JWT_SECRET}"
          
    register_service_factories:
      - "RegisterAuthServiceFactory"
      
    register_handlers:
      - "RegisterAuthHandlers"
      
    register_middleware:
      - "RegisterAuthMiddleware"
```

## Error Handling

The implementation provides detailed error messages for each stage:

- Missing required services
- Service creation failures  
- Plugin method lookup failures
- Method execution failures
- Invalid method signatures

## Configuration-Only Modules

Modules can be defined without a plugin path for configuration-only purposes:

```yaml
modules:
  - name: logging-setup
    # No path - configuration only
    create_services:
      - name: "app-logger"
        type: "lokstra.logger"
        config:
          level: "info"
```

This allows modules to serve as configuration containers for related services without requiring plugin code.
