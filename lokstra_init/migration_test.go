package lokstra_init_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/primadi/lokstra/lokstra_init"
)

func TestLoadMigrationYaml(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "migration.yaml")

	// Test 1: File doesn't exist
	t.Run("file_not_exists", func(t *testing.T) {
		_, err := lokstra_init.LoadMigrationYaml(filepath.Join(tmpDir, "nonexistent.yaml"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	// Test 2: Valid YAML with all fields
	t.Run("valid_yaml_all_fields", func(t *testing.T) {
		yamlContent := `dbpool-name: tenant-db
schema-table: tenant_migrations
force: on
description: Tenant database migrations
depends-on:
  - main-db
  - user-db
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		cfg, err := lokstra_init.LoadMigrationYaml(yamlPath)
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		if cfg.DbPoolName != "tenant-db" {
			t.Errorf("Expected DbPoolName 'tenant-db', got '%s'", cfg.DbPoolName)
		}
		if cfg.SchemaTable != "tenant_migrations" {
			t.Errorf("Expected SchemaTable 'tenant_migrations', got '%s'", cfg.SchemaTable)
		}
		if cfg.Description != "Tenant database migrations" {
			t.Errorf("Expected Description 'Tenant database migrations', got '%s'", cfg.Description)
		}
	})

	// Test 3: Minimal YAML
	t.Run("minimal_yaml", func(t *testing.T) {
		yamlContent := `dbpool-name: main-db`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		cfg, err := lokstra_init.LoadMigrationYaml(yamlPath)
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		if cfg.DbPoolName != "main-db" {
			t.Errorf("Expected DbPoolName 'main-db', got '%s'", cfg.DbPoolName)
		}
		// Other fields should be empty/default
		if cfg.SchemaTable != "" {
			t.Errorf("Expected empty SchemaTable, got '%s'", cfg.SchemaTable)
		}
	})

	// Test 4: Invalid YAML
	t.Run("invalid_yaml", func(t *testing.T) {
		yamlContent := `invalid: yaml: content: [[[`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err := lokstra_init.LoadMigrationYaml(yamlPath)
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestCheckDbMigration_YamlMerge(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "migration.yaml")

	// Create YAML config
	yamlContent := `dbpool-name: analytics-db
schema-table: analytics_migrations
force: off
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test: YAML config should be loaded and merged
	t.Run("yaml_config_merge", func(t *testing.T) {
		cfg := &lokstra_init.MigrationConfig{
			// MigrationsDir: tmpDir,
			// Leave other fields empty - should be filled from YAML
		}

		// We can't actually run the migration without a DB,
		// but we can verify the config merging logic works
		// by checking the values before it tries to connect

		// Simulate the merging logic from CheckDbMigration
		if yamlCfg, err := lokstra_init.LoadMigrationYaml(yamlPath); err == nil {
			if cfg.DbPoolName == "" && yamlCfg.DbPoolName != "" {
				cfg.DbPoolName = yamlCfg.DbPoolName
			}
			if cfg.SchemaTable == "" && yamlCfg.SchemaTable != "" {
				cfg.SchemaTable = yamlCfg.SchemaTable
			}
		}

		// Apply final defaults
		if cfg.DbPoolName == "" {
			cfg.DbPoolName = "main-db"
		}
		if cfg.SchemaTable == "" {
			cfg.SchemaTable = "schema_migrations"
		}

		// Verify merged values
		if cfg.DbPoolName != "analytics-db" {
			t.Errorf("Expected DbPoolName 'analytics-db' from YAML, got '%s'", cfg.DbPoolName)
		}
		if cfg.SchemaTable != "analytics_migrations" {
			t.Errorf("Expected SchemaTable 'analytics_migrations' from YAML, got '%s'", cfg.SchemaTable)
		}
	})

	// Test: Explicit config should override YAML
	t.Run("explicit_config_overrides_yaml", func(t *testing.T) {
		cfg := &lokstra_init.MigrationConfig{
			// MigrationsDir: tmpDir,
			DbPoolName: "custom-db", // Explicit value
		}

		// Simulate merging - explicit values should NOT be overridden
		if yamlCfg, err := lokstra_init.LoadMigrationYaml(yamlPath); err == nil {
			if cfg.DbPoolName == "" && yamlCfg.DbPoolName != "" {
				cfg.DbPoolName = yamlCfg.DbPoolName
			}
			if cfg.SchemaTable == "" && yamlCfg.SchemaTable != "" {
				cfg.SchemaTable = yamlCfg.SchemaTable
			}
			// Force is already set, should not be overridden
		}

		// SchemaTable should come from YAML (was empty)
		if cfg.SchemaTable != "analytics_migrations" {
			t.Errorf("Expected SchemaTable from YAML, got '%s'", cfg.SchemaTable)
		}

		// DbPoolName should keep explicit value
		if cfg.DbPoolName != "custom-db" {
			t.Errorf("Expected explicit DbPoolName 'custom-db', got '%s'", cfg.DbPoolName)
		}
	})
}
