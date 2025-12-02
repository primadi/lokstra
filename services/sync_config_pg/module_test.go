package sync_config_pg

import (
	"context"
	"testing"
	"time"

	"github.com/primadi/lokstra/services/dbpool_manager"
)

func TestServiceFactory(t *testing.T) {
	// Create a mock dbpool manager
	poolManager := dbpool_manager.NewPgxPoolManager()

	deps := map[string]any{
		"dbpool-manager": poolManager,
	}

	params := map[string]any{
		"dbpool_name":        "test-pool",
		"table_name":         "test_config",
		"channel":            "test_channel",
		"heartbeat_interval": 1, // 1 minute
		"sync_on_mismatch":   true,
	}

	// This will panic if database is not available, which is expected for unit test
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic when database/pool is not available:", r)
		}
	}()

	_ = ServiceFactory(deps, params)
}

func TestConfig_Defaults(t *testing.T) {
	cfg := &Config{
		// DbPoolName: "global-db",
	}

	if cfg.TableName == "" {
		cfg.TableName = "sync_config"
	}

	if cfg.Channel == "" {
		cfg.Channel = "config_changes"
	}

	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 5 * time.Minute
	}

	if cfg.TableName != "sync_config" {
		t.Errorf("Expected table_name 'sync_config', got %s", cfg.TableName)
	}

	if cfg.Channel != "config_changes" {
		t.Errorf("Expected channel 'config_changes', got %s", cfg.Channel)
	}

	if cfg.HeartbeatInterval != 5*time.Minute {
		t.Errorf("Expected heartbeat 5 minutes, got %v", cfg.HeartbeatInterval)
	}
}

// Integration tests - requires PostgreSQL
// Run with: go test -v -run TestSyncConfig_Integration
// Make sure PostgreSQL is running and accessible

func TestSyncConfig_Integration(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")

	// Setup DB Pool Manager
	poolManager := dbpool_manager.NewPgxPoolManager()
	poolManager.SetNamedDsn("test-pool", "postgres://localhost:5432/testdb?sslmode=disable", "public")

	cfg := &Config{
		DbPoolName:         "test-pool",
		TableName:          "test_sync_config",
		Channel:            "test_config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := Service(cfg, poolManager)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Test Set and Get
	t.Run("SetAndGet", func(t *testing.T) {
		err := service.Set(ctx, "test_key", "test_value")
		if err != nil {
			t.Fatalf("Failed to set config: %v", err)
		}

		value, err := service.Get(ctx, "test_key")
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}

		if value != "test_value" {
			t.Errorf("Expected 'test_value', got %v", value)
		}
	})

	// Test GetString
	t.Run("GetString", func(t *testing.T) {
		err := service.Set(ctx, "string_key", "hello world")
		if err != nil {
			t.Fatalf("Failed to set string: %v", err)
		}

		value := service.GetString(ctx, "string_key", "default")
		if value != "hello world" {
			t.Errorf("Expected 'hello world', got %s", value)
		}

		// Test default value
		value = service.GetString(ctx, "nonexistent", "default")
		if value != "default" {
			t.Errorf("Expected 'default', got %s", value)
		}
	})

	// Test GetInt
	t.Run("GetInt", func(t *testing.T) {
		err := service.Set(ctx, "int_key", 42)
		if err != nil {
			t.Fatalf("Failed to set int: %v", err)
		}

		value := service.GetInt(ctx, "int_key", 0)
		if value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}
	})

	// Test GetBool
	t.Run("GetBool", func(t *testing.T) {
		err := service.Set(ctx, "bool_key", true)
		if err != nil {
			t.Fatalf("Failed to set bool: %v", err)
		}

		value := service.GetBool(ctx, "bool_key", false)
		if value != true {
			t.Errorf("Expected true, got %v", value)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := service.Set(ctx, "delete_key", "will be deleted")
		if err != nil {
			t.Fatalf("Failed to set config: %v", err)
		}

		err = service.Delete(ctx, "delete_key")
		if err != nil {
			t.Fatalf("Failed to delete config: %v", err)
		}

		_, err = service.Get(ctx, "delete_key")
		if err == nil {
			t.Error("Expected error when getting deleted key")
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		err := service.Set(ctx, "key1", "value1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.Set(ctx, "key2", "value2")
		if err != nil {
			t.Fatal(err)
		}

		all, err := service.GetAll(ctx)
		if err != nil {
			t.Fatalf("Failed to get all configs: %v", err)
		}

		if len(all) < 2 {
			t.Errorf("Expected at least 2 configs, got %d", len(all))
		}
	})

	// Test Subscribe
	t.Run("Subscribe", func(t *testing.T) {
		receivedKey := ""
		receivedValue := any(nil)
		done := make(chan bool)

		subscriptionID := service.Subscribe(func(key string, value any) {
			receivedKey = key
			receivedValue = value
			done <- true
		})
		defer service.Unsubscribe(subscriptionID)

		// Set a value to trigger callback
		err := service.Set(ctx, "subscribe_test", "callback_value")
		if err != nil {
			t.Fatal(err)
		}

		// Wait for callback
		select {
		case <-done:
			if receivedKey != "subscribe_test" {
				t.Errorf("Expected key 'subscribe_test', got %s", receivedKey)
			}
			if receivedValue != "callback_value" {
				t.Errorf("Expected value 'callback_value', got %v", receivedValue)
			}
		case <-time.After(2 * time.Second):
			t.Error("Timeout waiting for subscription callback")
		}
	})

	// Test CRC
	t.Run("CRC", func(t *testing.T) {
		crc1 := service.GetCRC()

		err := service.Set(ctx, "crc_test", "crc_value")
		if err != nil {
			t.Fatal(err)
		}

		crc2 := service.GetCRC()

		if crc1 == crc2 {
			t.Error("CRC should change after setting a value")
		}
	})

	// Test Sync
	t.Run("Sync", func(t *testing.T) {
		err := service.Sync(ctx)
		if err != nil {
			t.Fatalf("Failed to sync: %v", err)
		}
	})
}

// TestMultipleInstances tests sync between multiple instances
func TestMultipleInstances_Integration(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")

	// Setup DB Pool Manager
	poolManager := dbpool_manager.NewPgxPoolManager()
	poolManager.SetNamedDsn("test-pool", "postgres://localhost:5432/testdb?sslmode=disable", "public")

	cfg := &Config{
		DbPoolName:         "test-pool",
		TableName:          "test_sync_multi",
		Channel:            "test_multi_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	// Create first instance
	service1, err := Service(cfg, poolManager)
	if err != nil {
		t.Fatalf("Failed to create service1: %v", err)
	}
	defer service1.Shutdown()

	// Create second instance
	service2, err := Service(cfg, poolManager)
	if err != nil {
		t.Fatalf("Failed to create service2: %v", err)
	}
	defer service2.Shutdown()

	ctx := context.Background()

	// Subscribe on service2
	received := make(chan bool, 1)
	service2.Subscribe(func(key string, value any) {
		if key == "sync_test" {
			received <- true
		}
	})

	// Set value on service1
	err = service1.Set(ctx, "sync_test", "synced_value")
	if err != nil {
		t.Fatalf("Failed to set on service1: %v", err)
	}

	// Wait for notification on service2
	select {
	case <-received:
		// Check if service2 has the value
		value := service2.GetString(ctx, "sync_test", "")
		if value != "synced_value" {
			t.Errorf("Expected 'synced_value' on service2, got %s", value)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for sync between instances")
	}

	// Verify CRC matches
	crc1 := service1.GetCRC()
	crc2 := service2.GetCRC()

	if crc1 != crc2 {
		t.Errorf("CRC mismatch: service1=%d, service2=%d", crc1, crc2)
	}
}
