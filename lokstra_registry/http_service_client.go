package lokstra_registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPServiceClient provides common HTTP client functionality for service communication
type HTTPServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewHTTPServiceClient creates a new HTTP service client
func NewHTTPServiceClient(baseURL string) *HTTPServiceClient {
	timeout := time.Duration(serviceIntegrationConfig.Timeout) * time.Second

	return &HTTPServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Timeout: timeout,
	}
}

// GET performs a GET request and unmarshals response to target
func (c *HTTPServiceClient) GET(endpoint string, target any) error {
	url := c.BaseURL + endpoint

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body failed: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return nil
}

// POST performs a POST request with JSON payload
func (c *HTTPServiceClient) POST(endpoint string, payload any, target any) error {
	url := c.BaseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload failed: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST %s returned status %d", url, resp.StatusCode)
	}

	if target != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body failed: %w", err)
		}

		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("unmarshaling response failed: %w", err)
		}
	}

	return nil
}

// PUT performs a PUT request with JSON payload
func (c *HTTPServiceClient) PUT(endpoint string, payload any, target any) error {
	url := c.BaseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload failed: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating PUT request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("PUT %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("PUT %s returned status %d", url, resp.StatusCode)
	}

	if target != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body failed: %w", err)
		}

		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("unmarshaling response failed: %w", err)
		}
	}

	return nil
}

// DELETE performs a DELETE request
func (c *HTTPServiceClient) DELETE(endpoint string) error {
	url := c.BaseURL + endpoint

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating DELETE request failed: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("DELETE %s returned status %d", url, resp.StatusCode)
	}

	return nil
}
