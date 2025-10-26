package config

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid complete config",
			yamlInput: `
servers:
  - name: main-server
    baseUrl: http://localhost
    deployment-id: dev
    apps:
      - name: api
        addr: :8080
        listener-type: default
        routers:
          - product-api
          - order-api

services:
  - name: database
    type: postgres
    enable: true
    config:
      host: localhost
      port: 5432

middlewares:
  - name: cors
    type: cors
    enable: true
    config:
      allowOrigins: ["*"]

configs:
  - name: app-version
    value: "1.0.0"
`,
			wantError: false,
		},
		{
			name: "missing required server name",
			yamlInput: `
servers:
  - baseUrl: http://localhost
    apps:
      - addr: :8080
`,
			wantError: true,
			errorMsg:  "name",
		},
		{
			name: "missing required app addr",
			yamlInput: `
servers:
  - name: test-server
    apps:
      - name: api
`,
			wantError: true,
			errorMsg:  "addr",
		},
		{
			name: "invalid baseUrl format",
			yamlInput: `
servers:
  - name: test-server
    baseUrl: invalid-url
    apps:
      - addr: :8080
`,
			wantError: true,
			errorMsg:  "baseUrl",
		},
		{
			name: "service without type (valid - defaults to name)",
			yamlInput: `
services:
  - name: database
    config:
      host: localhost
`,
			wantError: false,
		},
		{
			name: "empty server apps array",
			yamlInput: `
servers:
  - name: test-server
    apps: []
`,
			wantError: true,
			errorMsg:  "apps",
		},
		{
			name: "valid minimal config",
			yamlInput: `
servers:
  - name: test-server
    apps:
      - addr: :8080
`,
			wantError: false,
		},
		{
			name: "empty config is valid",
			yamlInput: `
# empty config
`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAMLString(tt.yamlInput)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateYAMLString() expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateYAMLString() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateYAMLString() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateConfigStruct(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config struct",
			config: &Config{
				Servers: []*Server{
					{
						Name:    "test-server",
						BaseUrl: "http://localhost",
						Apps: []*App{
							{
								Name: "api",
								Addr: ":8080",
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name:      "empty config struct",
			config:    &Config{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.wantError && err == nil {
				t.Errorf("ValidateConfig() expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateConfig() unexpected error = %v", err)
			}
		})
	}
}
