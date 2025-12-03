package internal

import (
	"fmt"
	"io"
)

type Config struct {
	// MaxSize is the maximum allowed body size in bytes
	MaxSize int64

	// SkipLargePayloads if true, skips reading body when it exceeds limit
	// If false, returns error immediately
	SkipLargePayloads bool

	// Message is the custom error message for oversized bodies
	Message string

	// StatusCode is the HTTP status code to return for oversized bodies
	StatusCode int

	// SkipOnPath is a list of path patterns to skip body limit check
	// Supports wildcards like /api/* or /public/** to skip multiple paths
	// Example: ["/api/*", "/public/**"]
	// will skip all paths under /api/ and /public/ (including subpaths)
	SkipOnPath []string
}

// LimitedReadCloser wraps the request body to enforce size limits during reading
type LimitedReadCloser struct {
	reader    io.ReadCloser
	remaining int64
	config    *Config
}

func NewLimitedReadCloser(reader io.ReadCloser, maxSize int64, config *Config) *LimitedReadCloser {
	return &LimitedReadCloser{
		reader:    reader,
		remaining: maxSize,
		config:    config,
	}
}

func (l *LimitedReadCloser) Read(p []byte) (int, error) {
	if l.remaining <= 0 {
		if l.config.SkipLargePayloads {
			// Return EOF to signal end of reading
			return 0, io.EOF
		}

		// Body size exceeded during reading
		return 0, fmt.Errorf("request body exceeds limit of %d bytes", l.config.MaxSize)
	}

	// Limit read to remaining bytes
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}

	n, err := l.reader.Read(p)
	l.remaining -= int64(n)

	return n, err
}

func (l *LimitedReadCloser) Close() error {
	return l.reader.Close()
}
