package body_limit

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
)

const NAME = "body_limit"

// Config holds configuration for body limit middleware
type Config struct {
	// MaxSize is the maximum allowed body size in bytes
	MaxSize int64 `json:"max_size" yaml:"max_size"`

	// SkipLargePayloads if true, skips reading body when it exceeds limit
	// If false, returns error immediately
	SkipLargePayloads bool `json:"skip_large_payloads" yaml:"skip_large_payloads"`

	// Message is the custom error message for oversized bodies
	Message string `json:"message" yaml:"message"`

	// StatusCode is the HTTP status code to return for oversized bodies
	StatusCode int `json:"status_code" yaml:"status_code"`

	// SkipOnPath is a list of path patterns to skip body limit check
	SkipOnPath []string `json:"skip_on_path" yaml:"skip_on_path"`
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MaxSize:           10 * 1024 * 1024, // 10MB default
		SkipLargePayloads: false,
		Message:           "Request body too large",
		StatusCode:        http.StatusRequestEntityTooLarge, // 413
		SkipOnPath:        []string{},
	}
}

// BodyLimitMiddleware creates a new body limit middleware with the given config
func BodyLimitMiddleware(config Config) midware.Func {
	if config.MaxSize <= 0 {
		config.MaxSize = DefaultConfig().MaxSize
	}
	if config.Message == "" {
		config.Message = DefaultConfig().Message
	}
	if config.StatusCode == 0 {
		config.StatusCode = DefaultConfig().StatusCode
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			// Check if current path should be skipped
			currentPath := ctx.Request.URL.Path
			for _, skipPath := range config.SkipOnPath {
				if matchPath(currentPath, skipPath) {
					return next(ctx)
				}
			}

			// Check Content-Length header first if available
			if ctx.Request.ContentLength > config.MaxSize {
				if config.SkipLargePayloads {
					// Don't read body, but continue processing
					return next(ctx)
				}

				// Set error response and return error
				ctx.SetStatusCode(config.StatusCode)
				ctx.WithMessage(config.Message).WithData(map[string]any{
					"maxSize": config.MaxSize,
					"actual":  ctx.Request.ContentLength,
				})
				// Return an actual error to stop the chain
				return fmt.Errorf("%s (maxSize: %d, actual: %d)",
					config.Message, config.MaxSize, ctx.Request.ContentLength)
			}

			// If no Content-Length or it's within limit, wrap the body reader
			if ctx.Request.Body != nil {
				ctx.Request.Body = &limitedReadCloser{
					reader:    ctx.Request.Body,
					remaining: config.MaxSize,
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
		l.ctx.WithMessage(l.config.Message).WithData(map[string]any{
			"maxSize": l.config.MaxSize,
		})

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

func (l *limitedReadCloser) Close() error {
	return l.reader.Close()
}

// Convenience functions for common size limits

// BodyLimit creates middleware with specified size limit
func BodyLimit(maxSize int64) midware.Func {
	config := DefaultConfig()
	config.MaxSize = maxSize
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
	config.MaxSize = maxSize
	config.SkipLargePayloads = true
	return BodyLimitMiddleware(config)
}

// factory creates body limit middleware from configuration
func factory(config any) midware.Func {
	cfg := DefaultConfig()

	if config != nil {
		switch c := config.(type) {
		case map[string]any:
			if maxSize, ok := c["max_size"]; ok {
				if size, ok := maxSize.(int64); ok {
					cfg.MaxSize = size
				} else if size, ok := maxSize.(int); ok {
					cfg.MaxSize = int64(size)
				} else if size, ok := maxSize.(float64); ok {
					cfg.MaxSize = int64(size)
				}
			}
			if statusCode, ok := c["status_code"]; ok {
				if code, ok := statusCode.(int); ok {
					cfg.StatusCode = code
				} else if code, ok := statusCode.(float64); ok {
					cfg.StatusCode = int(code)
				}
			}
			if message, ok := c["message"]; ok {
				if msg, ok := message.(string); ok {
					cfg.Message = msg
				}
			}
			if skipOnPath, ok := c["skip_on_path"]; ok {
				if paths, ok := skipOnPath.([]any); ok {
					cfg.SkipOnPath = make([]string, 0, len(paths))
					for _, path := range paths {
						if pathStr, ok := path.(string); ok {
							cfg.SkipOnPath = append(cfg.SkipOnPath, pathStr)
						}
					}
				} else if paths, ok := skipOnPath.([]string); ok {
					cfg.SkipOnPath = paths
				}
			}
		case *Config:
			cfg = *c
		case Config:
			cfg = c
		}
	}

	return BodyLimitMiddleware(cfg)
}

// matchPath checks if a path matches a pattern
// Supports basic wildcard patterns with * and **
func matchPath(path, pattern string) bool {
	// Direct match
	if path == pattern {
		return true
	}

	// Handle ** patterns (match any number of path segments)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			// Remove trailing slash from prefix if present
			prefix = strings.TrimSuffix(prefix, "/")
			// Remove leading slash from suffix if present
			suffix = strings.TrimPrefix(suffix, "/")

			prefixMatch := prefix == "" || strings.HasPrefix(path, prefix)
			suffixMatch := suffix == "" || strings.HasSuffix(path, suffix)

			return prefixMatch && suffixMatch
		}
	}

	// Handle single * patterns (should not match path separators)
	if strings.Contains(pattern, "*") && !strings.Contains(pattern, "**") {
		// Split pattern by /* to handle segments properly
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(path, "/")

		// Must have same number of segments for single * match
		if len(patternParts) != len(pathParts) {
			return false
		}

		for i, patternPart := range patternParts {
			if patternPart == "*" {
				continue // * matches any single segment
			}
			if matched, err := filepath.Match(patternPart, pathParts[i]); err != nil || !matched {
				return false
			}
		}
		return true
	}

	// Fallback to filepath.Match for patterns without wildcards
	if matched, err := filepath.Match(pattern, path); err == nil && matched {
		return true
	}

	return false
}
