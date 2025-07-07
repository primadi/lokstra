package server

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/meta"
	"lokstra/core/app"
	"sync"
	"time"
)

type Server struct {
	ctx      component.ComponentContext
	name     string
	apps     []*app.App
	settings map[string]any
}

func NewServer(ctx component.ComponentContext, name string) *Server {
	return &Server{
		ctx:      ctx,
		name:     name,
		apps:     make([]*app.App, 0),
		settings: make(map[string]any),
	}
}

func NewServerFromMeta(ctx component.ComponentContext, meta *meta.ServerMeta) *Server {
	svr := NewServer(ctx, meta.GetName())
	for _, appMeta := range meta.GetApps() {
		appInstance := app.NewAppFromMeta(ctx, appMeta)
		if appInstance == nil {
			panic(fmt.Sprintf("Failed to create app from meta: %s", appMeta.GetName()))
		}
		svr.AddApp(appInstance)
	}

	settings := meta.GetSettings()
	for key, value := range settings {
		svr.SetSetting(key, value)
	}

	return svr
}

func (s *Server) GetName() string {
	return s.name
}

func (s *Server) AddApp(app *app.App) {
	s.apps = append(s.apps, app)
}

// SetSetting sets a configuration setting for the server.
func (s *Server) SetSetting(key string, value any) {
	s.settings[key] = value
}

// GetSetting retrieves a configuration setting by key.
func (s *Server) GetSetting(key string) (any, bool) {
	value, exists := s.settings[key]
	return value, exists
}

func (s *Server) Start() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.apps))

	for _, ap := range s.apps {
		wg.Add(1)
		go func(a *app.App) {
			defer wg.Done()
			if err := a.Start(); err != nil {
				errCh <- fmt.Errorf("app '%s' failed: %w", ap.GetName(), err)
			}
		}(ap)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

func (s *Server) Shutdown(shutdownTimeout time.Duration) {
	for _, app := range s.apps {
		if err := app.Shutdown(shutdownTimeout); err != nil {
			fmt.Printf("Failed to shutdown app '%s': %v\n", app.GetName(), err)
		} else {
			fmt.Printf("App '%s' has been gracefully shutdown.\n", app.GetName())
		}
	}
	fmt.Println("Server has been gracefully shutdown.")
}
