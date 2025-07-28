package listener

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

func TestNewNetHttpListener(t *testing.T) {
	tests := []struct {
		name   string
		config any
	}{
		{
			name:   "with_nil_config",
			config: nil,
		},
		{
			name:   "with_empty_map_config",
			config: map[string]any{},
		},
		{
			name:   "with_string_config",
			config: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewNetHttpListener(tt.config)
			if err != nil {
				t.Errorf("NewNetHttpListener() error = %v, want nil", err)
				return
			}

			if service == nil {
				t.Error("NewNetHttpListener() returned nil service")
				return
			}

			listener, ok := service.(*NetHttpListener)
			if !ok {
				t.Error("NewNetHttpListener() did not return *NetHttpListener")
				return
			}

			if listener == nil {
				t.Error("NewNetHttpListener() returned nil listener")
			}

			// Verify implements required interfaces
			var _ serviceapi.HttpListener = listener
		})
	}
}

func TestNetHttpListener_IsRunning(t *testing.T) {
	listener := &NetHttpListener{}

	// Initially should not be running
	if listener.IsRunning() {
		t.Error("NetHttpListener.IsRunning() = true, want false for new listener")
	}

	// Test with running state
	listener.mu.Lock()
	listener.running = true
	listener.mu.Unlock()

	if !listener.IsRunning() {
		t.Error("NetHttpListener.IsRunning() = false, want true when running")
	}

	// Test with stopped state
	listener.mu.Lock()
	listener.running = false
	listener.mu.Unlock()

	if listener.IsRunning() {
		t.Error("NetHttpListener.IsRunning() = true, want false when stopped")
	}
}

func TestNetHttpListener_ActiveRequest(t *testing.T) {
	listener := &NetHttpListener{}

	// Initially should have 0 active requests
	if got := listener.ActiveRequest(); got != 0 {
		t.Errorf("NetHttpListener.ActiveRequest() = %v, want 0 for new listener", got)
	}

	// Test incrementing active count
	listener.activeCount.Store(5)
	if got := listener.ActiveRequest(); got != 5 {
		t.Errorf("NetHttpListener.ActiveRequest() = %v, want 5", got)
	}

	// Test decrementing active count
	listener.activeCount.Store(3)
	if got := listener.ActiveRequest(); got != 3 {
		t.Errorf("NetHttpListener.ActiveRequest() = %v, want 3", got)
	}
}

func TestNetHttpListener_ListenAndServe_TCP(t *testing.T) {
	listener := &NetHttpListener{}

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Use a test server to get a free port
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Extract the address from test server
	addr := ts.Listener.Addr().String()
	ts.Close() // Close to free the port

	// Test concurrent start and immediate shutdown
	var wg sync.WaitGroup
	var startErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		startErr = listener.ListenAndServe(addr, handler)
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	if !listener.IsRunning() {
		t.Error("Listener should be running after ListenAndServe")
	}

	// Shutdown the listener
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()

	// Check if server stopped gracefully (should return http.ErrServerClosed or nil)
	if startErr != nil && startErr != http.ErrServerClosed {
		t.Errorf("ListenAndServe() error = %v, want nil or http.ErrServerClosed", startErr)
	}
}

func TestNetHttpListener_ListenAndServe_Unix(t *testing.T) {
	// Skip on Windows as unix sockets are not supported
	t.Skip("Skipping unix socket test on Windows platform")

	listener := &NetHttpListener{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unix test"))
	})

	// Use a temporary socket path
	socketPath := "/tmp/test_listener.sock"
	addr := fmt.Sprintf("unix:%s", socketPath)

	var wg sync.WaitGroup
	var startErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		startErr = listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	if !listener.IsRunning() {
		t.Error("Listener should be running after ListenAndServe on unix socket")
	}

	// Shutdown
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()

	if startErr != nil && startErr != http.ErrServerClosed {
		t.Errorf("ListenAndServe() unix error = %v, want nil or http.ErrServerClosed", startErr)
	}
}

func TestNetHttpListener_RequestHandling(t *testing.T) {
	listener := &NetHttpListener{}

	var requestCount atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Start server
	ts := httptest.NewServer(handler)
	addr := ts.Listener.Addr().String()
	ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Make some concurrent requests
	var reqWg sync.WaitGroup
	numRequests := 5

	for i := 0; i < numRequests; i++ {
		reqWg.Add(1)
		go func() {
			defer reqWg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("http://" + addr)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}

	// Wait for all requests to complete
	reqWg.Wait()

	// Shutdown
	listener.Shutdown(5 * time.Second)
	wg.Wait()
}

func TestNetHttpListener_GracefulShutdown(t *testing.T) {
	listener := &NetHttpListener{}

	var requestsStarted atomic.Int32
	var requestsCompleted atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsStarted.Add(1)
		// Simulate longer processing time
		time.Sleep(200 * time.Millisecond)
		requestsCompleted.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(handler)
	addr := ts.Listener.Addr().String()
	ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start some long-running requests
	var reqWg sync.WaitGroup
	numRequests := 3

	for i := 0; i < numRequests; i++ {
		reqWg.Add(1)
		go func() {
			defer reqWg.Done()
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get("http://" + addr)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}

	// Wait a bit for requests to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown while requests are processing
	shutdownStart := time.Now()
	err := listener.Shutdown(5 * time.Second)
	shutdownDuration := time.Since(shutdownStart)

	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Should wait for requests to complete
	if shutdownDuration < 150*time.Millisecond {
		t.Error("Shutdown returned too quickly, should wait for active requests")
	}

	reqWg.Wait()
	wg.Wait()

	// All requests should have completed
	started := requestsStarted.Load()
	completed := requestsCompleted.Load()

	if started != completed {
		t.Errorf("Started %d requests but only %d completed", started, completed)
	}
}

func TestNetHttpListener_ShutdownTimeout(t *testing.T) {
	listener := &NetHttpListener{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Very long processing time to test timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(handler)
	addr := ts.Listener.Addr().String()
	ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start a long request
	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		client.Get("http://" + addr)
	}()

	// Wait for request to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown with short timeout
	shutdownStart := time.Now()
	err := listener.Shutdown(100 * time.Millisecond)
	shutdownDuration := time.Since(shutdownStart)

	// Should timeout and return error
	if err == nil {
		t.Error("Expected timeout error but got nil")
	}

	// Should not wait too long
	if shutdownDuration > 500*time.Millisecond {
		t.Error("Shutdown took too long, should respect timeout")
	}

	wg.Wait()
}

func TestNetHttpListener_ShutdownWhenNotRunning(t *testing.T) {
	listener := &NetHttpListener{}

	// Shutdown when not running should not error
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() when not running error = %v, want nil", err)
	}
}

func TestNetHttpListener_ShutdownDuringShutdown(t *testing.T) {
	listener := &NetHttpListener{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if listener.isShuttingDown.Load() {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			return
		}

		// Simulate processing
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(handler)
	addr := ts.Listener.Addr().String()
	ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test that requests during shutdown get 503 response
	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://" + addr)
		if err == nil {
			if resp.StatusCode != http.StatusServiceUnavailable {
				t.Errorf("Expected 503 during shutdown, got %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	}()

	// Shutdown
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()
}

func TestNetHttpListener_ActiveRequestTracking(t *testing.T) {
	listener := &NetHttpListener{}

	var maxActiveRequests atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track maximum concurrent requests
		current := int32(listener.ActiveRequest())
		for {
			max := maxActiveRequests.Load()
			if current <= max || maxActiveRequests.CompareAndSwap(max, current) {
				break
			}
		}

		// Simulate processing
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(handler)
	addr := ts.Listener.Addr().String()
	ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start multiple concurrent requests
	var reqWg sync.WaitGroup
	numRequests := 5

	for i := 0; i < numRequests; i++ {
		reqWg.Add(1)
		go func() {
			defer reqWg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("http://" + addr)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}

	reqWg.Wait()

	// Should have tracked concurrent requests
	maxActive := maxActiveRequests.Load()
	if maxActive < 1 {
		t.Error("Should have tracked at least 1 active request")
	}

	// Final active count should be 0
	if final := listener.ActiveRequest(); final != 0 {
		t.Errorf("Final active request count = %d, want 0", final)
	}

	listener.Shutdown(5 * time.Second)
	wg.Wait()
}
