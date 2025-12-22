package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

// Helper function to create service and test generated code
func testInjectCfgValue(t *testing.T, testName, serviceCode string, expectedStrings []string) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(serviceCode), 0644); err != nil {
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

	for _, expected := range expectedStrings {
		if !strings.Contains(generatedCode, expected) {
			t.Errorf("%s: generated code should contain: %s", testName, expected)
		}
	}
}

func TestInjectCfgValue_AllTypes(t *testing.T) {
	testCases := []struct {
		name            string
		serviceCode     string
		expectedStrings []string
	}{
		{
			name: "BasicTypes",
			serviceCode: `package testservice

import "time"

// @RouterService name="basic-service", prefix="/api/basic"
type BasicService struct {
	// @InjectCfgValue "app.name"
	AppName string
	// @InjectCfgValue "app.port"
	Port int
	// @InjectCfgValue "app.timeout"
	Timeout time.Duration
}

// @Route "GET /"
func (s *BasicService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"AppName", "Port", "Timeout", `"time"`},
		},
		{
			name: "ByteSlice",
			serviceCode: `package testservice

// @RouterService name="byte-service", prefix="/api/bytes"
type ByteService struct {
	// @InjectCfgValue "secret"
	Secret []byte
}

// @Route "GET /"
func (s *ByteService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"Secret", "[]byte"},
		},
		{
			name: "StringSlice",
			serviceCode: `package testservice

// @RouterService name="string-slice-service", prefix="/api/strings"
type StringSliceService struct {
	// @InjectCfgValue "hosts"
	Hosts []string
}

// @Route "GET /"
func (s *StringSliceService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"Hosts", `"strconv"`, `"strings"`},
		},
		{
			name: "IntSlices",
			serviceCode: `package testservice

// @RouterService name="int-slice-service", prefix="/api/ints"
type IntSliceService struct {
	// @InjectCfgValue "ports"
	Ports []int
	// @InjectCfgValue "delays"
	Delays []int64
	// @InjectCfgValue "rates"
	Rates []float64
}

// @Route "GET /"
func (s *IntSliceService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"Ports", "Delays", "Rates"},
		},
		{
			name: "StructType",
			serviceCode: `package testservice

type DatabaseConfig struct {
	Host string
	Port int
}

// @RouterService name="struct-service", prefix="/api/struct"
type StructService struct {
	// @InjectCfgValue "database"
	DBConfig DatabaseConfig
}

// @Route "GET /"
func (s *StructService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"DBConfig", "DatabaseConfig", "github.com/primadi/lokstra/common/cast"},
		},
		{
			name: "StructSlice",
			serviceCode: `package testservice

type ServerConfig struct {
	Host string
	Port int
}

// @RouterService name="server-service", prefix="/api/servers"
type ServerService struct {
	// @InjectCfgValue "servers"
	Servers []ServerConfig
}

// @Route "GET /"
func (s *ServerService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{"Servers", "[]ServerConfig", "github.com/primadi/lokstra/common/cast"},
		},
		{
			name: "MixedTypes",
			serviceCode: `package testservice

import "time"

type AuthConfig struct {
	Provider string
}

// @RouterService name="mixed-service", prefix="/api/mixed"
type MixedService struct {
	// @InjectCfgValue "name"
	Name string
	// @InjectCfgValue "timeout"
	Timeout time.Duration
	// @InjectCfgValue "secret"
	Secret []byte
	// @InjectCfgValue "hosts"
	Hosts []string
	// @InjectCfgValue "ports"
	Ports []int
	// @InjectCfgValue "auth"
	Auth AuthConfig
}

// @Route "GET /"
func (s *MixedService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Name", "Timeout", "Secret", "Hosts", "Ports", "Auth",
				`"time"`, `"strconv"`, `"strings"`, "github.com/primadi/lokstra/common/cast",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testInjectCfgValue(t, tc.name, tc.serviceCode, tc.expectedStrings)
		})
	}
}
