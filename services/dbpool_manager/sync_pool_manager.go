package dbpool_manager

import (
	"context"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/syncmap"
)

type SyncPoolManager struct {
	pools       *sync.Map // map[dsn]serviceapi.DbPool
	tenantPools *syncmap.SyncMap[*dsnSchema]
	namedPools  *syncmap.SyncMap[*dsnSchema]
	newPoolFunc func(dsn string) (serviceapi.DbPool, error)
}

var _ serviceapi.DbPoolManager = (*SyncPoolManager)(nil)

func NewSyncPoolManager(
	tenantPools *syncmap.SyncMap[*dsnSchema],
	namedPools *syncmap.SyncMap[*dsnSchema],
	newPoolFunc func(dsn string) (serviceapi.DbPool, error),
) serviceapi.DbPoolManager {
	return &SyncPoolManager{
		pools:       &sync.Map{},
		tenantPools: tenantPools,
		namedPools:  namedPools,
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxSyncPoolManager(
	tenantPools *syncmap.SyncMap[*dsnSchema],
	namedPools *syncmap.SyncMap[*dsnSchema],
) serviceapi.DbPoolManager {
	return &SyncPoolManager{
		pools:       &sync.Map{},
		tenantPools: tenantPools,
		namedPools:  namedPools,
		newPoolFunc: func(dsn string) (serviceapi.DbPool, error) {
			return dbpool_pg.NewPgxPostgresPool(context.Background(), dsn)
		},
	}
}

// AcquireNamedConn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
	dsn, schema, err := m.GetNamedDsn(name)
	if err != nil {
		return nil, err
	}
	pool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return pool.Acquire(ctx, schema)
}

// AcquireTenantConn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) AcquireTenantConn(ctx context.Context, tenant string) (serviceapi.DbConn, error) {
	dsn, schema, err := m.GetTenantDsn(tenant)
	if err != nil {
		return nil, err
	}
	pool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return pool.AcquireMultiTenant(ctx, schema, tenant)
}

// GetNamedDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) GetNamedDsn(name string) (string, string, error) {
	ds, err := m.namedPools.Get(context.Background(), name)
	if err != nil {
		return "", "", err
	}
	return ds.Dsn, ds.Schema, nil
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
	ds, err := m.tenantPools.Get(context.Background(), tenant)
	if err != nil {
		return "", "", err
	}
	return ds.Dsn, ds.Schema, nil
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
	_ = m.namedPools.Delete(context.Background(), name)
}

// RemoveTenant implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) RemoveTenant(tenant string) {
	_ = m.tenantPools.Delete(context.Background(), tenant)
}

// SetNamedDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) SetNamedDsn(name string, dsn string, schema string) {
	_ = m.namedPools.Set(context.Background(), name, &dsnSchema{Dsn: dsn, Schema: schema})
}

// SetTenantDsn implements serviceapi.DbPoolManager.
func (m *SyncPoolManager) SetTenantDsn(tenant string, dsn string, schema string) {
	_ = m.tenantPools.Set(context.Background(), tenant, &dsnSchema{Dsn: dsn, Schema: schema})
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
