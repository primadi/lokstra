package body_limit

import (
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
)

const MODULE_NAME = "body_limit"

type BodyLimitModule struct{}

// Description implements registration.Module.
func (b *BodyLimitModule) Description() string {
	return "Body limit middleware"
}

// Name implements registration.Module.
func (b *BodyLimitModule) Name() string {
	return MODULE_NAME
}

// Register implements registration.Module.
func (b *BodyLimitModule) Register(regCtx registration.Context) error {
	return regCtx.RegisterMiddlewareFactory(MODULE_NAME, factory)
}

var _ registration.Module = (*BodyLimitModule)(nil)

// GetModule returns the body limit module
func GetModule() registration.Module {
	return &BodyLimitModule{}
}

// Preferred way to get body limit middleware execution
func GetMidware(cfg *Config) *midware.Execution {
	if cfg == nil {
		cfg = &Config{
			MaxSize: 10 * 1024 * 1024, // Default to 10MB
		}
	}
	return &midware.Execution{
		Name:         MODULE_NAME,
		Config:       cfg,
		MiddlewareFn: BodyLimitMiddleware(cfg),
		Priority:     25,
	}
}
