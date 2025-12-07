package dbpool_manager

import (
	"context"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/syncmap"
)

type SyncPoolManager struct {
	pools       *sync.Map                    // map[dsn]serviceapi.DbPool
	aliasPools  *syncmap.SyncMap[*dsnSchema] // unified tenant and named pools
	newPoolFunc func(dsn string) (serviceapi.DbPool, error)
}

var _ serviceapi.DbPoolManager = (*SyncPoolManager)(nil)

func NewSyncPoolManager(
	aliasPools *syncmap.SyncMap[*dsnSchema],
	newPoolFunc func(dsn string) (serviceapi.DbPool, error),
) serviceapi.DbPoolManager {
	return &SyncPoolManager{
		pools:       &sync.Map{},
		aliasPools:  aliasPools,
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxSyncPoolManager(
	aliasPools *syncmap.SyncMap[*dsnSchema],
) serviceapi.DbPoolManager {
	return &SyncPoolManager{
		pools:      &sync.Map{},
		aliasPools: aliasPools,
		newPoolFunc: func(dsn string) (serviceapi.DbPool, error) {
			return dbpool_pg.NewPgxPostgresPool(context.Background(), dsn)
		},
	}
}

// AcquireNamedConn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
	return m.acquireAliasConn(ctx, "named:"+name, false)
}

// AcquireTenantConn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) AcquireTenantConn(ctx context.Context, tenant string) (serviceapi.DbConn, error) {
	return m.acquireAliasConn(ctx, "tenant:"+tenant, true)
}

// GetNamedDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) GetNamedDsn(name string) (string, string, error) {
	return m.getAliasDsn("named:" + name)
}

// GetNamedPool implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) GetNamedPool(name string) (serviceapi.DbPoolWithSchema, error) {
	dsn, schema, err := m.GetNamedDsn(name)
	if err != nil {
		return nil, err
	}
	dbPool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return dbpool_pg.NewDbPoolWithSchema(dbPool, schema), nil
}

// GetTenantDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) GetTenantDsn(tenant string) (string, string, error) {
	return m.getAliasDsn("tenant:" + tenant)
}

// GetTenantPool implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) GetTenantPool(tenant string) (serviceapi.DbPoolWithTenant, error) {
	dsn, schema, err := m.GetTenantDsn(tenant)
	if err != nil {
		return nil, err
	}
	dbPool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return dbpool_pg.NewDbPoolWithTenant(dbPool, schema, tenant), nil
}

// RemoveNamed implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) RemoveNamed(name string) {
	m.removeAlias("named:" + name)
}

// RemoveTenant implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) RemoveTenant(tenant string) {
	m.removeAlias("tenant:" + tenant)
}

// SetNamedDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) SetNamedDsn(name string, dsn string, schema string) {
	m.setAlias("named:"+name, dsn, schema)
}

// SetTenantDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) SetTenantDsn(tenant string, dsn string, schema string) {
	m.setAlias("tenant:"+tenant, dsn, schema)
}

func (m *SyncPoolManager) GetDsnPool(dsn string) (serviceapi.DbPool, error) {
	if pool, ok := m.pools.Load(dsn); ok {
		if ok {
			return pool.(serviceapi.DbPool), nil
		}
	}

	newPool, err := m.newPoolFunc(dsn)
	if err != nil {
		return nil, err
	}

	pool, _ := m.pools.LoadOrStore(dsn, newPool)
	return pool.(serviceapi.DbPool), nil
}

func (m *SyncPoolManager) Shutdown() error {
	m.pools.Range(func(key, value any) bool {
		pool := value.(serviceapi.DbPool)
		_ = pool.Shutdown()
		return true
	})

	return nil
}

// ========================================
// Internal helper methods
// ========================================

func (m *SyncPoolManager) setAlias(alias, dsn, schema string) {
	_ = m.aliasPools.Set(context.Background(), alias, &dsnSchema{Dsn: dsn, Schema: schema})
}

func (m *SyncPoolManager) getAliasDsn(alias string) (string, string, error) {
	ds, err := m.aliasPools.Get(context.Background(), alias)
	if err != nil {
		return "", "", err
	}
	return ds.Dsn, ds.Schema, nil
}

func (m *SyncPoolManager) removeAlias(alias string) {
	_ = m.aliasPools.Delete(context.Background(), alias)
}

func (m *SyncPoolManager) acquireAliasConn(ctx context.Context, alias string, isTenant bool) (serviceapi.DbConn, error) {
	dsn, schema, err := m.getAliasDsn(alias)
	if err != nil {
		return nil, err
	}
	pool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	if isTenant {
		// Extract tenant ID from alias (remove "tenant:" prefix)
		tenantID := alias[7:] // len("tenant:") = 7
		return pool.AcquireMultiTenant(ctx, schema, tenantID)
	}
	return pool.Acquire(ctx, schema)
}
