package lokstra

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_manager"
)

// auto create dbpool-manager service if not exists
func UsePgxDbPoolManager(useSync bool) {
	pm := lokstra_registry.GetService[serviceapi.DbPoolManager]("dbpool-manager")
	if pm != nil {
		return // Already registered
	}

	if useSync {
		pm = dbpool_manager.NewPgxSyncDbPoolManager()
		logger.LogDebug("[Lokstra] DbPoolManager initialized with distributed sync")
	} else {
		// Default: use regular pool manager (local sync.Map)
		pm = dbpool_manager.NewPgxPoolManager()
		logger.LogDebug("[Lokstra] DbPoolManager initialized with local sync")
	}

	lokstra_registry.RegisterService("dbpool-manager", pm)
}
