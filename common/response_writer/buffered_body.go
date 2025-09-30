package response_writer

import (
	"bytes"
	"net/http"
)

type BufferedBodyWriter struct {
	http.ResponseWriter
	Buf  bytes.Buffer
	Code int
}

func NewBufferedBodyWriter(w http.ResponseWriter) *BufferedBodyWriter {
	return &BufferedBodyWriter{ResponseWriter: w}
}

func (d *BufferedBodyWriter) Write(b []byte) (int, error) {
	d.Buf.Write(b)
	return len(b), nil
}

func (d *BufferedBodyWriter) WriteHeader(code int) {
	d.Code = code
}
