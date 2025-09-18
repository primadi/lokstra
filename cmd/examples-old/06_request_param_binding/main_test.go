package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primadi/lokstra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a single global registration context for all tests. Calling
// NewGlobalRegistrationContext multiple times causes a panic because the
// framework initializes global state. Initialize once at package init.
var testRegCtx = lokstra.NewGlobalRegistrationContext()

// normalizeResponse handles Lokstra's response envelope. The framework encodes
// a Response struct into JSON with top-level fields like {"code","success","data":{...}}.
// This helper extracts the inner `data` map when present so tests can assert on
// the actual payload written by handlers.
func normalizeResponse(t *testing.T, raw []byte) map[string]any {
	var outer map[string]any
	err := json.Unmarshal(raw, &outer)
	require.NoError(t, err)

	if d, ok := outer["data"]; ok {
		if dm, ok := d.(map[string]any); ok {
			return dm
		}
	}
	return outer
}

func setupTestApp() *lokstra.App {
	// Reuse the package-level registration context
	// Use default NewApp which picks up the repository default router engine
	app := lokstra.NewApp(testRegCtx, "test-binding", ":8080")
	setupRoutes(app)
	// Build router so ServeHTTP works without calling Start()
	if err := app.BuildRouter(); err != nil {
		panic(err)
	}
	return app
}

func TestHealthCheck(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("response body: %s", w.Body.String())
	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "request-binding-examples", response["service"])
	assert.NotNil(t, response["timestamp"])
}

func TestManualBinding(t *testing.T) {
	app := setupTestApp()

	// Test with various query parameters
	url := "/users/user123?page=2&limit=10&tags=web&tags=api&active=true"
	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "TestAgent/1.0")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any = normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "manual_binding", response["method"])

	data := response["data"].(map[string]any)
	assert.Equal(t, "user123", data["id"])
	assert.Equal(t, float64(2), data["page"]) // JSON numbers are float64
	assert.Equal(t, float64(10), data["limit"])
	assert.Equal(t, []any{"web", "api"}, data["tags"])
	assert.Equal(t, true, data["active"])
	assert.Equal(t, "Bearer token123", data["authorization"])
	assert.Equal(t, "TestAgent/1.0", data["user_agent"])
}

func TestManualBindingMissingPath(t *testing.T) {
	app := setupTestApp()

	// Test with missing path parameter (should be handled by router)
	req := httptest.NewRequest("GET", "/users/", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)
	// Should return 404 or be handled by router
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestSmartBinding(t *testing.T) {
	app := setupTestApp()

	// Create request body
	requestData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
		"preferences": map[string]any{
			"theme":         "dark",
			"language":      "en",
			"notifications": true,
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	url := "/users/user456/smart?page=1&limit=5&tags=premium&active=false"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer smarttoken")
	req.Header.Set("User-Agent", "SmartClient/2.0")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "smart_binding", response["method"])

	data := response["data"].(map[string]any)

	// Check path parameter
	assert.Equal(t, "user456", data["id"])

	// Check query parameters
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(5), data["limit"])
	assert.Equal(t, []any{"premium"}, data["tags"])
	assert.Equal(t, false, data["active"])

	// Check headers
	assert.Equal(t, "Bearer smarttoken", data["authorization"])
	assert.Equal(t, "SmartClient/2.0", data["user_agent"])

	// Check body parameters
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
	assert.Equal(t, float64(30), data["age"])

	preferences := data["preferences"].(map[string]any)
	assert.Equal(t, "dark", preferences["theme"])
	assert.Equal(t, "en", preferences["language"])
	assert.Equal(t, true, preferences["notifications"])
}

func TestBindBodySmartToMap(t *testing.T) {
	app := setupTestApp()

	// Test with various data types
	testCases := []struct {
		name     string
		data     map[string]any
		expected map[string]any
	}{
		{
			name: "simple_data",
			data: map[string]any{
				"name":  "Alice",
				"age":   25,
				"email": "alice@example.com",
			},
			expected: map[string]any{
				"name":  "Alice",
				"age":   float64(25),
				"email": "alice@example.com",
			},
		},
		{
			name: "complex_nested_data",
			data: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "Bob",
						"settings": map[string]any{
							"theme": "light",
							"lang":  "es",
						},
					},
				},
				"tags":   []string{"admin", "power-user"},
				"active": true,
			},
			expected: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "Bob",
						"settings": map[string]any{
							"theme": "light",
							"lang":  "es",
						},
					},
				},
				"tags":   []any{"admin", "power-user"},
				"active": true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.data)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/users/create-map", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			response := normalizeResponse(t, w.Body.Bytes())

			assert.Equal(t, "bind_body_smart_to_map", response["method"])
			assert.Equal(t, "map[string]interface {}", response["data_type"])
			assert.Equal(t, float64(len(tc.expected)), response["field_count"])

			receivedData := response["received_data"].(map[string]any)
			assert.Equal(t, tc.expected, receivedData)
		})
	}
}

func TestBindBodySmartToMapInvalidJSON(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("POST", "/users/create-map", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Contains(t, response["message"], "Body binding failed")
}

func TestBindAllSmartToMapLimitation(t *testing.T) {
	app := setupTestApp()

	requestData := map[string]any{
		"name": "Charlie",
		"age":  35,
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/users/user789/all-map", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "bind_all_smart_to_map", response["method"])

	// Should show error and recommendation
	if errorMsg, exists := response["error"]; exists {
		assert.NotEmpty(t, errorMsg)
		assert.Equal(t, "BindAllSmart to map[string]any failed as expected. Use hybrid approach instead.", response["message"])
		assert.Equal(t, "/users/:id/hybrid", response["see_endpoint"])
	}
}

func TestHybridBinding(t *testing.T) {
	app := setupTestApp()

	// Complex dynamic body data
	requestData := map[string]any{
		"profile": map[string]any{
			"firstName": "David",
			"lastName":  "Smith",
			"avatar":    "https://example.com/avatar.jpg",
		},
		"settings": map[string]any{
			"notifications": map[string]any{
				"email": true,
				"sms":   false,
				"push":  true,
			},
			"privacy": map[string]any{
				"showEmail":   false,
				"showProfile": true,
			},
		},
		"metadata": map[string]any{
			"source":    "api",
			"version":   "2.0",
			"timestamp": time.Now().Unix(),
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	url := "/users/hybrid123/hybrid?page=3&limit=20&tags=vip&tags=beta"
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer hybridtoken")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "hybrid_binding", response["method"])
	assert.Equal(t, "This is the recommended pattern for flexible APIs", response["recommendation"])

	// Check structured data
	structuredData := response["structured_data"].(map[string]any)
	assert.Equal(t, "hybrid123", structuredData["id"])
	assert.Equal(t, float64(3), structuredData["page"])
	assert.Equal(t, float64(20), structuredData["limit"])
	assert.Equal(t, []any{"vip", "beta"}, structuredData["tags"])
	assert.Equal(t, "Bearer hybridtoken", structuredData["authorization"])

	// Check dynamic body
	dynamicBody := response["dynamic_body"].(map[string]any)

	profile := dynamicBody["profile"].(map[string]any)
	assert.Equal(t, "David", profile["firstName"])
	assert.Equal(t, "Smith", profile["lastName"])

	settings := dynamicBody["settings"].(map[string]any)
	notifications := settings["notifications"].(map[string]any)
	assert.Equal(t, true, notifications["email"])
	assert.Equal(t, false, notifications["sms"])
}

func TestComplexQueryBinding(t *testing.T) {
	app := setupTestApp()

	// Test complex query parameters
	url := "/search?q=lokstra&filter=type:web&filter=lang:go&sort=name&page=1&limit=10&opt[format]=json&opt[include]=docs&date=2023-01-01&date=2023-12-31"
	req := httptest.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "complex_query_binding", response["method"])

	searchParams := response["search_parameters"].(map[string]any)
	assert.Equal(t, "lokstra", searchParams["query"])
	assert.Equal(t, []any{"type:web", "lang:go"}, searchParams["filters"])
	assert.Equal(t, "name", searchParams["sort"])
	assert.Equal(t, float64(1), searchParams["page"])
	assert.Equal(t, float64(10), searchParams["limit"])
	assert.Equal(t, []any{"2023-01-01", "2023-12-31"}, searchParams["date_range"])

	// Note: map[string]string might need special handling in Lokstra
	// This test might reveal how Lokstra handles nested query parameters
}

func TestFormDataBinding(t *testing.T) {
	app := setupTestApp()

	// Test form data instead of JSON
	formData := "name=FormUser&email=form@example.com&age=28"
	req := httptest.NewRequest("POST", "/users/create-map", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	response := normalizeResponse(t, w.Body.Bytes())

	assert.Equal(t, "bind_body_smart_to_map", response["method"])

	// Check that form data was properly parsed into map
	receivedData := response["received_data"].(map[string]any)
	assert.Equal(t, "FormUser", receivedData["name"])
	assert.Equal(t, "form@example.com", receivedData["email"])
	// Note: Form data might come as string, not number
	assert.Equal(t, "28", receivedData["age"])
}

// Benchmark tests to compare binding performance
func BenchmarkManualBinding(b *testing.B) {
	app := setupTestApp()

	url := "/users/bench123?page=1&limit=10&tags=test&active=true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer benchtoken")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkSmartBinding(b *testing.B) {
	app := setupTestApp()

	requestData := map[string]any{
		"name":  "BenchUser",
		"email": "bench@example.com",
		"age":   30,
	}
	jsonData, _ := json.Marshal(requestData)

	url := "/users/bench456/smart?page=1&limit=10&tags=test&active=true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer benchtoken")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkHybridBinding(b *testing.B) {
	app := setupTestApp()

	requestData := map[string]any{
		"dynamic_field_1": "value1",
		"dynamic_field_2": "value2",
		"nested": map[string]any{
			"field": "value",
		},
	}
	jsonData, _ := json.Marshal(requestData)

	url := "/users/bench789/hybrid?page=1&limit=10&tags=test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer benchtoken")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}
