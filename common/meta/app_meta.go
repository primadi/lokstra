package meta

import (
	"fmt"
	"lokstra/serviceapi/core_service"
)

// AppMeta represents an application instance (port + router) to be served.
// It must be resolved into a live App object later.
type AppMeta struct {
	*RouterMeta
	name string
	port int

	listenerType     string
	routerEngineType string
	settings         map[string]any
}

func NewApp(name string, port int) *AppMeta {
	return &AppMeta{
		name:             name,
		port:             port,
		listenerType:     core_service.DEFAULT_LISTENER_NAME,
		routerEngineType: core_service.DEFAULT_ROUTER_ENGINE_NAME,
		RouterMeta:       NewRouter(),
		settings:         map[string]any{},
	}
}

func (a *AppMeta) GetListenerType() string {
	return a.listenerType
}

func (a *AppMeta) WithListenerType(listenerType string) *AppMeta {
	a.listenerType = listenerType
	return a
}

func (a *AppMeta) GetRouterEngineType() string {
	return a.routerEngineType
}

func (a *AppMeta) WithRouterEngineType(routerEngineType string) *AppMeta {
	a.routerEngineType = routerEngineType
	return a
}

func (a *AppMeta) Mount(router *RouterMeta) {
	a.RouterMeta = router
}

func (a *AppMeta) SetSetting(key string, value any) {
	if a.settings == nil {
		a.settings = map[string]any{}
	}
	a.settings[key] = value
}

func (a *AppMeta) GetSetting(key string) (any, bool) {
	if a.settings == nil {
		return nil, false
	}
	value, exists := a.settings[key]
	return value, exists
}

func (a *AppMeta) GetRouter() *RouterMeta {
	return a.RouterMeta
}

func (a *AppMeta) Addr() string {
	return fmt.Sprintf(":%d", a.port)
}

func (a *AppMeta) GetName() string {
	return a.name
}

func (a *AppMeta) GetPort() int {
	return a.port
}
