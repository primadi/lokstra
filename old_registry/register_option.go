package old_registry

type registerOptions struct {
	allowOverride bool
}

type RegisterOption interface {
	apply(opt *registerOptions)
}

type allowOverrideOption struct {
	allowOverride bool
}

func (o *allowOverrideOption) apply(opt *registerOptions) {
	opt.allowOverride = o.allowOverride
}

func AllowOverride(enable bool) RegisterOption {
	return &allowOverrideOption{allowOverride: enable}
}
