package server

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/iface"
)

type Server struct {
	ctx      iface.RegistrationContext
	name     string
	apps     []*app.App
	settings map[string]any
}

// NewServer creates a new Server instance with the given context and name.
func NewServer(ctx iface.RegistrationContext, name string) *Server {
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

func (s *Server) RegisterModule(module iface.Module) error {
	if err := module.Register(s.ctx); err != nil {
		return fmt.Errorf("failed to register module '%s': %w", module.Name(), err)
	}
	return nil
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

// Start starts all registered apps concurrently.
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

// Shutdown gracefully stops all registered apps.
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

// WaitForShutdown listens for OS signals and gracefully shuts down the server.
func (s *Server) WaitForShutdown(shutdownTimeout time.Duration) {
	// Listen for OS signals to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// waith for shutdown signal
	<-stop
	fmt.Println("Received shutdown signal...")

	// Call shutdown with the specified timeout
	s.Shutdown(shutdownTimeout)
}

// StartAndWaitForShutdown starts the server and waits for shutdown signal.
func (s *Server) StartAndWaitForShutdown(shutdownTimeout time.Duration) error {
	// Start async
	go func() {
		if err := s.Start(); err != nil {
			fmt.Printf("Server start error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for signal
	s.WaitForShutdown(shutdownTimeout)
	return nil
}
