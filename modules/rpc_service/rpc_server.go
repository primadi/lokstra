package rpc_service

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

type RpcServerImpl struct{}

func NewRpcServer(_ any) (service.Service, error) {
	return &RpcServerImpl{}, nil
}

// HandleRequest implements serviceapi.RpcServer.
func (r *RpcServerImpl) HandleRequest(ctx *request.Context,
	svc service.Service, MethodName string) error {
	return HandleRpcRequest(ctx, svc, MethodName)
}

var _ service.Service = (*RpcServerImpl)(nil)
var _ serviceapi.RpcServer = (*RpcServerImpl)(nil)
