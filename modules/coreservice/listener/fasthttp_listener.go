package listener

import (
	"context"
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/utils"
	"lokstra/core/router"
	"lokstra/serviceapi"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// FastHttpListener implements the HttpListener interface using the fasthttp package.
// It provides a high-performance HTTP server with graceful shutdown capabilities.
type FastHttpListener struct {
	server *fasthttp.Server

	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	mu             sync.RWMutex
	addr           string
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

func (f *FastHttpListener) ListenerType() string {
	return serviceapi.FASTHTTP_LISTENER_NAME
}

func NewFastHttpListener(config any) (iface.Service, error) {
	var readTimeout, writeTimeout, idleTimeout time.Duration
	if cfg, ok := config.(map[string]any); ok {

		readTimeout = utils.GetDurationFromMap(cfg, READ_TIMEOUT_KEY, DEFAULT_READ_TIMEOUT)
		writeTimeout = utils.GetDurationFromMap(cfg, WRITE_TIMEOUT_KEY, DEFAULT_WRITE_TIMEOUT)
		idleTimeout = utils.GetDurationFromMap(cfg, IDLE_TIMEOUT_LEY, DEFAULT_IDLE_TIMEOUT)
	}
	return &FastHttpListener{
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		idleTimeout:  idleTimeout,
	}, nil
}

func (f *FastHttpListener) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.running
}

func (f *FastHttpListener) ActiveRequest() int {
	return int(f.activeCount.Load())
}

func (f *FastHttpListener) ListenAndServe(addr string, handler http.Handler) error {
	wrappedHandler := func(ctx *fasthttp.RequestCtx) {
		if f.isShuttingDown.Load() {
			ctx.Response.Header.Set("Retry-After", "5")
			ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
			ctx.SetBodyString("Server is shutting down")
			return
		}

		f.activeRequests.Add(1)
		f.activeCount.Add(1)
		defer func() {
			f.activeCount.Add(-1)
			f.activeRequests.Done()
		}()

		fasthttpadaptor.NewFastHTTPHandler(handler)(ctx)
	}

	f.mu.Lock()
	f.server = &fasthttp.Server{
		Handler:      wrappedHandler,
		ReadTimeout:  f.readTimeout,
		WriteTimeout: f.writeTimeout,
		IdleTimeout:  f.idleTimeout,
	}
	f.running = true
	f.mu.Unlock()

	f.addr = addr

	var listener net.Listener
	var err error

	if after, ok := strings.CutPrefix(addr, "unix:"); ok {
		socketPath := after

		// Remove existing socket file if exists
		if _, err := os.Stat(socketPath); err == nil {
			if err := os.Remove(socketPath); err != nil {
				return fmt.Errorf("failed to remove existing socket file: %w", err)
			}
		}

		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on unix socket: %w", err)
		}
		fmt.Printf("[FASTHTTP] Starting server on Unix socket %s\n", socketPath)
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on TCP address %s: %w", addr, err)
		}
		fmt.Printf("[FASTHTTP] Starting server on TCP %s\n", addr)
	}

	if r, ok := handler.(router.Router); ok {
		dumpRoutes(r)
	}

	err = f.server.Serve(listener)

	f.mu.Lock()
	f.running = false
	f.mu.Unlock()

	return err
}

func (f *FastHttpListener) Shutdown(shutdownTimeout time.Duration) error {
	f.isShuttingDown.Store(true)

	f.mu.RLock()
	server := f.server
	f.mu.RUnlock()

	if server == nil {
		return nil
	}

	fmt.Printf("[FASTHTTP] Initiating graceful shutdown for server at %s\n", f.addr)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := server.ShutdownWithContext(ctx)

	done := make(chan struct{})
	go func() {
		f.activeRequests.Wait()
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

var _ serviceapi.HttpListener = (*FastHttpListener)(nil)
