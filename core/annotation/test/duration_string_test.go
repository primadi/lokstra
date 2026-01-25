package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

func TestDurationStringFormat(t *testing.T) {
	tests := []struct {
		name        string
		serviceCode string
		expectValue string
	}{
		{
			name: "duration with string format 15m",
			serviceCode: `package testservice

import "time"

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:timeout", "15m"
	Timeout time.Duration
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectValue: `15*time.Minute`,
		},
		{
			name: "duration with string format 2h",
			serviceCode: `package testservice

import "time"

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:timeout", "2h"
	Timeout time.Duration
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectValue: `2*time.Hour`,
		},
		{
			name: "duration with string format 30s",
			serviceCode: `package testservice

import "time"

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:timeout", "30s"
	Timeout time.Duration
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectValue: `30*time.Second`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(tt.serviceCode), 0644); err != nil {
				t.Fatalf("Failed to write service file: %v", err)
			}

			_, err := annotation.ProcessPerFolder(tmpDir, annotation.GenerateCodeForFolder)
			if err != nil {
				t.Fatalf("ProcessPerFolder failed: %v", err)
			}

			generatedBytes, err := os.ReadFile(filepath.Join(tmpDir, "zz_generated.lokstra.go"))
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}
			generatedCode := string(generatedBytes)

			if !strings.Contains(generatedCode, tt.expectValue) {
				t.Errorf("Generated code should contain: %s", tt.expectValue)
				t.Logf("Generated code:\n%s", generatedCode)
			}
		})
	}
}
