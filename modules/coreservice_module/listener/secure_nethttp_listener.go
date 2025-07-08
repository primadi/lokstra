package listener

import (
	"context"
	"fmt"
	"lokstra/common/iface"
	"lokstra/core/router"
	"lokstra/serviceapi/core_service"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const CERT_FILE = "cert_file"
const KEY_FILE = "key_file"

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

// ListenerType implements listener_iface.HttpListener.
func (s *SecureNetHttpListener) ListenerType() string {
	return core_service.SECURE_NETHTTP_LISTENER_NAME
}

// NewSecureNetHttpListener returns a new SecureNetHttpListener instance.
func NewSecureNetHttpListener(config any) (iface.Service, error) {
	var certFile, keyFile string
	if cfg, ok := config.(map[string]any); ok {
		var ok bool
		certFile, ok = cfg[CERT_FILE].(string)
		if !ok || certFile == "" {
			return nil, fmt.Errorf("missing or invalid 'cert_file' in config")
		}
		keyFile, ok = cfg[KEY_FILE].(string)
		if !ok || keyFile == "" {
			return nil, fmt.Errorf("missing or invalid 'key_file' in config")
		}
	} else if arr, ok := config.([]any); ok {
		if len(arr) != 2 {
			return nil, fmt.Errorf("expected at least 2 elements in config array for cert and key files")
		}
		if certFile, ok = arr[0].(string); !ok || certFile == "" {
			return nil, fmt.Errorf("invalid or missing cert file in config array")
		}
		if keyFile, ok = arr[1].(string); !ok || keyFile == "" {
			return nil, fmt.Errorf("invalid or missing key file in config array")
		}
	} else if arrStr, ok := config.([]string); ok {
		if len(arrStr) != 2 {
			return nil, fmt.Errorf("expected at least 2 elements in config array for cert and key files")
		}
		if certFile = arrStr[0]; certFile == "" {
			return nil, fmt.Errorf("invalid or missing cert file in config array")
		}
		if keyFile = arrStr[1]; keyFile == "" {
			return nil, fmt.Errorf("invalid or missing key file in config array")
		}
	} else {
		return nil, fmt.Errorf("invalid configuration type: expected map[string]any")
	}

	return &SecureNetHttpListener{
		certFile: certFile,
		keyFile:  keyFile,
	}, nil
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
	dumpRoutes(handler.(router.Router))

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

var _ core_service.HttpListener = (*SecureNetHttpListener)(nil)
