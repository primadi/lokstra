package gzipcompression

import (
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
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
func (g *GzipCompressionMiddleware) Register(regCtx registration.Context) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

type Config struct {
	MinSize int `json:"min_size" yaml:"min_size"`
	Level   int `json:"level" yaml:"level"`
}

// Name implements registration.Module.
func (g *GzipCompressionMiddleware) Name() string {
	return NAME
}

// Factory implements iface.MiddlewareModule.
func factory(config any) midware.Func {
	cfg := &Config{
		MinSize: DEFAULT_MIN_SIZE,
		Level:   DEFAULT_LEVEL,
	}

	if config == nil {
		config = cfg
	}

	switch c := config.(type) {
	case map[string]any:
		if v, ok := c[MIN_SIZE_KEY]; ok {
			if size, ok := v.(int); ok && size > 0 {
				cfg.MinSize = size
			}
		}
		if v, ok := c[LEVEL_KEY]; ok {
			if lvl, ok := v.(int); ok && lvl >= 0 && lvl <= 9 {
				cfg.Level = lvl
			}
		}
	case []any:
		if len(c) >= 1 {
			if size, ok := c[0].(int); ok && size > 0 {
				cfg.MinSize = size
			}
		}
		if len(c) >= 2 {
			if lvl, ok := c[1].(int); ok && lvl >= 0 && lvl <= 9 {
				cfg.Level = lvl
			}
		}
	case *Config:
		*cfg = *c
	case Config:
		cfg = &c
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			// Skip if client does not accept gzip
			if !ctx.IsHeaderContainValue("Accept-Encoding", "gzip") {
				return next(ctx)
			}

			writer := &gzipResponseWriter{
				ResponseWriter: ctx.Writer,
				minSize:        cfg.MinSize,
				level:          cfg.Level,
				buffer:         make([]byte, 0, cfg.MinSize),
			}

			ctx.Writer = writer

			err := next(ctx)
			_ = writer.Close()
			return err
		}
	}
}

var _ registration.Module = (*GzipCompressionMiddleware)(nil)

// return GzipCompressionMiddleware with name "lokstra.gzipcompression"
func GetModule() registration.Module {
	return &GzipCompressionMiddleware{}
}

// Preferred way to get gzipcompression middleware execution
func GetMidware(cfg *Config) *midware.Execution {
	return &midware.Execution{
		Name:         NAME,
		Config:       cfg,
		MiddlewareFn: factory(cfg),
		Priority:     20,
	}
}
