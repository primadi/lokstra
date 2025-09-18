package server

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/primadi/lokstra/core/app"

	"github.com/primadi/lokstra/core/registration"
)

type Server struct {
	ctx      registration.Context
	name     string
	apps     []*app.App
	settings map[string]any
}

// NewServer creates a new Server instance with the given context and name.
func NewServer(ctx registration.Context, name string) *Server {
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

// AddApp registers an application to the server.
func (s *Server) AddApp(app *app.App) {
	s.apps = append(s.apps, app)
}

// NewApp creates a new application instance and registers it to the server.
func (s *Server) NewApp(name string, addr string) *app.App {
	newApp := app.NewApp(s.ctx, name, addr)
	s.AddApp(newApp)
	return newApp
}

func (s *Server) RegisterModule(module registration.Module) error {
	if err := module.Register(s.ctx); err != nil {
		return fmt.Errorf("failed to register module '%s': %w", module.Name(), err)
	}
	return nil
}

// SetSetting sets a configuration setting for the server.
func (s *Server) SetSetting(key string, value any) {
	s.settings[key] = value
}

// SetSettingsIfAbsent sets multiple configuration settings, only if they are not already set.
func (s *Server) SetSettingsIfAbsent(settings map[string]any) {
	for key, value := range settings {
		if _, exists := s.settings[key]; !exists {
			s.settings[key] = value
		}
	}
}

// GetSetting retrieves a configuration setting by key.
func (s *Server) GetSetting(key string) (any, bool) {
	value, exists := s.settings[key]
	return value, exists
}

func (s *Server) MergeAppsWithSameAddress() {
	for i, app := range s.apps {
		if app.IsMerged() {
			continue
		}
		for j := i + 1; j < len(s.apps); j++ {
			if s.apps[j].IsMerged() {
				continue
			}
			if app.GetAddr() == s.apps[j].GetAddr() {
				app.MergeOtherApp(s.apps[j])
			}
		}
	}
}

// Start starts all registered apps concurrently.
func (s *Server) Start() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.apps))

	s.MergeAppsWithSameAddress()

	// Start each app in its own goroutine
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

// Shutdown gracefully stops all registered apps.
func (s *Server) Shutdown(shutdownTimeout time.Duration) error {
	var errs []error
	for _, app := range s.apps {
		if err := app.Shutdown(shutdownTimeout); err != nil {
			fmt.Printf("Failed to shutdown app '%s': %v\n", app.GetName(), err)
			errs = append(errs, fmt.Errorf("app '%s': %w", app.GetName(), err))
		} else {
			fmt.Printf("App '%s' has been gracefully shutdown.\n", app.GetName())
		}
	}

	if err := s.ctx.ShutdownAllServices(); err != nil {
		fmt.Printf("Failed to shutdown all services: %v\n", err)
		errs = append(errs, fmt.Errorf("shutdown all services: %w", err))
	} else {
		fmt.Println("All services have been gracefully shutdown.")
	}

	fmt.Println("Server has been gracefully shutdown.")
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// StartAndWaitForShutdown starts the server and waits for shutdown signal.
func (s *Server) StartAndWaitForShutdown(shutdownTimeout time.Duration) error {
	// Run server in background
	errCh := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for signal or server error
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		fmt.Println("Received shutdown signal:", sig)
		if err := s.Shutdown(shutdownTimeout); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}

// ListApps returns the list of registered applications.
func (s *Server) ListApps() []*app.App {
	return s.apps
}
