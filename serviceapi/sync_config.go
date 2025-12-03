package serviceapi

import "context"

// ConfigChangeCallback is called when a config value changes
type ConfigChangeCallback func(key string, value any)

// SyncConfig provides a synchronized key-value configuration store
// with real-time updates across multiple instances
type SyncConfig interface {
	// Set sets a configuration value and notifies all listeners
	Set(ctx context.Context, key string, value any) error

	// Get retrieves a configuration value
	Get(ctx context.Context, key string) (any, error)

	// Delete removes a configuration value and notifies all listeners
	Delete(ctx context.Context, key string) error

	// GetAll retrieves all configuration values
	GetAll(ctx context.Context) (map[string]any, error)

	// Subscribe registers a callback for configuration changes
	Subscribe(callback ConfigChangeCallback) string

	// Unsubscribe removes a callback by subscription ID
	Unsubscribe(subscriptionID string)

	// GetCRC returns the current CRC32 checksum of all config data
	GetCRC() uint32

	// Sync forces a synchronization with the backend store
	Sync(ctx context.Context) error

	Shutdownable
}
