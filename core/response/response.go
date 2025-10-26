package response

import "net/http"

type Response struct {
	RespCode    string              // logical code, mapped to HTTP status
	RespData    any                 // payload (JSON-serializable)
	RespHeaders map[string][]string // custom headers

	// Output control
	RespStatusCode  int                             // HTTP status code
	RespContentType string                          // MIME type (default: application/json)
	WriterFunc      func(http.ResponseWriter) error // custom writer (streaming/file)
}

func NewResponse() *Response {
	return &Response{}
}
