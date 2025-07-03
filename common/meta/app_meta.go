package meta

import (
	"fmt"
	"lokstra/core/router/listener"
)

// AppMeta represents an application instance (port + router) to be served.
// It must be resolved into a live App object later.
type AppMeta struct {
	*RouterMeta
	name         string
	port         int
	listenerType listener.ListenerType // ListenerType indicates the type of listener to use (net/http or fasthttp).
	listener     listener.HttpListener // Listener is the actual HTTP listener instance.
	settings     map[string]any
}

func NewApp(name string, port int) *AppMeta {
	return &AppMeta{
		name:         name,
		port:         port,
		listenerType: listener.FastHttpListenerType, // Default to FastHttpListenerType
		RouterMeta:   NewRouterInfo(),
		settings:     map[string]any{},
	}
}

func (a *AppMeta) WithNetHttpListener() *AppMeta {
	a.listenerType = listener.NetHttpListenerType
	return a
}

func (a *AppMeta) WithFastHttpListener() *AppMeta {
	a.listenerType = listener.FastHttpListenerType
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
	if a.RouterMeta == nil {
		a.RouterMeta = NewRouterInfo()
	}
	return a.RouterMeta
}

func (a *AppMeta) Addr() string {
	return fmt.Sprintf(":%d", a.port)
}

func (a *AppMeta) Start() error {
	l := a.GetListener()
	router := a.RouterMeta.CreateNetHttpRouter()
	return l.ListenAndServe(a.Addr(), router)
}

func (a *AppMeta) GetListener() listener.HttpListener {
	if a.listener == nil {
		l, err := listener.NewHttpListener(a.listenerType)
		if err != nil {
			panic(fmt.Sprintf("Failed to create HTTP listener: %v", err))
		}
		a.listener = l
	}
	return a.listener
}

func (a *AppMeta) GetName() string {
	return a.name
}

func (a *AppMeta) GetPort() int {
	return a.port
}
