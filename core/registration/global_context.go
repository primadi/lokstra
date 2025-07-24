package registration

var globalContext = &ContextImpl{
	permission: &PermissionGranted{
		whitelistGetService: []string{"*"},

		allowRegisterHandler:    true,
		allowRegisterMiddleware: true,
		allowRegisterService:    true,

		contextSettings: make(map[string]any),
	},
}

var globalContextCreated = false

func NewGlobalContext() *ContextImpl {
	if globalContextCreated {
		panic("GlobalContext has already been created")
	}
	globalContextCreated = true
	return globalContext
}
