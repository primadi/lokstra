package dbpool_manager_test

import (
	"context"
	"testing"

	"github.com/primadi/lokstra/dbpool_manager"
	"github.com/primadi/lokstra/lokstra_init"
)

// Example: Add a new database pool at runtime
func ExampleAddDbPool() {
	// Initialize framework first
	lokstra_init.UsePgxDbPoolManager(true) // Enable distributed sync

	// Add new pool
	err := dbpool_manager.AddDbPool(dbpool_manager.DbPoolConfig{
		Name:   "db-analytics",
		DSN:    "postgres://user:pass@localhost:5432/analytics",
		Schema: "public",
	})

	if err != nil {
		panic(err)
	}

	// Pool is now available across all servers (if using distributed sync)
	conn, _ := dbpool_manager.AcquireDbConn(context.Background(), "db-analytics")
	defer conn.Release()

	// Use connection...
}

// Example: Update existing pool configuration
func ExampleUpdateDbPool() {
	err := dbpool_manager.UpdateDbPool(dbpool_manager.DbPoolConfig{
		Name:   "db-main",
		DSN:    "postgres://user:pass@new-host:5432/main",
		Schema: "public",
	})

	if err != nil {
		panic(err)
	}

	// New connections will use the updated configuration
	// Existing connections remain unchanged until closed
}

// Example: Remove a pool
func ExampleRemoveDbPool() {
	err := dbpool_manager.RemoveDbPool("db-old-tenant")
	if err != nil {
		panic(err)
	}

	// Pool removed from all servers (if using distributed sync)
}

// Example: List all pools
func ExampleListDbPools() {
	pools, err := dbpool_manager.ListDbPools()
	if err != nil {
		panic(err)
	}

	for _, poolName := range pools {
		info, _ := dbpool_manager.GetDbPoolInfo(poolName)
		println("Pool:", info.Name, "Schema:", info.Schema)
	}
}

// Example: Get pool info
func ExampleGetDbPoolInfo() {
	info, err := dbpool_manager.GetDbPoolInfo("db-main")
	if err != nil {
		panic(err)
	}

	println("DSN:", info.DSN)
	println("Schema:", info.Schema)
}

// Example: Direct pool access
func ExampleGetDbPool() {
	pool, err := dbpool_manager.GetDbPool("db-main")
	if err != nil {
		panic(err)
	}

	conn, _ := pool.Acquire(context.Background())
	defer conn.Release()

	// Use connection...
}

// Test CRUD operations
func TestDbPoolCRUD(t *testing.T) {
	// Setup
	lokstra_init.UsePgxDbPoolManager(false) // Use local sync for testing

	// Create
	err := dbpool_manager.AddDbPool(dbpool_manager.DbPoolConfig{
		Name:   "test-pool",
		DSN:    "postgres://localhost/test",
		Schema: "test_schema",
	})
	if err != nil {
		t.Fatalf("Failed to add pool: %v", err)
	}

	// Read
	info, err := dbpool_manager.GetDbPoolInfo("test-pool")
	if err != nil {
		t.Fatalf("Failed to get pool info: %v", err)
	}
	if info.Name != "test-pool" {
		t.Errorf("Expected name 'test-pool', got '%s'", info.Name)
	}
	if info.Schema != "test_schema" {
		t.Errorf("Expected schema 'test_schema', got '%s'", info.Schema)
	}

	// Update
	err = dbpool_manager.UpdateDbPool(dbpool_manager.DbPoolConfig{
		Name:   "test-pool",
		DSN:    "postgres://localhost/test2",
		Schema: "test_schema2",
	})
	if err != nil {
		t.Fatalf("Failed to update pool: %v", err)
	}

	info, _ = dbpool_manager.GetDbPoolInfo("test-pool")
	if info.Schema != "test_schema2" {
		t.Errorf("Expected schema 'test_schema2', got '%s'", info.Schema)
	}

	// Delete
	err = dbpool_manager.RemoveDbPool("test-pool")
	if err != nil {
		t.Fatalf("Failed to remove pool: %v", err)
	}

	// Verify deletion
	_, err = dbpool_manager.GetDbPoolInfo("test-pool")
	if err == nil {
		t.Error("Expected error when getting deleted pool")
	}
}
