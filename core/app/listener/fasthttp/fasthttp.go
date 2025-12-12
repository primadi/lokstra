package listener_fasthttp

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

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/app/listener"
	listener_utils "github.com/primadi/lokstra/core/app/listener/utils"
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
		logger.LogInfo("[FastHttp] Starting server on Unix socket %s\n", socketPath)
	} else {
		var err error
		listener, err = net.Listen("tcp", s.addr)
		if err != nil {
			return listener_utils.WrapListenError(s.addr, err)
		}
		logger.LogInfo("[FastHttp] Starting server on TCP %s\n", s.addr)
	}

	if s.secure {
		tlsConfig, err := listener_utils.CreateTLSConfig(s.certFile, s.keyFile, s.caFile)
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
		logger.LogInfo("[FastHttp] Initiating graceful shutdown for secure app at %s\n", s.addr)
	} else {
		logger.LogInfo("[FastHttp] Initiating graceful shutdown for app at %s\n", s.addr)
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

var _ listener.AppListener = (*FastHttp)(nil)

func NewFastHttp(config map[string]any, handler http.Handler) listener.AppListener {
	addr := utils.GetValueFromMap(config, "addr", ":8080")
	readTimeout := utils.GetValueFromMap(config, listener.READ_TIMEOUT_KEY, listener.DEFAULT_READ_TIMEOUT)
	writeTimeout := utils.GetValueFromMap(config, listener.WRITE_TIMEOUT_KEY, listener.DEFAULT_WRITE_TIMEOUT)
	idleTimeout := utils.GetValueFromMap(config, listener.IDLE_TIMEOUT_KEY, listener.DEFAULT_IDLE_TIMEOUT)

	secure := utils.GetValueFromMap(config, "secure", false)
	var certFile, keyFile, caFile string
	if secure {
		certFile = utils.GetValueFromMap(config, listener.CERT_FILE_KEY, "")
		keyFile = utils.GetValueFromMap(config, listener.KEY_FILE_KEY, "")
		caFile = utils.GetValueFromMap(config, listener.CA_FILE_KEY, "")
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

func init() {
	listener.RegisterListener("fasthttp", NewFastHttp)
}
