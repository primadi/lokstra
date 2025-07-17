package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/serviceapi"
)

const CERT_FILE_KEY = "cert_file"
const KEY_FILE_KEY = "key_file"
const CA_FILE_KEY = "ca_file"

// SecureNetHttpListener implements the HttpListener interface with TLS support.
type SecureNetHttpListener struct {
	server   *http.Server
	certFile string
	keyFile  string
	caFile   string

	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration

	mu             sync.RWMutex
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

// ListenerType implements listener_iface.HttpListener.
func (s *SecureNetHttpListener) ListenerType() string {
	return serviceapi.SECURE_NETHTTP_LISTENER_NAME
}

// NewSecureNetHttpListener returns a new SecureNetHttpListener instance.
func NewSecureNetHttpListener(config any) (iface.Service, error) {
	var certFile, keyFile, caFile string
	var readTimeout, writeTimeout, idleTimeout time.Duration

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

		readTimeout = utils.GetDurationFromMap(cfg, READ_TIMEOUT_KEY, DEFAULT_READ_TIMEOUT)
		writeTimeout = utils.GetDurationFromMap(cfg, WRITE_TIMEOUT_KEY, DEFAULT_WRITE_TIMEOUT)
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

	return &SecureNetHttpListener{
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,

		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		idleTimeout:  idleTimeout,
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
		Handler:      wrappedHandler,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}
	s.running = true
	s.mu.Unlock()

	var listener net.Listener
	var err error
	if socketPath, ok := strings.CutPrefix(addr, "unix:"); ok {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on unix socket: %w", err)
		}
		fmt.Printf("[SECURE-HTTP] Starting TLS server on Unix socket %s\n", socketPath)
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on TCP %s: %w", addr, err)
		}
		fmt.Printf("[SECURE-HTTP] Starting TLS server at %s\n", addr)
	}

	// Wrap with TLS
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS cert/key: %w", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}

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

	tlsListener := tls.NewListener(listener, tlsConfig)

	if r, ok := handler.(router.Router); ok {
		dumpRoutes(r)
	}

	err = s.server.Serve(tlsListener)

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

var _ serviceapi.HttpListener = (*SecureNetHttpListener)(nil)
