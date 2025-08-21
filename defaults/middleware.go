package defaults

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/middleware/body_limit"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/gzipcompression"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
)

func RegisterAllMiddleware(regCtx iface.RegistrationContext) {
	regCtx.RegisterModule(body_limit.GetModule)
	regCtx.RegisterModule(cors.GetModule)
	regCtx.RegisterModule(gzipcompression.GetModule)
	regCtx.RegisterModule(recovery.GetModule)
	regCtx.RegisterModule(request_logger.GetModule)
	regCtx.RegisterModule(slow_request_logger.GetModule)
}
