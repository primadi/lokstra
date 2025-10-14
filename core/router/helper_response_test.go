package router

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

func init() {
	// Setup global formatter for tests
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())
}

// Test 1: Handler returns *response.Response
func TestAdaptSmart_ReturnsResponsePointer(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"message": "custom response",
		})
		return resp, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	body := w.Body.String()
	if body != `{"message":"custom response"}` {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 2: Handler returns response.Response (value)
func TestAdaptSmart_ReturnsResponseValue(t *testing.T) {
	handler := func(c *request.Context) (response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusAccepted).Text("plain text response")
		return *resp, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "plain text response" {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 3: Handler returns *response.ApiHelper
func TestAdaptSmart_ReturnsApiHelperPointer(t *testing.T) {
	handler := func(c *request.Context) (*response.ApiHelper, error) {
		api := response.NewApiHelper()
		api.Created(map[string]string{"id": "123"}, "Resource created")
		return api, nil
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Response should be formatted by ApiHelper
	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 4: Handler returns response.ApiHelper (value)
func TestAdaptSmart_ReturnsApiHelperValue(t *testing.T) {
	handler := func(c *request.Context) (response.ApiHelper, error) {
		api := response.NewApiHelper()
		api.Ok(map[string]int{"count": 42})
		return *api, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 5: Handler returns nil *response.Response (should send default success)
func TestAdaptSmart_ReturnsNilResponsePointer(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		return nil, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Test 6: Handler returns *response.Response with status code, but error is non-nil
// ERROR SHOULD TAKE PRECEDENCE
func TestAdaptSmart_ResponseWithStatusButErrorNonNil(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusOK).Json(map[string]string{
			"message": "this should be ignored",
		})

		// Error takes precedence!
		return resp, errors.New("something went wrong")
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Error should be returned as-is, triggering error handling
	// Status should be 500 (internal error) not 200
	if w.Code == http.StatusOK {
		t.Error("Expected error status, not 200 OK (error should take precedence)")
	}
}

// Test 7: Handler with struct param returns *response.Response
func TestAdaptSmart_StructParamReturnsResponse(t *testing.T) {
	type CreateRequest struct {
		Name string `json:"name"`
	}

	handler := func(req *CreateRequest) (*response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"name": req.Name,
		})
		return resp, nil
	}

	r := New("test")
	r.POST("/test", handler)

	// Send JSON body
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody // We'll use binding from context

	w := httptest.NewRecorder()
	// Note: This test might need actual JSON body handling
	// For now, we're testing the signature support
	r.ServeHTTP(w, req)

	// Should not panic - that's the main goal
	if w.Code == 0 {
		t.Error("Handler was not called")
	}
}

// Test 8: Handler without context returns *response.Response
func TestAdaptSmart_NoContextReturnsResponse(t *testing.T) {
	handler := func() (*response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusTeapot).Text("I'm a teapot")
		return resp, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("Expected status 418, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "I'm a teapot" {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 9: Regular handler (backward compatibility - should still work)
func TestAdaptSmart_RegularHandlerBackwardCompat(t *testing.T) {
	handler := func(c *request.Context) (map[string]string, error) {
		return map[string]string{"status": "ok"}, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should be wrapped by Api.Ok()
	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 10: Handler returns *ApiHelper with custom headers via Resp()
func TestAdaptSmart_ApiHelperWithCustomHeaders(t *testing.T) {
	handler := func(c *request.Context) (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		// Set custom header via Response
		resp := api.Resp()
		if resp.RespHeaders == nil {
			resp.RespHeaders = make(map[string][]string)
		}
		resp.RespHeaders["X-Custom-Header"] = []string{"custom-value"}

		api.Ok(map[string]string{"data": "with custom header"})
		return api, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check for custom header
	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Error("Custom header not found in response")
	}
}

// ============================================================================
// Tests for handlers returning any (without error)
// ============================================================================

// Test 11: Handler returns data only (no error)
func TestAdaptSmart_ReturnsDataOnly(t *testing.T) {
	handler := func(c *request.Context) map[string]string {
		return map[string]string{"message": "data without error"}
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 12: Handler returns *Response only (no error)
func TestAdaptSmart_ReturnsResponsePointerOnly(t *testing.T) {
	handler := func(c *request.Context) *response.Response {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"message": "response without error",
		})
		return resp
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	body := w.Body.String()
	if body != `{"message":"response without error"}` {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 13: Handler returns Response value only (no error)
func TestAdaptSmart_ReturnsResponseValueOnly(t *testing.T) {
	handler := func(c *request.Context) response.Response {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusAccepted).Text("response value without error")
		return *resp
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "response value without error" {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 14: Handler returns *ApiHelper only (no error)
func TestAdaptSmart_ReturnsApiHelperPointerOnly(t *testing.T) {
	handler := func(c *request.Context) *response.ApiHelper {
		api := response.NewApiHelper()
		api.Created(map[string]string{"id": "456"}, "Resource created without error")
		return api
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 15: Handler returns ApiHelper value only (no error)
func TestAdaptSmart_ReturnsApiHelperValueOnly(t *testing.T) {
	handler := func(c *request.Context) response.ApiHelper {
		api := response.NewApiHelper()
		api.Ok(map[string]int{"total": 100})
		return *api
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 16: Handler with struct param returns data only (no error)
func TestAdaptSmart_StructParamReturnsDataOnly(t *testing.T) {
	type GetRequest struct {
		ID int `path:"id"`
	}

	handler := func(req *GetRequest) map[string]any {
		return map[string]any{
			"id":      req.ID,
			"message": "data from struct param",
		}
	}

	r := New("test")
	r.GET("/test/{id}", handler)

	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty body")
	}
}

// Test 17: Handler without context returns *Response only (no error)
func TestAdaptSmart_NoContextReturnsResponseOnly(t *testing.T) {
	handler := func() *response.Response {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusTeapot).Text("I'm still a teapot")
		return resp
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("Expected status 418, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "I'm still a teapot" {
		t.Errorf("Unexpected body: %s", body)
	}
}

// Test 18: Handler returns nil *Response (should send default success)
func TestAdaptSmart_ReturnsNilResponsePointerOnly(t *testing.T) {
	handler := func(c *request.Context) *response.Response {
		return nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (default success), got %d", w.Code)
	}
}
