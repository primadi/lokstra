package dbpool_manager

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

// DbPoolInfo holds database DSN, schema name, and RLS context
type DbPoolInfo struct {
	Dsn        string
	Schema     string
	RlsContext map[string]string
}

type DbPoolManager struct {
	pools       map[string]serviceapi.DbPool // key: dsn
	namedPools  map[string]*DbPoolInfo       // key: name
	mu          sync.RWMutex
	newPoolFunc func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)
}

// AcquireConn implements serviceapi.DbPoolManager.
func (p *DbPoolManager) AcquireConn(ctx context.Context, dsn string, schema string, rlsContext map[string]string) (serviceapi.DbConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPool, ok := p.pools[dsn]
	if !ok {
		var err error
		dbPool, err = p.newPoolFunc(dsn, schema, rlsContext)
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
	p.mu.RLock()
	dbPoolInfo, ok := p.namedPools[name]
	p.mu.RUnlock()
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.AcquireConn(ctx, dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetDbPool implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetDbPool(dsn string, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPool, ok := p.pools[dsn]
	if ok {
		return dbPool, nil
	}
	newPool, err := p.newPoolFunc(dsn, schema, rlsContext)
	if err != nil {
		return nil, err
	}
	p.pools[dsn] = newPool
	return newPool, nil
}

// GetNamedDbPool implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetNamedDbPool(name string) (serviceapi.DbPool, error) {
	p.mu.RLock()
	dbPoolInfo, ok := p.namedPools[name]
	p.mu.RUnlock()
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.GetDbPool(dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetNamedDbPoolInfo implements serviceapi.DbPoolManager.
func (p *DbPoolManager) GetNamedDbPoolInfo(name string) (string, string, map[string]string, error) {
	p.mu.RLock()
	dbPoolInfo, ok := p.namedPools[name]
	p.mu.RUnlock()
	if !ok {
		return "", "", nil, errors.New("dbpool: named pool not found: " + name)
	}
	return dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext, nil
}

// RemoveNamedDbPool implements serviceapi.DbPoolManager.
func (p *DbPoolManager) RemoveNamedDbPool(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.namedPools, name)
}

// SetNamedDbPool implements serviceapi.DbPoolManager.
func (p *DbPoolManager) SetNamedDbPool(name string, dsn string, schema string, rlsContext map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.namedPools[name] = &DbPoolInfo{
		Dsn:        dsn,
		Schema:     schema,
		RlsContext: rlsContext,
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

var _ serviceapi.DbPoolManager = (*DbPoolManager)(nil)

func NewPoolManager(newPoolFunc func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)) serviceapi.DbPoolManager {
	return &DbPoolManager{
		pools:       make(map[string]serviceapi.DbPool),
		namedPools:  make(map[string]*DbPoolInfo),
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxPoolManager() serviceapi.DbPoolManager {
	return NewPoolManager(func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
		return dbpool_pg.NewPgxPostgresPool(dsn, schema, rlsContext)
	},
	)
}
