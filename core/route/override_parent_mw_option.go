package route

// Sets whether to override parent middleware or not. Default is false (do not override).
func WithOverrideParentMwOption(override bool) RouteHandlerOption {
	return &withOverrideParentMwOption{override: override}
}

type withOverrideParentMwOption struct {
	override bool
}

// Apply implements RouteOption.
func (w *withOverrideParentMwOption) Apply(rt *Route) {
	rt.OverrideParentMw = w.override
}

var _ RouteHandlerOption = (*withOverrideParentMwOption)(nil)
