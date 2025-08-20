package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/middleware/body_limit"
)

func main() {
	// Demonstrate body limit middleware usage
	demonstrateBodyLimitMiddleware()
}

func demonstrateBodyLimitMiddleware() {
	fmt.Println("=== Body Limit Middleware Demo ===")

	// Test 1: Small payload within limit
	fmt.Println("1. Testing small payload (within 1MB limit):")
	testBodyLimit(body_limit.BodyLimit1MB(), "small payload", true)

	// Test 2: Large payload exceeding limit
	fmt.Println("\n2. Testing large payload (exceeding 1MB limit):")
	largePayload := strings.Repeat("a", 2*1024*1024) // 2MB
	testBodyLimit(body_limit.BodyLimit1MB(), largePayload, false)

	// Test 3: Custom configuration
	fmt.Println("\n3. Testing custom configuration (100 byte limit):")
	customConfig := body_limit.BodyLimitMiddleware(body_limit.Config{
		MaxSize:    100,
		Message:    "Custom: Payload too large",
		StatusCode: http.StatusRequestEntityTooLarge,
	})
	testBodyLimit(customConfig, "This is a longer payload that exceeds 100 bytes limit", false)

	// Test 4: Skip large payloads
	fmt.Println("\n4. Testing skip large payloads:")
	skipMiddleware := body_limit.BodyLimitWithSkip(50)
	testBodyLimit(skipMiddleware, "This payload is longer than 50 bytes and should be skipped", true)

	fmt.Println("\n=== Demo completed ===")
}

func testBodyLimit(middleware lokstra.MiddlewareFunc, payload string, expectSuccess bool) {
	// Create a test handler
	handler := func(ctx *request.Context) error {
		body, err := ctx.GetRawRequestBody()
		if err != nil {
			return ctx.ErrorBadRequest("Failed to read body: " + err.Error())
		}

		return ctx.Ok(map[string]any{
			"message":  "Request processed successfully",
			"bodySize": len(body),
		})
	}

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Create test request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")

	// Create context and execute
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)

	// Report results
	if expectSuccess {
		if err != nil {
			fmt.Printf("  ❌ Expected success but got error: %v\n", err)
		} else {
			fmt.Printf("  ✅ Success: Processed %d bytes\n", len(payload))
		}
	} else {
		if err == nil {
			fmt.Printf("  ❌ Expected error but request succeeded\n")
		} else {
			fmt.Printf("  ✅ Correctly rejected: %v\n", err)
			fmt.Printf("     Status: %d, Message: %s\n",
				ctx.Response.StatusCode, ctx.Response.Message)
		}
	}
}

/*
Usage in Real Applications:

1. Register middleware with registration context:

func setupMiddlewares(ctx lokstra.RegistrationContext) {
    // Standard limits
    ctx.RegisterMiddlewareFunc("body_limit_1mb", body_limit.BodyLimit1MB())
    ctx.RegisterMiddlewareFunc("body_limit_5mb", body_limit.BodyLimit5MB())
    ctx.RegisterMiddlewareFunc("body_limit_10mb", body_limit.BodyLimit10MB())
    ctx.RegisterMiddlewareFunc("body_limit_50mb", body_limit.BodyLimit50MB())

    // Custom webhook limit
    ctx.RegisterMiddlewareFunc("body_limit_webhook",
        body_limit.BodyLimitMiddleware(body_limit.Config{
            MaxBodySize:  512 * 1024, // 512KB
            ErrorMessage: "Webhook payload too large",
            StatusCode:   413,
        }))
}

2. Use in routes:

// Global limit
router.Use("body_limit_10mb")

// Per-route limits
router.POST("/api/upload", "upload_handler", "body_limit_50mb")
router.POST("/api/webhook", "webhook_handler", "body_limit_webhook")
router.POST("/api/users", "create_user", "body_limit_1mb")

// Group limits
api := router.Group("/api").Use("body_limit_5mb")
files := router.Group("/files").Use("body_limit_50mb")

3. Custom configurations:

smallLimit := body_limit.BodyLimitMiddleware(body_limit.Config{
    MaxBodySize:       1024,     // 1KB
    SkipLargePayloads: false,    // Return error
    ErrorMessage:      "Request too large",
    StatusCode:        413,
})

skipLimit := body_limit.BodyLimitMiddleware(body_limit.Config{
    MaxBodySize:       1024*1024, // 1MB
    SkipLargePayloads: true,      // Skip reading large bodies
    ErrorMessage:      "Body skipped",
    StatusCode:        200,
})

Expected Benefits:
- ✅ Prevents DoS attacks with large payloads
- ✅ Controls memory usage
- ✅ Configurable per route/group
- ✅ Early detection via Content-Length header
- ✅ Runtime protection during body reading
- ✅ Flexible skip behavior for admin operations
*/
