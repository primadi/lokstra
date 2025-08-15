package body_limit

import (
	"fmt"
	"io"
	"net/http"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
)

// Config holds configuration for body limit middleware
type Config struct {
	// MaxBodySize is the maximum allowed body size in bytes
	MaxBodySize int64

	// SkipLargePayloads if true, skips reading body when it exceeds limit
	// If false, returns error immediately
	SkipLargePayloads bool

	// ErrorMessage is the custom error message for oversized bodies
	ErrorMessage string

	// StatusCode is the HTTP status code to return for oversized bodies
	StatusCode int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MaxBodySize:       10 * 1024 * 1024, // 10MB default
		SkipLargePayloads: false,
		ErrorMessage:      "Request body too large",
		StatusCode:        http.StatusRequestEntityTooLarge, // 413
	}
}

// BodyLimitMiddleware creates a new body limit middleware with the given config
func BodyLimitMiddleware(config Config) midware.Func {
	if config.MaxBodySize <= 0 {
		config.MaxBodySize = DefaultConfig().MaxBodySize
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = DefaultConfig().ErrorMessage
	}
	if config.StatusCode == 0 {
		config.StatusCode = DefaultConfig().StatusCode
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			// Check Content-Length header first if available
			if ctx.Request.ContentLength > config.MaxBodySize {
				if config.SkipLargePayloads {
					// Don't read body, but continue processing
					return next(ctx)
				}

				// Set error response and return error
				ctx.SetStatusCode(config.StatusCode)
				ctx.WithMessage(config.ErrorMessage).WithData(map[string]any{
					"maxSize": config.MaxBodySize,
					"actual":  ctx.Request.ContentLength,
				})
				// Return an actual error to stop the chain
				return fmt.Errorf("%s (maxSize: %d, actual: %d)",
					config.ErrorMessage, config.MaxBodySize, ctx.Request.ContentLength)
			}

			// If no Content-Length or it's within limit, wrap the body reader
			if ctx.Request.Body != nil {
				ctx.Request.Body = &limitedReadCloser{
					reader:    ctx.Request.Body,
					remaining: config.MaxBodySize,
					config:    config,
					ctx:       ctx,
				}
			}

			return next(ctx)
		}
	}
}

// limitedReadCloser wraps the request body to enforce size limits during reading
type limitedReadCloser struct {
	reader    io.ReadCloser
	remaining int64
	config    Config
	ctx       *request.Context
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	if l.remaining <= 0 {
		if l.config.SkipLargePayloads {
			// Return EOF to signal end of reading
			return 0, io.EOF
		}

		// Set error response in context
		l.ctx.SetStatusCode(l.config.StatusCode)
		l.ctx.WithMessage(l.config.ErrorMessage).WithData(map[string]any{
			"maxSize": l.config.MaxBodySize,
		})

		return 0, fmt.Errorf("request body exceeds limit of %d bytes", l.config.MaxBodySize)
	}

	// Limit read to remaining bytes
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}

	n, err := l.reader.Read(p)
	l.remaining -= int64(n)

	return n, err
}

func (l *limitedReadCloser) Close() error {
	return l.reader.Close()
}

// Convenience functions for common size limits

// BodyLimit creates middleware with specified size limit
func BodyLimit(maxSize int64) midware.Func {
	config := DefaultConfig()
	config.MaxBodySize = maxSize
	return BodyLimitMiddleware(config)
}

// BodyLimit1MB creates middleware with 1MB limit
func BodyLimit1MB() midware.Func {
	return BodyLimit(1024 * 1024)
}

// BodyLimit5MB creates middleware with 5MB limit
func BodyLimit5MB() midware.Func {
	return BodyLimit(5 * 1024 * 1024)
}

// BodyLimit10MB creates middleware with 10MB limit
func BodyLimit10MB() midware.Func {
	return BodyLimit(10 * 1024 * 1024)
}

// BodyLimit50MB creates middleware with 50MB limit (for file uploads)
func BodyLimit50MB() midware.Func {
	return BodyLimit(50 * 1024 * 1024)
}

// BodyLimitWithSkip creates middleware that skips large payloads instead of erroring
func BodyLimitWithSkip(maxSize int64) midware.Func {
	config := DefaultConfig()
	config.MaxBodySize = maxSize
	config.SkipLargePayloads = true
	return BodyLimitMiddleware(config)
}
