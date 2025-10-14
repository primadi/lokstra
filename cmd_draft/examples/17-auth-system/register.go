package main

import (
	// Core services
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/kvstore_redis"
	"github.com/primadi/lokstra/services/redis"

	// Auth services
	"github.com/primadi/lokstra/services/auth_flow_otp"
	"github.com/primadi/lokstra/services/auth_flow_password"
	"github.com/primadi/lokstra/services/auth_service"
	"github.com/primadi/lokstra/services/auth_session_redis"
	"github.com/primadi/lokstra/services/auth_token_jwt"
	"github.com/primadi/lokstra/services/auth_user_repo_pg"
	"github.com/primadi/lokstra/services/auth_validator"

	// Middleware
	"github.com/primadi/lokstra/middleware/accesscontrol"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/jwtauth"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

func registerServices() {
	// Core services
	redis.Register()
	kvstore_redis.Register()
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

func registerMiddleware() {
	cors.Register()
	recovery.Register()
	request_logger.Register()
	jwtauth.Register()
	accesscontrol.Register()
}
