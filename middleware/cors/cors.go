package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "cors"

// Config represents CORS middleware configuration
type Config struct {
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// getDefaultConfig returns default CORS configuration
func getDefaultConfig() *Config {
	return &Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{},
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
func (c *CorsMiddleware) Register(regCtx iface.RegistrationContext) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 30)
}

// Name implements registration.Module.
func (c *CorsMiddleware) Name() string {
	return NAME
}

func factory(config any) lokstra.MiddlewareFunc {
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
			cfg.AllowedOrigins = origins
		}
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			origin := ctx.GetHeader("Origin")

			// Handle actual request
			if origin != "" && matchOrigin(cfg.AllowedOrigins, origin) {
				ctx.WithHeader("Access-Control-Allow-Origin", origin)

				if cfg.AllowCredentials {
					ctx.WithHeader("Access-Control-Allow-Credentials", "true")
				}

				if len(cfg.ExposedHeaders) > 0 {
					ctx.WithHeader("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
				}
			}

			// Handle preflight request
			if ctx.Request.Method == "OPTIONS" {
				// Set allowed methods
				if len(cfg.AllowedMethods) > 0 {
					ctx.WithHeader("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
				}

				// Set allowed headers
				if len(cfg.AllowedHeaders) > 0 {
					if contains(cfg.AllowedHeaders, "*") {
						// If wildcard is allowed, echo back the requested headers
						if reqHeaders := ctx.GetHeader("Access-Control-Request-Headers"); reqHeaders != "" {
							ctx.WithHeader("Access-Control-Allow-Headers", reqHeaders)
						}
					} else {
						ctx.WithHeader("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
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
	if v, ok := m["allowed_origins"]; ok {
		if origins := parseOrigins(v); len(origins) > 0 {
			cfg.AllowedOrigins = origins
		}
	}

	if v, ok := m["allowed_methods"]; ok {
		if methods := parseStringSlice(v); len(methods) > 0 {
			cfg.AllowedMethods = methods
		}
	}

	if v, ok := m["allowed_headers"]; ok {
		if headers := parseStringSlice(v); len(headers) > 0 {
			cfg.AllowedHeaders = headers
		}
	}

	if v, ok := m["exposed_headers"]; ok {
		if headers := parseStringSlice(v); len(headers) > 0 {
			cfg.ExposedHeaders = headers
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

// contains checks if slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var _ lokstra.Module = (*CorsMiddleware)(nil)

// GetModule returns the CORS middleware module
func GetModule() lokstra.Module {
	return &CorsMiddleware{}
}
