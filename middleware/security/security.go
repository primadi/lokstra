package security

import (
	"lokstra"
)

const NAME = "lokstra.security"

type SecurityMiddleware struct{}

func (s *SecurityMiddleware) Name() string {
	return NAME
}

func (s *SecurityMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    25,
		Description: "Security headers middleware for common web security protections",
		Tags:        []string{"security", "headers", "protection"},
	}
}

func (s *SecurityMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	enableHSTS := true
	if h, ok := configMap["enable_hsts"].(bool); ok {
		enableHSTS = h
	}

	enableCSP := true
	if c, ok := configMap["enable_csp"].(bool); ok {
		enableCSP = c
	}

	cspPolicy := "default-src 'self'"
	if p, ok := configMap["csp_policy"].(string); ok {
		cspPolicy = p
	}

	enableXFrameOptions := true
	if x, ok := configMap["enable_x_frame_options"].(bool); ok {
		enableXFrameOptions = x
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			if enableHSTS {
				ctx.Headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			if enableCSP {
				ctx.Headers.Set("Content-Security-Policy", cspPolicy)
			}

			if enableXFrameOptions {
				ctx.Headers.Set("X-Frame-Options", "DENY")
			}

			ctx.Headers.Set("X-Content-Type-Options", "nosniff")
			ctx.Headers.Set("X-XSS-Protection", "1; mode=block")
			ctx.Headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

			return next(ctx)
		}
	}
}

var _ lokstra.MiddlewareModule = (*SecurityMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
	return &SecurityMiddleware{}
}
