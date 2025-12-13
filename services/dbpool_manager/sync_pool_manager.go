package dbpool_manager

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/syncmap"
)

type SyncDbPoolManager struct {
	pools       map[string]serviceapi.DbPool  // key: dsn
	namedPools  *syncmap.SyncMap[*DbPoolInfo] // key: name
	mu          sync.RWMutex
	newPoolFunc func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)
}

// AcquireConn implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) AcquireConn(ctx context.Context, dsn string, schema string, rlsContext map[string]string) (serviceapi.DbConn, error) {
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
func (p *SyncDbPoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
	dbPoolInfo, ok := p.namedPools.Load(name)
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.AcquireConn(ctx, dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetDbPool implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) GetDbPool(dsn string, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
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
func (p *SyncDbPoolManager) GetNamedDbPool(name string) (serviceapi.DbPool, error) {
	dbPoolInfo, ok := p.namedPools.Load(name)
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.GetDbPool(dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetNamedDbPoolInfo implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) GetNamedDbPoolInfo(name string) (string, string, map[string]string, error) {
	dbPoolInfo, ok := p.namedPools.Load(name)
	if !ok {
		return "", "", nil, errors.New("dbpool: named pool not found: " + name)
	}
	return dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext, nil
}

// RemoveNamedDbPool implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) RemoveNamedDbPool(name string) {
	p.namedPools.Delete(context.Background(), name)
}

// SetNamedDbPool implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) SetNamedDbPool(name string, dsn string, schema string, rlsContext map[string]string) {
	p.namedPools.Store(name, &DbPoolInfo{
		Dsn:        dsn,
		Schema:     schema,
		RlsContext: rlsContext,
	})
}

// Shutdown implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		_ = pool.Shutdown()
	}

	return nil
}

var _ serviceapi.DbPoolManager = (*SyncDbPoolManager)(nil)

func NewSyncDbPoolManager(syncName string, newPoolFunc func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)) serviceapi.DbPoolManager {
	return &SyncDbPoolManager{
		pools:       make(map[string]serviceapi.DbPool),
		namedPools:  syncmap.NewSyncMap[*DbPoolInfo](syncName),
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxSyncDbPoolManager() serviceapi.DbPoolManager {
	return NewSyncDbPoolManager("dbpool",
		func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
			return dbpool_pg.NewPgxPostgresPool(dsn, schema, rlsContext)
		})
}
