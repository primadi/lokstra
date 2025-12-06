package application

import (
	"fmt"
	"log"
)

// Example with Init() method

// @Service name="cache-manager"
type CacheManager struct {
	// @InjectCfg key="cache.max-size", default=1000
	MaxSize int

	// @InjectCfg key="cache.ttl-seconds", default=300
	TTLSeconds int

	// Internal state (not injected)
	cache map[string]any
}

// Init is called after dependency injection
func (c *CacheManager) Init() error {
	// Initialize internal state
	c.cache = make(map[string]any, c.MaxSize)

	// Validation
	if c.MaxSize <= 0 {
		return fmt.Errorf("cache max size must be positive, got %d", c.MaxSize)
	}

	if c.TTLSeconds <= 0 {
		return fmt.Errorf("cache TTL must be positive, got %d", c.TTLSeconds)
	}

	log.Printf("✅ CacheManager initialized: max_size=%d, ttl=%ds", c.MaxSize, c.TTLSeconds)
	return nil
}

func (c *CacheManager) Set(key string, value any) {
	if c.cache == nil {
		c.cache = make(map[string]any)
	}
	c.cache[key] = value
}

func (c *CacheManager) Get(key string) (any, bool) {
	if c.cache == nil {
		return nil, false
	}
	val, ok := c.cache[key]
	return val, ok
}
