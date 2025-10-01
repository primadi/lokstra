package main

import (
	"encoding/json"
	"testing"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

type TestUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var testUsers = []TestUser{
	{1, "John Doe", "john@example.com"},
	{2, "Jane Smith", "jane@example.com"},
	{3, "Bob Johnson", "bob@example.com"},
}

func BenchmarkStructFormatter(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		response := &api_formatter.ApiResponse{
			Status: "success",
			Data:   testUsers,
		}
		_, _ = json.Marshal(response)
	}
}

func BenchmarkMapFormatter(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		response := map[string]any{
			"status": "success",
			"data":   testUsers,
		}
		_, _ = json.Marshal(response)
	}
}
