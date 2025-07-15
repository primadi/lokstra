package gzipcompression

import (
	"io"
	"net/http"

	"github.com/klauspost/compress/gzip"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	writer     io.Writer
	minSize    int
	level      int
	buffer     []byte
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// Buffer first, and compress only if larger than minSize
	w.buffer = append(w.buffer, b...)
	if len(w.buffer) < w.minSize {
		return len(b), nil // Simulate write but do nothing
	}
	if w.gzipWriter == nil {
		gz, err := gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err != nil {
			return 0, err
		}

		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.gzipWriter = gz
		w.writer = gz
		_, err = w.writer.Write(w.buffer)
		if err != nil {
			return 0, err
		}
		w.buffer = nil
	}
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter != nil {
		return w.gzipWriter.Close()
	}
	// Flush uncompressed if below threshold
	if len(w.buffer) > 0 {
		_, err := w.ResponseWriter.Write(w.buffer)
		return err
	}
	return nil
}
