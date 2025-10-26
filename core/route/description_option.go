package route

// Sets the description of the route, for identification and introspection purposes.
func WithDescriptionOption(description string) RouteHandlerOption {
	return &withDescriptionOption{description: description}
}

type withDescriptionOption struct {
	description string
}

// Apply implements RouteOption.
func (o *withDescriptionOption) Apply(rt *Route) {
	rt.Description = o.description
}

var _ RouteHandlerOption = (*withDescriptionOption)(nil)
