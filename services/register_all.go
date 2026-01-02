package services

// This file provides convenient functions to register all built-in Lokstra services

import (
	// Core services
	"time"

	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/email_smtp"
	"github.com/primadi/lokstra/services/kvstore/kvstore_inmemory"
	"github.com/primadi/lokstra/services/kvstore/kvstore_redis"
	"github.com/primadi/lokstra/services/metrics_prometheus"
	"github.com/primadi/lokstra/services/sync_config_pg"
)

// RegisterAllServices registers all built-in Lokstra service factories
// Note: Auth services have been moved to github.com/primadi/lokstra-auth
func RegisterAllServices() {
	// Core services
	kvstore_redis.Register()
	kvstore_inmemory.Register()
	metrics_prometheus.Register()
	dbpool_pg.Register()
	email_smtp.Register()
	sync_config_pg.Register("db_main", 5*time.Minute, 5*time.Second)
}
