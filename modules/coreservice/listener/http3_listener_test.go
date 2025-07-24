package listener

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

// Helper function to create test certificate files for HTTP/3
func createTestCertFilesHttp3(t *testing.T) (certFile, keyFile string, cleanup func()) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test HTTP3"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Test"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    []string{"localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Create temporary files
	certFile = t.TempDir() + "/http3_cert.pem"
	keyFile = t.TempDir() + "/http3_key.pem"

	// Write certificate file
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("Failed to write certificate: %v", err)
	}

	// Write key file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	cleanup = func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}

	return certFile, keyFile, cleanup
}

func TestNewHttp3Listener(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	tests := []struct {
		name        string
		config      any
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_map_config",
			config: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
			},
			wantErr: false,
		},
		{
			name: "valid_map_config_with_idle_timeout",
			config: map[string]any{
				CERT_FILE_KEY:    certFile,
				KEY_FILE_KEY:     keyFile,
				IDLE_TIMEOUT_LEY: "60s",
			},
			wantErr: false,
		},
		{
			name: "valid_map_config_with_ca",
			config: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
				CA_FILE_KEY:   certFile, // Using cert as CA for test
			},
			wantErr: false,
		},
		{
			name:    "valid_array_config",
			config:  []any{certFile, keyFile},
			wantErr: false,
		},
		{
			name:    "valid_string_array_config",
			config:  []string{certFile, keyFile},
			wantErr: false,
		},
		{
			name: "missing_cert_file_in_map",
			config: map[string]any{
				KEY_FILE_KEY: keyFile,
			},
			wantErr:     true,
			errContains: "missing or invalid 'cert_file'",
		},
		{
			name: "missing_key_file_in_map",
			config: map[string]any{
				CERT_FILE_KEY: certFile,
			},
			wantErr:     true,
			errContains: "missing or invalid 'key_file'",
		},
		{
			name:        "insufficient_array_elements",
			config:      []any{certFile},
			wantErr:     true,
			errContains: "expected at least 2 elements",
		},
		{
			name:        "insufficient_string_array_elements",
			config:      []string{certFile},
			wantErr:     true,
			errContains: "expected at least 2 elements",
		},
		{
			name:        "invalid_cert_file_in_array",
			config:      []any{"", keyFile},
			wantErr:     true,
			errContains: "invalid or missing cert file",
		},
		{
			name:        "invalid_key_file_in_array",
			config:      []any{certFile, ""},
			wantErr:     true,
			errContains: "invalid or missing key file",
		},
		{
			name:        "invalid_config_type",
			config:      "invalid",
			wantErr:     true,
			errContains: "invalid configuration type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewHttp3Listener("test", tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewHttp3Listener() expected error but got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("NewHttp3Listener() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewHttp3Listener() error = %v, want nil", err)
				return
			}

			if service == nil {
				t.Error("NewHttp3Listener() returned nil service")
				return
			}

			listener, ok := service.(*Http3Listener)
			if !ok {
				t.Error("NewHttp3Listener() did not return *Http3Listener")
				return
			}

			// Verify implements required interfaces
			var _ serviceapi.HttpListener = listener

			// Verify cert and key files are set
			if listener.certFile == "" {
				t.Error("certFile should not be empty")
			}
			if listener.keyFile == "" {
				t.Error("keyFile should not be empty")
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

func TestHttp3Listener_IsRunning(t *testing.T) {
	listener := &Http3Listener{}

	// Initially should not be running
	if listener.IsRunning() {
		t.Error("Http3Listener.IsRunning() = true, want false for new listener")
	}

	// Test with running state
	listener.mu.Lock()
	listener.running = true
	listener.mu.Unlock()

	if !listener.IsRunning() {
		t.Error("Http3Listener.IsRunning() = false, want true when running")
	}

	// Test with stopped state
	listener.mu.Lock()
	listener.running = false
	listener.mu.Unlock()

	if listener.IsRunning() {
		t.Error("Http3Listener.IsRunning() = true, want false when stopped")
	}
}

func TestHttp3Listener_ActiveRequest(t *testing.T) {
	listener := &Http3Listener{}

	// Initially should have 0 active requests
	if got := listener.ActiveRequest(); got != 0 {
		t.Errorf("Http3Listener.ActiveRequest() = %v, want 0 for new listener", got)
	}

	// Test incrementing active count
	listener.activeCount.Store(5)
	if got := listener.ActiveRequest(); got != 5 {
		t.Errorf("Http3Listener.ActiveRequest() = %v, want 5", got)
	}

	// Test decrementing active count
	listener.activeCount.Store(3)
	if got := listener.ActiveRequest(); got != 3 {
		t.Errorf("Http3Listener.ActiveRequest() = %v, want 3", got)
	}
}

func TestHttp3Listener_ListenAndServe(t *testing.T) {
	// Skip HTTP/3 tests if we don't have the required dependencies
	if testing.Short() {
		t.Skip("Skipping HTTP/3 test in short mode")
	}

	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	service, err := NewHttp3Listener("test", map[string]any{
		CERT_FILE_KEY:    certFile,
		KEY_FILE_KEY:     keyFile,
		IDLE_TIMEOUT_LEY: "10s",
	})
	if err != nil {
		t.Fatalf("NewHttp3Listener() error = %v", err)
	}

	listener := service.(*Http3Listener)

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("http3 test response"))
	})

	// Use localhost with a high port number
	addr := "localhost:9443"

	// Test concurrent start and immediate shutdown
	var wg sync.WaitGroup
	var startErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		startErr = listener.ListenAndServe(addr, handler)
	}()

	// Wait a bit for server to start
	time.Sleep(200 * time.Millisecond)

	if !listener.IsRunning() {
		t.Error("Listener should be running after ListenAndServe")
	}

	// Shutdown the listener
	err = listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()

	// HTTP/3 might have different error behavior
	if startErr != nil {
		t.Logf("ListenAndServe() returned: %v", startErr)
	}
}

func TestHttp3Listener_GracefulShutdown(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	service, err := NewHttp3Listener("test", map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewHttp3Listener() error = %v", err)
	}

	listener := service.(*Http3Listener)

	var requestsStarted atomic.Int32
	var requestsCompleted atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsStarted.Add(1)
		// Simulate longer processing time
		time.Sleep(200 * time.Millisecond)
		requestsCompleted.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	addr := "localhost:9444"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Start some mock requests
	var reqWg sync.WaitGroup
	numRequests := 3

	for i := 0; i < numRequests; i++ {
		reqWg.Add(1)
		go func() {
			defer reqWg.Done()
			// Simulate request processing time
			time.Sleep(100 * time.Millisecond)
		}()
	}

	// Wait a bit for requests to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown while requests might be processing
	shutdownStart := time.Now()
	err = listener.Shutdown(5 * time.Second)
	shutdownDuration := time.Since(shutdownStart)

	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Should complete reasonably quickly since no real requests
	if shutdownDuration > 2*time.Second {
		t.Error("Shutdown took too long")
	}

	reqWg.Wait()
	wg.Wait()
}

func TestHttp3Listener_ShutdownWhenNotRunning(t *testing.T) {
	listener := &Http3Listener{}

	// Shutdown when not running should not error
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() when not running error = %v, want nil", err)
	}
}

func TestHttp3Listener_ShutdownTimeout(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	service, err := NewHttp3Listener("test", map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewHttp3Listener() error = %v", err)
	}

	listener := service.(*Http3Listener)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Very long processing time to test timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	addr := "localhost:9445"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Shutdown with short timeout
	shutdownStart := time.Now()
	err = listener.Shutdown(100 * time.Millisecond)
	shutdownDuration := time.Since(shutdownStart)

	// Should timeout or complete quickly since no real connections
	if shutdownDuration > 1*time.Second {
		t.Error("Shutdown took much longer than expected")
	}

	wg.Wait()
}

func TestHttp3Listener_ActiveRequestTracking(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	service, err := NewHttp3Listener("test", map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewHttp3Listener() error = %v", err)
	}

	listener := service.(*Http3Listener)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	addr := "localhost:9446"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Final active count should be 0
	if final := listener.ActiveRequest(); final != 0 {
		t.Errorf("Final active request count = %d, want 0", final)
	}

	listener.Shutdown(5 * time.Second)
	wg.Wait()
}

func TestHttp3Listener_ConfigurationVariations(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	tests := []struct {
		name     string
		config   map[string]any
		wantIdle time.Duration
	}{
		{
			name: "default_idle_timeout",
			config: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
			},
			wantIdle: DEFAULT_IDLE_TIMEOUT,
		},
		{
			name: "custom_idle_timeout",
			config: map[string]any{
				CERT_FILE_KEY:    certFile,
				KEY_FILE_KEY:     keyFile,
				IDLE_TIMEOUT_LEY: "3m",
			},
			wantIdle: 3 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewHttp3Listener("test", tt.config)
			if err != nil {
				t.Fatalf("NewHttp3Listener() error = %v", err)
			}

			listener := service.(*Http3Listener)

			if listener.idleTimeout != tt.wantIdle {
				t.Errorf("idleTimeout = %v, want %v", listener.idleTimeout, tt.wantIdle)
			}
		})
	}
}

func TestHttp3Listener_RequestHandlingDuringShutdown(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFilesHttp3(t)
	defer cleanup()

	service, err := NewHttp3Listener("test", map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewHttp3Listener() error = %v", err)
	}

	listener := service.(*Http3Listener)

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

	addr := "localhost:9447"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.ListenAndServe(addr, handler)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Test that the shutdown flag is properly handled
	// (We can't easily test actual HTTP/3 requests without complex setup)

	// Shutdown
	err = listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()
}
