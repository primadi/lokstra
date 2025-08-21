package defaults

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/health_check"
	"github.com/primadi/lokstra/services/kvstore_mem"
	"github.com/primadi/lokstra/services/kvstore_redis"
	"github.com/primadi/lokstra/services/logger"
	"github.com/primadi/lokstra/services/metrics"
	"github.com/primadi/lokstra/services/redis"
)

func RegisterAllServices(regCtx iface.RegistrationContext) {
	// Database services
	regCtx.RegisterModule(dbpool_pg.GetModule)

	// Key-value store services
	regCtx.RegisterModule(kvstore_mem.GetModule)
	regCtx.RegisterModule(kvstore_redis.GetModule)

	// Observability services
	regCtx.RegisterModule(logger.GetModule)
	regCtx.RegisterModule(metrics.GetModule)
	regCtx.RegisterModule(health_check.GetModule)

	// Redis service
	regCtx.RegisterModule(redis.GetModule)
}
