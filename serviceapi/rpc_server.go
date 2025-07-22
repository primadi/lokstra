package serviceapi

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

type RpcServer interface {
	HandleRequest(ctx *request.Context, service service.Service, MethodName string) error
}
