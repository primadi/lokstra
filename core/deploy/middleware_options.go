package deploy

// MiddlewareTypeOption configures middleware type registration
type MiddlewareTypeOption func(*middlewareTypeOptions)

type middlewareTypeOptions struct {
	allowOverride bool
}

// WithAllowOverride allows overriding existing middleware type registration
func WithAllowOverride(allow bool) MiddlewareTypeOption {
	return func(opts *middlewareTypeOptions) {
		opts.allowOverride = allow
	}
}

// MiddlewareNameOption configures middleware name registration
type MiddlewareNameOption func(*middlewareNameOptions)

type middlewareNameOptions struct {
	allowOverride bool
}

// WithAllowOverrideForName allows overriding existing middleware name registration
func WithAllowOverrideForName(allow bool) MiddlewareNameOption {
	return func(opts *middlewareNameOptions) {
		opts.allowOverride = allow
	}
}
