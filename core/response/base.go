package response

import (
	"net/http"

	"github.com/primadi/lokstra/common/json"
)

func (r *Response) WithStatus(code int) *Response {
	r.StatusCode = code
	return r
}

func (r *Response) Json(data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.Raw("application/json", b)
}

func (r *Response) Html(html string) error {
	return r.Raw("text/html; charset=utf-8", []byte(html))
}

func (r *Response) Text(text string) error {
	return r.Raw("text/plain; charset=utf-8", []byte(text))
}

func (r *Response) Raw(contentType string, b []byte) error {
	r.ContentType = contentType
	r.WriterFunc = func(w http.ResponseWriter) error {
		_, err := w.Write(b)
		return err
	}
	return nil
}

func (r *Response) Stream(contentType string, fn func(w http.ResponseWriter) error) error {
	r.ContentType = contentType
	r.WriterFunc = func(w http.ResponseWriter) error {
		return fn(w)
	}
	return nil
}
