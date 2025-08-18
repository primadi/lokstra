package tenant_dbpool

import (
	"context"
	"sync"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

var tenantDSN = map[string]string{}            // map of tenantID to DSN
var dsnDbPool = map[string]serviceapi.DbPool{} // map of DSN to DbPool
var RWMutex = &sync.RWMutex{}

func RegisterTenantDSN(tenantID, dsn string) {
	RWMutex.RLock()
	tenantDSN[tenantID] = dsn
	RWMutex.RUnlock()
}

func GetTenantDSN(tenantID string) (string, bool) {
	RWMutex.RLock()
	dsn, ok := tenantDSN[tenantID]
	RWMutex.RUnlock()
	return dsn, ok
}

func GetTenantDbPool(tenantID string) (serviceapi.DbPool, bool) {
	RWMutex.RLock()
	if pool, ok := dsnDbPool[tenantID]; ok {
		RWMutex.RUnlock()
		return pool, true
	}

	dsn, ok := tenantDSN[tenantID]
	RWMutex.RUnlock()
	if !ok {
		return nil, false
	}

	p, err := dbpool_pg.NewPgxPostgresPool(context.Background(), dsn)
	if err != nil {
		return nil, false
	}

	RWMutex.Lock()
	defer RWMutex.Unlock()

	// Check if the pool already added by another goroutine
	if existing, exists := dsnDbPool[tenantID]; exists {
		return existing, true
	}

	dsnDbPool[tenantID] = p
	return p, ok
}
