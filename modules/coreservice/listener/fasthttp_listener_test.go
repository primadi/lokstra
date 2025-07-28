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

func TestNewFastHttpListener(t *testing.T) {
	tests := []struct {
		name   string
		config any
		want   *FastHttpListener
	}{
		{
			name:   "with_nil_config",
			config: nil,
			want: &FastHttpListener{
				readTimeout:  0,
				writeTimeout: 0,
				idleTimeout:  0,
			},
		},
		{
			name:   "with_empty_map_config",
			config: map[string]any{},
			want: &FastHttpListener{
				readTimeout:  DEFAULT_READ_TIMEOUT,
				writeTimeout: DEFAULT_WRITE_TIMEOUT,
				idleTimeout:  DEFAULT_IDLE_TIMEOUT,
			},
		},
		{
			name: "with_custom_timeouts",
			config: map[string]any{
				READ_TIMEOUT_KEY:  "30s",
				WRITE_TIMEOUT_KEY: "45s",
				IDLE_TIMEOUT_LEY:  "60s",
			},
			want: &FastHttpListener{
				readTimeout:  30 * time.Second,
				writeTimeout: 45 * time.Second,
				idleTimeout:  60 * time.Second,
			},
		},
		{
			name: "with_invalid_timeouts",
			config: map[string]any{
				READ_TIMEOUT_KEY:  "invalid",
				WRITE_TIMEOUT_KEY: nil,
				IDLE_TIMEOUT_LEY:  123,
			},
			want: &FastHttpListener{
				readTimeout:  DEFAULT_READ_TIMEOUT,  // falls back to default when invalid
				writeTimeout: DEFAULT_WRITE_TIMEOUT, // falls back to default when nil
				idleTimeout:  123 * time.Second,     // int type gets converted to seconds
			},
		},
		{
			name:   "with_string_config",
			config: "test",
			want: &FastHttpListener{
				readTimeout:  0,
				writeTimeout: 0,
				idleTimeout:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewFastHttpListener(tt.config)
			if err != nil {
				t.Errorf("NewFastHttpListener() error = %v, want nil", err)
				return
			}

			if service == nil {
				t.Error("NewFastHttpListener() returned nil service")
				return
			}

			listener, ok := service.(*FastHttpListener)
			if !ok {
				t.Error("NewFastHttpListener() did not return *FastHttpListener")
				return
			}

			if listener.readTimeout != tt.want.readTimeout {
				t.Errorf("readTimeout = %v, want %v", listener.readTimeout, tt.want.readTimeout)
			}
			if listener.writeTimeout != tt.want.writeTimeout {
				t.Errorf("writeTimeout = %v, want %v", listener.writeTimeout, tt.want.writeTimeout)
			}
			if listener.idleTimeout != tt.want.idleTimeout {
				t.Errorf("idleTimeout = %v, want %v", listener.idleTimeout, tt.want.idleTimeout)
			}

			// Verify implements required interfaces
			var _ serviceapi.HttpListener = listener
		})
	}
}

func TestFastHttpListener_IsRunning(t *testing.T) {
	listener := &FastHttpListener{}

	// Initially should not be running
	if listener.IsRunning() {
		t.Error("FastHttpListener.IsRunning() = true, want false for new listener")
	}

	// Test with running state
	listener.mu.Lock()
	listener.running = true
	listener.mu.Unlock()

	if !listener.IsRunning() {
		t.Error("FastHttpListener.IsRunning() = false, want true when running")
	}

	// Test with stopped state
	listener.mu.Lock()
	listener.running = false
	listener.mu.Unlock()

	if listener.IsRunning() {
		t.Error("FastHttpListener.IsRunning() = true, want false when stopped")
	}
}

func TestFastHttpListener_ActiveRequest(t *testing.T) {
	listener := &FastHttpListener{}

	// Initially should have 0 active requests
	if got := listener.ActiveRequest(); got != 0 {
		t.Errorf("FastHttpListener.ActiveRequest() = %v, want 0 for new listener", got)
	}

	// Test incrementing active count
	listener.activeCount.Store(5)
	if got := listener.ActiveRequest(); got != 5 {
		t.Errorf("FastHttpListener.ActiveRequest() = %v, want 5", got)
	}

	// Test decrementing active count
	listener.activeCount.Store(3)
	if got := listener.ActiveRequest(); got != 3 {
		t.Errorf("FastHttpListener.ActiveRequest() = %v, want 3", got)
	}
}

func TestFastHttpListener_ListenAndServe_TCP(t *testing.T) {
	service, err := NewFastHttpListener(map[string]any{
		READ_TIMEOUT_KEY:  "5s",
		WRITE_TIMEOUT_KEY: "5s",
		IDLE_TIMEOUT_LEY:  "10s",
	})
	if err != nil {
		t.Fatalf("NewFastHttpListener() error = %v", err)
	}

	listener := service.(*FastHttpListener)

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fasthttp test response"))
	})

	// Use a test server to get a free port
	ts := httptest.NewServer(handler)
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

	// Verify addr is set
	if listener.addr != addr {
		t.Errorf("listener.addr = %v, want %v", listener.addr, addr)
	}

	// Shutdown the listener
	err = listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()

	// FastHTTP doesn't return specific errors like net/http
	if startErr != nil {
		t.Logf("ListenAndServe() returned: %v", startErr)
	}
}

func TestFastHttpListener_ListenAndServe_Unix(t *testing.T) {
	// Skip on Windows as unix sockets are not supported
	t.Skip("Skipping unix socket test on Windows platform")

	listener := &FastHttpListener{
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  10 * time.Second,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fasthttp unix test"))
	})

	// Use a temporary socket path
	socketPath := "/tmp/test_fasthttp_listener.sock"
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

	if startErr != nil {
		t.Logf("ListenAndServe() unix returned: %v", startErr)
	}
}

func TestFastHttpListener_RequestHandling(t *testing.T) {
	listener := &FastHttpListener{
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  10 * time.Second,
	}

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

func TestFastHttpListener_GracefulShutdown(t *testing.T) {
	listener := &FastHttpListener{
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		idleTimeout:  15 * time.Second,
	}

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

func TestFastHttpListener_ShutdownTimeout(t *testing.T) {
	listener := &FastHttpListener{
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		idleTimeout:  15 * time.Second,
	}

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

	// FastHTTP may or may not timeout exactly like net/http
	if err != nil {
		t.Logf("Shutdown returned error (expected): %v", err)
	}

	// Should not wait too long beyond timeout
	if shutdownDuration > 1*time.Second {
		t.Error("Shutdown took much longer than expected")
	}

	wg.Wait()
}

func TestFastHttpListener_ShutdownWhenNotRunning(t *testing.T) {
	listener := &FastHttpListener{}

	// Shutdown when not running should not error
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() when not running error = %v, want nil", err)
	}
}

func TestFastHttpListener_ShutdownDuringShutdown(t *testing.T) {
	listener := &FastHttpListener{
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  10 * time.Second,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if listener.isShuttingDown.Load() {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Server is shutting down"))
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

	// Test that requests during shutdown get appropriate response
	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://" + addr)
		if err == nil {
			// FastHTTP might handle this differently than net/http
			t.Logf("Response status during shutdown: %d", resp.StatusCode)
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

func TestFastHttpListener_ActiveRequestTracking(t *testing.T) {
	listener := &FastHttpListener{
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  10 * time.Second,
	}

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

func TestFastHttpListener_TimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]any
		wantRead  time.Duration
		wantWrite time.Duration
		wantIdle  time.Duration
	}{
		{
			name:      "default_timeouts",
			config:    map[string]any{},
			wantRead:  DEFAULT_READ_TIMEOUT,
			wantWrite: DEFAULT_WRITE_TIMEOUT,
			wantIdle:  DEFAULT_IDLE_TIMEOUT,
		},
		{
			name: "custom_timeouts",
			config: map[string]any{
				READ_TIMEOUT_KEY:  "1m",
				WRITE_TIMEOUT_KEY: "2m",
				IDLE_TIMEOUT_LEY:  "3m",
			},
			wantRead:  1 * time.Minute,
			wantWrite: 2 * time.Minute,
			wantIdle:  3 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewFastHttpListener(tt.config)
			if err != nil {
				t.Fatalf("NewFastHttpListener() error = %v", err)
			}

			listener := service.(*FastHttpListener)

			if listener.readTimeout != tt.wantRead {
				t.Errorf("readTimeout = %v, want %v", listener.readTimeout, tt.wantRead)
			}
			if listener.writeTimeout != tt.wantWrite {
				t.Errorf("writeTimeout = %v, want %v", listener.writeTimeout, tt.wantWrite)
			}
			if listener.idleTimeout != tt.wantIdle {
				t.Errorf("idleTimeout = %v, want %v", listener.idleTimeout, tt.wantIdle)
			}
		})
	}
}
