package response

import (
	"encoding/json"
	"net/http"
)

// WriteHttp writes the response to http.ResponseWriter.
// Priority: WriterFunc > Data > empty.
func (r *Response) WriteHttp(w http.ResponseWriter) {
	// apply headers
	for k, values := range r.RespHeaders {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// determine status code
	status := r.RespStatusCode
	if status == 0 {
		status = http.StatusOK
	}

	// 1. Custom writer
	if r.WriterFunc != nil {
		if r.RespContentType != "" {
			w.Header().Set("Content-Type", r.RespContentType)
		}
		w.WriteHeader(status)
		_ = r.WriterFunc(w)
		return
	}

	// 2. JSON encoder
	if r.RespData != nil {
		ct := r.RespContentType
		if ct == "" {
			ct = "application/json"
		}
		w.Header().Set("Content-Type", ct)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(r.RespData)
		return
	}

	w.WriteHeader(status)
}
