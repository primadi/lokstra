package auth_module

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

type KVSessionStore struct {
	kv serviceapi.KvStore
}

// Delete implements auth.Session.
func (k *KVSessionStore) Delete(ctx context.Context, sessionID string) error {
	pattern := fmt.Sprintf("session:%s:*", sessionID)
	keys, _ := k.kv.Keys(ctx, pattern)
	if len(keys) == 0 {
		return nil
	}
	return k.kv.Delete(ctx, keys[0])
}

// DeleteByUser implements auth.Session.
func (k *KVSessionStore) DeleteByUser(ctx context.Context, tenantID, userID string) error {
	pattern := fmt.Sprintf("session:*:%s:%s", tenantID, userID)
	keys, _ := k.kv.Keys(ctx, pattern)
	return k.kv.DeleteKeys(ctx, keys...)
}

// Get implements auth.Session.
func (k *KVSessionStore) Get(ctx context.Context, sessionID string) (*auth.SessionData, error) {
	pattern := fmt.Sprintf("session:%s:*", sessionID)
	keys, err := k.kv.Keys(ctx, pattern)
	if err != nil || len(keys) == 0 {
		return nil, err
	}
	var data auth.SessionData
	if err = k.kv.Get(ctx, keys[0], &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ListByUser implements auth.Session.
func (k *KVSessionStore) ListByUser(ctx context.Context, tenantID, userID string) ([]string, error) {
	pattern := fmt.Sprintf("session:*:%s:%s", tenantID, userID)
	return k.kv.Keys(ctx, pattern)
}

// Set implements auth.Session.
func (k *KVSessionStore) Set(ctx context.Context, sessionID string, data *auth.SessionData, ttl time.Duration) error {
	key := sessionKey(sessionID, data.TenantID, data.UserID)
	return k.kv.Set(ctx, key, data, ttl)
}

func NewSession(kv serviceapi.KvStore) *KVSessionStore {
	return &KVSessionStore{kv: kv}
}

var _ auth.Session = (*KVSessionStore)(nil)

func sessionKey(sessionID, tenantID, userID string) string {
	return fmt.Sprintf("session:%s:%s:%s", sessionID, tenantID, userID)
}
