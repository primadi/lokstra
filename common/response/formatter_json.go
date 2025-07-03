package response

import (
	"io"
	"lokstra/common/json"
	"lokstra/common/response/response_iface"
	"net/http"
	"os"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f JSONFormatter) ContentType() string {
	return "application/json"
}

func (f JSONFormatter) WriteHttp(w http.ResponseWriter, r response_iface.Response) error {
	for k, v := range r.GetHeaders() {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.Header().Set("Content-Type", f.ContentType())
	w.WriteHeader(r.GetStatusCode())
	return json.NewEncoder(w).Encode(response_iface.DefaultTemplateFunc(r))
}

func (f JSONFormatter) WriteBuffer(w io.Writer, r response_iface.Response) error {
	return json.NewEncoder(w).Encode(response_iface.DefaultTemplateFunc(r))
}

func (f JSONFormatter) WriteStdout(r response_iface.Response) error {
	return f.WriteBuffer(os.Stdout, r)
}

var _ response_iface.ResponseFormatter = JSONFormatter{}

// func init() {
// 	RegisterFormatter("application/json", defaultJSONFormatter)
// 	RegisterFormatter("application/vnd.api+json", defaultJSONFormatter)
// 	RegisterFormatter("text/json", defaultJSONFormatter)
// 	RegisterFormatter("text/x-json", defaultJSONFormatter)
// 	RegisterFormatter("application/x-json", defaultJSONFormatter)
// }
