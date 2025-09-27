package request

import (
	"net/http"
)

// writerWrapper wraps http.ResponseWriter to detect direct writes
type writerWrapper struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	wroteBody   bool
}

func newWriterWrapper(w http.ResponseWriter) *writerWrapper {
	return &writerWrapper{ResponseWriter: w}
}

func (lw *writerWrapper) WriteHeader(code int) {
	if lw.wroteHeader {
		// status code already written → ignore subsequent calls
		return
	}
	lw.statusCode = code
	lw.wroteHeader = true
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *writerWrapper) Write(b []byte) (int, error) {
	if !lw.wroteHeader {
		// default 200 if WriteHeader not called yet
		lw.WriteHeader(http.StatusOK)
	}
	lw.wroteBody = true
	return lw.ResponseWriter.Write(b)
}

// Check if user wrote manually
func (lw *writerWrapper) ManualWritten() bool {
	return lw.wroteHeader || lw.wroteBody
}
