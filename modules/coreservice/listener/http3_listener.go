package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/quic-go/quic-go/http3"
)

// HTTP3Listner implements the HttpListener interface with HTTP/3 support.
type Http3Listener struct {
	server   *http3.Server
	certFile string
	keyFile  string
	caFile   string

	idleTimeout time.Duration

	mu             sync.RWMutex
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

// NewHttp3Listener returns a new Http3Listener instance.
func NewHttp3Listener(config any) (service.Service, error) {
	var certFile, keyFile, caFile string
	var idleTimeout time.Duration

	if cfg, ok := config.(map[string]any); ok {
		certFile = utils.GetValueFromMap(cfg, CERT_FILE_KEY, "")
		if certFile == "" {
			return nil, fmt.Errorf("missing or invalid 'cert_file' in config")
		}
		keyFile = utils.GetValueFromMap(cfg, KEY_FILE_KEY, "")
		if keyFile == "" {
			return nil, fmt.Errorf("missing or invalid 'key_file' in config")
		}
		caFile = utils.GetValueFromMap(cfg, CA_FILE_KEY, "")

		idleTimeout = utils.GetDurationFromMap(cfg, IDLE_TIMEOUT_LEY, DEFAULT_IDLE_TIMEOUT)
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

	return &Http3Listener{
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,

		idleTimeout: idleTimeout,
	}, nil
}

func (s *Http3Listener) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Http3Listener) ActiveRequest() int {
	return int(s.activeCount.Load())
}

func (s *Http3Listener) ListenAndServe(addr string, handler http.Handler) error {
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

	s.running = true
	s.mu.Unlock()

	var err error

	// fmt.Printf("[HTTP3] Starting TLS server at %s\n", addr)

	// Wrap with TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		NextProtos: []string{"h3"},
	}

	if s.caFile != "" {
		caCert, err := os.ReadFile(s.caFile)
		if err != nil {
			return fmt.Errorf("failed to read CA file %s: %w", s.caFile, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to append CA cert from %s", s.caFile)
		}
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// if r, ok := handler.(router.Router); ok {
	// 	dumpRoutes(r)
	// }

	s.server = &http3.Server{
		Addr:        addr,
		Handler:     wrappedHandler,
		IdleTimeout: s.idleTimeout,
		TLSConfig:   tlsConfig,
	}

	err = s.server.ListenAndServeTLS(s.certFile, s.keyFile)

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Http3Listener) Shutdown(shutdownTimeout time.Duration) error {
	s.isShuttingDown.Store(true)

	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	fmt.Printf("[HTTP3] Initiating graceful shutdown for TLS server at %s\n", server.Addr)
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

func (s *Http3Listener) GetStartMessage(addr string) string {
	return fmt.Sprintf("[HTTP3] Starting TLS server at %s\n", addr)
}

var _ serviceapi.HttpListener = (*Http3Listener)(nil)
var _ service.Service = (*Http3Listener)(nil)
