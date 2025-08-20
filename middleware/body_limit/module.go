package body_limit

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const MODULE_NAME = "body_limit"

type BodyLimitModule struct{}

// Description implements iface.Module.
func (b *BodyLimitModule) Description() string {
	return "Body limit middleware"
}

// Name implements iface.Module.
func (b *BodyLimitModule) Name() string {
	return MODULE_NAME
}

// Register implements iface.Module.
func (b *BodyLimitModule) Register(regCtx iface.RegistrationContext) error {
	return regCtx.RegisterMiddlewareFactory(MODULE_NAME, factory)
}

var _ lokstra.Module = (*BodyLimitModule)(nil)

// GetModule returns the body limit module
func GetModule() lokstra.Module {
	return &BodyLimitModule{}
}
