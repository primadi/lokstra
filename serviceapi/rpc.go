package serviceapi

import (
	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/core/request"
)

type RpcServer interface {
	RegisterRpcService(name string, impl any) error
	HandleRequest(ctx *request.Context, serviceName, MethodName string) error
}

type RpcClient interface {
	GetService(url string) (iface.Service, error)
}
