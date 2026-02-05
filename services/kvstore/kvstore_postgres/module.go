package kvstore_postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

var SERVICE_TYPE = "kvrepository_postgres"

var ErrKeyNotFound = errors.New("key not found")

type kvRepositoryPostgres struct {
	dbPool serviceapi.DbPool
	prefix string
}

func (k *kvRepositoryPostgres) prefixKey(key string) string {
	if k.prefix != "" {
		return k.prefix + ":" + key
	}
	return key
}

// Delete implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) Delete(ctx context.Context, key string) error {
	_, err := k.dbPool.Exec(ctx, "DELETE FROM kvrepository WHERE key = $1", k.prefixKey(key))
	return err
}

// DeleteKeys implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) DeleteKeys(ctx context.Context, keys ...string) error {
	_, err := k.dbPool.Exec(ctx, "DELETE FROM kvrepository WHERE key = ANY($1)", func() []string {
		prefixedKeys := make([]string, len(keys))
		for i, key := range keys {
			prefixedKeys[i] = k.prefixKey(key)
		}
		return prefixedKeys
	}())
	return err
}

// Get implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) Get(ctx context.Context, key string, dest any) error {
	err := k.dbPool.SelectMustOne(ctx,
		"SELECT value FROM kvrepository WHERE key = $1 AND (expiresat IS NULL OR expiresat > NOW())",
		[]any{k.prefixKey(key)}, &dest)
	if k.dbPool.IsErrorNoRows(err) {
		return ErrKeyNotFound
	}
	return err
}

// GetPrefix implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) GetPrefix() string {
	return k.prefix
}

// Keys implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) Keys(ctx context.Context, pattern string) ([]string, error) {
	pattern = strings.ReplaceAll(pattern, "*", "%")
	rows, err := k.dbPool.Query(ctx,
		"SELECT key FROM kvrepository WHERE key LIKE $1 AND (expiresat IS NULL OR expiresat > NOW())",
		k.prefixKey(pattern))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	startItem := len(k.prefix)
	if startItem > 0 {
		startItem++ // to account for the colon
	}

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key[startItem:])
	}
	return keys, nil
}

// Set implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	var expiresAt *time.Time
	if ttl > 0 {
		exp := time.Now().Add(ttl)
		expiresAt = &exp
	}
	res, err := k.dbPool.Exec(ctx, "UPDATE kvrepository SET value=$1, expiresAt=$2 WHERE key=$3",
		value, expiresAt, k.prefixKey(key))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		_, _ = k.dbPool.Exec(ctx, "INSERT INTO kvrepository (key, value, expiresAt) VALUES ($1, $2, $3)",
			k.prefixKey(key), value, expiresAt)
	}
	return nil
}

// SetPrefix implements [serviceapi.KvRepository].
func (k *kvRepositoryPostgres) SetPrefix(prefix string) {
	k.prefix = prefix
}

var _ serviceapi.KvRepository = (*kvRepositoryPostgres)(nil)

func Service(poolName, prefix string) *kvRepositoryPostgres {
	return &kvRepositoryPostgres{
		dbPool: lokstra_registry.GetService[serviceapi.DbPool](poolName),
		prefix: prefix,
	}
}

func ServiceFactory(config map[string]any) any {
	poolName := utils.GetValueFromMap(config, "pool_name", "db_main")
	prefix := utils.GetValueFromMap(config, "prefix", "")
	return Service(poolName, prefix)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
