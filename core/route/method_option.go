package route

// WithMethodOption sets the HTTP method for a route
func WithMethodOption(method string) RouteHandlerOption {
	return &withMethodOption{method: method}
}

type withMethodOption struct {
	method string
}

// Apply implements RouteHandlerOption.
func (w *withMethodOption) Apply(rt *Route) {
	rt.Method = w.method
}

var _ RouteHandlerOption = (*withMethodOption)(nil)
