package dbpool_manager

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra/common/syncmap"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

type SyncDbPoolManager struct {
	pools        map[string]serviceapi.DbPool             // key: dsn
	namedPools   *syncmap.SyncMap[*serviceapi.DbPoolInfo] // key: name
	syncMapName  string                                   // lazy init: sync map name
	mu           sync.RWMutex
	newPoolFunc  func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error)
	syncMapMu    sync.Mutex // lazy init mutex
	syncMapReady bool       // lazy init flag
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

// ensureSyncMapInitialized initializes the SyncMap lazily on first use
func (p *SyncDbPoolManager) ensureSyncMapInitialized() {
	p.syncMapMu.Lock()
	defer p.syncMapMu.Unlock()

	if p.syncMapReady {
		return
	}

	// Create SyncMap now that sync-config service should be available
	p.namedPools = syncmap.NewSyncMap[*serviceapi.DbPoolInfo](p.syncMapName)
	p.syncMapReady = true
}

// AcquireNamedConn implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
	p.ensureSyncMapInitialized()
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

// GetDbPoolManager implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) GetDbPoolManager(name string) (serviceapi.DbPool, error) {
	p.ensureSyncMapInitialized()
	dbPoolInfo, ok := p.namedPools.Load(name)
	if !ok {
		return nil, errors.New("dbpool: named pool not found: " + name)
	}
	return p.GetDbPool(dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext)
}

// GetDbPoolManagerInfo implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) GetDbPoolManagerInfo(name string) (string, string, map[string]string, error) {
	p.ensureSyncMapInitialized()
	dbPoolInfo, ok := p.namedPools.Load(name)
	if !ok {
		return "", "", nil, errors.New("dbpool: named pool not found: " + name)
	}
	return dbPoolInfo.Dsn, dbPoolInfo.Schema, dbPoolInfo.RlsContext, nil
}

// RemoveDbPoolManager implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) RemoveDbPoolManager(name string) {
	p.ensureSyncMapInitialized()
	p.namedPools.Delete(context.Background(), name)

	// Unregister service from registry
	if registry := getGlobalRegistry(); registry != nil {
		registry.UnregisterService(name)
	}
}

// SetDbPoolManager implements serviceapi.DbPoolManager.
func (p *SyncDbPoolManager) SetDbPoolManager(name string, dsn string, schema string, rlsContext map[string]string) {
	p.ensureSyncMapInitialized()
	p.namedPools.Store(name, &serviceapi.DbPoolInfo{
		Dsn:        dsn,
		Schema:     schema,
		RlsContext: rlsContext,
	})

	// Auto-register pool as a service (lazy, created on first access)
	// This makes pools accessible via lokstra_registry.GetService[DbPool](name)
	if registry := getGlobalRegistry(); registry != nil {
		registry.RegisterLazyService(name, func() any {
			pool, _ := p.GetDbPoolManager(name)
			return pool
		}, nil)
	}
}

func (p *SyncDbPoolManager) GetAllDbPoolManager() map[string]*serviceapi.DbPoolInfo {
	p.ensureSyncMapInitialized()
	all, err := p.namedPools.All(context.Background())
	if err != nil {
		return nil
	}
	return all
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
		syncMapName: syncName,
		newPoolFunc: newPoolFunc,
		// namedPools will be lazily initialized on first use
	}
}

func NewPgxSyncDbPoolManager() serviceapi.DbPoolManager {
	return NewSyncDbPoolManager("dbpool",
		func(dsn, schema string, rlsContext map[string]string) (serviceapi.DbPool, error) {
			return dbpool_pg.NewPgxPostgresPool(dsn, schema, rlsContext)
		})
}
