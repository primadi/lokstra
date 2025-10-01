package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

// Benchmark data
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{1, "John Doe", "john@example.com"},
	{2, "Jane Smith", "jane@example.com"},
	{3, "Bob Johnson", "bob@example.com"},
}

// Struct-based formatter (current implementation)
func formatWithStruct(data any) []byte {
	response := &api_formatter.ApiResponse{
		Status: "success",
		Data:   data,
	}

	jsonData, _ := json.Marshal(response)
	return jsonData
}

// Map-based formatter (old implementation)
func formatWithMap(data any) []byte {
	response := map[string]any{
		"status": "success",
		"data":   data,
	}

	jsonData, _ := json.Marshal(response)
	return jsonData
}

// Benchmark functions are in main_test.go

func main() {
	fmt.Println("🚀 Performance Comparison: Struct vs Map for JSON Response Formatting")
	fmt.Println()

	// Manual timing test
	iterations := 1000000

	// Test struct performance
	start := time.Now()
	for i := 0; i < iterations; i++ {
		formatWithStruct(users)
	}
	structDuration := time.Since(start)

	// Test map performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		formatWithMap(users)
	}
	mapDuration := time.Since(start)

	fmt.Printf("📊 Performance Results (%d iterations):\n", iterations)
	fmt.Printf("• Struct-based formatting: %v\n", structDuration)
	fmt.Printf("• Map-based formatting:    %v\n", mapDuration)

	if structDuration < mapDuration {
		improvement := float64(mapDuration-structDuration) / float64(mapDuration) * 100
		fmt.Printf("✅ Struct is %.1f%% faster than Map!\n", improvement)
	} else {
		degradation := float64(structDuration-mapDuration) / float64(structDuration) * 100
		fmt.Printf("❌ Struct is %.1f%% slower than Map\n", degradation)
	}

	fmt.Println()
	fmt.Println("📋 Why Struct is Better:")
	fmt.Println("• Compile-time type checking")
	fmt.Println("• Better memory layout")
	fmt.Println("• Faster JSON marshaling/unmarshaling")
	fmt.Println("• IDE autocompletion support")
	fmt.Println("• Less memory allocations")
	fmt.Println()
	fmt.Printf("Run benchmark with: go test -bench=.\n")
}
