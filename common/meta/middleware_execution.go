package meta

import "lokstra/common/iface"

type MiddlewareExecution struct {
	Name   string
	Config any // Configuration for the middleware

	MiddlewareFn   iface.MiddlewareFunc // Function to create the middleware
	Priority       int                  // Lower number means higher priority (1-100)
	ExecutionOrder int                  // Order of execution, lower number means earlier execution
}

func NamedMiddleware(name string, config ...any) *MiddlewareExecution {
	var cfg any
	switch len(config) {
	case 0:
		cfg = nil // No config provided, use nil
	case 1:
		cfg = config[0] // Single config provided
	default:
		cfg = config // Multiple configs provided, use as is
	}
	return &MiddlewareExecution{Name: name, Config: cfg, Priority: 5000}
}

func MiddlewareFn(fn iface.MiddlewareFunc) *MiddlewareExecution {
	return &MiddlewareExecution{MiddlewareFn: fn, Priority: 5000}
}
