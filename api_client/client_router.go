package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/primadi/lokstra/core/router"
)

// default timeout for HTTP client requests
var DefaultHTTPTimeout = 30 * time.Second

// ClientRouter stores information about where a router can be accessed
type ClientRouter struct {
	RouterName string
	ServerName string
	FullURL    string
	IsLocal    bool
	Router     router.Router

	Timeout time.Duration
}

// performs a GET request to the router with optional headers
func (c *ClientRouter) GET(path string, headers map[string]string) (*http.Response, error) {
	return c.makeRequest("GET", path, nil, headers)
}

// performs a POST request to the router with optional headers
func (c *ClientRouter) POST(path string, body any, headers map[string]string) (*http.Response, error) {
	return c.makeRequest("POST", path, body, headers)
}

// performs a PUT request to the router with optional headers
func (c *ClientRouter) PUT(path string, body any, headers map[string]string) (*http.Response, error) {
	return c.makeRequest("PUT", path, body, headers)
}

// performs a PATCH request to the router with optional headers
func (c *ClientRouter) PATCH(path string, body any, headers map[string]string) (*http.Response, error) {
	return c.makeRequest("PATCH", path, body, headers)
}

// performs a DELETE request to the router with optional headers
func (c *ClientRouter) DELETE(path string, headers map[string]string) (*http.Response, error) {
	return c.makeRequest("DELETE", path, nil, headers)
}

func (c *ClientRouter) Method(method, path string, body any, headers map[string]string) (*http.Response, error) {
	return c.makeRequest(method, path, body, headers)
}

// makeRequest handles both local (router.ServeHTTP) and remote (HTTP) calls, with headers
func (c *ClientRouter) makeRequest(method, path string, body any, headers map[string]string) (*http.Response, error) {
	if c.IsLocal && c.Router != nil {
		// Use router.ServeHTTP for same-server communication (faster than httptest)
		return c.makeLocalRequest(method, path, body, headers)
	}
	// Use HTTP for remote communication
	return c.makeRemoteRequest(method, path, body, headers)
}

// makeLocalRequest uses router.ServeHTTP for zero-overhead local calls, with headers
func (c *ClientRouter) makeLocalRequest(method, path string, body any,
	headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Use router.ServeHTTP directly (faster than httptest roundtrip)
	c.Router.ServeHTTP(w, req)

	return w.Result(), nil
}

// makeRemoteRequest uses standard HTTP client for remote calls, with headers
func (c *ClientRouter) makeRemoteRequest(method, path string, body any,
	headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	url := c.FullURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make HTTP call with timeout
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}
	client := &http.Client{
		Timeout: timeout,
	}

	return client.Do(req)
}
