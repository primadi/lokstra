package flow

// customStep implements Step interface for custom user-defined functions
type customStep struct {
	fn   func(*Context) error
	name string
}

func (s *customStep) Run(ctx *Context) error {
	return s.fn(ctx)
}

func (s *customStep) Meta() StepMeta {
	name := s.name
	if name == "" {
		name = "custom.function"
	}
	return StepMeta{
		Name: name,
		Kind: StepNormal,
	}
}

var _ Step = (*customStep)(nil)
