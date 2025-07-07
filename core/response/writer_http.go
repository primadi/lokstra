package response

import (
	"lokstra/common/json"

	"net/http"
)

func (r *Response) WriteHttp(w http.ResponseWriter) error {
	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)
	return json.NewEncoder(w).Encode(r)
}
