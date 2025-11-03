package main

import (
	"log"
	"sync"
	"time"

	"github.com/primadi/lokstra"
)

// ConfigStore simulates a dynamic configuration store
type ConfigStore struct {
	mu     sync.RWMutex
	values map[string]string
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		values: map[string]string{
			"feature_flag": "enabled",
			"max_requests": "100",
			"timeout":      "30s",
		},
	}
}

func (cs *ConfigStore) Get(key string) string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.values[key]
}

func (cs *ConfigStore) Set(key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.values[key] = value
	log.Printf("Config updated: %s = %s", key, value)
}

var configStore = NewConfigStore()

// Status handler
func StatusHandler() map[string]any {
	return map[string]any{
		"feature_flag": configStore.Get("feature_flag"),
		"max_requests": configStore.Get("max_requests"),
		"timeout":      configStore.Get("timeout"),
	}
}

// Update config handler
func UpdateConfigHandler() map[string]any {
	// Simulate config update
	configStore.Set("max_requests", "200")
	configStore.Set("feature_flag", "disabled")

	return map[string]any{
		"message": "Configuration updated",
		"new_values": map[string]string{
			"feature_flag": configStore.Get("feature_flag"),
			"max_requests": configStore.Get("max_requests"),
		},
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Dynamic Configuration Example</h1>
		<p>Configuration that can be updated at runtime</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - View current configuration</li>
			<li><a href="/update">Update</a> - Simulate configuration update</li>
		</ul>
		<p>Watch server logs to see configuration changes</p>
	</body>
	</html>
	`
}

func main() {
	// Start config watcher (simulates external config updates)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Simulate random config changes
			if configStore.Get("feature_flag") == "enabled" {
				configStore.Set("feature_flag", "disabled")
			} else {
				configStore.Set("feature_flag", "enabled")
			}
		}
	}()

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)
	router.POST("/update", UpdateConfigHandler)

	// Create and run app
	app := lokstra.NewApp("dynamic-config", ":3040", router)

	log.Println("Starting server on :3040")
	log.Println("Configuration will auto-update every 15 seconds")
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
