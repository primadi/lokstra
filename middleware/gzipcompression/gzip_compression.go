package gzipcompression

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/old_registry"
)

const GZIP_COMPRESSION_TYPE = "gzip_compression"
const PARAMS_MIN_SIZE = "min_size"
const PARAMS_COMPRESSION_LEVEL = "compression_level"
const PARAMS_EXCLUDED_CONTENT_TYPES = "excluded_content_types"

type Config struct {
	// MinSize is the minimum response size to compress (in bytes)
	// Responses smaller than this will not be compressed
	MinSize int

	// CompressionLevel is the gzip compression level (1-9)
	// 1 = fastest, 9 = best compression, -1 = default compression
	CompressionLevel int

	// ExcludedContentTypes is a list of content types that should not be compressed
	// Example: ["image/jpeg", "image/png", "video/mp4"]
	ExcludedContentTypes []string
}

func DefaultConfig() *Config {
	return &Config{
		MinSize:          1024,                    // 1KB minimum
		CompressionLevel: gzip.DefaultCompression, // -1
		ExcludedContentTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"video/mp4",
			"video/webm",
			"application/zip",
			"application/gzip",
		},
	}
}

// middleware to compress response bodies with gzip
func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg.MinSize <= 0 {
		cfg.MinSize = defConfig.MinSize
	}
	if cfg.CompressionLevel == 0 {
		cfg.CompressionLevel = defConfig.CompressionLevel
	}
	if cfg.ExcludedContentTypes == nil {
		cfg.ExcludedContentTypes = defConfig.ExcludedContentTypes
	}

	return request.HandlerFunc(func(c *request.Context) error {
		// Check if client accepts gzip encoding
		acceptEncoding := c.R.Header.Get("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			// Client doesn't support gzip, pass through
			return c.Next()
		}

		// Wrap the underlying response writer with gzip writer
		originalWriter := c.W.ResponseWriter
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: originalWriter,
			config:         cfg,
			context:        c,
		}

		// Replace the underlying response writer
		c.W.ResponseWriter = gzipWriter

		// Call next handler
		err := c.Next()

		// Close gzip writer if it was used
		if gzipWriter.gzipWriter != nil {
			gzipWriter.gzipWriter.Close()
		}

		// Restore original writer
		c.W.ResponseWriter = originalWriter

		return err
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		MinSize:              utils.GetValueFromMap(params, PARAMS_MIN_SIZE, defConfig.MinSize),
		CompressionLevel:     utils.GetValueFromMap(params, PARAMS_COMPRESSION_LEVEL, defConfig.CompressionLevel),
		ExcludedContentTypes: utils.GetValueFromMap(params, PARAMS_EXCLUDED_CONTENT_TYPES, defConfig.ExcludedContentTypes),
	}
	return Middleware(cfg)
}

func Register() {
	old_registry.RegisterMiddlewareFactory(GZIP_COMPRESSION_TYPE, MiddlewareFactory,
		old_registry.AllowOverride(true))
}

// gzipResponseWriter wraps http.ResponseWriter to compress response
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	config     *Config
	context    *request.Context
	statusCode int
	written    bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode

	// Check if content type should be excluded
	contentType := w.Header().Get("Content-Type")
	for _, excluded := range w.config.ExcludedContentTypes {
		if strings.Contains(contentType, excluded) {
			// Don't compress, write header directly
			w.ResponseWriter.WriteHeader(statusCode)
			w.written = true
			return
		}
	}

	// Don't write header yet, wait for Write() to determine if we compress
	// This allows us to check content size
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	// If already written without compression, pass through
	if w.written && w.gzipWriter == nil {
		return w.ResponseWriter.Write(data)
	}

	// Check content size
	if len(data) < w.config.MinSize && w.gzipWriter == nil {
		// Too small to compress, write directly
		if w.statusCode > 0 {
			w.ResponseWriter.WriteHeader(w.statusCode)
		}
		w.written = true
		return w.ResponseWriter.Write(data)
	}

	// Initialize gzip writer if not already done
	if w.gzipWriter == nil {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Remove content-length as it will change

		if w.statusCode > 0 {
			w.ResponseWriter.WriteHeader(w.statusCode)
		}

		var err error
		w.gzipWriter, err = gzip.NewWriterLevel(w.ResponseWriter, w.config.CompressionLevel)
		if err != nil {
			return 0, err
		}
		w.written = true
	}

	// Write compressed data
	return w.gzipWriter.Write(data)
}

func (w *gzipResponseWriter) Flush() {
	if w.gzipWriter != nil {
		w.gzipWriter.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Ensure gzipResponseWriter implements http.Flusher
var _ http.Flusher = (*gzipResponseWriter)(nil)
var _ io.WriteCloser = (*gzipResponseWriter)(nil)

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter != nil {
		return w.gzipWriter.Close()
	}
	return nil
}
