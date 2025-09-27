package response

import (
	"encoding/json"
	"net/http"
)

// WriteHttp writes the response to http.ResponseWriter.
// Priority: WriterFunc > Data > empty.
func (r *Response) WriteHttp(w http.ResponseWriter) {
	// apply headers
	for k, values := range r.Headers {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// determine status code
	status := r.StatusCode
	if status == 0 {
		status = http.StatusOK
	}

	// 1. Custom writer
	if r.WriterFunc != nil {
		if r.ContentType != "" {
			w.Header().Set("Content-Type", r.ContentType)
		}
		w.WriteHeader(status)
		_ = r.WriterFunc(w)
		return
	}

	// 2. JSON encoder
	if r.Data != nil {
		ct := r.ContentType
		if ct == "" {
			ct = "application/json"
		}
		w.Header().Set("Content-Type", ct)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(r.Data)
		return
	}

	w.WriteHeader(status)
}
