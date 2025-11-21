package lokstra_registry_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
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
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "global-db.dsn",
		Value: "postgres://localhost/test",
	})
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "global-db.schema",
		Value: "public",
	})
	if err := registry.ResolveConfigs(); err != nil {
		t.Fatalf("Failed to resolve configs: %v", err)
	}

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
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "app-db.dsn",
		Value: "postgres://localhost/test",
	})
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "app-db.schema",
		Value: "public",
	})
	if err := registry.ResolveConfigs(); err != nil {
		t.Fatalf("Failed to resolve configs: %v", err)
	}

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
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "database.dsn",
		Value: "postgres://localhost/mydb",
	})
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "database.schema",
		Value: "app",
	})
	if err := registry.ResolveConfigs(); err != nil {
		t.Fatalf("Failed to resolve configs: %v", err)
	}

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
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "server.host",
		Value: "localhost",
	})
	registry.DefineConfig(&schema.ConfigDef{
		Name:  "server.port",
		Value: 8080.0, // YAML numbers are float64
	})
	if err := registry.ResolveConfigs(); err != nil {
		t.Fatalf("Failed to resolve configs: %v", err)
	}

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
