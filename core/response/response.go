package response

import "net/http"

type Response struct {
	ResponseCode string              // logical code, mapped to HTTP status
	Data         any                 // payload (JSON-serializable)
	Headers      map[string][]string // custom headers

	// Output control
	StatusCode  int                             // HTTP status code
	ContentType string                          // MIME type (default: application/json)
	WriterFunc  func(http.ResponseWriter) error // custom writer (streaming/file)
}
