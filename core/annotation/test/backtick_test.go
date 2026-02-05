package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

func TestBacktickStringInAnnotation(t *testing.T) {
	tests := []struct {
		name        string
		serviceCode string
		expectError bool
		expectValue string
	}{
		{
			name: "backtick with simple string",
			serviceCode: `package testservice

type Config struct {
	Name string
	Port int
}

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:config", ` + "`Config{Name: \"myapp\", Port: 8080}`" + `
	Cfg Config
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectError: false,
			expectValue: `Config{Name: "myapp", Port: 8080}`,
		},
		{
			name: "double quote with escaped quotes",
			serviceCode: `package testservice

type Config struct {
	Name string
	Port int
}

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:config", "Config{Name: \"myapp\", Port: 8080}"
	Cfg Config
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectError: false,
			expectValue: `Config{Name: \"myapp\", Port: 8080}`,
		},
		{
			name: "backtick with complex struct",
			serviceCode: `package testservice

import "time"

type ScheduleConfig struct {
	EventName string
	StartDate string
	Duration  time.Duration
}

// @Handler name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:schedule", ` + "`ScheduleConfig{EventName: \"Meeting\", StartDate: \"2024-12-25\", Duration: 3600000000000}`" + `
	Config ScheduleConfig
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectError: false,
			expectValue: `ScheduleConfig{EventName: "Meeting", StartDate: "2024-12-25", Duration: 3600000000000}`, // No escaped quotes with backtick!
		},
		{
			name: "backtick in Route annotation",
			serviceCode: `package testservice

// @Handler name="test-service", prefix="/api"
type TestService struct {}

// @Route ` + "`POST /users/{id}`" + `
func (s *TestService) CreateUser(id string) string { return "created" }
`,
			expectError: false,
			expectValue: `POST /users/{id}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(tt.serviceCode), 0644); err != nil {
				t.Fatalf("Failed to write service file: %v", err)
			}

			_, err := annotation.ProcessPerFolder(tmpDir, annotation.GenerateCodeForFolder)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

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
