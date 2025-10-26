package body_limit

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const BODY_LIMIT_TYPE = "body_limit"
const PARAMS_MAX_SIZE = "max_size"
const PARAMS_SKIP_LARGE_PAYLOADS = "skip_large_payloads"
const PARAMS_MESSAGE = "message"
const PARAMS_STATUS_CODE = "status_code"
const PARAMS_SKIP_ON_PATH = "skip_on_path"

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

func DefaultConfig() *Config {
	return &Config{
		MaxSize:           10 * 1024 * 1024, // 10MB default
		SkipLargePayloads: false,
		Message:           "Request body too large",
		StatusCode:        http.StatusRequestEntityTooLarge, // 413
		SkipOnPath:        []string{},
	}
}

// BodyLimit middleware to enforce maximum request body size
func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = defConfig.MaxSize
	}
	if cfg.Message == "" {
		cfg.Message = defConfig.Message
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = defConfig.StatusCode
	}

	return request.HandlerFunc(func(c *request.Context) error {
		// Check if current path should skip body limit
		requestPath := path.Clean(c.R.URL.Path)
		for _, pattern := range cfg.SkipOnPath {
			if matched := matchPath(requestPath, pattern); matched {
				return c.Next()
			}
		}

		// Optional: Early validation using ContentLength header if available
		// Note: This only works if:
		// - Client sets ContentLength in HTTP request
		// - OR previous middleware sets ContentLength
		// If ContentLength is -1 (unknown), this check is skipped
		if c.R.ContentLength > 0 && c.R.ContentLength > cfg.MaxSize {
			if !cfg.SkipLargePayloads {
				// Reject immediately based on declared size
				return c.Api.Error(cfg.StatusCode, "BODY_TOO_LARGE", cfg.Message)
			}
			// If SkipLargePayloads=true, continue but body will be limited by reader
		}

		// Primary enforcement: Wrap body reader with limitedReadCloser
		// This provides reliable protection regardless of:
		// - ContentLength header presence or accuracy
		// - Middleware ordering
		// - When body is actually read
		// The limit is enforced during actual read operations
		if c.R.Body != nil {
			c.R.Body = &limitedReadCloser{
				reader:    c.R.Body,
				remaining: cfg.MaxSize,
				config:    cfg,
			}
		}

		// Continue to next handler
		// Body will be limited by limitedReadCloser when handler reads it
		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		MaxSize:           utils.GetValueFromMap(params, PARAMS_MAX_SIZE, defConfig.MaxSize),
		SkipLargePayloads: utils.GetValueFromMap(params, PARAMS_SKIP_LARGE_PAYLOADS, defConfig.SkipLargePayloads),
		Message:           utils.GetValueFromMap(params, PARAMS_MESSAGE, defConfig.Message),
		StatusCode:        utils.GetValueFromMap(params, PARAMS_STATUS_CODE, defConfig.StatusCode),
		SkipOnPath:        utils.GetValueFromMap(params, PARAMS_SKIP_ON_PATH, defConfig.SkipOnPath),
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(BODY_LIMIT_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}

// matchPath checks if a path matches a pattern
// Supports basic wildcard patterns with * and **
func matchPath(requestPath, pattern string) bool {
	// Direct match
	if requestPath == pattern {
		return true
	}

	// Handle ** patterns (match any number of path segments)
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			prefixMatch := prefix == "" || strings.HasPrefix(requestPath, prefix)
			suffixMatch := suffix == "" || strings.HasSuffix(requestPath, suffix)

			return prefixMatch && suffixMatch
		}
	}

	// Handle single * patterns (should not match path separators)
	if strings.Contains(pattern, "*") && !strings.Contains(pattern, "**") {
		// Split pattern by / to handle segments properly
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(requestPath, "/")

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
	if matched, err := filepath.Match(pattern, requestPath); err == nil && matched {
		return true
	}

	return false
}

// limitedReadCloser wraps the request body to enforce size limits during reading
type limitedReadCloser struct {
	reader    io.ReadCloser
	remaining int64
	config    *Config
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
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

func (l *limitedReadCloser) Close() error {
	return l.reader.Close()
}
