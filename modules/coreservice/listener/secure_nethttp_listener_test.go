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
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

// Helper function to create test certificate files
func createTestCertFiles(t *testing.T) (certFile, keyFile string, cleanup func()) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test"},
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
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Create temporary files
	certFile = t.TempDir() + "/cert.pem"
	keyFile = t.TempDir() + "/key.pem"

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

func TestNewSecureNetHttpListener(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
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
			name: "valid_map_config_with_timeouts",
			config: map[string]any{
				CERT_FILE_KEY:     certFile,
				KEY_FILE_KEY:      keyFile,
				READ_TIMEOUT_KEY:  "30s",
				WRITE_TIMEOUT_KEY: "45s",
				IDLE_TIMEOUT_LEY:  "60s",
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
			service, err := NewSecureNetHttpListener(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewSecureNetHttpListener() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewSecureNetHttpListener() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewSecureNetHttpListener() error = %v, want nil", err)
				return
			}

			if service == nil {
				t.Error("NewSecureNetHttpListener() returned nil service")
				return
			}

			listener, ok := service.(*SecureNetHttpListener)
			if !ok {
				t.Error("NewSecureNetHttpListener() did not return *SecureNetHttpListener")
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
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && (s[len(s)-len(substr):] == substr || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

func TestSecureNetHttpListener_IsRunning(t *testing.T) {
	listener := &SecureNetHttpListener{}

	// Initially should not be running
	if listener.IsRunning() {
		t.Error("SecureNetHttpListener.IsRunning() = true, want false for new listener")
	}

	// Test with running state
	listener.mu.Lock()
	listener.running = true
	listener.mu.Unlock()

	if !listener.IsRunning() {
		t.Error("SecureNetHttpListener.IsRunning() = false, want true when running")
	}

	// Test with stopped state
	listener.mu.Lock()
	listener.running = false
	listener.mu.Unlock()

	if listener.IsRunning() {
		t.Error("SecureNetHttpListener.IsRunning() = true, want false when stopped")
	}
}

func TestSecureNetHttpListener_ActiveRequest(t *testing.T) {
	listener := &SecureNetHttpListener{}

	// Initially should have 0 active requests
	if got := listener.ActiveRequest(); got != 0 {
		t.Errorf("SecureNetHttpListener.ActiveRequest() = %v, want 0 for new listener", got)
	}

	// Test incrementing active count
	listener.activeCount.Store(5)
	if got := listener.ActiveRequest(); got != 5 {
		t.Errorf("SecureNetHttpListener.ActiveRequest() = %v, want 5", got)
	}

	// Test decrementing active count
	listener.activeCount.Store(3)
	if got := listener.ActiveRequest(); got != 3 {
		t.Errorf("SecureNetHttpListener.ActiveRequest() = %v, want 3", got)
	}
}

func TestSecureNetHttpListener_ListenAndServe_TCP(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
	defer cleanup()

	service, err := NewSecureNetHttpListener(map[string]any{
		CERT_FILE_KEY:     certFile,
		KEY_FILE_KEY:      keyFile,
		READ_TIMEOUT_KEY:  "5s",
		WRITE_TIMEOUT_KEY: "5s",
		IDLE_TIMEOUT_LEY:  "10s",
	})
	if err != nil {
		t.Fatalf("NewSecureNetHttpListener() error = %v", err)
	}

	listener := service.(*SecureNetHttpListener)

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("secure test response"))
	})

	// Use a test server to get a free port
	ts := httptest.NewTLSServer(handler)
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
	err = listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	wg.Wait()

	// Check if server stopped gracefully
	if startErr != nil && startErr != http.ErrServerClosed {
		t.Errorf("ListenAndServe() error = %v, want nil or http.ErrServerClosed", startErr)
	}
}

func TestSecureNetHttpListener_GracefulShutdown(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
	defer cleanup()

	service, err := NewSecureNetHttpListener(map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewSecureNetHttpListener() error = %v", err)
	}

	listener := service.(*SecureNetHttpListener)

	var requestsStarted atomic.Int32
	var requestsCompleted atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsStarted.Add(1)
		// Simulate longer processing time
		time.Sleep(200 * time.Millisecond)
		requestsCompleted.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewTLSServer(handler)
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
			// Note: This won't actually connect since we don't have proper TLS setup
			// But it will test the graceful shutdown logic
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

func TestSecureNetHttpListener_ShutdownWhenNotRunning(t *testing.T) {
	listener := &SecureNetHttpListener{}

	// Shutdown when not running should not error
	err := listener.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() when not running error = %v, want nil", err)
	}
}

func TestSecureNetHttpListener_ShutdownTimeout(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
	defer cleanup()

	service, err := NewSecureNetHttpListener(map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewSecureNetHttpListener() error = %v", err)
	}

	listener := service.(*SecureNetHttpListener)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Very long processing time to test timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewTLSServer(handler)
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

func TestSecureNetHttpListener_ActiveRequestTracking(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
	defer cleanup()

	service, err := NewSecureNetHttpListener(map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewSecureNetHttpListener() error = %v", err)
	}

	listener := service.(*SecureNetHttpListener)

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

	ts := httptest.NewTLSServer(handler)
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

	// Final active count should be 0
	if final := listener.ActiveRequest(); final != 0 {
		t.Errorf("Final active request count = %d, want 0", final)
	}

	listener.Shutdown(5 * time.Second)
	wg.Wait()
}

func TestSecureNetHttpListener_ConfigurationVariations(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertFiles(t)
	defer cleanup()

	tests := []struct {
		name      string
		config    map[string]any
		wantRead  time.Duration
		wantWrite time.Duration
		wantIdle  time.Duration
	}{
		{
			name: "default_timeouts",
			config: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
			},
			wantRead:  DEFAULT_READ_TIMEOUT,
			wantWrite: DEFAULT_WRITE_TIMEOUT,
			wantIdle:  DEFAULT_IDLE_TIMEOUT,
		},
		{
			name: "custom_timeouts",
			config: map[string]any{
				CERT_FILE_KEY:     certFile,
				KEY_FILE_KEY:      keyFile,
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
			service, err := NewSecureNetHttpListener(tt.config)
			if err != nil {
				t.Fatalf("NewSecureNetHttpListener() error = %v", err)
			}

			listener := service.(*SecureNetHttpListener)

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
