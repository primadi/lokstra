// package meta contains the code-time setup structures that will be resolved
// into live runtime components when StartServer is called.
package meta

import (
	"fmt"
	"sync"
	"time"
)

// ServerMeta is a code-defined structure representing a server setup.
type ServerMeta struct {
	name string

	apps     []*AppMeta
	settings map[string]any
}

// NewServer creates a new empty server meta struct with initialized maps.
func NewServer(name string) *ServerMeta {
	return &ServerMeta{
		name:     name,
		apps:     []*AppMeta{},
		settings: map[string]any{},
	}
}

// AddApp adds a new app to the server.
func (s *ServerMeta) AddApp(app *AppMeta) {
	s.apps = append(s.apps, app)
}

// SetSetting sets a configuration setting for the server.
func (s *ServerMeta) SetSetting(key string, value any) {
	if s.settings == nil {
		s.settings = make(map[string]any)
	}
	s.settings[key] = value
}

// GetSetting retrieves a configuration setting by key.
func (s *ServerMeta) GetSetting(key string) (any, bool) {
	if s.settings == nil {
		return nil, false
	}
	value, exists := s.settings[key]
	return value, exists
}

func (s *ServerMeta) Start() error {
	// server := core.GetServer()

	// for _, appInfo := range s.apps {
	// 	app := core.NewApp(appInfo.GetName(), appInfo.GetPort()).
	// 	server.AddApp(app)
	// }

	var wg sync.WaitGroup
	errCh := make(chan error, len(s.apps))

	for _, app := range s.apps {
		wg.Add(1)
		go func(a *AppMeta) {
			defer wg.Done()
			if err := a.Start(); err != nil {
				errCh <- fmt.Errorf("app '%s' failed: %w", app.GetName(), err)
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

func (s *ServerMeta) GetName() string {
	return s.name
}

func (s *ServerMeta) GetApps() []*AppMeta {
	if s.apps == nil {
		return []*AppMeta{}
	}
	return s.apps
}

func (s *ServerMeta) Shutdown(shutdownTimeout time.Duration) {
	for _, app := range s.apps {
		if err := app.GetListener().Shutdown(shutdownTimeout); err != nil {
			fmt.Printf("Failed to shutdown app '%s': %v\n", app.GetName(), err)
		} else {
			fmt.Printf("App '%s' has been gracefully shutdown.\n", app.GetName())
		}
	}
	fmt.Println("Server has been gracefully shutdown.")
}
