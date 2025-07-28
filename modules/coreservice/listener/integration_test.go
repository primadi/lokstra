package listener

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

// TestListener_InterfaceCompliance tests that all listeners implement required interfaces
func TestListener_InterfaceCompliance(t *testing.T) {
	// Test NetHttpListener
	netListener, err := NewNetHttpListener(nil)
	if err != nil {
		t.Fatalf("Failed to create NetHttpListener: %v", err)
	}
	var _ serviceapi.HttpListener = netListener.(*NetHttpListener)

	// Test FastHttpListener
	fastListener, err := NewFastHttpListener(nil)
	if err != nil {
		t.Fatalf("Failed to create FastHttpListener: %v", err)
	}
	var _ serviceapi.HttpListener = fastListener.(*FastHttpListener)

	// Test SecureNetHttpListener with test certs
	certFile, keyFile, cleanup := createTestCertsIntegration(t)
	defer cleanup()

	secureListener, err := NewSecureNetHttpListener(map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create SecureNetHttpListener: %v", err)
	}
	var _ serviceapi.HttpListener = secureListener.(*SecureNetHttpListener)

	// Test Http3Listener with test certs
	http3Listener, err := NewHttp3Listener(map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create Http3Listener: %v", err)
	}
	var _ serviceapi.HttpListener = http3Listener.(*Http3Listener)
}

// TestListener_BasicFunctionality tests basic operations for all listeners
func TestListener_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name           string
		createListener func() serviceapi.HttpListener
	}{
		{
			name: "NetHttpListener",
			createListener: func() serviceapi.HttpListener {
				listener, _ := NewNetHttpListener(nil)
				return listener.(*NetHttpListener)
			},
		},
		{
			name: "FastHttpListener",
			createListener: func() serviceapi.HttpListener {
				listener, _ := NewFastHttpListener(nil)
				return listener.(*FastHttpListener)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := tt.createListener()

			// Test initial state
			if listener.IsRunning() {
				t.Error("Listener should not be running initially")
			}

			if active := listener.ActiveRequest(); active != 0 {
				t.Errorf("Initial active requests should be 0, got %d", active)
			}

			// Test shutdown when not running
			if err := listener.Shutdown(time.Second); err != nil {
				t.Errorf("Shutdown when not running should not error: %v", err)
			}
		})
	}
}

// TestListener_TLSListeners tests TLS-enabled listeners separately due to certificate requirements
func TestListener_TLSListeners(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertsIntegration(t)
	defer cleanup()

	tests := []struct {
		name           string
		createListener func() serviceapi.HttpListener
	}{
		{
			name: "SecureNetHttpListener",
			createListener: func() serviceapi.HttpListener {
				listener, err := NewSecureNetHttpListener(map[string]any{
					CERT_FILE_KEY: certFile,
					KEY_FILE_KEY:  keyFile,
				})
				if err != nil {
					t.Fatalf("Failed to create SecureNetHttpListener: %v", err)
				}
				return listener.(*SecureNetHttpListener)
			},
		},
		{
			name: "Http3Listener",
			createListener: func() serviceapi.HttpListener {
				listener, err := NewHttp3Listener(map[string]any{
					CERT_FILE_KEY: certFile,
					KEY_FILE_KEY:  keyFile,
				})
				if err != nil {
					t.Fatalf("Failed to create Http3Listener: %v", err)
				}
				return listener.(*Http3Listener)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := tt.createListener()

			// Test initial state
			if listener.IsRunning() {
				t.Error("TLS Listener should not be running initially")
			}

			if active := listener.ActiveRequest(); active != 0 {
				t.Errorf("Initial active requests should be 0, got %d", active)
			}

			// Test shutdown when not running
			if err := listener.Shutdown(time.Second); err != nil {
				t.Errorf("Shutdown when not running should not error: %v", err)
			}
		})
	}
}

// TestListener_ConfigurationParsing tests configuration parsing across different listeners
func TestListener_ConfigurationParsing(t *testing.T) {
	certFile, keyFile, cleanup := createTestCertsIntegration(t)
	defer cleanup()

	tests := []struct {
		name          string
		factory       func(any) (any, error)
		validConfig   any
		invalidConfig any
	}{
		{
			name:          "NetHttpListener",
			factory:       func(c any) (any, error) { return NewNetHttpListener(c) },
			validConfig:   map[string]any{},
			invalidConfig: nil, // NetHttpListener accepts any config
		},
		{
			name:    "FastHttpListener",
			factory: func(c any) (any, error) { return NewFastHttpListener(c) },
			validConfig: map[string]any{
				READ_TIMEOUT_KEY:  "5s",
				WRITE_TIMEOUT_KEY: "5s",
				IDLE_TIMEOUT_LEY:  "10s",
			},
			invalidConfig: nil, // FastHttpListener accepts any config
		},
		{
			name:    "SecureNetHttpListener",
			factory: func(c any) (any, error) { return NewSecureNetHttpListener(c) },
			validConfig: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
			},
			invalidConfig: map[string]any{
				"invalid": "config",
			},
		},
		{
			name:    "Http3Listener",
			factory: func(c any) (any, error) { return NewHttp3Listener(c) },
			validConfig: map[string]any{
				CERT_FILE_KEY: certFile,
				KEY_FILE_KEY:  keyFile,
			},
			invalidConfig: map[string]any{
				"invalid": "config",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_valid_config", func(t *testing.T) {
			_, err := tt.factory(tt.validConfig)
			if err != nil {
				t.Errorf("Valid config should not error: %v", err)
			}
		})

		if tt.invalidConfig != nil {
			t.Run(tt.name+"_invalid_config", func(t *testing.T) {
				_, err := tt.factory(tt.invalidConfig)
				if err == nil {
					t.Error("Invalid config should error")
				}
			})
		}
	}
}

// TestListener_TimeoutBehavior tests timeout configurations across listeners
func TestListener_TimeoutBehavior(t *testing.T) {
	tests := []struct {
		name          string
		timeoutConfig map[string]any
		expectedRead  time.Duration
		expectedWrite time.Duration
		expectedIdle  time.Duration
	}{
		{
			name:          "default_timeouts",
			timeoutConfig: map[string]any{},
			expectedRead:  DEFAULT_READ_TIMEOUT,
			expectedWrite: DEFAULT_WRITE_TIMEOUT,
			expectedIdle:  DEFAULT_IDLE_TIMEOUT,
		},
		{
			name: "custom_timeouts",
			timeoutConfig: map[string]any{
				READ_TIMEOUT_KEY:  "30s",
				WRITE_TIMEOUT_KEY: "45s",
				IDLE_TIMEOUT_LEY:  "60s",
			},
			expectedRead:  30 * time.Second,
			expectedWrite: 45 * time.Second,
			expectedIdle:  60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run("FastHttpListener_"+tt.name, func(t *testing.T) {
			service, err := NewFastHttpListener(tt.timeoutConfig)
			if err != nil {
				t.Fatalf("Failed to create FastHttpListener: %v", err)
			}

			listener := service.(*FastHttpListener)
			if listener.readTimeout != tt.expectedRead {
				t.Errorf("Read timeout = %v, want %v", listener.readTimeout, tt.expectedRead)
			}
			if listener.writeTimeout != tt.expectedWrite {
				t.Errorf("Write timeout = %v, want %v", listener.writeTimeout, tt.expectedWrite)
			}
			if listener.idleTimeout != tt.expectedIdle {
				t.Errorf("Idle timeout = %v, want %v", listener.idleTimeout, tt.expectedIdle)
			}
		})

		// Test secure listener timeouts
		certFile, keyFile, cleanup := createTestCertsIntegration(t)
		defer cleanup()

		secureConfig := map[string]any{
			CERT_FILE_KEY: certFile,
			KEY_FILE_KEY:  keyFile,
		}
		for k, v := range tt.timeoutConfig {
			secureConfig[k] = v
		}

		t.Run("SecureNetHttpListener_"+tt.name, func(t *testing.T) {
			service, err := NewSecureNetHttpListener(secureConfig)
			if err != nil {
				t.Fatalf("Failed to create SecureNetHttpListener: %v", err)
			}

			listener := service.(*SecureNetHttpListener)
			if listener.readTimeout != tt.expectedRead {
				t.Errorf("Read timeout = %v, want %v", listener.readTimeout, tt.expectedRead)
			}
			if listener.writeTimeout != tt.expectedWrite {
				t.Errorf("Write timeout = %v, want %v", listener.writeTimeout, tt.expectedWrite)
			}
			if listener.idleTimeout != tt.expectedIdle {
				t.Errorf("Idle timeout = %v, want %v", listener.idleTimeout, tt.expectedIdle)
			}
		})
	}
}

// Helper function to create test certificate files for integration tests
func createTestCertsIntegration(t *testing.T) (certFile, keyFile string, cleanup func()) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Integration"},
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
	certFile = t.TempDir() + "/integration_cert.pem"
	keyFile = t.TempDir() + "/integration_key.pem"

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
