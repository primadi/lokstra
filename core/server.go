package core

import (
	"fmt"
	"lokstra/iface"
	"net/http"
	"sync"
)

// Server is the main struct that holds multiple Apps and shared configuration.
type Server struct {
	name     string
	apps     map[string]*App
	settings map[string]any
}

var _ iface.Server = (*Server)(nil)

// NewServer creates a new server instance with a unique name.
func NewServer(name string) *Server {
	return &Server{
		name:     name,
		apps:     make(map[string]*App),
		settings: make(map[string]any),
	}
}

// Name returns the name of the server.
func (s *Server) Name() string {
	return s.name
}

// GetSetting retrieves a global configuration value by key.
func (s *Server) GetSetting(key string) any {
	return s.settings[key]
}

// SetSetting stores a key-value pair accessible by all apps and handlers.
func (s *Server) SetSetting(key string, value any) *Server {
	s.settings[key] = value
	return s
}

// NewApp creates and mounts a new App with a given name and port.
func (s *Server) NewApp(name string, port int) *App {
	app := NewApp(name, port)
	app.server = s
	s.apps[app.name] = app
	return app
}

// MountApp mounts an existing App to the server.
func (s *Server) MountApp(app *App) *Server {
	s.apps[app.name] = app
	app.server = s
	return s
}

// Start runs all mounted Apps concurrently.
// Each App will listen on its configured port.
// Returns error if any app fails to start.
func (s *Server) Start() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.apps))

	for _, app := range s.apps {
		wg.Add(1)
		go func(app *App) {
			defer wg.Done()
			if rtr, ok := app.Router.(*RouterImpl); ok {
				rtr.DumpRoutes() // Optional: Print registered routes
			}
			fmt.Printf("[INFO] Starting app '%s' on port %d\n", app.name, app.port)
			err := http.ListenAndServe(app.Addr(), app)
			if err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("app '%s' failed: %w", app.name, err)
			}
		}(app)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}
