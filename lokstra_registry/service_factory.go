package lokstra_registry

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/primadi/lokstra/common/json"
)

// Service factory support 3 forms service factory function:
// 1. func (config map[string]any) any
// 2. func (config any) any
// 3. func () any  - no config
type ServiceFactory = func(config map[string]any) any
type AnyServiceFactory = any

// Service factory registry - separate for local and remote
type serviceFactoryEntry struct {
	localFactory  ServiceFactory
	remoteFactory ServiceFactory
}

var serviceFactoryRegistry sync.Map

func RegisterServiceFactoryLocalAndRemote(serviceType string,
	localFactory, remoteFactory AnyServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && (entry.localFactory != nil || entry.remoteFactory != nil) {
			panic("service factory for type " + serviceType + " already registered")
		}
	}

	entry.localFactory = adaptServiceFactory(localFactory)
	entry.remoteFactory = adaptServiceFactory(remoteFactory)

	serviceFactoryRegistry.Store(serviceType, entry)
}

// RegisterServiceFactory registers BOTH local and remote factory (backward compatibility)
// Deprecated: Use RegisterServiceFactoryLocal and RegisterServiceFactoryRemote instead
func RegisterServiceFactory(serviceType string, factory AnyServiceFactory,
	opts ...RegisterOption) {
	RegisterServiceFactoryLocal(serviceType, factory, opts...)
}

// RegisterServiceFactoryLocal registers a factory for creating LOCAL service instances
// This factory will be used when the service needs to run in the same process
func RegisterServiceFactoryLocal(serviceType string, factory AnyServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && entry.localFactory != nil {
			panic("local service factory for type " + serviceType + " already registered")
		}
	}

	entry.localFactory = adaptServiceFactory(factory)
	serviceFactoryRegistry.Store(serviceType, entry)
}

// RegisterServiceFactoryRemote registers a factory for creating REMOTE service client instances
// This factory will be used when the service needs to call a remote instance via HTTP
func RegisterServiceFactoryRemote(serviceType string, factory AnyServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && entry.remoteFactory != nil {
			panic("remote service factory for type " + serviceType + " already registered")
		}
	}

	entry.remoteFactory = adaptServiceFactory(factory)
	serviceFactoryRegistry.Store(serviceType, entry)
}

// GetServiceFactory retrieves the appropriate factory (local or remote) based on service location
// Returns local factory if service should run locally, remote factory otherwise
// Framework automatically decides based on:
// - Router registration (same server = local, different server = remote)
// - ClientRouter metadata (IsLocal flag)
// serviceName is the instance name (e.g., "user-service"), serviceType is the factory type (e.g., "user_service")
func GetServiceFactory(serviceType string, serviceName string) ServiceFactory {
	entryAny, ok := serviceFactoryRegistry.Load(serviceType)
	if !ok {
		return nil
	}

	entry := entryAny.(*serviceFactoryEntry)

	// Determine if we need local or remote factory
	// Strategy: Check if there's a ClientRouter for this service name (NOT serviceType)
	// If ClientRouter exists and IsLocal=true -> use local
	// If ClientRouter exists and IsLocal=false -> use remote
	// If no ClientRouter -> default to local

	client := GetClientRouter(serviceName) // Use serviceName instead of serviceType
	// fmt.Printf("[DEBUG GetServiceFactory] serviceType=%s, serviceName=%s, client=%v, IsLocal=%v",
	// 	serviceType, serviceName, client != nil, client != nil && client.IsLocal)
	// if client != nil {
	// 	fmt.Printf(", ServerName=%s, currentServer=%s\n", client.ServerName, GetCurrentServerName())
	// } else {
	// 	fmt.Printf("\n")
	// }
	if client != nil {
		if client.IsLocal {
			// Service is on same server - use local factory
			if entry.localFactory == nil {
				panic(fmt.Sprintf("service %s requires local factory but only remote is registered", serviceType))
			}
			return entry.localFactory
		}
		// Service is on different server - use remote factory
		if entry.remoteFactory == nil {
			panic(fmt.Sprintf("service %s requires remote factory but only local is registered", serviceType))
		}
		return entry.remoteFactory
	}

	// No ClientRouter found - default to local
	if entry.localFactory != nil {
		return entry.localFactory
	}

	// Fallback to remote if local not available
	return entry.remoteFactory
}

// GetServiceFactoryLocal explicitly gets the local factory (for testing/debugging)
func GetServiceFactoryLocal(serviceType string) ServiceFactory {
	if entryAny, ok := serviceFactoryRegistry.Load(serviceType); ok {
		entry := entryAny.(*serviceFactoryEntry)
		return entry.localFactory
	}
	return nil
}

// GetServiceFactoryRemote explicitly gets the remote factory (for testing/debugging)
func GetServiceFactoryRemote(serviceType string) ServiceFactory {
	if entryAny, ok := serviceFactoryRegistry.Load(serviceType); ok {
		entry := entryAny.(*serviceFactoryEntry)
		return entry.remoteFactory
	}
	return nil
}

func adaptServiceFactory(factory any) ServiceFactory {
	switch f := factory.(type) {
	case func(map[string]any) any:
		return f
	case func() any:
		return func(_ map[string]any) any {
			return f()
		}
	default:
		t := reflect.TypeOf(factory)
		if t.Kind() != reflect.Func {
			panic("service factory must be a function, got: " + t.String())
		}

		if t.NumOut() != 1 {
			panic("service factory must return exactly one value, got: " + t.String())
		}

		if t.NumIn() == 0 {
			fVal := reflect.ValueOf(factory)
			return func(_ map[string]any) any {
				return fVal.Call([]reflect.Value{})[0].Interface()
			}
		}

		if t.NumIn() != 1 {
			panic(fmt.Sprintf("service factory must have exactly 1 input parameter, got: %d", t.NumIn()))
		}

		paramType := t.In(0)
		paramIsPtr := paramType.Kind() == reflect.Ptr
		structType := paramType
		if paramIsPtr {
			structType = paramType.Elem()
		}

		if structType.Kind() != reflect.Struct {
			panic(fmt.Sprintf("service factory parameter must be a struct or pointer to struct, got: %s", paramType))
		}

		// validate all exported fields have json tag
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if field.PkgPath != "" { // unexported
				continue
			}
			if field.Tag.Get("json") == "" {
				panic(fmt.Sprintf("field %s in %s must have json tag", field.Name, structType.Name()))
			}
		}

		fVal := reflect.ValueOf(factory)
		return func(config map[string]any) any {
			// Convert config map to struct using JSON marshal/unmarshal
			b, err := json.Marshal(config)
			if err != nil {
				panic(fmt.Errorf("failed to marshal config: %w", err))
			}

			argPtr := reflect.New(structType).Interface()
			if err := json.Unmarshal(b, argPtr); err != nil {
				panic(fmt.Errorf("failed to unmarshal config: %w", err))
			}

			// prepare input parameter
			var in []reflect.Value
			if paramIsPtr {
				in = []reflect.Value{reflect.ValueOf(argPtr)}
			} else {
				in = []reflect.Value{reflect.ValueOf(argPtr).Elem()}
			}

			out := fVal.Call(in)
			return out[0].Interface()
		}
	}
}
