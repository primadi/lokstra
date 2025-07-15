package requestid

import (
	"crypto/rand"
	"encoding/hex"
	"lokstra"
)

const NAME = "lokstra.requestid"

type RequestIDMiddleware struct{}

func (r *RequestIDMiddleware) Name() string {
	return NAME
}

func (r *RequestIDMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    5,
		Description: "Request ID middleware for tracing and debugging requests",
		Tags:        []string{"requestid", "tracing", "debugging"},
	}
}

func (r *RequestIDMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	headerName := "X-Request-ID"
	if h, ok := configMap["header_name"].(string); ok {
		headerName = h
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			requestID := ctx.Headers.Get(headerName)
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx.Headers.Set(headerName, requestID)
			ctx.Set("request_id", requestID)

			return next(ctx)
		}
	}
}

func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

var _ lokstra.MiddlewareModule = (*RequestIDMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
	return &RequestIDMiddleware{}
}
