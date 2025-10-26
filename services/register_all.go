package services

// This file provides convenient functions to register all built-in Lokstra services

import (
	// Core services
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/kvstore_redis"
	"github.com/primadi/lokstra/services/metrics_prometheus"
	"github.com/primadi/lokstra/services/redis"

	// Auth services
	"github.com/primadi/lokstra/services/auth_flow_otp"
	"github.com/primadi/lokstra/services/auth_flow_password"
	"github.com/primadi/lokstra/services/auth_service"
	"github.com/primadi/lokstra/services/auth_session_redis"
	"github.com/primadi/lokstra/services/auth_token_jwt"
	"github.com/primadi/lokstra/services/auth_user_repo_pg"
	"github.com/primadi/lokstra/services/auth_validator"
)

// RegisterAllServices registers all built-in Lokstra service factories
func RegisterAllServices() {
	// Core services
	redis.Register()
	kvstore_redis.Register()
	metrics_prometheus.Register()
	dbpool_pg.Register()

	// Auth services
	auth_session_redis.Register()
	auth_token_jwt.Register()
	auth_user_repo_pg.Register()
	auth_flow_password.Register()
	auth_flow_otp.Register()
	auth_service.Register()
	auth_validator.Register()
}

// RegisterCoreServices registers only core infrastructure services
func RegisterCoreServices() {
	redis.Register()
	kvstore_redis.Register()
	metrics_prometheus.Register()
	dbpool_pg.Register()
}

// RegisterAuthServices registers only authentication-related services
func RegisterAuthServices() {
	auth_session_redis.Register()
	auth_token_jwt.Register()
	auth_user_repo_pg.Register()
	auth_flow_password.Register()
	auth_flow_otp.Register()
	auth_service.Register()
	auth_validator.Register()
}
