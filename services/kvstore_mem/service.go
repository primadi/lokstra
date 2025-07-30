package kvstore_mem

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

type memoryEntry struct {
	value     []byte
	expiresAt time.Time
}

type Service struct {
	data map[string]*memoryEntry
	mu   sync.RWMutex
	stop chan struct{}
}

// Delete implements serviceapi.KvStore.
func (s *Service) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// DeleteKeys implements serviceapi.KvStore.
func (s *Service) DeleteKeys(ctx context.Context, keys ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.data, key)
	}
	return nil
}

// Get implements serviceapi.KvStore.
func (s *Service) Get(ctx context.Context, key string, dest any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok || (!entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)) {
		return nil
	}
	return json.Unmarshal(entry.value, dest)
}

// Keys implements serviceapi.KvStore.
func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prefix := strings.TrimRight(pattern, "*")
	var keys []string
	for k, v := range s.data {
		if !v.expiresAt.IsZero() && time.Now().After(v.expiresAt) {
			continue
		}
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// Set implements serviceapi.KvStore.
func (s *Service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &memoryEntry{
		value:     data,
		expiresAt: expiresAt,
	}
	return nil
}

func (s *Service) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *Service) cleanupExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, entry := range s.data {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(s.data, key)
		}
	}
}

func (s *Service) StopCleanup() {
	close(s.stop)
}

func New() *Service {
	return &Service{
		data: make(map[string]*memoryEntry),
	}
}

var _ serviceapi.KvStore = (*Service)(nil)
