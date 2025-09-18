package cors

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
)

const MODULE_NAME = "cors"

// Config represents CORS middleware configuration
type Config struct {
	AllowOrigins     []string `json:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string `json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// getDefaultConfig returns default CORS configuration
func getDefaultConfig() *Config {
	return &Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

type CorsMiddleware struct{}

// Description implements registration.Module.
func (c *CorsMiddleware) Description() string {
	return "CORS middleware for handling cross-origin requests"
}

// Register implements registration.Module.
func (c *CorsMiddleware) Register(regCtx registration.Context) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(MODULE_NAME, factory, 30)
}

// Name implements registration.Module.
func (c *CorsMiddleware) Name() string {
	return MODULE_NAME
}

func factory(config any) midware.Func {
	cfg := getDefaultConfig()

	// Parse configuration
	switch c := config.(type) {
	case map[string]any:
		parseMapConfig(cfg, c)
	case *Config:
		if c != nil {
			*cfg = *c
		}
	case Config:
		cfg = &c
	case nil:
		// Use default config
	default:
		// For backward compatibility, treat as allowed_origins
		if origins := parseOrigins(config); len(origins) > 0 {
			cfg.AllowOrigins = origins
		}
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			origin := ctx.GetHeader("Origin")

			// Handle actual request
			if origin != "" && matchOrigin(cfg.AllowOrigins, origin) {
				ctx.WithHeader("Access-Control-Allow-Origin", origin)

				if cfg.AllowCredentials {
					ctx.WithHeader("Access-Control-Allow-Credentials", "true")
				}

				if len(cfg.ExposeHeaders) > 0 {
					ctx.WithHeader("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
				}
			}

			// Handle preflight request
			if ctx.Request.Method == "OPTIONS" {
				// Set allowed methods
				if len(cfg.AllowMethods) > 0 {
					ctx.WithHeader("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
				}

				// Set allowed headers
				if len(cfg.AllowHeaders) > 0 {
					if slices.Contains(cfg.AllowHeaders, "*") {
						// If wildcard is allowed, echo back the requested headers
						if reqHeaders := ctx.GetHeader("Access-Control-Request-Headers"); reqHeaders != "" {
							ctx.WithHeader("Access-Control-Allow-Headers", reqHeaders)
						}
					} else {
						ctx.WithHeader("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
					}
				}

				// Set max age for preflight cache
				if cfg.MaxAge > 0 {
					ctx.WithHeader("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
				}

				ctx.SetStatusCode(http.StatusNoContent)
				return nil
			}

			return next(ctx)
		}
	}
}

// parseMapConfig parses map[string]any configuration into Config struct
func parseMapConfig(cfg *Config, m map[string]any) {
	if v, ok := m["allow_origins"]; ok {
		if origins := parseOrigins(v); len(origins) > 0 {
			cfg.AllowOrigins = origins
		}
	}

	if v, ok := m["allow_methods"]; ok {
		if methods := parseStringSlice(v); len(methods) > 0 {
			cfg.AllowMethods = methods
		}
	}

	if v, ok := m["allow_headers"]; ok {
		if headers := parseStringSlice(v); len(headers) > 0 {
			cfg.AllowHeaders = headers
		}
	}

	if v, ok := m["expose_headers"]; ok {
		if headers := parseStringSlice(v); len(headers) > 0 {
			cfg.ExposeHeaders = headers
		}
	}

	if v, ok := m["allow_credentials"]; ok {
		if b, ok := v.(bool); ok {
			cfg.AllowCredentials = b
		}
	}

	if v, ok := m["max_age"]; ok {
		if i, ok := v.(int); ok {
			cfg.MaxAge = i
		} else if f, ok := v.(float64); ok {
			cfg.MaxAge = int(f)
		}
	}
}

// parseOrigins parses various formats into string slice for origins
func parseOrigins(v any) []string {
	return parseStringSlice(v)
}

// parseStringSlice parses various formats into string slice
func parseStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		var result []string
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case string:
		// Split by comma for single string with multiple values
		if strings.Contains(val, ",") {
			parts := strings.Split(val, ",")
			var result []string
			for _, part := range parts {
				if trimmed := strings.TrimSpace(part); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		return []string{val}
	}
	return nil
}

var _ registration.Module = (*CorsMiddleware)(nil)

// GetModule returns the CORS middleware module
func GetModule() registration.Module {
	return &CorsMiddleware{}
}

// Preferred way to get cors middleware execution
func GetMidware(cfg *Config) *midware.Execution {
	return &midware.Execution{
		Name:         MODULE_NAME,
		Config:       cfg,
		MiddlewareFn: factory(cfg),
		Priority:     25,
	}
}
