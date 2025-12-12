package lokstra

import "github.com/primadi/lokstra/services/sync_config_pg"

func UsePgxSyncConfig(syncDbPoolName string) {
	// Register SyncConfigPG service type (SyncMap based)
	sync_config_pg.Register(syncDbPoolName)
}
