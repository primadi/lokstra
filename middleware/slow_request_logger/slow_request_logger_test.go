package slow_request_logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/request"
)

func TestSlowRequestLogger_DefaultConfig(t *testing.T) {
	config := &Config{
		Threshold:           500 * time.Millisecond,
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
	}
	mw := GetMidware(config)

	handler := func(ctx *request.Context) error {
		time.Sleep(600 * time.Millisecond)
		ctx.Writer.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	ctx, _ := request.NewContext(nil, w, req)

	// Simulate middleware execution
	err := mw.MiddlewareFn(handler)(ctx)
	if err != nil {
		t.Errorf("Handler error: %v", err)
	}

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK")
	}
	// Manual log check required for slow request output
}

func TestSlowRequestLogger_Threshold(t *testing.T) {
	config := &Config{
		Threshold:           100 * time.Millisecond,
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
	}
	mw := GetMidware(config)

	handler := func(ctx *request.Context) error {
		time.Sleep(150 * time.Millisecond)
		ctx.Writer.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest("GET", "/threshold", nil)
	w := httptest.NewRecorder()
	ctx, _ := request.NewContext(nil, w, req)

	err := mw.MiddlewareFn(handler)(ctx)
	if err != nil {
		t.Errorf("Handler error: %v", err)
	}

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK")
	}
}
