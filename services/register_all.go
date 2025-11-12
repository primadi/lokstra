package services

// This file provides convenient functions to register all built-in Lokstra services

import (
	// Core services
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/kvstore_redis"
	"github.com/primadi/lokstra/services/metrics_prometheus"
	"github.com/primadi/lokstra/services/redis"
)

// RegisterAllServices registers all built-in Lokstra service factories
// Note: Auth services have been moved to github.com/primadi/lokstra-auth
func RegisterAllServices() {
	// Core services
	redis.Register()
	kvstore_redis.Register()
	metrics_prometheus.Register()
	dbpool_pg.Register()
}

// RegisterCoreServices registers only core infrastructure services
func RegisterCoreServices() {
	redis.Register()
	kvstore_redis.Register()
	metrics_prometheus.Register()
	dbpool_pg.Register()
}
