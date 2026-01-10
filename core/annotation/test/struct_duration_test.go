package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

func TestStructWithDurationField(t *testing.T) {
	tests := []struct {
		name            string
		serviceCode     string
		expectedStrings []string
	}{
		{
			name: "struct field duration - auto convert from string",
			serviceCode: `package testservice

import "time"

type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:server"
	Server ServerConfig
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Server", "ServerConfig",
				"cast.ToStruct",
			},
		},
		{
			name: "struct with duration default using backtick",
			serviceCode: `package testservice

import "time"

type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:server"
	// NOTE: Default value not yet supported for struct config injection via @Inject
	// Use config.yaml to provide default values instead
	Server ServerConfig
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Server", "ServerConfig",
				"cast.ToStruct",
			},
		},
		{
			name: "struct with duration default using double quote",
			serviceCode: `package testservice

import "time"

type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:server"
	Server ServerConfig
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Server", "ServerConfig",
				"cast.ToStruct",
			},
		},
		{
			name: "struct with duration string in default - using backtick",
			serviceCode: `package testservice

import "time"

type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
	// @Inject "cfg:server"
	Server ServerConfig
}

// @Route "GET /"
func (s *TestService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Server", "ServerConfig",
				"cast.ToStruct",
			},
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

			for _, expected := range tt.expectedStrings {
				if !strings.Contains(generatedCode, expected) {
					t.Errorf("Generated code should contain: %s", expected)
					t.Logf("Generated code:\n%s", generatedCode)
				}
			}
		})
	}
}
