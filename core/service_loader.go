package core

import (
	"fmt"
	"lokstra/internal"
)

type ServiceConfig struct {
	Type    string         `yaml:"type"`    // Type of the service (e.g., "db", "redis")
	Name    string         `yaml:"name"`    // Named instance of the service (e.g., "default", "cache")
	Enabled string         `yaml:"enabled"` // Optional: "true" or "false", supports env/template (default is "true")
	Config  map[string]any `yaml:"config"`  // Configuration specific to the service
}

// Example services.yaml:
//
// services:
//   - type: db
//     name: default
//     enabled: true           # Optional, default is true
//     config:
//       dsn: postgres://localhost/maindb
//
//   - type: db
//     name: analytics
//     config:
//       dsn: postgres://localhost/analytics
//
//   - type: redis
//     name: cache
//     config:
//       addr: localhost:6379
//       db: 0

// startServicesFromConfig starts all services defined in the parsed YAML.
// It uses StartService(...) internally, and skips services marked as disabled.
// Returns an error if any service fails to start.
func startServicesFromConfig(data []ServiceConfig) error {
	for _, entry := range data {
		if !internal.IsEnabled(entry.Enabled, true) {
			continue
		}
		if err := StartServiceFromConfig(&entry); err != nil {
			return fmt.Errorf("failed to start service %s:%s: %w", entry.Type, entry.Name, err)
		}
	}
	return nil
}
