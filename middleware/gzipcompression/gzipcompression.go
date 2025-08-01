package gzipcompression

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "gzipcompression"
const MIN_SIZE_KEY = "min_size"
const LEVEL_KEY = "level"
const DEFAULT_MIN_SIZE = 1024 // Minimum size in bytes to apply Gzip compression
const DEFAULT_LEVEL = 5       // Default compression level if not specified

type GzipCompressionMiddleware struct{}

// Description implements registration.Module.
func (g *GzipCompressionMiddleware) Description() string {
	return "Gzip Compression Middleware for Lokstra, should be the outermost middleware"
}

// Register implements registration.Module.
func (g *GzipCompressionMiddleware) Register(regCtx iface.RegistrationContext) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (g *GzipCompressionMiddleware) Name() string {
	return NAME
}

// Factory implements iface.MiddlewareModule.
func factory(config any) lokstra.MiddlewareFunc {
	minSize := DEFAULT_MIN_SIZE
	level := DEFAULT_LEVEL

	switch cfg := config.(type) {
	case map[string]any:
		if v, ok := cfg[MIN_SIZE_KEY]; ok {
			if size, ok := v.(int); ok && size > 0 {
				minSize = size
			}
		}
		if v, ok := cfg[LEVEL_KEY]; ok {
			if lvl, ok := v.(int); ok && lvl >= 0 && lvl <= 9 {
				level = lvl
			}
		}
	case []any:
		if len(cfg) >= 1 {
			if size, ok := cfg[0].(int); ok && size > 0 {
				minSize = size
			}
		}
		if len(cfg) >= 2 {
			if lvl, ok := cfg[1].(int); ok && lvl >= 0 && lvl <= 9 {
				level = lvl
			}
		}
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			// Skip if client does not accept gzip
			if !ctx.IsHeaderContainValue("Accept-Encoding", "gzip") {
				return next(ctx)
			}

			writer := &gzipResponseWriter{
				ResponseWriter: ctx.Writer,
				minSize:        minSize,
				level:          level,
				buffer:         make([]byte, 0, minSize),
			}

			ctx.Writer = writer

			err := next(ctx)
			_ = writer.Close()
			return err
		}
	}
}

var _ lokstra.Module = (*GzipCompressionMiddleware)(nil)

// return GzipCompressionMiddleware with name "lokstra.gzipcompression"
func GetModule() lokstra.Module {
	return &GzipCompressionMiddleware{}
}
