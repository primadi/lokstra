package deploy

import (
	"fmt"

	"github.com/primadi/lokstra/core/deploy/internal"
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
// Supports inline parameters syntax: "middleware-name param1="value1", param2="value2"
//
// Examples:
//   - "recovery" - Load middleware without params
//   - "cors" - Load from RegisterMiddlewareName if exists, or factory with nil config
//   - "rate-limit max=100, window="1m"" - Load factory with inline params
func (g *GlobalRegistry) CreateMiddleware(name string) request.HandlerFunc {
	// Parse name and extract inline parameters
	middlewareName, inlineParams := internal.ParseMiddlewareName(name)

	// Step 1: First check if already instantiated
	cacheKey := name // Use full name as cache key to support different params
	if mw, ok := g.middlewareInstances.Load(cacheKey); ok {
		return mw.(request.HandlerFunc)
	}

	// Step 2: Check if it's registered via RegisterMiddlewareName (factory pattern)
	if entryAny, ok := g.middlewareEntries.Load(middlewareName); ok {
		entry := entryAny.(*MiddlewareEntry)
		factory := g.GetMiddlewareFactory(entry.Type)
		if factory != nil {
			// Merge inline params with registered config (inline takes precedence)
			config := internal.MergeConfig(entry.Config, inlineParams)
			mw := factory(config)

			// Try type assertion first (named type)
			if handlerFunc, ok := mw.(request.HandlerFunc); ok {
				g.middlewareInstances.Store(cacheKey, handlerFunc)
				return handlerFunc
			}

			// Fallback: Try converting from unnamed func signature
			if fnErr, ok := mw.(func(*request.Context) error); ok {
				handlerFunc := request.HandlerFunc(fnErr)
				g.middlewareInstances.Store(cacheKey, handlerFunc)
				return handlerFunc
			}
		}
		return nil
	}

	// Step 3: If not found in entries, assume middlewareName is a factory type
	// Create directly from factory with inline params (or nil)
	factory := g.GetMiddlewareFactory(middlewareName)
	if factory != nil {
		var config map[string]any
		if len(inlineParams) > 0 {
			config = inlineParams
		}
		mw := factory(config)

		// Try type assertion first (named type)
		if handlerFunc, ok := mw.(request.HandlerFunc); ok {
			g.middlewareInstances.Store(cacheKey, handlerFunc)
			return handlerFunc
		}

		// Fallback: Try converting from unnamed func signature
		if fnErr, ok := mw.(func(*request.Context) error); ok {
			handlerFunc := request.HandlerFunc(fnErr)
			g.middlewareInstances.Store(cacheKey, handlerFunc)
			return handlerFunc
		}
	}

	return nil
}
