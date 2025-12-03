package gzipcompression_test

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/middleware/gzipcompression"
)

func TestGzipCompression(t *testing.T) {
	tests := []struct {
		name              string
		config            *gzipcompression.Config
		responseBody      string
		acceptEncoding    string
		expectCompressed  bool
		expectContentType string
	}{
		{
			name: "compress large response",
			config: &gzipcompression.Config{
				MinSize:          100,
				CompressionLevel: gzip.BestSpeed,
			},
			responseBody:     strings.Repeat("a", 500),
			acceptEncoding:   "gzip, deflate",
			expectCompressed: true,
		},
		{
			name: "skip small response",
			config: &gzipcompression.Config{
				MinSize:          1000,
				CompressionLevel: gzip.BestSpeed,
			},
			responseBody:     "small response",
			acceptEncoding:   "gzip, deflate",
			expectCompressed: false,
		},
		{
			name: "client doesn't support gzip",
			config: &gzipcompression.Config{
				MinSize:          100,
				CompressionLevel: gzip.BestSpeed,
			},
			responseBody:     strings.Repeat("a", 500),
			acceptEncoding:   "deflate",
			expectCompressed: false,
		},
		{
			name: "exclude image content type",
			config: &gzipcompression.Config{
				MinSize:              100,
				CompressionLevel:     gzip.BestSpeed,
				ExcludedContentTypes: []string{"image/jpeg", "image/png"},
			},
			responseBody:      strings.Repeat("a", 500),
			acceptEncoding:    "gzip, deflate",
			expectCompressed:  false,
			expectContentType: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup formatter
			api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

			// Create router
			r := router.New("test-router")

			// Add gzip middleware
			r.Use(gzipcompression.Middleware(tt.config))

			// Add test handler
			r.GET("/test", func(c *request.Context) error {
				if tt.expectContentType != "" {
					c.W.Header().Set("Content-Type", tt.expectContentType)
				}
				c.W.Write([]byte(tt.responseBody))
				return nil
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check if compressed
			isCompressed := w.Header().Get("Content-Encoding") == "gzip"
			if isCompressed != tt.expectCompressed {
				t.Errorf("Expected compressed=%v, got %v", tt.expectCompressed, isCompressed)
			}

			// Verify content
			if isCompressed {
				// Decompress and verify
				gr, err := gzip.NewReader(w.Body)
				if err != nil {
					t.Fatalf("Failed to create gzip reader: %v", err)
				}
				defer gr.Close()

				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("Failed to decompress: %v", err)
				}

				if string(decompressed) != tt.responseBody {
					t.Errorf("Decompressed content doesn't match original")
				}
			} else {
				// Verify uncompressed content
				body := w.Body.String()
				if body != tt.responseBody {
					t.Errorf("Response body doesn't match. Expected %d bytes, got %d bytes", len(tt.responseBody), len(body))
				}
			}
		})
	}
}

func TestGzipCompressionWithDefaultConfig(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	r := router.New("test-router")
	r.Use(gzipcompression.Middleware(&gzipcompression.Config{}))

	r.GET("/test", func(c *request.Context) error {
		// Large response (> 1KB default)
		c.W.Write([]byte(strings.Repeat("test data ", 200)))
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should be compressed with default config
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected response to be compressed with default config")
	}

	// Verify content-length header is removed
	if w.Header().Get("Content-Length") != "" {
		t.Error("Content-Length header should be removed for compressed responses")
	}
}

func TestGzipCompressionFactory(t *testing.T) {
	// Test with nil params (should use defaults)
	middleware1 := gzipcompression.MiddlewareFactory(nil)
	if middleware1 == nil {
		t.Error("Expected middleware with nil params")
	}

	// Test with custom params
	params := map[string]any{
		gzipcompression.PARAMS_MIN_SIZE:               2048,
		gzipcompression.PARAMS_COMPRESSION_LEVEL:      gzip.BestCompression,
		gzipcompression.PARAMS_EXCLUDED_CONTENT_TYPES: []string{"text/html"},
	}
	middleware2 := gzipcompression.MiddlewareFactory(params)
	if middleware2 == nil {
		t.Error("Expected middleware with custom params")
	}
}
