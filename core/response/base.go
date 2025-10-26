package response

import (
	"net/http"

	"github.com/primadi/lokstra/common/json"
)

// set status code for the response
func (r *Response) WithStatus(code int) *Response {
	r.RespStatusCode = code
	return r
}

// return JSON response from data
// if data is nil, it will return empty object {}
func (r *Response) Json(data any) error {
	if data == nil {
		data = map[string]any{}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.Raw("application/json", b)
}

// return HTML response
func (r *Response) Html(html string) error {
	return r.Raw("text/html; charset=utf-8", []byte(html))
}

// return plain text response
func (r *Response) Text(text string) error {
	return r.Raw("text/plain; charset=utf-8", []byte(text))
}

// return raw response with specified content type
func (r *Response) Raw(contentType string, b []byte) error {
	r.RespContentType = contentType
	r.WriterFunc = func(w http.ResponseWriter) error {
		_, err := w.Write(b)
		return err
	}
	return nil
}

// return stream response with specified content type
func (r *Response) Stream(contentType string, fn func(w http.ResponseWriter) error) error {
	r.RespContentType = contentType
	r.WriterFunc = func(w http.ResponseWriter) error {
		return fn(w)
	}
	return nil
}
