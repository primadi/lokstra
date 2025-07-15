package response

import (
	"lokstra/common/json"

	"net/http"
)

func (r *Response) WriteHttp(w http.ResponseWriter) error {
	contentTypeExists := false

	for k, v := range r.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
			if k == "Content-Type" {
				contentTypeExists = true
			}
		}
	}

	if !contentTypeExists {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(r.StatusCode)

	if r.RawData != nil {
		_, err := w.Write(r.RawData)
		return err
	}

	return json.NewEncoder(w).Encode(r)
}
