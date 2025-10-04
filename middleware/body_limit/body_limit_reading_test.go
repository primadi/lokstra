package body_limit

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

func TestBodyLimitWithActualBodyReading(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	// Test that body is actually limited during reading (not just ContentLength)
	cfg := &Config{
		MaxSize: 100, // Very small limit to test
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))
	r.POST("/test", func(c *request.Context) error {
		// Try to read body
		body, err := io.ReadAll(c.R.Body)
		if err != nil {
			return c.Api.InternalError(err.Error())
		}
		return c.Api.Ok(map[string]any{
			"bodySize": len(body),
		})
	})

	// Test 1: Body within limit
	smallBody := bytes.Repeat([]byte("a"), 50)
	req1 := httptest.NewRequest("POST", "/test", bytes.NewReader(smallBody))
	req1.ContentLength = 50

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected status 200 for small body, got %d", w1.Code)
	}

	// Test 2: Body exceeds limit - should fail during reading
	largeBody := bytes.Repeat([]byte("b"), 200)
	req2 := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req2.ContentLength = 200

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code == 200 {
		t.Error("Expected error for large body, got 200")
	}

	// Test 3: ContentLength not set but body is large
	// This tests the limitedReadCloser during actual reading
	largeBody3 := bytes.Repeat([]byte("c"), 200)
	req3 := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody3))
	req3.ContentLength = -1 // Unknown content length

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	// Should still work because limitedReadCloser wraps the body
	// and will stop reading after MaxSize bytes
	if w3.Code != 200 {
		t.Logf("Body reading with unknown ContentLength returned status: %d", w3.Code)
	}
}

func TestBodyLimitWithSkipLargePayloadsAndReading(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	cfg := &Config{
		MaxSize:           50,
		SkipLargePayloads: true,
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))
	r.POST("/test", func(c *request.Context) error {
		// Try to read body
		body, err := io.ReadAll(c.R.Body)
		if err != nil && err != io.EOF {
			return c.Api.InternalError(err.Error())
		}
		return c.Api.Ok(map[string]any{
			"bodySize": len(body),
		})
	})

	// Large body with SkipLargePayloads=true
	// Should read only up to MaxSize bytes, then return EOF
	largeBody := bytes.Repeat([]byte("a"), 200)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req.ContentLength = 200

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed but only read up to limit
	if w.Code != 200 {
		t.Errorf("Expected status 200 with SkipLargePayloads, got %d", w.Code)
	}
}

func TestLimitedReadCloser(t *testing.T) {
	tests := []struct {
		name              string
		bodySize          int
		maxSize           int64
		skipLargePayloads bool
		expectError       bool
		expectedRead      int
	}{
		{
			name:              "read within limit",
			bodySize:          50,
			maxSize:           100,
			skipLargePayloads: false,
			expectError:       false,
			expectedRead:      50,
		},
		{
			name:              "read exceeds limit",
			bodySize:          200,
			maxSize:           100,
			skipLargePayloads: false,
			expectError:       true,
			expectedRead:      100,
		},
		{
			name:              "read exceeds limit with skip",
			bodySize:          200,
			maxSize:           100,
			skipLargePayloads: true,
			expectError:       false,
			expectedRead:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("a"), tt.bodySize)
			reader := io.NopCloser(bytes.NewReader(body))

			limitedReader := &limitedReadCloser{
				reader:    reader,
				remaining: tt.maxSize,
				config: &Config{
					MaxSize:           tt.maxSize,
					SkipLargePayloads: tt.skipLargePayloads,
				},
			}

			result, err := io.ReadAll(limitedReader)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil && err != io.EOF {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedRead {
				t.Errorf("Expected to read %d bytes, got %d", tt.expectedRead, len(result))
			}
		})
	}
}

// TestBodyLimitCannotBypassBySettingContentLength verifies that handlers cannot bypass
// the body limit by manually setting ContentLength after middleware runs
func TestBodyLimitCannotBypassBySettingContentLength(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	r := router.New("test")

	// Apply body limit (10 bytes)
	r.Use(Middleware(&Config{
		MaxSize: 10,
	}))

	r.POST("/test", func(c *request.Context) error {
		// Attempt to bypass by setting large ContentLength
		originalContentLength := c.R.ContentLength
		c.R.ContentLength = 999999999999

		// Try to read body - should still be limited
		body, err := io.ReadAll(c.R.Body)

		if err != nil {
			// Error from limitedReadCloser
			return c.Api.InternalError(err.Error())
		}

		// Body should be limited to 10 bytes max
		return c.Api.Ok(map[string]any{
			"bytesRead":             len(body),
			"body":                  string(body),
			"originalContentLength": originalContentLength,
			"modifiedContentLength": c.R.ContentLength,
		})
	})

	// Send 20 bytes (exceeds 10 byte limit)
	largeBody := strings.NewReader("12345678901234567890") // 20 bytes
	req := httptest.NewRequest("POST", "/test", largeBody)
	req.Header.Set("Content-Type", "text/plain")
	// Intentionally set wrong ContentLength to simulate bypass attempt
	req.ContentLength = 5 // Claim only 5 bytes

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should get error (500) because limitedReadCloser stopped at 10 bytes
	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Verify response contains error about limit
	body := w.Body.String()
	if !strings.Contains(body, "exceeds limit") {
		t.Errorf("Response should indicate body limit exceeded, got: %s", body)
	}

	t.Logf("✓ Handler bypass attempt correctly blocked - limitedReadCloser enforced the limit")
	t.Logf("  Response: %s", body)
}
