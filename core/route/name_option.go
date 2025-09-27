package route

// Sets the name of the route, for identification and introspection purposes.
// if not set, default name will be METHOD[path]_handler
func WithNameOption(name string) RouteHandlerOption {
	return &withNameOption{name: name}
}

type withNameOption struct {
	name string
}

// Apply implements RouteOption.
func (o *withNameOption) Apply(rt *Route) {
	rt.Name = o.name
}

var _ RouteHandlerOption = (*withNameOption)(nil)
