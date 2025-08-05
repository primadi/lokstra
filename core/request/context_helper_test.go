package request_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

func TestContext_ShouldStopMiddlewareChain(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		statusCode   int
		expectedStop bool
		description  string
	}{
		{
			name:         "no_error_ok_status",
			err:          nil,
			statusCode:   200,
			expectedStop: false,
			description:  "Should continue when no error and status < 400",
		},
		{
			name:         "no_error_created_status",
			err:          nil,
			statusCode:   201,
			expectedStop: false,
			description:  "Should continue when no error and status < 400",
		},
		{
			name:         "no_error_no_status",
			err:          nil,
			statusCode:   0,
			expectedStop: false,
			description:  "Should continue when no error and no status set",
		},
		{
			name:         "error_present",
			err:          errors.New("test error"),
			statusCode:   200,
			expectedStop: true,
			description:  "Should stop when error present regardless of status",
		},
		{
			name:         "bad_request_status",
			err:          nil,
			statusCode:   400,
			expectedStop: true,
			description:  "Should stop when status >= 400",
		},
		{
			name:         "unauthorized_status",
			err:          nil,
			statusCode:   401,
			expectedStop: true,
			description:  "Should stop when status >= 400",
		},
		{
			name:         "not_found_status",
			err:          nil,
			statusCode:   404,
			expectedStop: true,
			description:  "Should stop when status >= 400",
		},
		{
			name:         "internal_error_status",
			err:          nil,
			statusCode:   500,
			expectedStop: true,
			description:  "Should stop when status >= 400",
		},
		{
			name:         "error_and_bad_status",
			err:          errors.New("test error"),
			statusCode:   400,
			expectedStop: true,
			description:  "Should stop when both error and bad status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			// Set the status code
			ctx.StatusCode = tt.statusCode

			// Test
			shouldStop := ctx.ShouldStopMiddlewareChain(tt.err)

			// Assertions
			if shouldStop != tt.expectedStop {
				t.Errorf("Expected %v, got %v - %s", tt.expectedStop, shouldStop, tt.description)
			}
		})
	}
}

func TestContext_ShouldStopMiddlewareChain_NilContext(t *testing.T) {
	// This test ensures the method doesn't panic with nil context
	// Although in practice this shouldn't happen with proper usage
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShouldStopMiddlewareChain panicked: %v", r)
		}
	}()

	// Setup minimal context
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test various combinations
	testCases := []struct {
		err    error
		status int
		expect bool
	}{
		{nil, 0, false},
		{nil, 200, false},
		{nil, 400, true},
		{errors.New("error"), 0, true},
		{errors.New("error"), 200, true},
		{errors.New("error"), 400, true},
	}

	for _, tc := range testCases {
		ctx.StatusCode = tc.status
		result := ctx.ShouldStopMiddlewareChain(tc.err)
		if result != tc.expect {
			t.Errorf("Error: %v, Status: %d - expected %v, got %v",
				tc.err, tc.status, tc.expect, result)
		}
	}
}

func TestContext_ShouldStopMiddlewareChain_EdgeCases(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test boundary values
	boundaryTests := []struct {
		status int
		expect bool
		desc   string
	}{
		{399, false, "Status 399 should not stop (< 400)"},
		{400, true, "Status 400 should stop (>= 400)"},
		{401, true, "Status 401 should stop (>= 400)"},
		{-1, false, "Negative status should not stop"},
		{999, true, "Status 999 should stop (>= 400)"},
	}

	for _, test := range boundaryTests {
		ctx.StatusCode = test.status
		result := ctx.ShouldStopMiddlewareChain(nil)
		if result != test.expect {
			t.Errorf("%s: expected %v, got %v", test.desc, test.expect, result)
		}
	}
}
