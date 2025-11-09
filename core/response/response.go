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

func NewJsonResponse(data any) *Response {
	r := NewResponse()
	r.Json(data)
	return r
}

func NewHtmlResponse(html string) *Response {
	r := NewResponse()
	r.Html(html)
	return r
}
func NewTextResponse(text string) *Response {
	r := NewResponse()
	r.Text(text)
	return r
}

func NewRawResponse(contentType string, b []byte) *Response {
	r := NewResponse()
	r.Raw(contentType, b)
	return r
}

func NewStreamResponse(contentType string, fn func(w http.ResponseWriter) error) *Response {
	r := NewResponse()
	r.Stream(contentType, fn)
	return r
}
