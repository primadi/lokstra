package meta

import (
	"github.com/primadi/lokstra/core/service"
)

type RpcServiceMeta struct {
	MethodParam string // default ":method"
	ServiceURI  string // lokstra://rpc_service/hello
	ServiceInst service.Service
}
