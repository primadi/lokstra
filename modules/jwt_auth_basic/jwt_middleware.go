package jwt_auth_basic

import (
	"lokstra"
	"strings"
)

const MIDDLEWARE_NAME = "lokstra.jwt_auth"

type JWTAuthMiddleware struct{}

func (j *JWTAuthMiddleware) Name() string {
	return MIDDLEWARE_NAME
}

func (j *JWTAuthMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    20,
		Description: "JWT authentication middleware for protecting routes",
		Tags:        []string{"auth", "jwt", "security"},
	}
}

func (j *JWTAuthMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	serviceName := "jwt_auth"
	if name, ok := configMap["service_name"].(string); ok {
		serviceName = name
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			authHeader := ctx.Headers.Get("Authorization")
			if authHeader == "" {
				return ctx.ErrorUnauthorized("Missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return ctx.ErrorUnauthorized("Invalid authorization header format")
			}

			service, err := ctx.GetService(serviceName)
			if err != nil {
				return ctx.ErrorInternal("JWT service not available")
			}

			jwtService, ok := service.(*JWTAuthService)
			if !ok {
				return ctx.ErrorInternal("Invalid JWT service type")
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				return ctx.ErrorUnauthorized("Invalid token")
			}

			ctx.Set("user_id", (*claims)["user_id"])
			ctx.Set("role_id", (*claims)["role_id"])
			ctx.Set("jwt_claims", claims)

			return next(ctx)
		}
	}
}

var _ lokstra.MiddlewareModule = (*JWTAuthMiddleware)(nil)

func GetMiddlewareModule() lokstra.MiddlewareModule {
	return &JWTAuthMiddleware{}
}
