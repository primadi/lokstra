package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// ============================================================================
// API HELPER RESPONSES (Opinionated JSON API)
// ============================================================================

// Method 1: NewApiOk - Success response (RECOMMENDED for APIs)
func ApiSuccess() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"user": "John Doe",
		"role": "admin",
	})
}

// Method 2: NewApiOkWithMessage - Success with custom message
func ApiSuccessWithMessage() *response.ApiHelper {
	return response.NewApiOkWithMessage(
		map[string]string{"id": "123", "name": "Product"},
		"Product retrieved successfully",
	)
}

// Method 3: NewApiCreated - 201 Created
func ApiCreated() *response.ApiHelper {
	return response.NewApiCreated(
		map[string]any{"id": 456, "status": "active"},
		"Resource created successfully",
	)
}

// Method 4: NewApiError - Error responses
func ApiBadRequest() *response.ApiHelper {
	return response.NewApiBadRequest(
		"INVALID_INPUT",
		"The provided input is invalid",
	)
}

func ApiUnauthorized() *response.ApiHelper {
	return response.NewApiUnauthorized("Authentication required")
}

func ApiNotFound() *response.ApiHelper {
	return response.NewApiNotFound("Resource not found")
}

// Manual method (less common)
func ApiManual() *response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]string{
		"message": "Manual method",
		"note":    "Use helpers instead",
	})
	return api
}

// ============================================================================
// RESPONSE TYPES DEMONSTRATION
// ============================================================================

// Method 1: NewJsonResponse - Quick JSON response (RECOMMENDED)
func JsonQuick() *response.Response {
	return response.NewJsonResponse(map[string]any{
		"message": "Using NewJsonResponse",
		"method":  "Quick and clean",
		"status":  "success",
	})
}

// Method 2: Manual JSON with chainable methods
func JsonManual() *response.Response {
	r := response.NewResponse()
	r.WithStatus(200).Json(map[string]any{
		"message": "Manual JSON with chaining",
		"method":  "More control",
	})
	return r
}

// Method 3: Using ctx.Resp (context helper)
func JsonWithContext(ctx *request.Context) error {
	// ctx.Resp is alias for response.NewResponse()
	return ctx.Resp.WithStatus(200).Json(map[string]any{
		"message": "Using ctx.Resp helper",
		"method":  ctx.R.Method,
		"path":    ctx.R.URL.Path,
	})
}

// ============================================================================
// HTML RESPONSES
// ============================================================================

// Method 1: NewHtmlResponse - Quick HTML
func HtmlQuick() *response.Response {
	html := `
<!DOCTYPE html>
<html>
<head><title>Quick HTML</title></head>
<body>
	<h1>Using NewHtmlResponse</h1>
	<p>Quick and clean HTML response</p>
</body>
</html>`
	return response.NewHtmlResponse(html)
}

// Method 2: Manual HTML with status code
func HtmlManual() *response.Response {
	html := `
<!DOCTYPE html>
<html>
<head><title>Manual HTML</title></head>
<body>
	<h1>Manual HTML Response</h1>
	<p>Using chainable methods for more control</p>
</body>
</html>`
	r := response.NewResponse()
	r.WithStatus(200)
	r.Html(html)
	return r
}

// Method 3: Dynamic HTML with context
func HtmlDynamic(ctx *request.Context) *response.Response {
	name := ctx.R.URL.Query().Get("name")
	if name == "" {
		name = "Guest"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>Dynamic HTML</title></head>
<body>
	<h1>Hello, %s!</h1>
	<p>Request Method: %s</p>
	<p>Request Path: %s</p>
	<p>Time: %s</p>
</body>
</html>`, name, ctx.R.Method, ctx.R.URL.Path, time.Now().Format("2006-01-02 15:04:05"))

	return response.NewHtmlResponse(html)
}

// ============================================================================
// TEXT RESPONSES
// ============================================================================

// Plain text response
func TextPlain() *response.Response {
	return response.NewTextResponse("This is a plain text response\nLine 2\nLine 3")
}

// Dynamic text with context
func TextDynamic(ctx *request.Context) *response.Response {
	text := fmt.Sprintf(`Request Information:
Method: %s
Path: %s
Host: %s
User-Agent: %s
Time: %s`,
		ctx.R.Method,
		ctx.R.URL.Path,
		ctx.R.Host,
		ctx.R.UserAgent(),
		time.Now().Format(time.RFC3339))

	return response.NewTextResponse(text)
}

// ============================================================================
// RAW RESPONSES (Custom Content-Type)
// ============================================================================

// CSV Response
func CsvResponse() *response.Response {
	csv := `name,age,city
John Doe,30,Jakarta
Jane Smith,25,Bandung
Bob Johnson,35,Surabaya`

	return response.NewRawResponse("text/csv", []byte(csv))
}

// XML Response
func XmlResponse() *response.Response {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<response>
	<status>success</status>
	<data>
		<message>Using NewRawResponse for XML</message>
		<timestamp>` + time.Now().Format(time.RFC3339) + `</timestamp>
	</data>
</response>`

	return response.NewRawResponse("application/xml", []byte(xml))
}

// PDF-like binary response (simulated)
func BinaryResponse() *response.Response {
	// In real app, you would read actual PDF file
	fakeData := []byte("%PDF-1.4 Fake PDF content here...")

	r := response.NewRawResponse("application/pdf", fakeData)
	r.WithStatus(200)
	// Add download header
	if r.RespHeaders == nil {
		r.RespHeaders = make(map[string][]string)
	}
	r.RespHeaders["Content-Disposition"] = []string{"attachment; filename=document.pdf"}
	return r
}

// ============================================================================
// STREAM RESPONSES (Server-Sent Events, Chunked Transfer)
// ============================================================================

// Server-Sent Events (SSE)
func StreamSSE(ctx *request.Context) *response.Response {
	return response.NewStreamResponse("text/event-stream", func(w http.ResponseWriter) error {
		// Set SSE headers
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send 5 events
		for i := 1; i <= 5; i++ {
			fmt.Fprintf(w, "data: Event %d at %s\n\n", i, time.Now().Format("15:04:05"))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Fprintf(w, "data: Stream complete\n\n")
		return nil
	})
}

// Chunked text stream
func StreamChunked() *response.Response {
	return response.NewStreamResponse("text/plain", func(w http.ResponseWriter) error {
		lines := []string{
			"Starting stream...",
			"Processing data...",
			"Fetching results...",
			"Compiling response...",
			"Stream complete!",
		}

		for i, line := range lines {
			fmt.Fprintf(w, "[%d] %s\n", i+1, line)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(300 * time.Millisecond)
		}
		return nil
	})
}

// ============================================================================
// COMPARISON EXAMPLES
// ============================================================================

// All three methods produce identical JSON output
func CompareJSON1() *response.Response {
	// Method 1: NewJsonResponse (recommended)
	return response.NewJsonResponse(map[string]string{
		"method":  "NewJsonResponse",
		"message": "Quick and clean",
	})
}

func CompareJSON2() *response.Response {
	// Method 2: Manual with chainable
	r := response.NewResponse()
	r.Json(map[string]string{
		"method":  "Manual chaining",
		"message": "More control",
	})
	return r
}

func CompareJSON3(ctx *request.Context) error {
	// Method 3: Using ctx.Resp
	return ctx.Resp.Json(map[string]string{
		"method":  "ctx.Resp",
		"message": "With context",
	})
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	r := router.New("response-types-demo")

	// JSON responses
	r.GET("/json/quick", JsonQuick)
	r.GET("/json/manual", JsonManual)
	r.GET("/json/context", JsonWithContext)

	// API Helper responses (structured JSON)
	r.GET("/api/success", ApiSuccess)
	r.GET("/api/success-message", ApiSuccessWithMessage)
	r.GET("/api/created", ApiCreated)
	r.GET("/api/bad-request", ApiBadRequest)
	r.GET("/api/unauthorized", ApiUnauthorized)
	r.GET("/api/not-found", ApiNotFound)
	r.GET("/api/manual", ApiManual)

	// HTML responses
	r.GET("/html/quick", HtmlQuick)
	r.GET("/html/manual", HtmlManual)
	r.GET("/html/dynamic", HtmlDynamic)

	// Text responses
	r.GET("/text/plain", TextPlain)
	r.GET("/text/dynamic", TextDynamic)

	// Raw responses (custom content-type)
	r.GET("/raw/csv", CsvResponse)
	r.GET("/raw/xml", XmlResponse)
	r.GET("/raw/binary", BinaryResponse)

	// Stream responses
	r.GET("/stream/sse", StreamSSE)
	r.GET("/stream/chunked", StreamChunked)

	// Comparison endpoints
	r.GET("/compare/json1", CompareJSON1)
	r.GET("/compare/json2", CompareJSON2)
	r.GET("/compare/json3", CompareJSON3)

	// Info endpoint
	r.GET("/", func() *response.Response {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Response Types Demo</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		h1 { color: #333; }
		h2 { color: #666; margin-top: 30px; }
		ul { line-height: 1.8; }
		a { color: #0066cc; text-decoration: none; }
		a:hover { text-decoration: underline; }
		code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
	</style>
</head>
<body>
	<h1>Response Types Demo</h1>
	<p>Demonstrating all response types in Lokstra</p>

	<h2>JSON Responses</h2>
	<ul>
		<li><a href="/json/quick">/json/quick</a> - NewJsonResponse (recommended)</li>
		<li><a href="/json/manual">/json/manual</a> - Manual with chainable methods</li>
		<li><a href="/json/context">/json/context</a> - Using ctx.Resp helper</li>
	</ul>

	<h2>API Helper Responses (Structured JSON)</h2>
	<ul>
		<li><a href="/api/success">/api/success</a> - NewApiOk (success)</li>
		<li><a href="/api/success-message">/api/success-message</a> - With custom message</li>
		<li><a href="/api/created">/api/created</a> - 201 Created</li>
		<li><a href="/api/bad-request">/api/bad-request</a> - 400 Bad Request</li>
		<li><a href="/api/unauthorized">/api/unauthorized</a> - 401 Unauthorized</li>
		<li><a href="/api/not-found">/api/not-found</a> - 404 Not Found</li>
		<li><a href="/api/manual">/api/manual</a> - Manual method (old way)</li>
	</ul>

	<h2>HTML Responses</h2>
	<ul>
		<li><a href="/html/quick">/html/quick</a> - NewHtmlResponse</li>
		<li><a href="/html/manual">/html/manual</a> - Manual HTML</li>
		<li><a href="/html/dynamic?name=John">/html/dynamic?name=John</a> - Dynamic HTML</li>
	</ul>

	<h2>Text Responses</h2>
	<ul>
		<li><a href="/text/plain">/text/plain</a> - Plain text</li>
		<li><a href="/text/dynamic">/text/dynamic</a> - Dynamic text with info</li>
	</ul>

	<h2>Raw Responses (Custom Content-Type)</h2>
	<ul>
		<li><a href="/raw/csv">/raw/csv</a> - CSV file</li>
		<li><a href="/raw/xml">/raw/xml</a> - XML response</li>
		<li><a href="/raw/binary">/raw/binary</a> - Binary (PDF simulation)</li>
	</ul>

	<h2>Stream Responses</h2>
	<ul>
		<li><a href="/stream/sse">/stream/sse</a> - Server-Sent Events</li>
		<li><a href="/stream/chunked">/stream/chunked</a> - Chunked transfer</li>
	</ul>

	<h2>Comparison (All produce same JSON)</h2>
	<ul>
		<li><a href="/compare/json1">/compare/json1</a> - Method 1</li>
		<li><a href="/compare/json2">/compare/json2</a> - Method 2</li>
		<li><a href="/compare/json3">/compare/json3</a> - Method 3</li>
	</ul>
</body>
</html>`
		return response.NewHtmlResponse(html)
	})

	fmt.Println("=== RESPONSE TYPES DEMO ===")
	fmt.Println("All response types and methods demonstrated")
	fmt.Println("Open browser: http://localhost:8080")
	fmt.Println("\nServer starting on :8080...")

	a := app.New("response-demo", ":8080", r)
	srv := server.New("demo-server", a)
	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
