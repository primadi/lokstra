package lokstra_registry_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Test struct for nested config
type DBConfig struct {
	DSN    string `json:"dsn"`
	Schema string `json:"schema"`
}

// Test struct with pointer
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func TestGetConfig_FlatAccess(t *testing.T) {
	// Setup
	registry := deploy.Global()
	registry.SetConfig("global-db.dsn", "postgres://localhost/test")
	registry.SetConfig("global-db.schema", "public")
	// Test flat access
	dsn := lokstra_registry.GetConfig("global-db.dsn", "default")
	if dsn != "postgres://localhost/test" {
		t.Errorf("Expected 'postgres://localhost/test', got '%s'", dsn)
	}

	schema := lokstra_registry.GetConfig("global-db.schema", "default")
	if schema != "public" {
		t.Errorf("Expected 'public', got '%s'", schema)
	}
}

func TestGetConfig_NestedMap(t *testing.T) {
	// Setup
	registry := deploy.Global()
	registry.SetConfig("app-db.dsn", "postgres://localhost/test")
	registry.SetConfig("app-db.schema", "public")
	// Test nested access as map
	dbConfig := lokstra_registry.GetConfig[map[string]any]("app-db", nil)
	if dbConfig == nil {
		t.Fatal("Expected map, got nil")
	}

	if dbConfig["dsn"] != "postgres://localhost/test" {
		t.Errorf("Expected 'postgres://localhost/test', got '%v'", dbConfig["dsn"])
	}

	if dbConfig["schema"] != "public" {
		t.Errorf("Expected 'public', got '%v'", dbConfig["schema"])
	}
}

func TestGetConfig_StructBinding(t *testing.T) {
	// Setup
	registry := deploy.Global()
	registry.SetConfig("database.dsn", "postgres://localhost/mydb")
	registry.SetConfig("database.schema", "app")
	// Test struct binding
	config := lokstra_registry.GetConfig("database", DBConfig{})
	if config.DSN != "postgres://localhost/mydb" {
		t.Errorf("Expected 'postgres://localhost/mydb', got '%s'", config.DSN)
	}
	if config.Schema != "app" {
		t.Errorf("Expected 'app', got '%s'", config.Schema)
	}
}

func TestGetConfig_StructBindingWithPointer(t *testing.T) {
	// Setup
	registry := deploy.Global()
	registry.SetConfig("server.host", "localhost")
	registry.SetConfig("server.port", 8080)

	// Test struct pointer binding
	config := lokstra_registry.GetConfig[*ServerConfig]("server", nil)
	if config == nil {
		t.Fatal("Expected *ServerConfig, got nil")
	}
	if config.Host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", config.Host)
	}
	if config.Port != 8080 {
		t.Errorf("Expected 8080, got %d", config.Port)
	}
}

func TestGetConfig_DefaultValue(t *testing.T) {
	// Test non-existent config returns default
	dsn := lokstra_registry.GetConfig("nonexistent.dsn", "default-value")
	if dsn != "default-value" {
		t.Errorf("Expected 'default-value', got '%s'", dsn)
	}

	config := lokstra_registry.GetConfig("nonexistent", DBConfig{DSN: "default", Schema: "public"})
	if config.DSN != "default" {
		t.Errorf("Expected default struct, got %+v", config)
	}
}

func TestSetConfig_RuntimeValues(t *testing.T) {
	// Test setting runtime config
	lokstra_registry.SetConfig("runtime.mode", "dev")
	lokstra_registry.SetConfig("computed.license-key", "ABC123XYZ")

	// Test retrieval
	mode := lokstra_registry.GetConfig("runtime.mode", "prod")
	if mode != "dev" {
		t.Errorf("Expected runtime.mode='dev', got '%s'", mode)
	}

	key := lokstra_registry.GetConfig("computed.license-key", "")
	if key != "ABC123XYZ" {
		t.Errorf("Expected license-key='ABC123XYZ', got '%s'", key)
	}
}

func TestSetConfig_CaseInsensitive(t *testing.T) {
	// Set with lowercase
	lokstra_registry.SetConfig("my.config.value", "test123")

	// Get with different cases
	val1 := lokstra_registry.GetConfig("my.config.value", "")
	val2 := lokstra_registry.GetConfig("My.Config.Value", "")
	val3 := lokstra_registry.GetConfig("MY.CONFIG.VALUE", "")

	if val1 != "test123" || val2 != "test123" || val3 != "test123" {
		t.Errorf("Case-insensitive access failed: %s, %s, %s", val1, val2, val3)
	}
}

func TestSetConfig_ComplexTypes(t *testing.T) {
	// Test with map
	configMap := map[string]any{
		"host": "localhost",
		"port": 5432,
	}
	lokstra_registry.SetConfig("database.settings", configMap)

	// Retrieve as map
	retrieved := lokstra_registry.GetConfig[map[string]any]("database.settings", nil)
	if retrieved == nil {
		t.Fatal("Expected map, got nil")
	}

	if retrieved["host"] != "localhost" {
		t.Errorf("Expected host='localhost', got '%v'", retrieved["host"])
	}
}

func TestSetConfig_Overwrite(t *testing.T) {
	// Set initial value
	lokstra_registry.SetConfig("test.overwrite", "initial")
	val1 := lokstra_registry.GetConfig("test.overwrite", "")

	// Overwrite
	lokstra_registry.SetConfig("test.overwrite", "updated")
	val2 := lokstra_registry.GetConfig("test.overwrite", "")

	if val1 != "initial" {
		t.Errorf("Expected initial value='initial', got '%s'", val1)
	}
	if val2 != "updated" {
		t.Errorf("Expected updated value='updated', got '%s'", val2)
	}
}

func TestSetConfig_LeafValue(t *testing.T) {
	// Test setting leaf values (non-map) with dot notation
	lokstra_registry.SetConfig("db.host", "localhost")
	lokstra_registry.SetConfig("db.port", 5432)
	lokstra_registry.SetConfig("db.schema", "public")

	// Verify individual access
	host := lokstra_registry.GetConfig("db.host", "")
	port := lokstra_registry.GetConfig("db.port", 0)
	schema := lokstra_registry.GetConfig("db.schema", "")

	if host != "localhost" {
		t.Errorf("Expected host='localhost', got '%s'", host)
	}
	if port != 5432 {
		t.Errorf("Expected port=5432, got %d", port)
	}
	if schema != "public" {
		t.Errorf("Expected schema='public', got '%s'", schema)
	}

	// Verify nested access returns reconstructed map
	dbConfig := lokstra_registry.GetConfig[map[string]any]("db", nil)
	if dbConfig == nil {
		t.Fatal("Expected map, got nil")
	}
	if dbConfig["host"] != "localhost" {
		t.Errorf("Expected nested host='localhost', got '%v'", dbConfig["host"])
	}
	if dbConfig["port"] != 5432 {
		t.Errorf("Expected nested port=5432, got '%v'", dbConfig["port"])
	}
}

func TestSetConfig_MapWithCleanup(t *testing.T) {
	registry := deploy.Global()

	// First: Set complex nested structure
	lokstra_registry.SetConfig("db", map[string]any{
		"host": "localhost",
		"port": 5432,
		"pool": map[string]any{
			"min": 2,
			"max": 10,
		},
	})

	// Verify all keys exist
	if _, ok := registry.GetConfig("db"); !ok {
		t.Error("Expected 'db' to exist")
	}
	if _, ok := registry.GetConfig("db.host"); !ok {
		t.Error("Expected 'db.host' to exist")
	}
	if _, ok := registry.GetConfig("db.port"); !ok {
		t.Error("Expected 'db.port' to exist")
	}
	if _, ok := registry.GetConfig("db.pool"); !ok {
		t.Error("Expected 'db.pool' to exist")
	}
	if _, ok := registry.GetConfig("db.pool.min"); !ok {
		t.Error("Expected 'db.pool.min' to exist")
	}
	if _, ok := registry.GetConfig("db.pool.max"); !ok {
		t.Error("Expected 'db.pool.max' to exist")
	}

	// Second: Overwrite with simpler structure (should delete stale keys)
	lokstra_registry.SetConfig("db", map[string]any{
		"host": "newhost",
	})

	// Verify only new keys exist
	host := lokstra_registry.GetConfig("db.host", "")
	if host != "newhost" {
		t.Errorf("Expected host='newhost', got '%s'", host)
	}

	// Verify stale keys are deleted
	if _, ok := registry.GetConfig("db.port"); ok {
		t.Error("Expected 'db.port' to be deleted (stale)")
	}
	if _, ok := registry.GetConfig("db.pool"); ok {
		t.Error("Expected 'db.pool' to be deleted (stale)")
	}
	if _, ok := registry.GetConfig("db.pool.min"); ok {
		t.Error("Expected 'db.pool.min' to be deleted (stale)")
	}
	if _, ok := registry.GetConfig("db.pool.max"); ok {
		t.Error("Expected 'db.pool.max' to be deleted (stale)")
	}

	// Verify nested access only returns new structure
	dbConfig := lokstra_registry.GetConfig[map[string]any]("db", nil)
	if dbConfig == nil {
		t.Fatal("Expected map, got nil")
	}
	if len(dbConfig) != 1 {
		t.Errorf("Expected 1 key in nested map, got %d: %+v", len(dbConfig), dbConfig)
	}
	if dbConfig["host"] != "newhost" {
		t.Errorf("Expected nested host='newhost', got '%v'", dbConfig["host"])
	}
}
