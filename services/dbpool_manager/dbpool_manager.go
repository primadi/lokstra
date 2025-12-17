package dbpool_manager

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

// Helper to get global registry (avoid circular import)
func getGlobalRegistry() *deploy.GlobalRegistry {
	return deploy.Global()
}

type DbPoolManager struct {
	pools       map[string]serviceapi.DbPool      // key: dsn
	namedPools  map[string]*serviceapi.DbPoolInfo // key: name
	mu          sync.RWMutex
	newPoolFunc func(poolName, dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)
}

// AcquireConn implements serviceapi.DbPoolManager.
func (p *DbPoolManager) AcquireConn(ctx context.Context, dsn, schema string, rlsContext map[string]string) (serviceapi.DbConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPool, ok := p.pools[dsn]
	if !ok {
		var err error
		// For dynamic pools (not named), use empty string for poolName
		dbPool, err = p.newPoolFunc("", dsn, schema, rlsContext)
		if err != nil {
			return nil, err
		}
		p.pools[dsn] = dbPool
	}

	dbPoolWithRls, ok := dbPool.(serviceapi.DbPoolSchemaRls)
	if !ok {
		return nil, errors.New("dbpool: pool does not support schema or RLS")
	}
	dbPoolWithRls.SetSchemaRls(schema, rlsContext)

	return dbPool.Acquire(ctx)
}

// AcquireNamedConn implements serviceapi.DbPoolManager.
func (p *DbPoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPoolInfo, ok := p.namedPools[name]
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}

	dbPool, ok := p.pools[dbPoolInfo.Dsn]
	if !ok {
		var err error
		// For dynamic pools (not named), use empty string for poolName
		dbPool, err = p.newPoolFunc(name, dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
		if err != nil {
			return nil, err
		}
		p.pools[dbPoolInfo.Dsn] = dbPool
	}

	dbPoolWithRls, ok := dbPool.(serviceapi.DbPoolSchemaRls)
	if !ok {
		return nil, errors.New("dbpool: pool does not support schema or RLS")
	}
	dbPoolWithRls.SetSchemaRls(dbPoolInfo.Schema, dbPoolInfo.RlsContext)

	return dbPool.Acquire(ctx)
}

// GetDbPool implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetDbPool(dsn string, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPool, ok := p.pools[dsn]
	if ok {
		return dbPool, nil
	}
	// For dynamic pools (not named), use empty poolName
	newPool, err := p.newPoolFunc("", dsn, schema, rlsContext)
	if err != nil {
		return nil, err
	}
	p.pools[dsn] = newPool
	return newPool, nil
}

// GetNamedDbPool gets or creates a pool with a specific name
func (p *DbPoolManager) GetNamedDbPool(poolName, dsn string, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPool, ok := p.pools[dsn]
	if ok {
		return dbPool, nil
	}
	// For named pools, pass the poolName for transaction tracking
	newPool, err := p.newPoolFunc(poolName, dsn, schema, rlsContext)
	if err != nil {
		return nil, err
	}
	p.pools[dsn] = newPool
	return newPool, nil
}

// GetDbPoolManager implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetDbPoolManager(name string) (serviceapi.DbPool, error) {
	p.mu.RLock()
	dbPoolInfo, ok := p.namedPools[name]
	p.mu.RUnlock()
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.GetNamedDbPool(name, dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetDbPoolManagerInfo implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetDbPoolManagerInfo(name string) (string, string, map[string]string, error) {
	p.mu.RLock()
	dbPoolInfo, ok := p.namedPools[name]
	p.mu.RUnlock()
	if !ok {
		return "", "", nil, errors.New("dbpool: named pool not found: " + name)
	}
	return dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext, nil
}

// RemoveDbPoolManager implements serviceapi.DbPoolManager.
func (p *DbPoolManager) RemoveDbPoolManager(name string) {
	p.mu.Lock()
	delete(p.namedPools, name)
	p.mu.Unlock()

	// Unregister service from registry
	if registry := getGlobalRegistry(); registry != nil {
		registry.UnregisterService(name)
	}
}

// SetDbPoolManager implements serviceapi.DbPoolManager.
func (p *DbPoolManager) SetDbPoolManager(name string, dsn string, schema string, rlsContext map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.namedPools[name] = &serviceapi.DbPoolInfo{
		Dsn:        dsn,
		Schema:     schema,
		RlsContext: rlsContext,
	}

	// Auto-register pool as a service (lazy, created on first access)
	// This makes pools accessible via lokstra_registry.GetService[DbPool](name)
	if registry := getGlobalRegistry(); registry != nil {
		registry.RegisterLazyService(name, func() any {
			pool, _ := p.GetDbPoolManager(name)
			return pool
		}, nil)
	}
}

// Shutdown implements serviceapi.DbPoolManager.
func (p *DbPoolManager) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		_ = pool.Shutdown()
	}

	return nil
}

func (p *DbPoolManager) GetAllDbPoolManager() map[string]*serviceapi.DbPoolInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make(map[string]*serviceapi.DbPoolInfo)
	for name, info := range p.namedPools {
		result[name] = info
	}
	return result
}

var _ serviceapi.DbPoolManager = (*DbPoolManager)(nil)

func NewPoolManager(newPoolFunc func(poolName, dsn, schema string,
	rlsContext map[string]string) (serviceapi.DbPool, error)) serviceapi.DbPoolManager {
	return &DbPoolManager{
		pools:       make(map[string]serviceapi.DbPool),
		namedPools:  make(map[string]*serviceapi.DbPoolInfo),
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxPoolManager() serviceapi.DbPoolManager {
	return NewPoolManager(func(poolName, dsn, schema string,
		rlsContext map[string]string) (serviceapi.DbPool, error) {
		return dbpool_pg.NewPgxPostgresPool(poolName, dsn, schema, rlsContext)
	},
	)
}
