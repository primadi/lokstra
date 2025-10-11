package lokstra_registry

import (
	"sync"

	"github.com/primadi/lokstra/core/request"
)

type MiddlewareFactory = func(config map[string]any) request.HandlerFunc

type middlewareEntry struct {
	mwType string
	config map[string]any
}

var mwFactoryRegistry sync.Map

var mwEntryRegistry sync.Map

// Registers a middleware factory function for a given middleware type.
// If allowOverride is false and a factory for the same type already exists, it panics.
func RegisterMiddlewareFactory(mwType string, factory MiddlewareFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	if !options.allowOverride {
		if _, exists := mwFactoryRegistry.Load(mwType); exists {
			panic("middleware factory for type " + mwType + " already registered")
		}
	}
	mwFactoryRegistry.Store(mwType, factory)
}

// Registers a middleware entry by name, associating it with a type and config.
// If allowOverride is false and an entry with the same name already exists, it panics.
func RegisterMiddlewareName(mwName string, mwType string, config map[string]any,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	if !options.allowOverride {
		if _, exists := mwEntryRegistry.Load(mwName); exists {
			panic("middleware name " + mwName + " already registered")
		}
	}
	mwEntryRegistry.Store(mwName, middlewareEntry{
		mwType: mwType,
		config: config,
	})
}

// Creates a middleware instance by name using the registered factory and config.
// Returns nil if the name or type is not found.
func CreateMiddleware(mwName string) request.HandlerFunc {
	entryAny, entryExists := mwEntryRegistry.Load(mwName)

	if entryExists {
		entry := entryAny.(middlewareEntry)
		factoryAny, factoryExists := mwFactoryRegistry.Load(entry.mwType)

		if factoryExists {
			factory := factoryAny.(MiddlewareFactory)
			return factory(entry.config)
		}
	}
	return nil
}
