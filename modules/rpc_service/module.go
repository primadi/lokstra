package rpc_service

import (
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "rpc_service"

type module struct{}

// Description implements registration.Module.
func (r *module) Description() string {
	return "RPC Service Module provides RPC service functionality"
}

// Name implements registration.Module.
func (r *module) Name() string {
	return NAME
}

// Register implements registration.Module.
func (r *module) Register(regCtx iface.RegistrationContext) error {
	// Register the RPC service factory
	regCtx.RegisterServiceFactory(r.Name(), NewRpcServer)

	return nil
}

var _ iface.Module = (*module)(nil)

func GetModule() iface.Module {
	return &module{}
}
