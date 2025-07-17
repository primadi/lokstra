package gzipcompression

import (
	"github.com/primadi/lokstra"
)

const NAME = "lokstra.gzipcompression"
const MIN_SIZE_KEY = "min_size"
const LEVEL_KEY = "level"
const DEFAULT_MIN_SIZE = 1024 // Minimum size in bytes to apply Gzip compression
const DEFAULT_LEVEL = 5       // Default compression level if not specified

type GzipCompressionMiddleware struct{}

// Name implements iface.MiddlewareModule.
func (g *GzipCompressionMiddleware) Name() string {
	return NAME
}

// Meta implements iface.MiddlewareModule.
func (g *GzipCompressionMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    20,
		Description: "Compress response using Gzip compression. Should be the outermost middleware.",
		Tags:        []string{"compression", "gzip"},
	}
}

// Factory implements iface.MiddlewareModule.
func (g *GzipCompressionMiddleware) Factory(config any) lokstra.MiddlewareFunc {
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

var _ lokstra.MiddlewareModule = (*GzipCompressionMiddleware)(nil)

// return GzipCompressionMiddleware with name "lokstra.gzipcompression"
func GetModule() lokstra.MiddlewareModule {
	return &GzipCompressionMiddleware{}
}
