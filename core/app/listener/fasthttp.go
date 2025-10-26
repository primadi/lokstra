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
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type FastHttp struct {
	server  *fasthttp.Server
	handler http.Handler
	addr    string

	waitRequest sync.WaitGroup
	activeCount atomic.Int32

	secure   bool
	certFile string
	keyFile  string
	caFile   string
}

// ActiveRequests implements AppListener.
func (s *FastHttp) ActiveRequests() int {
	return int(s.activeCount.Load())
}

// ListenAndServe implements AppListener.
func (s *FastHttp) ListenAndServe() error {
	wrappedHandler := func(ctx *fasthttp.RequestCtx) {
		s.waitRequest.Add(1)
		s.activeCount.Add(1)
		defer func() {
			s.activeCount.Add(-1)
			s.waitRequest.Done()
		}()

		fasthttpadaptor.NewFastHTTPHandler(s.handler)(ctx)
	}

	s.server.Handler = wrappedHandler

	var listener net.Listener

	if after, ok := strings.CutPrefix(s.addr, "unix:"); ok {
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
		// fmt.Printf("[FastHttp] Starting server on Unix socket %s\n", socketPath)
	} else {
		var err error
		listener, err = net.Listen("tcp", s.addr)
		if err != nil {
			return wrapListenError(s.addr, err)
		}
		// fmt.Printf("[FastHttp] Starting server on TCP %s\n", addr)
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
func (s *FastHttp) Shutdown(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if s.secure {
		fmt.Printf("[FastHttp] Initiating graceful shutdown for secure app at %s\n", s.addr)
	} else {
		fmt.Printf("[FastHttp] Initiating graceful shutdown for app at %s\n", s.addr)
	}
	shutdownErr := s.server.ShutdownWithContext(ctx)

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

var _ AppListener = (*FastHttp)(nil)

func NewFastHttp(config map[string]any, handler http.Handler) AppListener {
	addr := utils.GetValueFromMap(config, "addr", ":8080")
	readTimeout := utils.GetValueFromMap(config, READ_TIMEOUT_KEY, DEFAULT_READ_TIMEOUT)
	writeTimeout := utils.GetValueFromMap(config, WRITE_TIMEOUT_KEY, DEFAULT_WRITE_TIMEOUT)
	idleTimeout := utils.GetValueFromMap(config, IDLE_TIMEOUT_KEY, DEFAULT_IDLE_TIMEOUT)

	secure := utils.GetValueFromMap(config, "secure", false)
	var certFile, keyFile, caFile string
	if secure {
		certFile = utils.GetValueFromMap(config, CERT_FILE_KEY, "")
		keyFile = utils.GetValueFromMap(config, KEY_FILE_KEY, "")
		caFile = utils.GetValueFromMap(config, CA_FILE_KEY, "")
	}

	return &FastHttp{
		addr:     addr,
		handler:  handler,
		secure:   secure,
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,
		server: &fasthttp.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
}
