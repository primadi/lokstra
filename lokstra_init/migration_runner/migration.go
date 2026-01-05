package migration_runner

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/primadi/lokstra/serviceapi"
)

//go:embed lokstra_core.sql
var lokstra_core_sql string

// Migration represents a single database migration
type Migration struct {
	Version     int
	Description string
	UpSQL       string
	DownSQL     string
}

// Runner manages database migrations
type Runner struct {
	dbPool        serviceapi.DbPool
	migrationsDir string
	migrations    []*Migration
	schemaTable   string
}

// New creates a new migration runner
func New(dbPool serviceapi.DbPool, migrationsDir string) *Runner {
	return &Runner{
		dbPool:        dbPool,
		migrationsDir: migrationsDir,
		schemaTable:   "schema_migrations",
	}
}

// WithSchemaTable sets a custom schema migrations table name
func (r *Runner) WithSchemaTable(tableName string) *Runner {
	r.schemaTable = tableName
	return r
}

// load reads migration files from the migrations directory
// If minVersion > 0, only loads migrations with version >= minVersion (optimization)
func (r *Runner) load() error {
	return r.loadFrom(0) // Load all migrations
}

// loadFrom reads migration files starting from a specific version
// This is an optimization to avoid loading already-applied migrations
func (r *Runner) loadFrom(minVersion int) error {
	entries, err := os.ReadDir(r.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found: %s", r.migrationsDir)
		}
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Group migrations by version
	migrationMap := make(map[int]*Migration)

	// Detect structure type by checking for migration.yaml in subfolders
	hasSubfolderStructure := false
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this subfolder has migration.yaml
			yamlPath := filepath.Join(r.migrationsDir, entry.Name(), "migration.yaml")
			if _, err := os.Stat(yamlPath); err == nil {
				hasSubfolderStructure = true
				break
			}
		}
	}

	if hasSubfolderStructure {
		// Load from subfolder structure: migrations/001_name/migration.yaml + SQL files
		if err := r.loadFromSubfolders(entries, minVersion, migrationMap); err != nil {
			return err
		}
	} else {
		// Load from flat structure: migrations/001_name.up.sql
		if err := r.loadFromFlat(entries, minVersion, migrationMap); err != nil {
			return err
		}
	}

	// Convert map to sorted slice
	r.migrations = make([]*Migration, 0, len(migrationMap))
	for _, m := range migrationMap {
		r.migrations = append(r.migrations, m)
	}

	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].Version < r.migrations[j].Version
	})

	return nil
}

// loadFromFlat loads migrations from flat structure: migrations/001_name.up.sql
func (r *Runner) loadFromFlat(entries []os.DirEntry, minVersion int, migrationMap map[int]*Migration) error {
	// Pattern: {version}_{description}.{up|down}.sql
	pattern := regexp.MustCompile(`^(\d+)_([a-z0-9_]+)\.(up|down)\.sql$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := pattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue // Skip non-migration files
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("invalid version in file %s: %w", entry.Name(), err)
		}

		// OPTIMIZATION: Skip files with version < minVersion
		if minVersion > 0 && version < minVersion {
			continue
		}

		description := matches[2]
		direction := matches[3]

		// Read SQL file
		filePath := filepath.Join(r.migrationsDir, entry.Name())
		sqlBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		sql := string(sqlBytes)

		// Get or create migration
		migration, exists := migrationMap[version]
		if !exists {
			migration = &Migration{
				Version:     version,
				Description: description,
			}
			migrationMap[version] = migration
		} else {
			// CONFLICT DETECTION: Version exists but description differs
			if migration.Description != description {
				return fmt.Errorf(
					"❌ Migration conflict detected for version %03d:\n"+
						"  Found: %03d_%s.*.sql\n"+
						"  Found: %03d_%s.*.sql\n"+
						"  → Same version cannot have different descriptions!\n"+
						"  → Please rename one of the migrations to use a different version number",
					version,
					version, migration.Description,
					version, description,
				)
			}
		}

		// Set UP or DOWN SQL
		if direction == "up" {
			if migration.UpSQL != "" {
				return fmt.Errorf(
					"❌ Duplicate UP migration file for version %03d_%s\n"+
						"  → Only one .up.sql file allowed per version",
					version, description,
				)
			}
			migration.UpSQL = sql
		} else {
			if migration.DownSQL != "" {
				return fmt.Errorf(
					"❌ Duplicate DOWN migration file for version %03d_%s\n"+
						"  → Only one .down.sql file allowed per version",
					version, description,
				)
			}
			migration.DownSQL = sql
		}
	}

	return nil
}

// loadFromSubfolders loads migrations from subfolder structure: migrations/001_name/migration.yaml + SQL files
func (r *Runner) loadFromSubfolders(entries []os.DirEntry, minVersion int, migrationMap map[int]*Migration) error {
	// Pattern for folder: {version}_{description}
	folderPattern := regexp.MustCompile(`^(\d+)_([a-z0-9_]+)$`)
	// Pattern for SQL files inside folder: {version}_{description}.{up|down}.sql or just {description}.{up|down}.sql
	filePattern := regexp.MustCompile(`^(?:\d+_)?([a-z0-9_]+)\.(up|down)\.sql$`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files in root migrations folder
		}

		// Check if this subfolder has migration.yaml (REQUIRED)
		folderPath := filepath.Join(r.migrationsDir, entry.Name())
		yamlPath := filepath.Join(folderPath, "migration.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			continue // Skip folders without migration.yaml
		}

		// Parse folder name
		folderMatches := folderPattern.FindStringSubmatch(entry.Name())
		if folderMatches == nil {
			continue // Skip non-migration folders
		}

		folderVersion, err := strconv.Atoi(folderMatches[1])
		if err != nil {
			return fmt.Errorf("invalid version in folder %s: %w", entry.Name(), err)
		}

		// OPTIMIZATION: Skip folders with version < minVersion
		if minVersion > 0 && folderVersion < minVersion {
			continue
		}

		folderDesc := folderMatches[2]

		// Read files inside subfolder
		subEntries, err := os.ReadDir(folderPath)
		if err != nil {
			return fmt.Errorf("failed to read migration folder %s: %w", entry.Name(), err)
		}

		// Get or create migration
		migration, exists := migrationMap[folderVersion]
		if !exists {
			migration = &Migration{
				Version:     folderVersion,
				Description: folderDesc,
			}
			migrationMap[folderVersion] = migration
		}

		// Process each SQL file in the subfolder
		for _, subEntry := range subEntries {
			if subEntry.IsDir() {
				continue
			}

			// Skip migration.yaml metadata file
			if subEntry.Name() == "migration.yaml" {
				continue
			}

			fileMatches := filePattern.FindStringSubmatch(subEntry.Name())
			if fileMatches == nil {
				continue // Skip non-SQL files
			}

			fileDesc := fileMatches[1]
			direction := fileMatches[2]

			// Read SQL file
			sqlPath := filepath.Join(folderPath, subEntry.Name())
			sqlBytes, err := os.ReadFile(sqlPath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s/%s: %w", entry.Name(), subEntry.Name(), err)
			}

			sql := string(sqlBytes)

			// Set UP or DOWN SQL
			// For subfolder structure, we concatenate multiple SQL files
			if direction == "up" {
				if migration.UpSQL != "" {
					// Concatenate multiple UP migrations
					migration.UpSQL += "\n\n-- " + fileDesc + "\n" + sql
				} else {
					migration.UpSQL = sql
				}
			} else {
				if migration.DownSQL != "" {
					// Concatenate multiple DOWN migrations
					migration.DownSQL += "\n\n-- " + fileDesc + "\n" + sql
				} else {
					migration.DownSQL = sql
				}
			}
		}
	}

	return nil
}

// EnsureSchemaTable creates the schema_migrations table if it doesn't exist
func (r *Runner) EnsureSchemaTable(ctx context.Context) error {
	_, err := r.dbPool.Exec(ctx, fmt.Sprintf(lokstra_core_sql, r.schemaTable))
	if err != nil {
		return fmt.Errorf("failed to create schema table: %w", err)
	}

	return nil
}

// getAppliedVersions returns a set of applied migration versions
func (r *Runner) getAppliedVersions(ctx context.Context) (map[int]bool, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	query := fmt.Sprintf("SELECT version FROM lokstra_core.%s", r.schemaTable)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// Up runs all pending UP migrations
func (r *Runner) Up(ctx context.Context) error {
	if err := r.EnsureSchemaTable(ctx); err != nil {
		return err
	}

	// OPTIMIZATION: Get current version first, then only load newer migrations
	currentVersion, err := r.getCurrentVersion(ctx)
	if err != nil {
		return err
	}

	// Load only migrations with version > currentVersion
	if err := r.loadFrom(currentVersion + 1); err != nil {
		return err
	}

	// If no pending migrations loaded, we're done
	if len(r.migrations) == 0 {
		fmt.Println("✅ No pending migrations")
		return nil
	}

	fmt.Printf("🔄 Running %d migration(s)...\n", len(r.migrations))

	// All loaded migrations are pending (because we filtered by version)
	for _, m := range r.migrations {
		if err := r.runUpMigration(ctx, m); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}
	}

	fmt.Println("✅ All migrations completed")
	return nil
}

// getCurrentVersion returns the highest applied migration version
// Returns 0 if no migrations have been applied
func (r *Runner) getCurrentVersion(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(version), 0) FROM lokstra_core.%s", r.schemaTable)

	var maxVersion int
	err := r.dbPool.QueryRow(ctx, query).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return maxVersion, nil
}

// runUpMigration runs a single UP migration in a transaction
func (r *Runner) runUpMigration(ctx context.Context, m *Migration) error {
	if m.UpSQL == "" {
		return fmt.Errorf("no UP SQL for migration %d", m.Version)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	return conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
		// Execute migration SQL
		_, err := tx.Exec(ctx, m.UpSQL)
		if err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record migration
		insertSQL := fmt.Sprintf(
			"INSERT INTO lokstra_core.%s (version, description) VALUES ($1, $2)",
			r.schemaTable,
		)
		_, err = tx.Exec(ctx, insertSQL, m.Version, m.Description)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		fmt.Printf("  ✓ Migration %03d: %s\n", m.Version, m.Description)
		return nil
	})
}

// Down rolls back the last applied migration
func (r *Runner) Down(ctx context.Context) error {
	return r.DownN(ctx, 1)
}

// DownN rolls back the last N applied migrations
func (r *Runner) DownN(ctx context.Context, n int) error {
	if err := r.load(); err != nil {
		return err
	}

	if err := r.EnsureSchemaTable(ctx); err != nil {
		return err
	}

	applied, err := r.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	// Get applied migrations in reverse order
	toRollback := make([]*Migration, 0)
	for i := len(r.migrations) - 1; i >= 0; i-- {
		m := r.migrations[i]
		if applied[m.Version] {
			toRollback = append(toRollback, m)
			if len(toRollback) >= n {
				break
			}
		}
	}

	if len(toRollback) == 0 {
		fmt.Println("✅ No migrations to rollback")
		return nil
	}

	fmt.Printf("🔄 Rolling back %d migration(s)...\n", len(toRollback))

	for _, m := range toRollback {
		if err := r.runDownMigration(ctx, m); err != nil {
			return fmt.Errorf("rollback of migration %d failed: %w", m.Version, err)
		}
	}

	fmt.Println("✅ Rollback completed")
	return nil
}

// runDownMigration runs a single DOWN migration in a transaction
func (r *Runner) runDownMigration(ctx context.Context, m *Migration) error {
	if m.DownSQL == "" {
		return fmt.Errorf(
			"❌ Cannot rollback migration %03d_%s\n"+
				"  → Missing file: %03d_%s.down.sql\n"+
				"  → Create the DOWN migration file or manually remove this version from schema_migrations table",
			m.Version, m.Description,
			m.Version, m.Description,
		)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	return conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
		// Execute rollback SQL
		_, err := tx.Exec(ctx, m.DownSQL)
		if err != nil {
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}

		// Remove migration record
		deleteSQL := fmt.Sprintf(
			"DELETE FROM lokstra_core.%s WHERE version = $1",
			r.schemaTable,
		)
		_, err = tx.Exec(ctx, deleteSQL, m.Version)
		if err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		fmt.Printf("  ✓ Rolled back %03d: %s\n", m.Version, m.Description)
		return nil
	})
}

// Version returns the current migration version
func (r *Runner) Version(ctx context.Context) (int, error) {
	if err := r.EnsureSchemaTable(ctx); err != nil {
		return 0, err
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	query := fmt.Sprintf("SELECT COALESCE(MAX(version), 0) FROM lokstra_core.%s", r.schemaTable)
	var version int
	err = conn.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// Status returns migration status information
func (r *Runner) Status(ctx context.Context) (string, error) {
	if err := r.load(); err != nil {
		return "", err
	}

	if err := r.EnsureSchemaTable(ctx); err != nil {
		return "", err
	}

	applied, err := r.getAppliedVersions(ctx)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("\nMigration Status:\n")
	sb.WriteString("=================\n\n")

	if len(r.migrations) == 0 {
		sb.WriteString("No migrations found\n")
		return sb.String(), nil
	}

	for _, m := range r.migrations {
		status := "[ ]"
		if applied[m.Version] {
			status = "[✓]"
		}
		sb.WriteString(fmt.Sprintf("%s %03d: %s\n", status, m.Version, m.Description))
	}

	pendingCount := 0
	for _, m := range r.migrations {
		if !applied[m.Version] {
			pendingCount++
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d migrations, %d applied, %d pending\n",
		len(r.migrations), len(applied), pendingCount))

	return sb.String(), nil
}

// Create generates a new migration file pair (up and down)
func (r *Runner) Create(name string) error {
	// Validate name format (must be snake_case alphanumeric)
	validName := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf(
			"❌ Invalid migration name: '%s'\n"+
				"  → Use snake_case with lowercase letters, numbers, and underscores only\n"+
				"  → Example: create_users_table, add_email_index",
			name,
		)
	}

	// Find highest version number in existing migrations
	nextVersion := 1
	entries, err := os.ReadDir(r.migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	pattern := regexp.MustCompile(`^(\d+)_`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := pattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			version, _ := strconv.Atoi(matches[1])
			if version >= nextVersion {
				nextVersion = version + 1
			}
		}
	}

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(r.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate file names
	upFile := filepath.Join(r.migrationsDir, fmt.Sprintf("%03d_%s.up.sql", nextVersion, name))
	downFile := filepath.Join(r.migrationsDir, fmt.Sprintf("%03d_%s.down.sql", nextVersion, name))

	// Check if files already exist
	if _, err := os.Stat(upFile); err == nil {
		return fmt.Errorf("❌ Migration file already exists: %s", filepath.Base(upFile))
	}
	if _, err := os.Stat(downFile); err == nil {
		return fmt.Errorf("❌ Migration file already exists: %s", filepath.Base(downFile))
	}

	// Create UP migration template
	upTemplate := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Version: %03d

-- Add your UP migration SQL here
-- Example:
-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     email VARCHAR(255) UNIQUE NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
`, name, "auto-generated", nextVersion)

	// Create DOWN migration template
	downTemplate := fmt.Sprintf(`-- Migration: %s (ROLLBACK)
-- Created: %s
-- Version: %03d

-- Add your DOWN migration SQL here (reverse of UP migration)
-- Example:
-- DROP TABLE IF EXISTS users;
`, name, "auto-generated", nextVersion)

	// Write files
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create UP migration file: %w", err)
	}

	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		// Cleanup UP file if DOWN file creation fails
		os.Remove(upFile)
		return fmt.Errorf("failed to create DOWN migration file: %w", err)
	}

	fmt.Printf("✅ Created migration files:\n")
	fmt.Printf("   → %s\n", filepath.Base(upFile))
	fmt.Printf("   → %s\n", filepath.Base(downFile))
	fmt.Printf("\n")
	fmt.Printf("📝 Next steps:\n")
	fmt.Printf("   1. Edit the migration files with your SQL\n")
	fmt.Printf("   2. Run: go run main.go --cmd=up\n")

	return nil
}
