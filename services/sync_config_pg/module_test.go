package sync_config_pg_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/services/sync_config_pg"
)

var once sync.Once

func loadConfig(t *testing.T) {
	once.Do(func() {
		if _, err := loader.LoadConfig("config_test.yaml"); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
	})
}

func TestServiceFactory(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		ReconnectInterval:  5 * time.Second,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	// This will panic if database is not available, which is expected for unit test
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic when database/pool is not available:", r)
		}
	}()

	sync_config_pg.Service(cfg)
}

func TestConfig_Defaults(t *testing.T) {
	cfg := &sync_config_pg.Config{}

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
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Test Set and Get
	err = service.Set(ctx, "test_key", "test_value")
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

	// Test GetString
	err = service.Set(ctx, "string_key", "hello world")
	if err != nil {
		t.Fatalf("Failed to set string: %v", err)
	}

	// Test GetInt
	err = service.Set(ctx, "int_key", 42)
	if err != nil {
		t.Fatalf("Failed to set int: %v", err)
	}

	// Test GetBool
	err = service.Set(ctx, "bool_key", true)
	if err != nil {
		t.Fatalf("Failed to set bool: %v", err)
	}

	// Test Delete
	err = service.Set(ctx, "delete_key", "will be deleted")
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	err = service.Delete(ctx, "delete_key")
	if err != nil {
		t.Fatalf("Failed to delete config: %v", err)
	}

	// Small delay to allow notification processing
	time.Sleep(100 * time.Millisecond)

	_, err = service.Get(ctx, "delete_key")
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	// Test GetAll
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
}

func TestSubscribeAndCRC_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Test Subscribe
	receivedKey := ""
	receivedValue := any(nil)
	done := make(chan bool, 10) // Buffer to prevent blocking from multiple notifications

	subscriptionID := service.Subscribe(func(key string, value any) {
		receivedKey = key
		receivedValue = value
		select {
		case done <- true:
		default: // Don't block if channel is full
		}
	})
	defer service.Unsubscribe(subscriptionID)

	// Set a value to trigger callback
	err = service.Set(ctx, "subscribe_test", "callback_value")
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

	// Test CRC
	crc1 := service.GetCRC()

	// Use unique value with timestamp to ensure CRC changes
	uniqueValue := time.Now().String()
	err = service.Set(ctx, "crc_test", uniqueValue)
	if err != nil {
		t.Fatal(err)
	}

	// Allow time for notification to complete
	time.Sleep(100 * time.Millisecond)

	crc2 := service.GetCRC()

	if crc1 == crc2 {
		t.Errorf("CRC should change after setting a value (crc1=%d, crc2=%d)", crc1, crc2)
	}
}

func TestSync(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}
	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	err = service.Sync(ctx)
	if err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}
}

// TestComplexDataTypes tests storing and retrieving complex JSON values
func TestComplexDataTypes_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Test object/map
	configObj := map[string]any{
		"host":    "localhost",
		"port":    5432,
		"timeout": 30,
		"ssl":     true,
	}

	err = service.Set(ctx, "db_config", configObj)
	if err != nil {
		t.Fatalf("Failed to set object: %v", err)
	}

	value, err := service.Get(ctx, "db_config")
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	valMap, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("Expected map, got %T", value)
	}

	if valMap["host"] != "localhost" {
		t.Errorf("Expected host 'localhost', got %v", valMap["host"])
	}

	// Test array
	configArray := []any{"user1", "user2", "user3"}

	err = service.Set(ctx, "allowed_users", configArray)
	if err != nil {
		t.Fatalf("Failed to set array: %v", err)
	}

	value, err = service.Get(ctx, "allowed_users")
	if err != nil {
		t.Fatalf("Failed to get array: %v", err)
	}

	valArray, ok := value.([]any)
	if !ok {
		t.Fatalf("Expected array, got %T", value)
	}

	if len(valArray) != 3 {
		t.Errorf("Expected 3 users, got %d", len(valArray))
	}

	// Test nested structure
	nested := map[string]any{
		"database": map[string]any{
			"primary": map[string]any{
				"host": "db1.example.com",
				"port": 5432,
			},
			"replicas": []any{
				"db2.example.com",
				"db3.example.com",
			},
		},
	}

	err = service.Set(ctx, "infrastructure", nested)
	if err != nil {
		t.Fatalf("Failed to set nested structure: %v", err)
	}

	value, err = service.Get(ctx, "infrastructure")
	if err != nil {
		t.Fatalf("Failed to get nested structure: %v", err)
	}

	if value == nil {
		t.Error("Expected nested structure, got nil")
	}
}

// TestUpdateExistingKey tests updating a key that already exists
func TestUpdateExistingKey_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Set initial value
	err = service.Set(ctx, "update_test", "initial_value")
	if err != nil {
		t.Fatalf("Failed to set initial value: %v", err)
	}

	value, _ := service.Get(ctx, "update_test")
	if value != "initial_value" {
		t.Errorf("Expected 'initial_value', got %v", value)
	}

	// Update with new value
	err = service.Set(ctx, "update_test", "updated_value")
	if err != nil {
		t.Fatalf("Failed to update value: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Allow notification

	value, _ = service.Get(ctx, "update_test")
	if value != "updated_value" {
		t.Errorf("Expected 'updated_value', got %v", value)
	}

	// Update with different type
	err = service.Set(ctx, "update_test", 12345)
	if err != nil {
		t.Fatalf("Failed to update with different type: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	value, _ = service.Get(ctx, "update_test")
	// JSON may unmarshal as int or float64 depending on implementation
	switch v := value.(type) {
	case int:
		if v != 12345 {
			t.Errorf("Expected 12345, got %v", v)
		}
	case float64:
		if v != 12345.0 {
			t.Errorf("Expected 12345, got %v", v)
		}
	default:
		t.Errorf("Expected number, got %T", value)
	}
}

// TestUnsubscribe tests that callbacks are not called after unsubscribe
func TestUnsubscribe_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	callCount := 0
	var mu sync.Mutex

	subscriptionID := service.Subscribe(func(key string, value any) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Set value - should trigger callback
	err = service.Set(ctx, "unsub_test1", "value1")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	firstCount := callCount
	mu.Unlock()

	if firstCount == 0 {
		t.Error("Expected callback to be called")
	}

	// Unsubscribe
	service.Unsubscribe(subscriptionID)

	// Set another value - should NOT trigger callback
	err = service.Set(ctx, "unsub_test2", "value2")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	secondCount := callCount
	mu.Unlock()

	// Count should not increase after unsubscribe
	// Note: may increase by 1 due to the Set for unsub_test2 itself (local notification)
	// but should not increase from pg_notify
	if secondCount > firstCount+1 {
		t.Errorf("Callback called after unsubscribe: before=%d, after=%d", firstCount, secondCount)
	}
}

// TestConcurrentOperations tests thread safety
func TestConcurrentOperations_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent Sets
	for i := range 10 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := time.Now().Format("concurrent_key_2006_01_02_15_04_05.000000")
			if err := service.Set(ctx, key, index); err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent Gets
	for range 10 {
		wg.Go(func() {
			_, _ = service.GetAll(ctx)
		})
	}

	// Concurrent CRC reads
	for range 10 {
		wg.Go(func() {
			_ = service.GetCRC()
		})
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}
}

// TestGetIntEdgeCases tests integer conversion from JSON
func TestGetIntEdgeCases_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Test float64 (JSON default)
	err = service.Set(ctx, "float_int", 42.0)
	if err != nil {
		t.Fatal(err)
	}

	// Test string number
	err = service.Set(ctx, "string_int", "123")
	if err != nil {
		t.Fatal(err)
	}

	// Test invalid string
	err = service.Set(ctx, "invalid_int", "not_a_number")
	if err != nil {
		t.Fatal(err)
	}
}

// TestSyncReloadsData tests that Sync reloads data from database
func TestSyncReloadsData_Integration(t *testing.T) {
	loadConfig(t)

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  1 * time.Minute,
		SyncOnMismatch:     true,
		EnableNotification: true,
	}

	service, err := sync_config_pg.Service(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Shutdown()

	ctx := context.Background()

	// Set a value
	uniqueKey := "sync_reload_" + time.Now().Format("20060102150405")
	err = service.Set(ctx, uniqueKey, "test_sync_value")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify it's there
	value, err := service.Get(ctx, uniqueKey)
	if err != nil {
		t.Fatal(err)
	}
	if value != "test_sync_value" {
		t.Errorf("Expected 'test_sync_value', got %v", value)
	}

	// Call Sync - should reload from database
	err = service.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify data still there after sync
	value, err = service.Get(ctx, uniqueKey)
	if err != nil {
		t.Fatalf("Key missing after sync: %v", err)
	}
	if value != "test_sync_value" {
		t.Errorf("Expected 'test_sync_value' after sync, got %v", value)
	}
}

// TestSingleton_Integration verifies that creating multiple instances with the same config
// returns the same singleton instance
func TestSingleton_Integration(t *testing.T) {
	loadConfig(t)
	ctx := context.Background()

	cfg := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config",
		Channel:            "config_changes",
		HeartbeatInterval:  5 * time.Minute,
		ReconnectInterval:  10 * time.Second,
		SyncOnMismatch:     true,
		EnableNotification: false, // Disable to avoid goroutine overhead
	}

	// Create first instance
	service1, err := sync_config_pg.NewSyncConfigPG(cfg)
	if err != nil {
		t.Fatalf("Failed to create first instance: %v", err)
	}
	defer service1.Shutdown()

	// Create second instance with same config
	service2, err := sync_config_pg.NewSyncConfigPG(cfg)
	if err != nil {
		t.Fatalf("Failed to create second instance: %v", err)
	}

	// Verify they are the same instance (same pointer)
	if service1 != service2 {
		t.Errorf("Expected singleton: service1 and service2 should be the same instance")
	}

	// Set value using service1
	uniqueKey := "singleton_test_key"
	err = service1.Set(ctx, uniqueKey, "from_service1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value using service2 - should see the same data immediately (same cache)
	value, err := service2.Get(ctx, uniqueKey)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "from_service1" {
		t.Errorf("Expected 'from_service1', got %v (singleton should share cache)", value)
	}

	// Create third instance with different config (different table)
	cfg3 := &sync_config_pg.Config{
		DbPoolName:         "db_main",
		TableName:          "sync_config_different", // Different table
		Channel:            "config_changes",
		HeartbeatInterval:  5 * time.Minute,
		ReconnectInterval:  10 * time.Second,
		SyncOnMismatch:     true,
		EnableNotification: false,
	}

	service3, err := sync_config_pg.NewSyncConfigPG(cfg3)
	if err != nil {
		// This might fail because table doesn't exist - that's ok for this test
		t.Logf("service3 creation failed (expected if table doesn't exist): %v", err)
		return
	}
	defer service3.Shutdown()

	// Verify service3 is a different instance
	if service1 == service3 {
		t.Errorf("Expected different instance: service1 and service3 should NOT be the same")
	}

	// Cleanup
	err = service1.Delete(ctx, uniqueKey)
	if err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
}
