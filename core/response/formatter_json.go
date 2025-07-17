package response

import (
	"io"
	"net/http"
	"os"

	"github.com/primadi/lokstra/common/iface/response_iface"
	"github.com/primadi/lokstra/common/json"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

func (f *JSONFormatter) WriteHttp(w http.ResponseWriter, r response_iface.Response) error {
	for k, v := range r.GetHeaders() {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.Header().Set("Content-Type", f.ContentType())
	w.WriteHeader(r.GetStatusCode())
	return json.NewEncoder(w).Encode(r)
}

func (f *JSONFormatter) WriteBuffer(w io.Writer, r response_iface.Response) error {
	return json.NewEncoder(w).Encode(r)
}

func (f *JSONFormatter) WriteStdout(r response_iface.Response) error {
	return f.WriteBuffer(os.Stdout, r)
}

var _ response_iface.ResponseFormatter = (*JSONFormatter)(nil)
