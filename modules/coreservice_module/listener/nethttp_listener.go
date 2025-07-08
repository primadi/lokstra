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

// NetHttpListener implements the HttpListener interface using the net/http package.
// It provides a standard HTTP server with graceful shutdown capabilities.
type NetHttpListener struct {
	server *http.Server

	mu             sync.RWMutex
	running        bool
	isShuttingDown atomic.Bool
	activeRequests sync.WaitGroup
	activeCount    atomic.Int32
}

var _ iface.Service = (*NetHttpListener)(nil)

func NewNetHttpListener(_ any) (iface.Service, error) {
	return &NetHttpListener{}, nil
}

// ListenerType implements listener_iface.HttpListener.
func (n *NetHttpListener) ListenerType() string {
	return core_service.NETHTTP_LISTENER_NAME
}

func (n *NetHttpListener) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

func (n *NetHttpListener) ActiveRequest() int {
	return int(n.activeCount.Load())
}

func (n *NetHttpListener) ListenAndServe(addr string, handler http.Handler) error {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.isShuttingDown.Load() {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			return
		}

		n.activeRequests.Add(1)
		n.activeCount.Add(1)
		defer func() {
			n.activeCount.Add(-1)
			n.activeRequests.Done()
		}()

		handler.ServeHTTP(w, r)
	})

	n.mu.Lock()
	n.server = &http.Server{
		Addr:    addr,
		Handler: wrappedHandler,
	}
	n.running = true
	n.mu.Unlock()

	fmt.Printf("[NETHTTP] Starting server at %s\n", addr)
	dumpRoutes(handler.(router.Router))

	err := n.server.ListenAndServe()

	n.mu.Lock()
	n.running = false
	n.mu.Unlock()

	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (n *NetHttpListener) Shutdown(shutdownTimeout time.Duration) error {
	n.isShuttingDown.Store(true)

	n.mu.RLock()
	server := n.server
	n.mu.RUnlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	fmt.Printf("[NETHTTP] Initiating graceful shutdown for server at %s\n", server.Addr)
	shutdownErr := server.Shutdown(ctx)

	done := make(chan struct{})
	go func() {
		n.activeRequests.Wait()
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

var _ core_service.HttpListener = (*NetHttpListener)(nil)
