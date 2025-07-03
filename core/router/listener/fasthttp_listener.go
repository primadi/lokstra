package listener

import (
	"context"
	"fmt"
	"net/http"
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

	mu             sync.RWMutex
	addr           string
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

func NewFastHttpListener() HttpListener {
	return &FastHttpListener{}
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
		Handler: wrappedHandler,
	}
	f.running = true
	f.mu.Unlock()

	f.addr = addr
	fmt.Printf("[FASTHTTP] Starting server at %s\n", addr)
	err := f.server.ListenAndServe(addr)

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

var _ HttpListener = (*FastHttpListener)(nil)
