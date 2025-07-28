package rpc_service

import "github.com/primadi/lokstra/core/registration"

type RpcServiceModule struct{}

// Description implements registration.Module.
func (r *RpcServiceModule) Description() string {
	return "RPC Service Module provides RPC service functionality"
}

// Name implements registration.Module.
func (r *RpcServiceModule) Name() string {
	return "rpc_server"
}

// Register implements registration.Module.
func (r *RpcServiceModule) Register(regCtx registration.Context) error {
	// Register the RPC service factory
	regCtx.RegisterServiceFactory("rpc_service.rpc_server", NewRpcServer)

	return nil
}

var _ registration.Module = (*RpcServiceModule)(nil)

func GetModule() registration.Module {
	return &RpcServiceModule{}
}
