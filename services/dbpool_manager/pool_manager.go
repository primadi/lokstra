package dbpool_manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

type dsnSchema struct {
	dsn    string
	schema string
}

type PoolManager struct {
	pools       *sync.Map //map[dsn]serviceapi.DbPool
	tenantPools *sync.Map // map[tenant]dsn, schema
	namedPools  *sync.Map // map[name]dsn, schema
	newPoolFunc func(dsn string) (serviceapi.DbPool, error)
}

func NewPoolManager(newPoolFunc func(dsn string) (serviceapi.DbPool, error)) serviceapi.DbPoolManager {
	return &PoolManager{
		pools:       &sync.Map{},
		tenantPools: &sync.Map{},
		namedPools:  &sync.Map{},
		newPoolFunc: newPoolFunc,
	}
}

func NewPgxPoolManager() serviceapi.DbPoolManager {
	return &PoolManager{
		pools:       &sync.Map{},
		tenantPools: &sync.Map{},
		namedPools:  &sync.Map{},
		newPoolFunc: func(dsn string) (serviceapi.DbPool, error) {
			return dbpool_pg.NewPgxPostgresPool(context.Background(), dsn)
		},
	}
}

// AcquireNamedConn implements serviceapi.DbPoolManager.
func (m *PoolManager) AcquireNamedConn(ctx context.Context, name string) (serviceapi.DbConn, error) {
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
func (m *PoolManager) AcquireTenantConn(ctx context.Context, tenant string) (serviceapi.DbConn, error) {
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
func (m *PoolManager) GetNamedDsn(name string) (string, string, error) {
	_ds, ok := m.namedPools.Load(name)
	if !ok {
		return "", "", fmt.Errorf("named pool not found: %s", name)
	}
	ds := _ds.(dsnSchema)
	return ds.dsn, ds.schema, nil
}

// GetNamedPool implements serviceapi.DbPoolManager.
func (m *PoolManager) GetNamedPool(name string) (serviceapi.DbPoolWithSchema, error) {
	dsn, schema, err := m.GetNamedDsn(name)
	if err != nil {
		return nil, err
	}
	dbPool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return &pgxDbPoolWithSchema{
		pool:   dbPool,
		schema: schema,
	}, nil
}

// GetTenantDsn implements serviceapi.DbPoolManager.
func (m *PoolManager) GetTenantDsn(tenant string) (string, string, error) {
	_ds, ok := m.tenantPools.Load(tenant)
	if !ok {
		return "", "", fmt.Errorf("tenant pool not found: %s", tenant)
	}
	ds := _ds.(dsnSchema)
	return ds.dsn, ds.schema, nil
}

// GetTenantPool implements serviceapi.DbPoolManager.
func (m *PoolManager) GetTenantPool(tenant string) (serviceapi.DbPoolWithTenant, error) {
	dsn, schema, err := m.GetTenantDsn(tenant)
	if err != nil {
		return nil, err
	}
	dbPool, err := m.GetDsnPool(dsn)
	if err != nil {
		return nil, err
	}
	return &pgxDbPoolWithTenant{
		pool:     dbPool,
		schema:   schema,
		tenantID: tenant,
	}, nil
}

// RemoveNamed implements serviceapi.DbPoolManager.
func (m *PoolManager) RemoveNamed(name string) {
	m.namedPools.Delete(name)
}

// RemoveTenant implements serviceapi.DbPoolManager.
func (m *PoolManager) RemoveTenant(tenant string) {
	m.tenantPools.Delete(tenant)
}

// SetNamedDsn implements serviceapi.DbPoolManager.
func (m *PoolManager) SetNamedDsn(name string, dsn string, schema string) {
	m.namedPools.Store(name, dsnSchema{dsn: dsn, schema: schema})
}

// SetTenantDsn implements serviceapi.DbPoolManager.
func (m *PoolManager) SetTenantDsn(tenant string, dsn string, schema string) {
	m.tenantPools.Store(tenant, dsnSchema{dsn: dsn, schema: schema})
}

func (m *PoolManager) GetDsnPool(dsn string) (serviceapi.DbPool, error) {
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

func (m *PoolManager) Shutdown() error {
	m.pools.Range(func(key, value any) bool {
		pool := value.(serviceapi.DbPool)
		_ = pool.Shutdown()
		return true
	})

	return nil
}
