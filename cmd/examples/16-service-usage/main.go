package main

import (
	"context"
	"fmt"
	"log"

	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
	"github.com/primadi/lokstra/services"
)

func main() {
	// Register all services
	services.RegisterAllServices()

	// Example 1: Create database pool
	dbPool := lokstra_registry.NewService[any](
		"my_db", "dbpool_pg",
		map[string]any{
			"host":     "localhost",
			"port":     5432,
			"database": "myapp",
			"username": "postgres",
			"password": "password",
		},
	)
	fmt.Println("DbPool created:", dbPool != nil)

	// Example 2: Create KvStore
	kvStore := lokstra_registry.NewService[any](
		"my_kvstore", "kvstore_redis",
		map[string]any{
			"addr":   "localhost:6379",
			"prefix": "myapp",
		},
	)
	fmt.Println("KvStore created:", kvStore != nil)

	// Example 3: Create metrics
	metrics := lokstra_registry.NewService[any](
		"my_metrics", "metrics_prometheus",
		map[string]any{
			"namespace": "myapp",
			"subsystem": "api",
		},
	)
	fmt.Println("Metrics created:", metrics != nil)

	// Example 4: Setup complete auth system
	setupAuthSystem()
}

func setupAuthSystem() {
	ctx := context.Background()

	// 1. Create Token Issuer
	lokstra_registry.NewService[any](
		"my_token_issuer", "auth_token_jwt",
		map[string]any{
			"secret_key": "my-super-secret-key-change-in-production",
			"issuer":     "myapp",
		},
	)

	// 2. Create Session Store
	lokstra_registry.NewService[any](
		"my_session", "auth_session_redis",
		map[string]any{
			"addr":   "localhost:6379",
			"prefix": "myapp_auth",
		},
	)

	// 3. Create User Repository (requires DB Pool first)
	lokstra_registry.NewService[any](
		"my_db", "dbpool_pg",
		map[string]any{
			"host":     "localhost",
			"port":     5432,
			"database": "myapp",
			"username": "postgres",
			"password": "password",
		},
	)

	lokstra_registry.NewService[any](
		"my_user_repo", "auth_user_repo_pg",
		map[string]any{
			"dbpool_service_name": "my_db",
			"schema":              "public",
			"table_name":          "users",
		},
	)

	// 4. Create Auth Flows
	lokstra_registry.NewService[any](
		"my_password_flow", "auth_flow_password",
		map[string]any{
			"user_repo_service_name": "my_user_repo",
		},
	)

	lokstra_registry.NewService[any](
		"my_kvstore", "kvstore_redis",
		map[string]any{
			"addr":   "localhost:6379",
			"prefix": "myapp",
		},
	)

	lokstra_registry.NewService[any](
		"my_otp_flow", "auth_flow_otp",
		map[string]any{
			"user_repo_service_name": "my_user_repo",
			"kvstore_service_name":   "my_kvstore",
			"otp_length":             6,
			"otp_ttl_seconds":        300,
		},
	)

	// 5. Create Main Auth Service
	authSvc := lokstra_registry.NewService[auth.Service](
		"my_auth", "auth_service",
		map[string]any{
			"token_issuer_service_name": "my_token_issuer",
			"session_service_name":      "my_session",
			"flow_service_names": map[string]string{
				"password": "my_password_flow",
				"otp":      "my_otp_flow",
			},
		},
	)

	// 6. Create Auth Validator
	lokstra_registry.NewService[any](
		"my_validator", "auth_validator",
		map[string]any{
			"token_issuer_service_name": "my_token_issuer",
		},
	)

	// Example usage: Login with password
	resp, err := authSvc.Login(ctx, auth.LoginRequest{
		Flow: "password",
		Payload: map[string]any{
			"tenant_id": "tenant-123",
			"username":  "user@example.com",
			"password":  "password123",
		},
	})

	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		log.Printf("Login successful! Access token: %s...", resp.AccessToken[:20])
	}
}
