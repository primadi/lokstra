package body_limit

import (
	"net/http"
	"path"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/body_limit/internal"
)

const BODY_LIMIT_TYPE = "body_limit"
const PARAMS_MAX_SIZE = "max_size"
const PARAMS_SKIP_LARGE_PAYLOADS = "skip_large_payloads"
const PARAMS_MESSAGE = "message"
const PARAMS_STATUS_CODE = "status_code"
const PARAMS_SKIP_ON_PATH = "skip_on_path"

type Config = internal.Config

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
			if matched := internal.MatchPath(requestPath, pattern); matched {
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
			c.R.Body = internal.NewLimitedReadCloser(c.R.Body, cfg.MaxSize, cfg)
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
