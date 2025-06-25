package iface

// Hook interfaces

type OnServerStartHook interface {
	OnServerStart(s Server)
}

type OnAppStartHook interface {
	OnAppStart(a App)
}

type OnShutdownHook interface {
	OnShutdown()
}
