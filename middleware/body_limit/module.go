package body_limit

import (
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

var _ iface.Module = (*BodyLimitModule)(nil)

// GetModule returns the body limit module
func GetModule() iface.Module {
	return &BodyLimitModule{}
}
