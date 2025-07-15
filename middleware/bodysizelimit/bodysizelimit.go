package bodysizelimit

import (
	"net/http"
	"lokstra"
)

const NAME = "lokstra.bodysizelimit"

type BodySizeLimitMiddleware struct{}

func (b *BodySizeLimitMiddleware) Name() string {
	return NAME
}

func (b *BodySizeLimitMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    35,
		Description: "Body size limit middleware to prevent large request payloads",
		Tags:        []string{"bodysize", "security", "performance"},
	}
}

func (b *BodySizeLimitMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	maxSizeMB := 10
	if m, ok := configMap["max_size_mb"].(int); ok {
		maxSizeMB = m
	}

	maxSize := int64(maxSizeMB * 1024 * 1024)

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			if ctx.Request.ContentLength > maxSize {
				return ctx.ErrorRequestEntityTooLarge("Request body too large")
			}

			ctx.Request.Body = http.MaxBytesReader(ctx.ResponseWriter, ctx.Request.Body, maxSize)

			return next(ctx)
		}
	}
}

var _ lokstra.MiddlewareModule = (*BodySizeLimitMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
	return &BodySizeLimitMiddleware{}
}
