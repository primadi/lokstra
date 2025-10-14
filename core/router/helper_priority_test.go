package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

func init() {
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())
}

// Test: Priority Rules for Mixed Usage

// Test 1: Return value overrides c.Resp
func TestPriority_ReturnOverridesContextResp(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		// Set via c.Resp (should be IGNORED)
		c.Resp.WithStatus(http.StatusOK).Json(map[string]string{
			"source": "c.Resp",
		})

		// Return Response (should be USED)
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"source": "return",
		})
		return resp, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use return value (201), NOT c.Resp (200)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 from return value, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return") {
		t.Errorf("Expected body from return value, got: %s", body)
	}
	if strings.Contains(body, "c.Resp") {
		t.Error("Body should NOT contain c.Resp data (should be overridden)")
	}
}

// Test 2: Return ApiHelper overrides c.Api
func TestPriority_ReturnApiHelperOverridesContextApi(t *testing.T) {
	handler := func(c *request.Context) (*response.ApiHelper, error) {
		// Set via c.Api (should be IGNORED)
		c.Api.Ok(map[string]string{
			"source": "c.Api",
		})

		// Return ApiHelper (should be USED)
		api := response.NewApiHelper()
		api.Created(map[string]string{
			"source": "return",
		}, "Created via return")
		return api, nil
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use return value (201), NOT c.Api (200)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 from return value, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return") {
		t.Errorf("Expected body from return value, got: %s", body)
	}
}

// Test 3: Error overrides both c.Resp and return value
func TestPriority_ErrorOverridesEverything(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		// Set via c.Resp (should be IGNORED)
		c.Resp.WithStatus(http.StatusOK).Json(map[string]string{
			"source": "c.Resp",
		})

		// Return Response with success (should be IGNORED)
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"source": "return",
		})

		// Return error (should be USED)
		return resp, http.ErrAbortHandler
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should NOT be 200 or 201, error should take precedence
	if w.Code == http.StatusOK || w.Code == http.StatusCreated {
		t.Error("Error should take precedence, NOT success status codes")
	}
}

// Test 4: c.Resp used when no return value
func TestPriority_ContextRespUsedWhenNoReturn(t *testing.T) {
	handler := func(c *request.Context) error {
		// Set via c.Resp (should be USED - no return value)
		c.Resp.WithStatus(http.StatusAccepted).Json(map[string]string{
			"source": "c.Resp",
		})
		return nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use c.Resp (202)
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202 from c.Resp, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "c.Resp") {
		t.Errorf("Expected body from c.Resp, got: %s", body)
	}
}

// Test 5: Nil return value falls back to c.Resp
func TestPriority_NilReturnFallsBackToContextResp(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		// Set via c.Resp
		c.Resp.WithStatus(http.StatusAccepted).Json(map[string]string{
			"source": "c.Resp",
		})

		// Return nil (should use c.Resp as fallback? NO!)
		// Actually nil return should trigger default success Api.Ok(nil)
		return nil, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Nil return triggers Api.Ok(nil), which is 200, NOT c.Resp's 202
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (default success), got %d", w.Code)
	}
}

// Test 6: Return value overrides even if c.Resp has WriterFunc set
func TestPriority_ReturnOverridesWriterFunc(t *testing.T) {
	handler := func(c *request.Context) (*response.Response, error) {
		// Set via c.Resp with WriterFunc (should be IGNORED)
		c.Resp.Stream("text/plain", func(w http.ResponseWriter) error {
			w.Write([]byte("from c.Resp stream"))
			return nil
		})

		// Return Response (should be USED)
		resp := response.NewResponse()
		resp.WithStatus(http.StatusOK).Text("from return")
		return resp, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "from return" {
		t.Errorf("Expected 'from return', got: %s", body)
	}
	if strings.Contains(body, "c.Resp") {
		t.Error("Body should NOT contain c.Resp data")
	}
}

// Test 7: Mixed c.Api call and return ApiHelper
func TestPriority_ReturnApiHelperOverridesMultipleApiCalls(t *testing.T) {
	handler := func(c *request.Context) (*response.ApiHelper, error) {
		// Multiple c.Api calls (should be IGNORED)
		c.Api.Ok(map[string]string{"first": "call"})
		c.Api.Created(map[string]string{"second": "call"}, "message")

		// Return ApiHelper (should be USED)
		api := response.NewApiHelper()
		api.Ok(map[string]string{
			"source": "return value wins",
		})
		return api, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use return value (200), not the Created (201)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 from return value, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return value wins") {
		t.Errorf("Expected return value in body, got: %s", body)
	}
}

// Test 8: Regular data return overrides c.Api
func TestPriority_RegularReturnOverridesContextApi(t *testing.T) {
	handler := func(c *request.Context) (map[string]string, error) {
		// Set via c.Api (should be IGNORED)
		c.Api.Created(map[string]string{
			"source": "c.Api",
		}, "Created")

		// Return data (should be USED and wrapped in Api.Ok)
		return map[string]string{
			"source": "return data",
		}, nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Return data is wrapped with Api.Ok (200), NOT Created (201)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 from return data, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return data") {
		t.Errorf("Expected return data in body, got: %s", body)
	}
}

// ============================================================================
// Priority Tests for handlers returning any (without error)
// ============================================================================

// Test 9: Return *Response only overrides c.Resp
func TestPriority_ReturnResponseOnlyOverridesContextResp(t *testing.T) {
	handler := func(c *request.Context) *response.Response {
		// Set via c.Resp (should be IGNORED)
		c.Resp.WithStatus(http.StatusOK).Json(map[string]string{
			"source": "c.Resp",
		})

		// Return Response only (should be USED)
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]string{
			"source": "return only",
		})
		return resp
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use return value (201), NOT c.Resp (200)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 from return value, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return only") {
		t.Errorf("Expected body from return value, got: %s", body)
	}
	if strings.Contains(body, "c.Resp") {
		t.Error("Body should NOT contain c.Resp data")
	}
}

// Test 10: Return *ApiHelper only overrides c.Api
func TestPriority_ReturnApiHelperOnlyOverridesContextApi(t *testing.T) {
	handler := func(c *request.Context) *response.ApiHelper {
		// Set via c.Api (should be IGNORED)
		c.Api.Ok(map[string]string{
			"source": "c.Api",
		})

		// Return ApiHelper only (should be USED)
		api := response.NewApiHelper()
		api.Created(map[string]string{
			"source": "return only",
		}, "Created via return")
		return api
	}

	r := New("test")
	r.POST("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should use return value (201), NOT c.Api (200)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 from return value, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return only") {
		t.Errorf("Expected body from return value, got: %s", body)
	}
}

// Test 11: Return data only overrides c.Api
func TestPriority_ReturnDataOnlyOverridesContextApi(t *testing.T) {
	handler := func(c *request.Context) map[string]string {
		// Set via c.Api (should be IGNORED)
		c.Api.Created(map[string]string{
			"source": "c.Api",
		}, "Created")

		// Return data only (should be USED and wrapped in Api.Ok)
		return map[string]string{
			"source": "return data only",
		}
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Return data is wrapped with Api.Ok (200), NOT Created (201)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 from return data, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return data only") {
		t.Errorf("Expected return data in body, got: %s", body)
	}
}

// Test 12: Nil *Response return sends default success
func TestPriority_NilResponseReturnSendsDefaultSuccess(t *testing.T) {
	handler := func(c *request.Context) *response.Response {
		// Set via c.Resp (might be IGNORED if return is nil)
		c.Resp.WithStatus(http.StatusAccepted).Json(map[string]string{
			"source": "c.Resp",
		})

		// Return nil (should trigger default success)
		return nil
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Nil return triggers Api.Ok(nil), which is 200
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (default success), got %d", w.Code)
	}
}

// Test 13: Return Response only overrides WriterFunc in c.Resp
func TestPriority_ReturnResponseOnlyOverridesWriterFunc(t *testing.T) {
	handler := func(c *request.Context) *response.Response {
		// Set via c.Resp with WriterFunc (should be IGNORED)
		c.Resp.Stream("text/plain", func(w http.ResponseWriter) error {
			w.Write([]byte("from c.Resp stream"))
			return nil
		})

		// Return Response only (should be USED)
		resp := response.NewResponse()
		resp.WithStatus(http.StatusOK).Text("from return only")
		return resp
	}

	r := New("test")
	r.GET("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "from return only" {
		t.Errorf("Expected 'from return only', got: %s", body)
	}
	if strings.Contains(body, "c.Resp") {
		t.Error("Body should NOT contain c.Resp data")
	}
}
