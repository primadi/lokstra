package lokstra_init

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/primadi/lokstra/common/dbpool_crud"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_init/migration_runner"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"gopkg.in/yaml.v3"
)

// MigrationYamlConfig represents the migration.yaml file structure
// This file is optional and located in the migrations directory
type MigrationYamlConfig struct {
	// DbPoolName is the database pool name from config.yaml dbpool-definitions
	DbPoolName string `yaml:"dbpool-name"`

	// SchemaTable is the table name for tracking migrations
	// Default: "schema_migrations"
	SchemaTable string `yaml:"schema-table"`

	// Enabled controls whether migrations are enabled
	// nil = enabled by default (true)
	// true = explicitly enabled
	// false = explicitly disabled
	Enabled *bool `yaml:"enabled"`

	// Description for documentation purposes
	Description string `yaml:"description"`
}

// MigrationConfig holds configuration for database migrations
type MigrationConfig struct {
	// MigrationsDir is the directory containing migration files
	// Default: "migrations"
	MigrationsDir string

	// DbPoolName is the name of the database pool from dbpool-manager
	// Default: "main-db"
	// Can be overridden by migration.yaml
	DbPoolName string

	// SchemaTable is the table name for tracking applied migrations
	// Default: "schema_migrations"
	// Can be overridden by migration.yaml
	SchemaTable string
}

// CheckDbMigration runs database migrations based on runtime mode
//
// Behavior based on Force setting:
//   - "true"/MigrationForceTrue: Always run migrations (even in prod)
//   - "false"/MigrationForceFalse: Never run migrations (use CLI instead)
//   - "auto"/MigrationForceAuto (default): Auto-detect based on runtime.mode
//   - dev/debug mode → run migrations automatically
//   - prod mode → skip migrations (use lokstra CLI)
//
// Example usage in main():
//
//	func main() {
//	    lokstra.Bootstrap()
//
//	    // Auto mode - runs in dev/debug, skips in prod (recommended)
//	    lokstra.CheckDbMigration(&lokstra.MigrationConfig{
//	        MigrationsDir: "migrations",
//	        Force: lokstra.MigrationForceAuto, // or just leave empty (default)
//	    })
//
//	    // Always run (even in prod)
//	    lokstra.CheckDbMigration(lokstra.MigrationConfig{
//	        Force: lokstra.MigrationForceTrue,
//	    })
//
//	    // Never run (use CLI only)
//	    lokstra.CheckDbMigration(lokstra.MigrationConfig{
//	        Force: lokstra.MigrationForceFalse,
//	    })
//
//	    lokstra_registry.RunServerFromConfig()
//	}
func CheckDbMigration(cfg *MigrationConfig) error {
	if cfg == nil {
		cfg = &MigrationConfig{}
	}
	// Set defaults
	if cfg.MigrationsDir == "" {
		cfg.MigrationsDir = "migrations"
	}

	// Try to load migration.yaml from migrations directory
	yamlPath := filepath.Join(cfg.MigrationsDir, "migration.yaml")
	if yamlCfg, err := LoadMigrationYaml(yamlPath); err == nil {
		// Merge YAML config with provided config (YAML takes precedence if not set)
		if cfg.DbPoolName == "" && yamlCfg.DbPoolName != "" {
			cfg.DbPoolName = yamlCfg.DbPoolName
		}
		if cfg.SchemaTable == "" && yamlCfg.SchemaTable != "" {
			cfg.SchemaTable = yamlCfg.SchemaTable
		}
	}

	// Apply final defaults if still empty
	if cfg.DbPoolName == "" {
		names, err := dbpool_crud.ListDbPools()
		if err != nil || len(names) == 0 {
			return fmt.Errorf("no database pools defined - check your config.yaml dbpool-definitions section")
		}
		cfg.DbPoolName = names[0] // Use first available pool
	}
	if cfg.SchemaTable == "" {
		cfg.SchemaTable = "schema_migrations"
	}

	// Get database pool
	pool, ok := lokstra_registry.GetServiceAny(cfg.DbPoolName)
	if !ok {
		return fmt.Errorf("database pool '%s' not found - check your config.yaml dbpool-definitions section", cfg.DbPoolName)
	}

	dbPool, ok := pool.(serviceapi.DbPool)
	if !ok {
		return fmt.Errorf("service '%s' is not a DbPoolWithSchema", cfg.DbPoolName)
	}

	// Create migration runner with custom schema table if specified
	runner := migration_runner.New(dbPool, cfg.MigrationsDir)
	if cfg.SchemaTable != "" && cfg.SchemaTable != "schema_migrations" {
		runner = runner.WithSchemaTable(cfg.SchemaTable)
	}

	// Run migrations
	ctx := context.Background()

	logger.LogInfo("[Lokstra] Running migrations (dir=%s, db=%s, schema=%s)",
		cfg.MigrationsDir, cfg.DbPoolName, cfg.SchemaTable)

	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.LogInfo("[Lokstra] Migrations completed successfully")

	return nil
}

// LoadMigrationYaml loads and parses migration.yaml file
// Returns error if file doesn't exist or cannot be parsed
func LoadMigrationYaml(path string) (*MigrationYamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err // File doesn't exist or cannot be read
	}

	var cfg MigrationYamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse migration.yaml: %w", err)
	}

	return &cfg, nil
}

// CheckDbMigrationsAuto scans a directory for migration folders and runs them in alphabetical order
//
// This function is designed for multi-database systems where each database has its own
// migration folder with migration.yaml configuration.
//
// Directory structure:
//
//	migrations/
//	├── 01_main-db/           # Runs first (requires migration.yaml)
//	│   ├── migration.yaml    # REQUIRED for subfolder detection
//	│   └── 001_create_users.up.sql
//	├── 02_tenant-db/         # Runs second (requires migration.yaml)
//	│   ├── migration.yaml    # REQUIRED for subfolder detection
//	│   └── 001_create_tenants.up.sql
//	└── 03_ledger-db/         # Runs third (requires migration.yaml)
//	    ├── migration.yaml    # REQUIRED for subfolder detection
//	    └── 001_create_accounts.up.sql
//
// Execution order is determined by folder name (alphabetical sort).
// Use numeric prefixes (01_, 02_, 03_) for explicit ordering.
//
// Each subfolder MUST contain migration.yaml for detection.
// Subfolders without migration.yaml will be ignored.
//
// Example usage:
//
//	func main() {
//	    lokstra.Bootstrap()
//
//	    // Auto-scan and run all database migrations
//	    if err := lokstra.CheckDbMigrationsAuto("migrations"); err != nil {
//	        logger.LogPanic("Migrations failed: %v", err)
//	    }
//
//	    lokstra_registry.RunServerFromConfig()
//	}
//
// Example migration.yaml:
//
// dbpool-name: main-db
// schema-table: schema_migrations
// force: auto
// description: Main application database with users, orders, products
func CheckDbMigrationsAuto(configFolder string) error {
	// Read all subdirectories sorted alphabetically
	basePath := utils.GetBasePath()
	rootDir := filepath.Join(basePath, configFolder)

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory '%s': %w", rootDir, err)
	}

	// Collect migration folders that have migration.yaml (explicit marker)
	var migrationFolders []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if migration.yaml exists (REQUIRED for subfolder detection)
			yamlPath := filepath.Join(rootDir, entry.Name(), "migration.yaml")
			if _, err := os.Stat(yamlPath); err == nil {
				migrationFolders = append(migrationFolders, entry.Name())
			}
		}
	}

	// If no subdirectories with migration.yaml found, treat rootDir as single migration folder
	if len(migrationFolders) == 0 {
		return CheckDbMigration(&MigrationConfig{
			MigrationsDir: rootDir,
		})
	}

	// Run migrations for each folder in order
	successCount := 0
	errorCount := 0
	skippedCount := 0

	for _, folder := range migrationFolders {
		folderPath := filepath.Join(rootDir, folder)

		logger.LogInfo("[Lokstra] Processing migration folder: %s", folder)

		// Check if migration is enabled (default: true if not specified)
		yamlPath := filepath.Join(folderPath, "migration.yaml")
		yamlCfg, _ := LoadMigrationYaml(yamlPath)

		// Default enabled to true if not explicitly set to false
		enabled := true
		if yamlCfg != nil && yamlCfg.Enabled != nil {
			enabled = *yamlCfg.Enabled
		}

		if !enabled {
			logger.LogInfo("[Lokstra] Skipping disabled migration folder: %s", folder)
			skippedCount++
			continue
		}

		// Run migration for this folder
		err := CheckDbMigration(&MigrationConfig{
			MigrationsDir: folderPath,
			DbPoolName:    yamlCfg.DbPoolName,
			SchemaTable:   yamlCfg.SchemaTable,
		})

		if err != nil {
			errorCount++
			logger.LogError("[Lokstra] Migration failed for '%s': %v", folder, err)
			// Continue with other folders instead of returning error
			continue
		}

		successCount++
	}

	logger.LogInfo("[Lokstra] Multi-database migrations completed: %d successful, %d errors, %d skipped", successCount, errorCount, skippedCount)
	return nil
}
