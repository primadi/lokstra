---
name: advanced-lokstra-validate-consistency
description: Validate application consistency - circular dependencies, schema validation, config checks, annotation validation, and service registration. Use after all code is implemented to identify issues before deployment.
phase: advanced
order: 3
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Advanced: Validate Consistency

## Overview

This skill provides comprehensive validation tools for Lokstra applications to ensure:
- Code quality and dependency correctness
- Configuration completeness and validity
- Database schema consistency
- Annotation correctness
- Service registration and injection validity

**Validation Categories:**
1. **Static Analysis** - Runs without starting the app (annotations, imports, config files)
2. **Runtime Validation** - Runs at application startup (service resolution, DI)
3. **Database Validation** - Requires database connection (schema, migrations)

## When to Use

Use this skill when:
- Before merging code to production branch
- Checking for circular dependencies between services
- Validating configuration completeness
- Ensuring database schema matches migrations
- Detecting configuration mismatches
- Pre-deployment validation
- CI/CD pipeline integration

Prerequisites:
- ✅ All code implemented (Phase 1-2 complete)
- ✅ Configuration finalized (config.yaml, configs/*.yaml)
- ✅ Database migrations created
- ✅ Ready for deployment

---

## Quick Validation Commands

```bash
# 1. Compile-time check (catches most errors)
go build ./...

# 2. Run with --generate-only (validates annotations without running server)
go run . --generate-only

# 3. Run all tests
go test ./... -v

# 4. Run specific validation scripts
go run scripts/validate_config.go
go run scripts/validate_deps.go
go run scripts/validate_annotations.go

# 5. Full pre-deployment check
bash scripts/pre_deploy_check.sh
```

---

## 1. Annotation Validation

### Validate @Handler, @Service, @Route, @Inject Annotations

Lokstra generates code from annotations. Invalid annotations cause runtime failures.

**File:** `scripts/validate_annotations.go`

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ValidationError struct {
	File    string
	Line    int
	Message string
}

var (
	handlerPattern = regexp.MustCompile(`@Handler\s+(?:name\s*=\s*"([^"]+)")?`)
	servicePattern = regexp.MustCompile(`@Service\s+"([^"]+)"`)
	routePattern   = regexp.MustCompile(`@Route\s+"(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s+([^"]+)"`)
	injectPattern  = regexp.MustCompile(`@Inject\s+"([^"]+)"`)
)

func main() {
	root := "./modules"
	errors := validateAnnotations(root)

	if len(errors) > 0 {
		fmt.Println("❌ ANNOTATION VALIDATION ERRORS:")
		for _, err := range errors {
			fmt.Printf("  %s:%d - %s\n", err.File, err.Line, err.Message)
		}
		os.Exit(1)
	}

	fmt.Println("✅ All annotations are valid")
}

func validateAnnotations(root string) []ValidationError {
	var errors []ValidationError
	handlerNames := make(map[string]string) // name -> file
	serviceNames := make(map[string]string) // name -> file

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip generated files
		if strings.HasSuffix(path, "_lokstra_gen.go") {
			return nil
		}

		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			errors = append(errors, ValidationError{
				File:    path,
				Line:    0,
				Message: fmt.Sprintf("Parse error: %v", parseErr),
			})
			return nil
		}

		// Extract comments and validate annotations
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				line := fset.Position(c.Pos()).Line
				text := c.Text

				// Validate @Handler
				if strings.Contains(text, "@Handler") {
					if match := handlerPattern.FindStringSubmatch(text); match != nil {
						name := match[1]
						if name == "" {
							errors = append(errors, ValidationError{
								File:    path,
								Line:    line,
								Message: "@Handler missing required 'name' parameter",
							})
						} else if existing, exists := handlerNames[name]; exists {
							errors = append(errors, ValidationError{
								File:    path,
								Line:    line,
								Message: fmt.Sprintf("Duplicate @Handler name '%s' (already in %s)", name, existing),
							})
						} else {
							handlerNames[name] = path
						}
					}
				}

				// Validate @Service
				if strings.Contains(text, "@Service") {
					if match := servicePattern.FindStringSubmatch(text); match != nil {
						name := match[1]
						if existing, exists := serviceNames[name]; exists {
							errors = append(errors, ValidationError{
								File:    path,
								Line:    line,
								Message: fmt.Sprintf("Duplicate @Service name '%s' (already in %s)", name, existing),
							})
						} else {
							serviceNames[name] = path
						}
					}
				}

				// Validate @Route
				if strings.Contains(text, "@Route") {
					if !routePattern.MatchString(text) {
						// Check for common mistakes
						if strings.Contains(text, `@Route "`) {
							errors = append(errors, ValidationError{
								File:    path,
								Line:    line,
								Message: "@Route format should be: @Route \"METHOD /path\" (e.g., @Route \"GET /users\")",
							})
						}
					}
				}

				// Validate @Inject
				if strings.Contains(text, "@Inject") {
					if match := injectPattern.FindStringSubmatch(text); match != nil {
						value := match[1]
						// Check for empty inject
						if strings.TrimSpace(value) == "" {
							errors = append(errors, ValidationError{
								File:    path,
								Line:    line,
								Message: "@Inject value cannot be empty",
							})
						}
					}
				}
			}
		}

		return nil
	})

	return errors
}
```

Run with: `go run scripts/validate_annotations.go`

---

## 2. Circular Dependency Detection

### Module-Level Dependencies

Lokstra uses DDD bounded contexts (modules). Cross-module dependencies should be unidirectional.

**File:** `scripts/validate_deps.go`

```go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Dependency struct {
	From string
	To   string
}

func main() {
	root := "./modules"
	deps := findDependencies(root)
	cycles := findCycles(deps)

	if len(cycles) > 0 {
		fmt.Println("❌ CIRCULAR DEPENDENCIES DETECTED:")
		for _, cycle := range cycles {
			fmt.Println("  ", strings.Join(cycle, " -> "))
		}
		fmt.Println("")
		fmt.Println("💡 Solutions:")
		fmt.Println("  1. Extract shared types to modules/shared/domain/")
		fmt.Println("  2. Use interfaces for cross-module communication")
		fmt.Println("  3. Use event-driven patterns for decoupling")
		os.Exit(1)
	}

	fmt.Println("✅ No circular dependencies found")
}

func findDependencies(root string) []Dependency {
	var deps []Dependency

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}

		module := extractModuleName(path)
		if module == "" || module == "shared" {
			return nil // Skip shared module
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if strings.Contains(importPath, "/modules/") {
				importedModule := extractModuleFromPath(importPath)
				if importedModule != "" && importedModule != module && importedModule != "shared" {
					deps = append(deps, Dependency{
						From: module,
						To:   importedModule,
					})
				}
			}
		}
		return nil
	})

	return deps
}

func extractModuleName(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		if part == "modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractModuleFromPath(importPath string) string {
	parts := strings.Split(importPath, "/")
	for i, part := range parts {
		if part == "modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func findCycles(deps []Dependency) [][]string {
	// Build adjacency list
	graph := make(map[string]map[string]bool)
	for _, dep := range deps {
		if graph[dep.From] == nil {
			graph[dep.From] = make(map[string]bool)
		}
		graph[dep.From][dep.To] = true
	}

	// DFS to find cycles
	var cycles [][]string
	visited := make(map[string]int) // 0=unvisited, 1=in-progress, 2=done

	var dfs func(node string, path []string) bool
	dfs = func(node string, path []string) bool {
		if visited[node] == 1 {
			// Found cycle - extract cycle from path
			cycleStart := -1
			for i, n := range path {
				if n == node {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(path[cycleStart:], node)
				cycles = append(cycles, cycle)
			}
			return true
		}
		if visited[node] == 2 {
			return false
		}

		visited[node] = 1
		path = append(path, node)

		for neighbor := range graph[node] {
			dfs(neighbor, path)
		}

		visited[node] = 2
		return false
	}

	for module := range graph {
		if visited[module] == 0 {
			dfs(module, nil)
		}
	}

	return cycles
}
```

---

## 3. Configuration Validation

### Validate config.yaml and configs/*.yaml

**File:** `scripts/validate_config.go`

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigValidationError struct {
	File    string
	Path    string
	Message string
}

func main() {
	errors := []ConfigValidationError{}

	// Load all config files
	configFiles := []string{"config.yaml"}
	if entries, err := os.ReadDir("configs"); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml") {
				configFiles = append(configFiles, filepath.Join("configs", entry.Name()))
			}
		}
	}

	merged := make(map[string]any)

	// Parse and merge all configs
	for _, file := range configFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			if file == "config.yaml" {
				errors = append(errors, ConfigValidationError{
					File:    file,
					Message: fmt.Sprintf("Cannot read config file: %v", err),
				})
			}
			continue
		}

		var cfg map[string]any
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			errors = append(errors, ConfigValidationError{
				File:    file,
				Message: fmt.Sprintf("YAML parse error: %v", err),
			})
			continue
		}

		mergeConfig(merged, cfg)
	}

	// Validate required sections
	requiredSections := []string{"deployments", "service-definitions"}
	for _, section := range requiredSections {
		if _, exists := merged[section]; !exists {
			errors = append(errors, ConfigValidationError{
				Path:    section,
				Message: fmt.Sprintf("Missing required section: %s", section),
			})
		}
	}

	// Validate service-definitions
	if services, ok := merged["service-definitions"].(map[string]any); ok {
		for name, def := range services {
			svc, ok := def.(map[string]any)
			if !ok {
				errors = append(errors, ConfigValidationError{
					Path:    fmt.Sprintf("service-definitions.%s", name),
					Message: "Service definition must be an object",
				})
				continue
			}

			// Check required 'type' field
			if _, hasType := svc["type"]; !hasType {
				errors = append(errors, ConfigValidationError{
					Path:    fmt.Sprintf("service-definitions.%s", name),
					Message: "Missing required 'type' field",
				})
			}

			// Validate depends-on references
			if deps, ok := svc["depends-on"].([]any); ok {
				for _, dep := range deps {
					depName, _ := dep.(string)
					if _, exists := services[depName]; !exists {
						errors = append(errors, ConfigValidationError{
							Path:    fmt.Sprintf("service-definitions.%s.depends-on", name),
							Message: fmt.Sprintf("Dependency '%s' not found in service-definitions", depName),
						})
					}
				}
			}
		}
	}

	// Validate deployments
	if deployments, ok := merged["deployments"].(map[string]any); ok {
		for deployName, deployDef := range deployments {
			deploy, ok := deployDef.(map[string]any)
			if !ok {
				continue
			}

			servers, hasServers := deploy["servers"].(map[string]any)
			if !hasServers || len(servers) == 0 {
				errors = append(errors, ConfigValidationError{
					Path:    fmt.Sprintf("deployments.%s", deployName),
					Message: "Deployment must have at least one server",
				})
				continue
			}

			for serverName, serverDef := range servers {
				server, ok := serverDef.(map[string]any)
				if !ok {
					continue
				}

				// Check server has addr
				if _, hasAddr := server["addr"]; !hasAddr {
					errors = append(errors, ConfigValidationError{
						Path:    fmt.Sprintf("deployments.%s.servers.%s", deployName, serverName),
						Message: "Server missing 'addr' field",
					})
				}

				// Validate published-services exist
				if pubServices, ok := server["published-services"].([]any); ok {
					services, _ := merged["service-definitions"].(map[string]any)
					for _, svc := range pubServices {
						svcName, _ := svc.(string)
						if _, exists := services[svcName]; !exists {
							errors = append(errors, ConfigValidationError{
								Path:    fmt.Sprintf("deployments.%s.servers.%s.published-services", deployName, serverName),
								Message: fmt.Sprintf("Published service '%s' not found in service-definitions", svcName),
							})
						}
					}
				}
			}
		}
	}

	// Validate middleware-definitions references
	if middlewares, ok := merged["middleware-definitions"].(map[string]any); ok {
		for name, def := range middlewares {
			mw, ok := def.(map[string]any)
			if !ok {
				continue
			}
			if _, hasType := mw["type"]; !hasType {
				errors = append(errors, ConfigValidationError{
					Path:    fmt.Sprintf("middleware-definitions.%s", name),
					Message: "Missing required 'type' field",
				})
			}
		}
	}

	// Output results
	if len(errors) > 0 {
		fmt.Println("❌ CONFIGURATION ERRORS:")
		for _, err := range errors {
			location := err.File
			if location == "" {
				location = err.Path
			}
			fmt.Printf("  [%s] %s\n", location, err.Message)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Configuration is valid")
	fmt.Printf("   Loaded %d config files\n", len(configFiles))
}

func mergeConfig(dst, src map[string]any) {
	for key, srcVal := range src {
		if dstVal, exists := dst[key]; exists {
			if dstMap, ok := dstVal.(map[string]any); ok {
				if srcMap, ok := srcVal.(map[string]any); ok {
					mergeConfig(dstMap, srcMap)
					continue
				}
			}
		}
		dst[key] = srcVal
	}
}
```

---

## 4. Injection Validation

### Validate @Inject References

Ensure all `@Inject` annotations reference existing services or valid config paths.

**File:** `scripts/validate_inject.go`

```go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var injectPattern = regexp.MustCompile(`@Inject\s+"([^"]+)"`)

type InjectRef struct {
	File   string
	Line   int
	Target string
}

func main() {
	// Load all inject references
	injects := findInjectAnnotations("./modules")

	// Load service definitions from config
	services := loadServiceDefinitions()

	// Load config keys
	configKeys := loadConfigKeys()

	errors := []string{}

	for _, inject := range injects {
		target := inject.Target

		switch {
		case strings.HasPrefix(target, "cfg:"):
			// Config value injection
			key := strings.TrimPrefix(target, "cfg:")
			key = strings.TrimPrefix(key, "@") // Handle indirect reference
			if !hasConfigKey(configKeys, key) {
				errors = append(errors, fmt.Sprintf(
					"%s:%d - Config key '%s' not found",
					inject.File, inject.Line, key,
				))
			}

		case strings.HasPrefix(target, "@"):
			// Config-based service reference
			key := strings.TrimPrefix(target, "@")
			if !hasConfigKey(configKeys, key) {
				errors = append(errors, fmt.Sprintf(
					"%s:%d - Config reference '%s' not found",
					inject.File, inject.Line, key,
				))
			}

		default:
			// Direct service reference
			if _, exists := services[target]; !exists {
				errors = append(errors, fmt.Sprintf(
					"%s:%d - Service '%s' not found in service-definitions",
					inject.File, inject.Line, target,
				))
			}
		}
	}

	if len(errors) > 0 {
		fmt.Println("❌ INJECTION VALIDATION ERRORS:")
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("✅ All %d @Inject references are valid\n", len(injects))
}

func findInjectAnnotations(root string) []InjectRef {
	var refs []InjectRef

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_lokstra_gen.go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if matches := injectPattern.FindAllStringSubmatch(c.Text, -1); matches != nil {
					for _, match := range matches {
						refs = append(refs, InjectRef{
							File:   path,
							Line:   fset.Position(c.Pos()).Line,
							Target: match[1],
						})
					}
				}
			}
		}

		return nil
	})

	return refs
}

func loadServiceDefinitions() map[string]any {
	services := make(map[string]any)

	// Load from config.yaml
	if data, err := os.ReadFile("config.yaml"); err == nil {
		var cfg map[string]any
		if yaml.Unmarshal(data, &cfg) == nil {
			if svcDefs, ok := cfg["service-definitions"].(map[string]any); ok {
				for k, v := range svcDefs {
					services[k] = v
				}
			}
		}
	}

	// Load from configs/*.yaml
	if entries, err := os.ReadDir("configs"); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				path := filepath.Join("configs", entry.Name())
				if data, err := os.ReadFile(path); err == nil {
					var cfg map[string]any
					if yaml.Unmarshal(data, &cfg) == nil {
						if svcDefs, ok := cfg["service-definitions"].(map[string]any); ok {
							for k, v := range svcDefs {
								services[k] = v
							}
						}
					}
				}
			}
		}
	}

	return services
}

func loadConfigKeys() map[string]any {
	configs := make(map[string]any)

	loadFile := func(path string) {
		if data, err := os.ReadFile(path); err == nil {
			var cfg map[string]any
			if yaml.Unmarshal(data, &cfg) == nil {
				if cfgSection, ok := cfg["configs"].(map[string]any); ok {
					flattenConfig("", cfgSection, configs)
				}
			}
		}
	}

	loadFile("config.yaml")
	if entries, err := os.ReadDir("configs"); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				loadFile(filepath.Join("configs", entry.Name()))
			}
		}
	}

	return configs
}

func flattenConfig(prefix string, cfg map[string]any, result map[string]any) {
	for k, v := range cfg {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		result[key] = v
		if nested, ok := v.(map[string]any); ok {
			flattenConfig(key, nested, result)
		}
	}
}

func hasConfigKey(configs map[string]any, key string) bool {
	_, exists := configs[key]
	return exists
}
```

---

## 5. Database Schema Validation

### Validate Migrations Match Database

**File:** `scripts/validate_schema.go`

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

var createTablePattern = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?`)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("⚠️  DATABASE_URL not set - skipping database schema validation")
		fmt.Println("   Set DATABASE_URL to validate against actual database")
		os.Exit(0)
	}

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("❌ Cannot connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verify connection
	if err := db.PingContext(context.Background()); err != nil {
		fmt.Println("❌ Database connection failed:", err)
		os.Exit(1)
	}

	// Get list of created tables
	dbTables, err := getTablesFromDB(db)
	if err != nil {
		fmt.Println("❌ Cannot query database tables:", err)
		os.Exit(1)
	}

	// Get expected tables from migrations
	expectedTables, err := getExpectedTables("./migrations")
	if err != nil {
		fmt.Println("❌ Cannot parse migration files:", err)
		os.Exit(1)
	}

	errors := []string{}
	warnings := []string{}

	// Check all expected tables exist
	for table := range expectedTables {
		if _, exists := dbTables[table]; !exists {
			errors = append(errors, fmt.Sprintf("Expected table '%s' not found in database", table))
		}
	}

	// Warn about unexpected tables (may be OK - could be from other sources)
	for table := range dbTables {
		if _, expected := expectedTables[table]; !expected && !isSystemTable(table) {
			warnings = append(warnings, fmt.Sprintf("Unexpected table '%s' in database (not in migrations)", table))
		}
	}

	// Output warnings
	for _, warn := range warnings {
		fmt.Printf("⚠️  %s\n", warn)
	}

	if len(errors) > 0 {
		fmt.Println("")
		fmt.Println("❌ SCHEMA VALIDATION FAILED:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println("")
		fmt.Println("💡 Solutions:")
		fmt.Println("  1. Run migrations: lokstra migration up")
		fmt.Println("  2. Check migration files in ./migrations/")
		os.Exit(1)
	}

	fmt.Printf("✅ Database schema is valid (%d tables verified)\n", len(expectedTables))
}

func getTablesFromDB(db *sql.DB) (map[string]bool, error) {
	tables := make(map[string]bool)

	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables[strings.ToLower(tableName)] = true
	}

	return tables, rows.Err()
}

func getExpectedTables(migrationsDir string) (map[string]bool, error) {
	tables := make(map[string]bool)

	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Find all CREATE TABLE statements
		matches := createTablePattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 2 {
				tableName := strings.ToLower(match[1])
				tables[tableName] = true
			}
		}

		return nil
	})

	return tables, err
}

func isSystemTable(name string) bool {
	systemTables := map[string]bool{
		"_prisma_migrations": true,
		"schema_migrations":  true,
		"migrations":         true,
		"goose_db_version":   true,
	}
	return systemTables[strings.ToLower(name)]
}
```

---

## 6. Migration Status Check

### Validate Migration State

**File:** `scripts/validate_migrations.go`

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("⚠️  DATABASE_URL not set - checking migration files only")
		checkMigrationFilesOnly()
		return
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("❌ Cannot connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Get applied migrations from database
	applied, err := getAppliedMigrations(db)
	if err != nil {
		fmt.Println("⚠️  Cannot query migration table (may not exist yet)")
		checkMigrationFilesOnly()
		return
	}

	// Get local migration files
	local, err := getLocalMigrations("./migrations")
	if err != nil {
		fmt.Println("❌ Cannot read migration files:", err)
		os.Exit(1)
	}

	// Compare
	errors := []string{}
	pending := []string{}

	for _, migration := range local {
		if _, exists := applied[migration]; !exists {
			pending = append(pending, migration)
		}
	}

	for migration := range applied {
		found := false
		for _, localMig := range local {
			if localMig == migration {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, fmt.Sprintf("Applied migration '%s' not found in local files", migration))
		}
	}

	if len(errors) > 0 {
		fmt.Println("❌ MIGRATION VALIDATION FAILED:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	if len(pending) > 0 {
		fmt.Printf("⚠️  %d pending migrations:\n", len(pending))
		for _, mig := range pending {
			fmt.Printf("  - %s\n", mig)
		}
		fmt.Println("")
		fmt.Println("💡 Run: lokstra migration up")
	} else {
		fmt.Printf("✅ All %d migrations are applied\n", len(applied))
	}
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	migrations := make(map[string]bool)

	rows, err := db.QueryContext(context.Background(), `
		SELECT version FROM schema_migrations ORDER BY version
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations[version] = true
	}

	return migrations, rows.Err()
}

func getLocalMigrations(dir string) ([]string, error) {
	var migrations []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".up.sql") {
			// Extract version from filename (e.g., "001_create_users.up.sql" -> "001")
			base := filepath.Base(path)
			parts := strings.SplitN(base, "_", 2)
			if len(parts) >= 1 {
				migrations = append(migrations, parts[0])
			}
		}
		return nil
	})

	sort.Strings(migrations)
	return migrations, err
}

func checkMigrationFilesOnly() {
	local, err := getLocalMigrations("./migrations")
	if err != nil {
		fmt.Println("❌ Cannot read migration files:", err)
		os.Exit(1)
	}

	if len(local) == 0 {
		fmt.Println("⚠️  No migration files found in ./migrations/")
		return
	}

	// Check for gaps in version numbers
	fmt.Printf("📋 Found %d migration files:\n", len(local))
	for _, mig := range local {
		fmt.Printf("  - %s\n", mig)
	}

	// Check for matching .up.sql and .down.sql pairs
	upFiles := make(map[string]bool)
	downFiles := make(map[string]bool)

	filepath.Walk("./migrations", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".up.sql") {
			prefix := strings.TrimSuffix(base, ".up.sql")
			upFiles[prefix] = true
		} else if strings.HasSuffix(base, ".down.sql") {
			prefix := strings.TrimSuffix(base, ".down.sql")
			downFiles[prefix] = true
		}
		return nil
	})

	// Check for missing pairs
	warnings := []string{}
	for up := range upFiles {
		if !downFiles[up] {
			warnings = append(warnings, fmt.Sprintf("Missing .down.sql for %s", up))
		}
	}
	for down := range downFiles {
		if !upFiles[down] {
			warnings = append(warnings, fmt.Sprintf("Missing .up.sql for %s", down))
		}
	}

	if len(warnings) > 0 {
		fmt.Println("")
		fmt.Println("⚠️  Migration file warnings:")
		for _, warn := range warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
}
```

---

## 7. Environment Variable Validation

### Validate Required Environment Variables

**File:** `scripts/validate_env.go`

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	// Find all ${VAR} references in config files
	requiredVars := findEnvVarReferences()

	// Check which are missing
	missing := []string{}
	set := []string{}

	for varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missing = append(missing, varName)
		} else {
			set = append(set, varName)
		}
	}

	if len(set) > 0 {
		fmt.Printf("✅ Set environment variables (%d):\n", len(set))
		for _, v := range set {
			fmt.Printf("   ✓ %s\n", v)
		}
	}

	if len(missing) > 0 {
		fmt.Println("")
		fmt.Printf("⚠️  Missing environment variables (%d):\n", len(missing))
		for _, v := range missing {
			fmt.Printf("   ✗ %s\n", v)
		}
		fmt.Println("")
		fmt.Println("💡 These may be required for production deployment")
	} else if len(set) == 0 {
		fmt.Println("ℹ️  No environment variable references found in config")
	}
}

func findEnvVarReferences() map[string]bool {
	vars := make(map[string]bool)

	processFile := func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}

		// Find ${VAR} and ${VAR:-default} patterns
		content := string(data)
		for i := 0; i < len(content); i++ {
			if i+1 < len(content) && content[i:i+2] == "${" {
				end := strings.Index(content[i:], "}")
				if end > 0 {
					varExpr := content[i+2 : i+end]
					// Handle ${VAR:-default} syntax
					if colonIdx := strings.Index(varExpr, ":"); colonIdx > 0 {
						varExpr = varExpr[:colonIdx]
					}
					if varExpr != "" {
						vars[varExpr] = true
					}
				}
			}
		}
	}

	processFile("config.yaml")
	if entries, err := os.ReadDir("configs"); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".yaml") {
				processFile("configs/" + entry.Name())
			}
		}
	}

	return vars
}
```

---

## 8. Pre-Deployment Checklist Script

### Comprehensive Pre-Deployment Validation

**File:** `scripts/pre_deploy_check.sh` (Linux/Mac)

```bash
#!/bin/bash

set -e  # Exit on first error

echo "🔍 Running pre-deployment checks..."
echo ""

ERRORS=0

# Function to run a check
run_check() {
    local name="$1"
    local cmd="$2"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "▶ $name"
    echo ""
    if eval "$cmd"; then
        echo ""
    else
        echo ""
        echo "❌ $name FAILED"
        ERRORS=$((ERRORS + 1))
    fi
}

# Check 1: Go compilation
run_check "1. Compile Check (go build)" "go build ./..."

# Check 2: Code generation
run_check "2. Code Generation (--generate-only)" "go run . --generate-only"

# Check 3: Annotation validation
if [ -f "scripts/validate_annotations.go" ]; then
    run_check "3. Annotation Validation" "go run scripts/validate_annotations.go"
else
    echo "▶ 3. Annotation Validation (skipped - script not found)"
fi

# Check 4: Dependency analysis
if [ -f "scripts/validate_deps.go" ]; then
    run_check "4. Dependency Analysis" "go run scripts/validate_deps.go"
else
    echo "▶ 4. Dependency Analysis (skipped - script not found)"
fi

# Check 5: Configuration
if [ -f "scripts/validate_config.go" ]; then
    run_check "5. Configuration Validation" "go run scripts/validate_config.go"
else
    echo "▶ 5. Configuration Validation (skipped - script not found)"
fi

# Check 6: Injection validation
if [ -f "scripts/validate_inject.go" ]; then
    run_check "6. Injection Validation" "go run scripts/validate_inject.go"
else
    echo "▶ 6. Injection Validation (skipped - script not found)"
fi

# Check 7: Tests
run_check "7. Unit Tests" "go test ./... -v -timeout 120s"

# Check 8: Database schema (if DATABASE_URL set)
if [ -n "$DATABASE_URL" ]; then
    if [ -f "scripts/validate_schema.go" ]; then
        run_check "8. Database Schema Validation" "go run scripts/validate_schema.go"
    fi
else
    echo "▶ 8. Database Schema Validation (skipped - DATABASE_URL not set)"
fi

# Check 9: Build binary
run_check "9. Build Binary" "go build -o ./tmp/app ./cmd/app 2>/dev/null || go build -o ./tmp/app ."

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -gt 0 ]; then
    echo "❌ Pre-deployment check FAILED with $ERRORS error(s)"
    exit 1
else
    echo "✅ All pre-deployment checks passed!"
    echo ""
    echo "Ready to deploy. Run: ./tmp/app"
fi
```

**File:** `scripts/pre_deploy_check.ps1` (Windows PowerShell)

```powershell
#!/usr/bin/env pwsh

$ErrorActionPreference = "Stop"

Write-Host "🔍 Running pre-deployment checks..." -ForegroundColor Cyan
Write-Host ""

$errors = 0

function Run-Check {
    param(
        [string]$Name,
        [string]$Command
    )
    
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host "▶ $Name" -ForegroundColor Yellow
    Write-Host ""
    
    try {
        Invoke-Expression $Command
        if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne $null) {
            throw "Command failed with exit code $LASTEXITCODE"
        }
        Write-Host ""
    }
    catch {
        Write-Host ""
        Write-Host "❌ $Name FAILED" -ForegroundColor Red
        $script:errors++
    }
}

# Check 1: Go compilation
Run-Check "1. Compile Check (go build)" "go build ./..."

# Check 2: Code generation
Run-Check "2. Code Generation (--generate-only)" "go run . --generate-only"

# Check 3: Annotation validation
if (Test-Path "scripts/validate_annotations.go") {
    Run-Check "3. Annotation Validation" "go run scripts/validate_annotations.go"
} else {
    Write-Host "▶ 3. Annotation Validation (skipped - script not found)" -ForegroundColor Gray
}

# Check 4: Dependency analysis
if (Test-Path "scripts/validate_deps.go") {
    Run-Check "4. Dependency Analysis" "go run scripts/validate_deps.go"
} else {
    Write-Host "▶ 4. Dependency Analysis (skipped - script not found)" -ForegroundColor Gray
}

# Check 5: Configuration
if (Test-Path "scripts/validate_config.go") {
    Run-Check "5. Configuration Validation" "go run scripts/validate_config.go"
} else {
    Write-Host "▶ 5. Configuration Validation (skipped - script not found)" -ForegroundColor Gray
}

# Check 6: Injection validation
if (Test-Path "scripts/validate_inject.go") {
    Run-Check "6. Injection Validation" "go run scripts/validate_inject.go"
} else {
    Write-Host "▶ 6. Injection Validation (skipped - script not found)" -ForegroundColor Gray
}

# Check 7: Tests
Run-Check "7. Unit Tests" "go test ./... -v -timeout 120s"

# Check 8: Database schema
if ($env:DATABASE_URL) {
    if (Test-Path "scripts/validate_schema.go") {
        Run-Check "8. Database Schema Validation" "go run scripts/validate_schema.go"
    }
} else {
    Write-Host "▶ 8. Database Schema Validation (skipped - DATABASE_URL not set)" -ForegroundColor Gray
}

# Check 9: Build binary
$null = New-Item -ItemType Directory -Path "./tmp" -Force
Run-Check "9. Build Binary" "go build -o ./tmp/app.exe ."

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

if ($errors -gt 0) {
    Write-Host "❌ Pre-deployment check FAILED with $errors error(s)" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✅ All pre-deployment checks passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Ready to deploy. Run: ./tmp/app.exe" -ForegroundColor Cyan
}
```

---

## 9. CI/CD Integration

### GitHub Actions Workflow

**File:** `.github/workflows/validation.yml`

```yaml
name: Validation

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  validate:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      
      - name: Install dependencies
        run: go mod download
      
      - name: Compile check
        run: go build ./...
      
      - name: Code generation
        run: go run . --generate-only
      
      - name: Validate annotations
        run: |
          if [ -f "scripts/validate_annotations.go" ]; then
            go run scripts/validate_annotations.go
          fi
      
      - name: Check circular dependencies
        run: |
          if [ -f "scripts/validate_deps.go" ]; then
            go run scripts/validate_deps.go
          fi
      
      - name: Validate configuration
        run: |
          if [ -f "scripts/validate_config.go" ]; then
            go run scripts/validate_config.go
          fi
      
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test_db?sslmode=disable
      
      - name: Validate database schema
        run: |
          if [ -f "scripts/validate_schema.go" ]; then
            go run scripts/validate_schema.go
          fi
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test_db?sslmode=disable
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
```

---

## 10. Runtime Validation (Application Startup)

### Built-in Lokstra Validation

Lokstra performs automatic validation at startup:

```go
package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Bootstrap validates:
	// ✅ Annotation parsing
	// ✅ Code generation
	// ✅ Import cycles
	lokstra_init.Bootstrap()

	// RunServerFromConfig validates:
	// ✅ Config file parsing
	// ✅ Service type registration
	// ✅ Dependency resolution
	// ✅ Service instantiation
	// ✅ Router registration
	lokstra_registry.RunServerFromConfig()
}
```

### Custom Startup Validation

```go
package main

import (
	"log"

	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	lokstra_init.Bootstrap()

	// Custom validation before starting server
	if err := validateRequiredServices(); err != nil {
		log.Fatalf("Service validation failed: %v", err)
	}

	if err := validateDatabaseConnections(); err != nil {
		log.Fatalf("Database validation failed: %v", err)
	}

	lokstra_registry.RunServerFromConfig()
}

func validateRequiredServices() error {
	required := []string{"user-handler", "auth-handler", "db-main"}
	
	for _, name := range required {
		if !lokstra_registry.HasService(name) {
			return fmt.Errorf("required service '%s' not registered", name)
		}
	}
	return nil
}

func validateDatabaseConnections() error {
	// Get all database pools and test connections
	db := lokstra_registry.GetService[serviceapi.DbPool]("db-main")
	if db == nil {
		return fmt.Errorf("database pool 'db-main' not available")
	}
	
	// Test query
	if err := db.Pool().Ping(context.Background()); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}
```

---

## Best Practices Summary

### 1. Validation Order

```
1. Static Analysis (no app start required)
   ├── go build ./...
   ├── validate_annotations.go
   ├── validate_deps.go
   ├── validate_config.go
   └── validate_inject.go

2. Code Generation
   └── go run . --generate-only

3. Tests
   └── go test ./...

4. Database (requires connection)
   ├── validate_migrations.go
   └── validate_schema.go

5. Full Application Start
   └── go run .
```

### 2. Fail Fast Principles

| Principle | Implementation |
|-----------|----------------|
| ✅ Exit on first critical error | Use `os.Exit(1)` |
| ✅ Clear error messages | Include file:line and suggested fix |
| ✅ Actionable suggestions | Provide "💡 Solutions:" section |
| ✅ Separate warnings from errors | Use ⚠️ vs ❌ |

### 3. Validation Categories

| Category | When to Run | Tools |
|----------|------------|-------|
| Annotations | Every build | `validate_annotations.go` |
| Dependencies | Every PR | `validate_deps.go` |
| Configuration | Every deploy | `validate_config.go` |
| Injection | Every deploy | `validate_inject.go` |
| Schema | Before deploy | `validate_schema.go` |
| Migrations | Before deploy | `validate_migrations.go` |
| Environment | Before deploy | `validate_env.go` |

---

## Troubleshooting Common Issues

### Issue: Duplicate Handler/Service Names

```
❌ Duplicate @Handler name 'user-handler' (already in modules/user/...)
```

**Solution:** Each @Handler and @Service must have a unique name across the entire application.

### Issue: Circular Dependency Detected

```
❌ CIRCULAR DEPENDENCIES: user -> order -> user
```

**Solutions:**
1. Extract shared types to `modules/shared/domain/`
2. Use interfaces for cross-module communication
3. Use event-driven patterns for decoupling

### Issue: Missing Service Reference

```
❌ Service 'user-repository' not found in service-definitions
```

**Solutions:**
1. Add service to `config.yaml` under `service-definitions`
2. Check spelling matches exactly
3. Ensure config file is being loaded

### Issue: Config Key Not Found

```
❌ Config key 'app.timeout' not found
```

**Solutions:**
1. Add key to `configs` section in config.yaml
2. Check nested path structure
3. Verify config files are merged correctly

---

## Related Skills

- [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md) - Comprehensive testing
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Configuration management
- [design-lokstra-schema-design](../design-lokstra-schema-design/SKILL.md) - Database schema design
- [implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md) - Migration files

---

## Next Steps

After validation passes:

1. ✅ Merge to production branch
2. ✅ Deploy with confidence
3. ✅ Monitor application logs
4. ✅ Set up health checks in production
