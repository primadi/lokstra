package core

import "lokstra/common/response"

// Service is the minimal interface all services must implement.
type Service interface {
	InstanceName() string
	GetSetting(key string) any
}

// Hook interfaces

type OnServerStartHook interface {
	OnServerStart(s Server)
}

type OnAppStartHook interface {
	OnAppStart(a App)
}

type OnContextHook interface {
	OnContext(ctx *RequestContext)
}

type OnResponseHook interface {
	OnResponse(resp *response.Response, ctx *RequestContext)
}

type OnShutdownHook interface {
	OnShutdown()
}

type ReloadableService interface {
	Reload(config map[string]any) error
}
