package lokstra_registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// RouterIntegrationMode defines how routers communicate
type RouterIntegrationMode string

const (
	RouterModeMonolith      RouterIntegrationMode = "monolith"      // Local router calls via httptest
	RouterModeMicroservices RouterIntegrationMode = "microservices" // HTTP calls to remote servers
	RouterModeHybrid        RouterIntegrationMode = "hybrid"        // Mixed mode based on router location
)

// RouterIntegrationConfig configures how routers communicate
type RouterIntegrationConfig struct {
	Mode           RouterIntegrationMode
	RouterURLs     map[string]string // router-name -> base-url mapping
	Timeout        time.Duration     // HTTP timeout
	RetryCount     int               // Number of retries for HTTP calls
	CircuitBreaker bool              // Enable circuit breaker pattern
}

var routerIntegrationConfig = RouterIntegrationConfig{
	Mode:           RouterModeMonolith, // Default to monolith
	RouterURLs:     make(map[string]string),
	Timeout:        30 * time.Second,
	RetryCount:     3,
	CircuitBreaker: true,
}

// SetRouterIntegrationMode configures how routers communicate globally
func SetRouterIntegrationMode(mode RouterIntegrationMode, config ...map[string]any) {
	routerIntegrationConfig.Mode = mode

	if len(config) > 0 {
		cfg := config[0]

		// Set router URLs for microservices mode
		if urls, ok := cfg["router_urls"].(map[string]string); ok {
			routerIntegrationConfig.RouterURLs = urls
		}

		// Set HTTP timeout
		if timeout, ok := cfg["timeout"].(time.Duration); ok {
			routerIntegrationConfig.Timeout = timeout
		} else if timeoutInt, ok := cfg["timeout"].(int); ok {
			routerIntegrationConfig.Timeout = time.Duration(timeoutInt) * time.Second
		}

		// Set retry count
		if retryCount, ok := cfg["retry_count"].(int); ok {
			routerIntegrationConfig.RetryCount = retryCount
		}

		// Set circuit breaker
		if circuitBreaker, ok := cfg["circuit_breaker"].(bool); ok {
			routerIntegrationConfig.CircuitBreaker = circuitBreaker
		}
	}

	fmt.Printf("🔧 Router Integration Mode: %s\n", mode)
	if mode == RouterModeMicroservices {
		fmt.Printf("🌐 Router URLs: %v\n", routerIntegrationConfig.RouterURLs)
	}
}

// AutoConfigureRouterIntegration automatically configures router integration from config
func AutoConfigureRouterIntegration() {
	deploymentType := GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		SetRouterIntegrationMode(RouterModeMonolith)
		fmt.Println("📍 Router Integration: All routers are local (httptest)")

	case "microservices":
		SetRouterIntegrationMode(RouterModeMicroservices, map[string]any{
			"timeout":         GetConfigInt("router-timeout", 30),
			"retry_count":     GetConfigInt("router-retry-count", 3),
			"circuit_breaker": GetConfigBool("router-circuit-breaker", true),
		})

		fmt.Println("📍 Router Integration: Auto-configuring remote routers from config")

		// Auto-register remote router URLs from config
		serviceToRouter := map[string]string{
			"product-service-url":   "product-api",
			"order-service-url":     "order-api",
			"user-service-url":      "user-api",
			"payment-service-url":   "payment-api",
			"analytics-service-url": "analytics-api",
		}

		for configKey, routerName := range serviceToRouter {
			if serviceURL := GetConfigString(configKey, ""); serviceURL != "" {
				RegisterRouterURL(routerName, serviceURL)
				fmt.Printf("   ↳ %s -> %s\n", routerName, serviceURL)
			}
		}

	case "hybrid":
		SetRouterIntegrationMode(RouterModeHybrid)
	}
}

// RouterClient provides seamless local/remote router communication
type RouterClient struct {
	routerName  string
	baseURL     string
	isLocal     bool
	localRouter http.Handler // For local httptest calls
}

// RegisterRouterURL registers a remote router URL
func RegisterRouterURL(routerName, baseURL string) {
	routerIntegrationConfig.RouterURLs[routerName] = baseURL
	fmt.Printf("📍 Registered router '%s' at %s\n", routerName, baseURL)
}

// GetRouterClient creates a client for seamless router communication
func GetRouterClient(routerName string) *RouterClient {
	client := &RouterClient{
		routerName: routerName,
	}

	// Check if router is registered locally
	if router, exists := GetRouterRegistry()[routerName]; exists {
		client.isLocal = true
		client.localRouter = router
		fmt.Printf("🏠 Router '%s': Local (httptest)\n", routerName)
	} else if baseURL, exists := routerIntegrationConfig.RouterURLs[routerName]; exists {
		client.isLocal = false
		client.baseURL = baseURL
		fmt.Printf("🌐 Router '%s': Remote at %s\n", routerName, baseURL)
	} else {
		fmt.Printf("⚠️ Router '%s' not found locally or remotely\n", routerName)
		return nil
	}

	return client
}

// GET performs a GET request to the router (local or remote)
func (c *RouterClient) GET(path string) (*http.Response, error) {
	return c.makeRequest("GET", path, nil)
}

// POST performs a POST request to the router (local or remote)
func (c *RouterClient) POST(path string, body interface{}) (*http.Response, error) {
	return c.makeRequest("POST", path, body)
}

// PUT performs a PUT request to the router (local or remote)
func (c *RouterClient) PUT(path string, body interface{}) (*http.Response, error) {
	return c.makeRequest("PUT", path, body)
}

// DELETE performs a DELETE request to the router (local or remote)
func (c *RouterClient) DELETE(path string) (*http.Response, error) {
	return c.makeRequest("DELETE", path, nil)
}

// makeRequest handles both local (httptest) and remote (HTTP) calls
func (c *RouterClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	if c.isLocal {
		return c.makeLocalRequest(method, path, body)
	}
	return c.makeRemoteRequest(method, path, body)
}

// makeLocalRequest uses httptest for zero-overhead local calls
func (c *RouterClient) makeLocalRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create httptest request
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Call local router directly (zero network overhead)
	c.localRouter.ServeHTTP(w, req)

	return w.Result(), nil
}

// makeRemoteRequest uses standard HTTP client for remote calls
func (c *RouterClient) makeRemoteRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make HTTP call with timeout
	client := &http.Client{
		Timeout: routerIntegrationConfig.Timeout,
	}

	return client.Do(req)
}

// GetIntegrationMode returns current router integration mode
func GetRouterIntegrationMode() RouterIntegrationMode {
	return routerIntegrationConfig.Mode
}

// Helper function to parse JSON response
func ParseJSONResponse[T any](resp *http.Response, target *T) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, target)
}
