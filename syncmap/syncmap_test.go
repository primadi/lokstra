package syncmap_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/syncmap"
)

var once sync.Once

func loadConfig(t *testing.T) {
	once.Do(func() {
		if _, err := loader.LoadConfig("../services/sync_config_pg/config_test.yaml"); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
	})
}

func TestSyncMapPG_BasicOperations(t *testing.T) {
	loadConfig(t)

	// Create SyncMap for testing
	testMap := syncmap.NewSyncMap[string]("test")
	ctx := context.Background()

	// Test Set
	err := testMap.Set(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	val, err := testMap.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%s'", val)
	}

	// Test Has
	if !testMap.Has(ctx, "key1") {
		t.Error("Has should return true for existing key")
	}
	if testMap.Has(ctx, "nonexistent") {
		t.Error("Has should return false for nonexistent key")
	}

	// Test Delete
	err = testMap.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = testMap.Get(ctx, "key1")
	if err == nil {
		t.Error("Key should not exist after deletion")
	}
}

func TestSyncMapPG_ComplexTypes(t *testing.T) {
	loadConfig(t)

	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	testMap := syncmap.NewSyncMap[TestStruct]("users")
	ctx := context.Background()

	// Test complex type
	user := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	err := testMap.Set(ctx, "john", user)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := testMap.Get(ctx, "john")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != user.Name || retrieved.Age != user.Age {
		t.Errorf("Retrieved user doesn't match: %+v", retrieved)
	}

	// Cleanup
	testMap.Clear(ctx)
}

func TestSyncMapPG_PrefixIsolation(t *testing.T) {
	loadConfig(t)

	// Create two maps with different prefixes
	map1 := syncmap.NewSyncMap[string]("prefix1")
	map2 := syncmap.NewSyncMap[string]("prefix2")
	ctx := context.Background()

	// Set values in both maps with same key
	map1.Set(ctx, "shared_key", "value_from_map1")
	map2.Set(ctx, "shared_key", "value_from_map2")

	// Verify isolation
	val1, err1 := map1.Get(ctx, "shared_key")
	val2, err2 := map2.Get(ctx, "shared_key")

	if err1 != nil || err2 != nil {
		t.Fatal("Both values should exist")
	}
	if val1 != "value_from_map1" {
		t.Errorf("Map1 expected 'value_from_map1', got '%s'", val1)
	}
	if val2 != "value_from_map2" {
		t.Errorf("Map2 expected 'value_from_map2', got '%s'", val2)
	}

	// Verify Keys returns only prefixed keys
	keys1, _ := map1.Keys(ctx)
	keys2, _ := map2.Keys(ctx)

	if len(keys1) != 1 || keys1[0] != "shared_key" {
		t.Errorf("Map1 keys incorrect: %v", keys1)
	}
	if len(keys2) != 1 || keys2[0] != "shared_key" {
		t.Errorf("Map2 keys incorrect: %v", keys2)
	}

	// Cleanup
	map1.Clear(ctx)
	map2.Clear(ctx)
}

func TestSyncMapPG_BulkOperations(t *testing.T) {
	loadConfig(t)

	testMap := syncmap.NewSyncMap[int]("numbers")
	ctx := context.Background()

	// Set multiple values
	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("num%d", i)
		if err := testMap.Set(ctx, key, i*10); err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Test Len
	length := testMap.Len(ctx)
	if length != 10 {
		t.Errorf("Expected length 10, got %d", length)
	}

	// Test Keys
	keys, err := testMap.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}
	if len(keys) != 10 {
		t.Errorf("Expected 10 keys, got %d", len(keys))
	}

	// Test Values
	values, err := testMap.Values(ctx)
	if err != nil {
		t.Fatalf("Values failed: %v", err)
	}
	if len(values) != 10 {
		t.Errorf("Expected 10 values, got %d", len(values))
	}

	// Test All
	all, err := testMap.All(ctx)
	if err != nil {
		t.Fatalf("All failed: %v", err)
	}
	if len(all) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(all))
	}

	// Test Range
	count := 0
	testMap.Range(ctx, func(key string, value int) bool {
		count++
		if value%10 != 0 {
			t.Errorf("Invalid value %d for key %s", value, key)
		}
		return true
	})
	if count != 10 {
		t.Errorf("Range should iterate 10 times, got %d", count)
	}

	// Test Clear
	err = testMap.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if testMap.Len(ctx) != 0 {
		t.Error("Map should be empty after Clear")
	}
}

func TestSyncMapPG_Subscribe(t *testing.T) {
	loadConfig(t)

	testMap := syncmap.NewSyncMap[string]("subscribe_test")
	ctx := context.Background()

	// Subscribe to changes
	changes := make(chan string, 10)
	subID := testMap.Subscribe(func(key string, value string) {
		changes <- key + "=" + value
	})
	defer testMap.Unsubscribe(subID)

	// Set a value
	testMap.Set(ctx, "test_key", "test_value")

	// Wait for notification
	select {
	case change := <-changes:
		if change != "test_key=test_value" {
			t.Errorf("Expected 'test_key=test_value', got '%s'", change)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for subscription notification")
	}

	// Cleanup
	testMap.Clear(ctx)
}
