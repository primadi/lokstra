package response

import "net/http"

// ResponseWriterWrapper wraps http.ResponseWriter to track if WriteHeader has been called
type ResponseWriterWrapper struct {
	http.ResponseWriter
	written bool
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{ResponseWriter: w}
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	if !w.written {
		w.ResponseWriter.WriteHeader(code)
		w.written = true
	}
}

func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
