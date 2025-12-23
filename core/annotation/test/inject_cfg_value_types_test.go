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
		{
			name: "StructWithComplexProperties",
			serviceCode: `package testservice

import "time"

type NestedConfig struct {
	Host string
	Port int
}

type ComplexConfig struct {
	Name     string
	Secret   []byte
	Timeout  time.Duration
	Hosts    []string
	Ports    []int
	Nested   NestedConfig
	Servers  []NestedConfig
}

// @RouterService name="complex-service", prefix="/api/complex"
type ComplexService struct {
	// @InjectCfgValue "config"
	Config ComplexConfig
}

// @Route "GET /"
func (s *ComplexService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Config", "ComplexConfig",
				"github.com/primadi/lokstra/common/cast",
				"cast.ToStruct", // Uses cast.ToStruct for struct conversion
			},
		},
		{
			name: "StructWithDefaultValue",
			serviceCode: `package testservice

type AppConfig struct {
	Name string
	Port int
}

// @RouterService name="default-service", prefix="/api/default"
type DefaultService struct {
	// @InjectCfgValue key="appconfig", default="AppConfig{Name: \"myapp\", Port: 8080}"
	Config AppConfig
}

// @Route "GET /"
func (s *DefaultService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Config", "AppConfig",
				"github.com/primadi/lokstra/common/cast",
				"cast.ToStruct",
				`return AppConfig{Name: \"myapp\", Port: 8080}`, // Default value (with escaped quotes)
			},
		},
		{
			name: "StructWithCustomUnmarshalJSON",
			serviceCode: `package testservice

import (
	"encoding/json"
	"time"
)

// CustomDate implements json.Unmarshaler for flexible date parsing
type CustomDate struct {
	time.Time
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if t, err := time.Parse("2006-01-02", str); err == nil {
			cd.Time = t
			return nil
		}
	}
	return json.Unmarshal(data, &cd.Time)
}

type ScheduleConfig struct {
	EventName string
	StartDate CustomDate
	Duration  time.Duration
}

// @RouterService name="schedule-service", prefix="/api/schedule"
type ScheduleService struct {
	// @InjectCfgValue "schedule"
	Config ScheduleConfig
}

// @Route "GET /"
func (s *ScheduleService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Config", "ScheduleConfig",
				"github.com/primadi/lokstra/common/cast",
				"cast.ToStruct",
			},
		},
		{
			name: "StructWithCustomUnmarshalJSON_AndDefault",
			serviceCode: `package testservice

import (
	"encoding/json"
	"time"
)

// CustomDate implements json.Unmarshaler for flexible date parsing
type CustomDate struct {
	time.Time
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if t, err := time.Parse("2006-01-02", str); err == nil {
			cd.Time = t
			return nil
		}
	}
	return json.Unmarshal(data, &cd.Time)
}

type ScheduleConfig struct {
	EventName string
	StartDate CustomDate
	Duration  time.Duration
}

// @RouterService name="schedule-service", prefix="/api/schedule"
type ScheduleService struct {
	// @InjectCfgValue key="schedule", default="ScheduleConfig{EventName: \"DefaultEvent\", Duration: 3600000000000}"
	Config ScheduleConfig
}

// @Route "GET /"
func (s *ScheduleService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Config", "ScheduleConfig",
				"github.com/primadi/lokstra/common/cast",
				"cast.ToStruct",
				`return ScheduleConfig{EventName: \"DefaultEvent\", Duration: 3600000000000}`,
			},
		},
		{
			name: "StructWithDefaultValue_UsingBacktick",
			serviceCode: `package testservice

type AppConfig struct {
	Name string
	Port int
}

// @RouterService name="default-service", prefix="/api/default"
type DefaultService struct {
	// @InjectCfgValue key="appconfig", default=` + "`AppConfig{Name: \"myapp\", Port: 8080}`" + `
	Config AppConfig
}

// @Route "GET /"
func (s *DefaultService) GetInfo() string { return "info" }
`,
			expectedStrings: []string{
				"Config", "AppConfig",
				"github.com/primadi/lokstra/common/cast",
				"cast.ToStruct",
				`return AppConfig{Name: "myapp", Port: 8080}`, // Backtick: no escaped quotes!
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testInjectCfgValue(t, tc.name, tc.serviceCode, tc.expectedStrings)
		})
	}
}
