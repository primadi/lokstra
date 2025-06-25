package core

import (
	"fmt"
	"lokstra/iface"
	"net/http"
)

type App struct {
	Router
	name   string
	port   int
	config map[string]any
	server *Server
}

// GetConfig implements iface.App.
func (a *App) GetConfig(key string) any {
	return a.config[key]
}

// GetServer implements iface.App.
func (a *App) GetServer() iface.Server {
	return a.server
}

// Name implements iface.App.
func (a *App) Name() string {
	return a.name
}

// Ensure App implements iface.App
var _ iface.App = (*App)(nil)

func NewApp(name string, port int) *App {
	return &App{
		Router: NewRouter(),
		name:   name,
		port:   port,
		config: make(map[string]any),
	}
}

func (a *App) UseRouter(r Router) {
	a.Router = r
}

func (a *App) Addr() string {
	return fmt.Sprintf(":%d", a.port)
}

func (a *App) SetConfig(key string, value any) {
	a.config[key] = value
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := NewRequestContext(w, r)
	defer cancel()

	ctx.App = a

	a.Router.ServeHTTP(w, ctx.Request)
}

func (a *App) Start() {
	http.ListenAndServe(a.Addr(), a.Router)
}
