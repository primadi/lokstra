package listener

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/serviceapi"
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
	return serviceapi.NETHTTP_LISTENER_NAME
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
		Handler: wrappedHandler,
	}
	n.running = true
	n.mu.Unlock()

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
		fmt.Printf("[NETHTTP] Starting server on Unix socket %s\n", socketPath)
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on TCP address %s: %w", addr, err)
		}
		fmt.Printf("[NETHTTP] Starting server on TCP %s\n", addr)
	}

	if r, ok := handler.(router.Router); ok {
		dumpRoutes(r)
	}

	err = n.server.Serve(listener)

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

var _ serviceapi.HttpListener = (*NetHttpListener)(nil)
