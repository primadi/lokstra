package dbpool_crud

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

// DbPoolConfig represents database pool configuration
type DbPoolConfig struct {
	Name       string            `json:"name"`
	DSN        string            `json:"dsn"`
	Schema     string            `json:"schema"`
	RlsContext map[string]string `json:"rls_context,omitempty"`
}

var dbpm = service.LazyLoad[serviceapi.DbPoolManager]("dbpool-manager")

// AddDbPool adds a new database pool at runtime
// This will persist to sync_config database if using distributed sync
// Note: If pool already exists, it will be updated (upsert behavior)
func AddDbPool(config DbPoolConfig) error {
	if config.Name == "" {
		return fmt.Errorf("pool name is required")
	}
	if config.DSN == "" {
		return fmt.Errorf("DSN is required")
	}
	if config.Schema == "" {
		config.Schema = "public" // Default schema
	}

	dpm := dbpm.Get()
	if dpm == nil {
		return fmt.Errorf("dbpool-manager service not found")
	}

	// Set the pool configuration (upsert: works for both new and existing pools)
	// This also auto-registers the pool as a lazy service
	dpm.SetDbPoolManager(config.Name, config.DSN, config.Schema, config.RlsContext)

	// Validate configuration by attempting to get the pool
	_, err := dpm.GetDbPoolManager(config.Name)
	if err != nil {
		// Rollback on validation failure
		dpm.RemoveDbPoolManager(config.Name)
		return fmt.Errorf("failed to create pool '%s': %w", config.Name, err)
	}

	logger.LogInfo("✅ Added/Updated DB pool: %s (schema: %s)", config.Name, config.Schema)
	return nil
}

// UpdateDbPool updates an existing database pool configuration
// This is an alias for AddDbPool (both do upsert)
// Note: This will NOT close existing connections, new connections will use new config
func UpdateDbPool(config DbPoolConfig) error {
	return AddDbPool(config) // Same implementation - upsert behavior
}

// RemoveDbPool removes a database pool at runtime
// This will persist to sync_config database if using distributed sync
// The pool service will also be unregistered from the service registry
func RemoveDbPool(name string) error {
	if name == "" {
		return fmt.Errorf("pool name is required")
	}

	dpm := dbpm.Get()
	if dpm == nil {
		return fmt.Errorf("dbpool-manager service not found")
	}

	// Check if pool exists
	_, _, _, err := dpm.GetDbPoolManagerInfo(name)
	if err != nil {
		return fmt.Errorf("pool '%s' not found: %w", name, err)
	}

	// Remove from manager (also unregisters service automatically)
	dpm.RemoveDbPoolManager(name)

	logger.LogInfo("✅ Removed DB pool: %s", name)
	return nil
}

// GetDbPoolInfo retrieves database pool configuration
func GetDbPoolInfo(name string) (*DbPoolConfig, error) {
	if name == "" {
		return nil, fmt.Errorf("pool name is required")
	}

	dpm := dbpm.Get()
	if dpm == nil {
		return nil, fmt.Errorf("dbpool-manager service not found")
	}

	dsn, schema, rlsContext, err := dpm.GetDbPoolManagerInfo(name)
	if err != nil {
		return nil, fmt.Errorf("pool '%s' not found: %w", name, err)
	}

	return &DbPoolConfig{
		Name:       name,
		DSN:        dsn,
		Schema:     schema,
		RlsContext: rlsContext,
	}, nil
}

// ListDbPools returns all configured database pools
// Works with both regular and sync DbPoolManager
func ListDbPools() ([]string, error) {
	dpm := dbpm.Get()
	if dpm == nil {
		return nil, fmt.Errorf("dbpool-manager service not found")
	}

	// Use GetAllDbPoolManager from DbPoolManager interface
	allPools := dpm.GetAllDbPoolManager()
	if allPools == nil {
		return []string{}, nil
	}

	poolNames := make([]string, 0, len(allPools))
	for poolName := range allPools {
		poolNames = append(poolNames, poolName)
	}

	return poolNames, nil
}

// GetDbPool returns the actual DbPool instance for direct use
func GetDbPool(name string) (serviceapi.DbPool, error) {
	if name == "" {
		return nil, fmt.Errorf("pool name is required")
	}

	dpm := dbpm.Get()
	if dpm == nil {
		return nil, fmt.Errorf("dbpool-manager service not found")
	}

	return dpm.GetDbPoolManager(name)
}

// AcquireDbConn acquires a connection from a named pool
func AcquireDbConn(ctx context.Context, poolName string) (serviceapi.DbConn, error) {
	if poolName == "" {
		return nil, fmt.Errorf("pool name is required")
	}

	dpm := dbpm.Get()
	if dpm == nil {
		return nil, fmt.Errorf("dbpool-manager service not found")
	}

	return dpm.AcquireNamedConn(ctx, poolName)
}
