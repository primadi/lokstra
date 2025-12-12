package server

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/app"
)

// Callback to shutdown services - set by registry to avoid circular dependency
var shutdownServicesCallback func()

// SetShutdownServicesCallback allows registry to set the callback function
func SetShutdownServicesCallback(callback func()) {
	shutdownServicesCallback = callback
}

type Server struct {
	Name         string
	BaseUrl      string // Base URL of the server
	DeploymentID string // Deployment ID for grouping servers
	Apps         []*app.App

	built bool
}

// GetName returns the server name (implements ServerInterface)
func (s *Server) GetName() string {
	return s.Name
}

// Create a new Server instance with given apps
func New(name string, apps ...*app.App) *Server {
	return &Server{
		Name: name,
		Apps: apps,
	}
}

// Print server start information, including each app's details
func (s *Server) PrintStartInfo() {
	s.build()
	logger.LogInfo("Server '%s' starting with %d app(s):\n", s.Name, len(s.Apps))
	for _, a := range s.Apps {
		a.PrintStartInfo()
	}
	logger.LogInfo("Press CTRL+C to stop the server...")
}

func (s *Server) AddApp(a *app.App) {
	if s.built {
		logger.LogPanic("Cannot add app after server is built")
	}
	s.Apps = append(s.Apps, a)
}

func (s *Server) build() {
	if s.built {
		return
	}
	s.built = true
	addrMap := make(map[string]*app.App)
	var mergedApps []*app.App

	for _, a := range s.Apps {
		addr := a.GetAddress()
		if existing, ok := addrMap[addr]; ok {
			// Merge the existing app with the new one
			existing.AddRouter(a.GetRouter())
		} else {
			addrMap[addr] = a
			mergedApps = append(mergedApps, a)
		}
	}

	s.Apps = mergedApps
}

// Start the server. It blocks until the server stops or returns an error.
// Shutdown must be called separately.
func (s *Server) Start() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.Apps))

	s.build()

	// Start each app in its own goroutine
	for _, ap := range s.Apps {
		wg.Add(1)
		go func(a *app.App) {
			defer wg.Done()
			if err := a.Start(); err != nil {
				errCh <- fmt.Errorf("app '%s' failed: %w", a.GetName(), err)
			}
		}(ap)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Shutdown gracefully all apps within the given timeout.
func (s *Server) Shutdown(timeout any) error {
	// Convert timeout to time.Duration
	var duration time.Duration
	switch t := timeout.(type) {
	case time.Duration:
		duration = t
	case int:
		duration = time.Duration(t) * time.Second
	case string:
		var err error
		duration, err = time.ParseDuration(t)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %v", err)
		}
	default:
		duration = 5 * time.Second // default
	}

	return s.shutdown(duration)
}

// Internal shutdown method with time.Duration
func (s *Server) shutdown(timeout time.Duration) error {
	var wg sync.WaitGroup

	errCh := make(chan error, len(s.Apps))
	for _, ap := range s.Apps {
		wg.Add(1)
		go func(a *app.App) {
			defer wg.Done()
			if err := a.Shutdown(timeout); err != nil {
				logger.LogError("Failed to shutdown app '%s': %v\n", a.GetName(), err)
				errCh <- fmt.Errorf("app '%s': %w", a.GetName(), err)
			} else {
				logger.LogInfo("App '%s' has been gracefully shutdown.\n", a.GetName())
			}
		}(ap)
	}

	wg.Wait()
	close(errCh)

	// Shutdown any remaining services via callback to avoid circular dependency
	if shutdownServicesCallback != nil {
		shutdownServicesCallback()
	}

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Starts the server and blocks until a termination signal is received.
// It shuts down gracefully with the given timeout.
func (s *Server) Run(timeout time.Duration) error {
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
		logger.LogInfo("Received shutdown signal:", sig)
		if err := s.shutdown(timeout); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

type ServerInterface interface {
	GetName() string
	Start() error
	Shutdown(timeout any) error
}

var _ ServerInterface = (*Server)(nil)
