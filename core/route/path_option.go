package route

// WithPathOption sets the path for a route
func WithPathOption(path string) RouteHandlerOption {
	return &withPathOption{path: path}
}

type withPathOption struct {
	path string
}

// Apply implements RouteHandlerOption.
func (w *withPathOption) Apply(rt *Route) {
	rt.Path = w.path
}

var _ RouteHandlerOption = (*withPathOption)(nil)
