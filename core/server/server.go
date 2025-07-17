package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/module"
	"github.com/primadi/lokstra/core/app"
)

type Server struct {
	ctx      module.RegistrationContext
	name     string
	apps     []*app.App
	settings map[string]any
}

func NewServer(ctx module.RegistrationContext, name string) *Server {
	return &Server{
		ctx:      ctx,
		name:     name,
		apps:     make([]*app.App, 0),
		settings: make(map[string]any),
	}
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
