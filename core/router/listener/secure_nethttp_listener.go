package listener

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// SecureNetHttpListener implements the HttpListener interface with TLS support.
type SecureNetHttpListener struct {
	server   *http.Server
	certFile string
	keyFile  string

	mu             sync.RWMutex
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

// NewSecureNetHttpListener returns a new SecureNetHttpListener instance.
func NewSecureNetHttpListener(certFile, keyFile string) HttpListener {
	return &SecureNetHttpListener{
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (s *SecureNetHttpListener) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *SecureNetHttpListener) ActiveRequest() int {
	return int(s.activeCount.Load())
}

func (s *SecureNetHttpListener) ListenAndServe(addr string, handler http.Handler) error {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.isShuttingDown.Load() {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			return
		}

		s.activeRequests.Add(1)
		s.activeCount.Add(1)
		defer func() {
			s.activeCount.Add(-1)
			s.activeRequests.Done()
		}()

		handler.ServeHTTP(w, r)
	})

	s.mu.Lock()
	s.server = &http.Server{
		Addr:    addr,
		Handler: wrappedHandler,
	}
	s.running = true
	s.mu.Unlock()

	fmt.Printf("[SECURE-HTTP] Starting TLS server at %s\n", addr)
	err := s.server.ListenAndServeTLS(s.certFile, s.keyFile)

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *SecureNetHttpListener) Shutdown(shutdownTimeout time.Duration) error {
	s.isShuttingDown.Store(true)

	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	fmt.Printf("[SECURE-HTTP] Initiating graceful shutdown for TLS server at %s\n", server.Addr)
	shutdownErr := server.Shutdown(ctx)

	done := make(chan struct{})
	go func() {
		s.activeRequests.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		if shutdownErr == nil {
			return ctx.Err()
		}
	}

	return shutdownErr
}

var _ HttpListener = (*SecureNetHttpListener)(nil)
