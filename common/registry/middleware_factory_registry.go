package registry

import (
	"lokstra/common/iface"
	"lokstra/common/permission"
	"lokstra/core/request"
	"strings"
)

type MiddlewareFactory = func(config any) iface.MiddlewareFunc

var middlewareFactories = make(map[string]MiddlewareFactory) // map of middlewareType to MiddlewareFactory

// RegisterMiddlewareFactory registers a new middleware factory with the given middlewareType.
func RegisterMiddlewareFactory(middlewareType string, middlewareFactory func(config any) iface.MiddlewareFunc,
	lic ...*permission.PermissionLicense) {
	if permission.GlobalAccessLocked() {
		if lic == nil || lic[0] == nil || !strings.HasPrefix(middlewareType, lic[0].GetModuleName()+":") {
			panic("cannot register middleware after global access is locked or middleware is not created in the same module")
		}
	}

	if middlewareFactory == nil {
		panic("middleware factory cannot be nil")
	}
	if middlewareType == "" {
		panic("middlewareType cannot be empty")
	}

	if !strings.Contains(middlewareType, ".") {
		middlewareType = "main." + middlewareType
	}

	if _, exists := middlewareFactories[middlewareType]; exists {
		panic("middleware with middlewareType '" + middlewareType + "' already exists")
	}

	middlewareFactories[middlewareType] = middlewareFactory
}

// RegisterMiddlewareFunc registers a middleware function with the given middlewareType.
func RegisterMiddlewareFunc(middlewareType string,
	middlewareFunc func(next request.HandlerFunc) request.HandlerFunc) {
	RegisterMiddlewareFactory(middlewareType, func(_ any) iface.MiddlewareFunc {
		return middlewareFunc
	})
}

// ResetMiddlewareFactories clears all registered middleware factories.
// This is useful for testing or reinitializing the registry.
func ResetMiddlewareFactories() {
	if permission.GlobalAccessLocked() {
		panic("cannot reset middleware after global access is locked")
	}

	middlewareFactories = make(map[string]MiddlewareFactory)
}
