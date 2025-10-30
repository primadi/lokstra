package deploy

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
)

// RegisterMiddlewareType registers a middleware factory
// Supports optional AllowOverride option
func (g *GlobalRegistry) RegisterMiddlewareType(middlewareType string, factory MiddlewareFactory, opts ...MiddlewareTypeOption) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var options middlewareTypeOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareFactories[middlewareType]; exists {
			panic(fmt.Sprintf("middleware type %s already registered", middlewareType))
		}
	}

	g.middlewareFactories[middlewareType] = factory
}

// RegisterMiddlewareName registers a middleware entry by name, associating it with a type and config.
// This allows creating multiple middleware instances from the same factory with different configurations.
//
// Example:
//
//	g.RegisterMiddlewareType("logger", loggerFactory)
//	g.RegisterMiddlewareName("logger-debug", "logger", map[string]any{"level": "debug"})
//	g.RegisterMiddlewareName("logger-info", "logger", map[string]any{"level": "info"})
func (g *GlobalRegistry) RegisterMiddlewareName(name, middlewareType string, config map[string]any, opts ...MiddlewareNameOption) {
	var options middlewareNameOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareEntries.Load(name); exists {
			panic(fmt.Sprintf("middleware name %s already registered", name))
		}
	}

	g.middlewareEntries.Store(name, &MiddlewareEntry{
		Type:   middlewareType,
		Config: config,
	})
}

// GetMiddlewareFactory returns the middleware factory
func (g *GlobalRegistry) GetMiddlewareFactory(middlewareType string) MiddlewareFactory {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.middlewareFactories[middlewareType]
}

// RegisterMiddleware registers a middleware instance by name (direct registration)
func (g *GlobalRegistry) RegisterMiddleware(name string, mw request.HandlerFunc) {
	if _, exists := g.middlewareInstances.Load(name); exists {
		panic(fmt.Sprintf("middleware %s already registered", name))
	}
	g.middlewareInstances.Store(name, mw)
}

// GetMiddleware retrieves a middleware instance by name
func (g *GlobalRegistry) GetMiddleware(name string) (request.HandlerFunc, bool) {
	if v, ok := g.middlewareInstances.Load(name); ok {
		return v.(request.HandlerFunc), true
	}
	return nil, false
}

// CreateMiddleware creates a middleware instance from definition
func (g *GlobalRegistry) CreateMiddleware(name string) request.HandlerFunc {
	// Step 1: First check if already instantiated
	if mw, ok := g.middlewareInstances.Load(name); ok {
		return mw.(request.HandlerFunc)
	}

	// Step 2: Check if it's registered via RegisterMiddlewareName (factory pattern)
	if entryAny, ok := g.middlewareEntries.Load(name); ok {
		entry := entryAny.(*MiddlewareEntry)
		factory := g.GetMiddlewareFactory(entry.Type)
		if factory != nil {
			mw := factory(entry.Config)
			if handlerFunc, ok := mw.(request.HandlerFunc); ok {
				// Cache it
				g.middlewareInstances.Store(name, handlerFunc)
				return handlerFunc
			}
		}
		return nil
	}

	// Step 3: If not found in entries, assume name is a factory type
	// Create directly from factory without config
	factory := g.GetMiddlewareFactory(name)
	if factory != nil {
		mw := factory(nil) // No config
		if handlerFunc, ok := mw.(request.HandlerFunc); ok {
			// Cache it
			g.middlewareInstances.Store(name, handlerFunc)
			return handlerFunc
		}
	}

	return nil
}
