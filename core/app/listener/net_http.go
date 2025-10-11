package listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra/common/utils"
)

const READ_TIMEOUT_KEY = "read_timeout"
const READ_HEADER_TIMEOUT_KEY = "read_header_timeout"
const WRITE_TIMEOUT_KEY = "write_timeout"
const IDLE_TIMEOUT_KEY = "idle_timeout"
const CERT_FILE_KEY = "cert_file"
const KEY_FILE_KEY = "key_file"
const CA_FILE_KEY = "ca_file"

const DEFAULT_READ_TIMEOUT = 10 * time.Second
const DEFAULT_READ_HEADER_TIMEOUT = 2 * time.Second
const DEFAULT_WRITE_TIMEOUT = 5 * time.Minute
const DEFAULT_IDLE_TIMEOUT = 2 * time.Minute

type NetHttp struct {
	server  *http.Server
	handler http.Handler

	waitRequest sync.WaitGroup
	activeCount atomic.Int32

	secure   bool
	certFile string
	keyFile  string
	caFile   string
}

// ActiveRequests implements AppListener.
func (s *NetHttp) ActiveRequests() int {
	return int(s.activeCount.Load())
}

// ListenAndServe implements AppListener.
func (s *NetHttp) ListenAndServe() error {
	s.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.waitRequest.Add(1)
		s.activeCount.Add(1)
		defer func() {
			s.activeCount.Add(-1)
			s.waitRequest.Done()
		}()

		s.handler.ServeHTTP(w, r)
	})

	var listener net.Listener

	if after, ok := strings.CutPrefix(s.server.Addr, "unix:"); ok {
		socketPath := after

		// Remove existing socket file if exists
		if _, err := os.Stat(socketPath); err == nil {
			if err := os.Remove(socketPath); err != nil {
				return fmt.Errorf("failed to remove existing socket file: %w", err)
			}
		}

		var err error
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on unix socket: %w", err)
		}
		// fmt.Printf("[NETHTTP] Starting server on Unix socket %s\n", socketPath)
	} else {
		var err error
		listener, err = net.Listen("tcp", s.server.Addr)
		if err != nil {
			return wrapListenError(s.server.Addr, err)
		}
		// fmt.Printf("[NETHTTP] Starting server on TCP %s\n", addr)
	}

	if s.secure {
		tlsConfig, err := createTLSConfig(s.certFile, s.keyFile, s.caFile)
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		listener = tls.NewListener(listener, tlsConfig)
	}

	// Start serving
	if err := s.server.Serve(listener); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown implements AppListener.
func (s *NetHttp) Shutdown(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if s.secure {
		fmt.Printf("[NETHTTP] Initiating graceful shutdown for secure app at %s\n", s.server.Addr)
	} else {
		fmt.Printf("[NETHTTP] Initiating graceful shutdown for app at %s\n", s.server.Addr)
	}
	shutdownErr := s.server.Shutdown(ctx)

	done := make(chan struct{})
	go func() {
		s.waitRequest.Wait()
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

var _ AppListener = (*NetHttp)(nil)

func NewNetHttp(config map[string]any, handler http.Handler) AppListener {
	addr := utils.GetValueFromMap(config, "addr", ":8080")
	readTimeout := utils.GetValueFromMap(config, READ_TIMEOUT_KEY, DEFAULT_READ_TIMEOUT)
	readHeaderTimeout := utils.GetValueFromMap(config, READ_HEADER_TIMEOUT_KEY, DEFAULT_READ_HEADER_TIMEOUT)
	if readHeaderTimeout > 0 && readHeaderTimeout < readTimeout {
		readTimeout = readHeaderTimeout
	}
	writeTimeout := utils.GetValueFromMap(config, WRITE_TIMEOUT_KEY, DEFAULT_WRITE_TIMEOUT)
	idleTimeout := utils.GetValueFromMap(config, IDLE_TIMEOUT_KEY, DEFAULT_IDLE_TIMEOUT)

	secure := utils.GetValueFromMap(config, "secure", false)
	var certFile, keyFile, caFile string
	if secure {
		certFile = utils.GetValueFromMap(config, CERT_FILE_KEY, "")
		keyFile = utils.GetValueFromMap(config, KEY_FILE_KEY, "")
		caFile = utils.GetValueFromMap(config, CA_FILE_KEY, "")
	}

	return &NetHttp{
		handler:  handler,
		secure:   secure,
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,
		server: &http.Server{
			Addr:              addr,
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}
