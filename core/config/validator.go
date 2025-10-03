package config

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// ValidateConfig validates a Config struct against the JSON schema
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Convert Config to JSON for validation
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Load schema
	schemaLoader := gojsonschema.NewStringLoader(configSchema)
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		errMsg := "YAML configuration validation failed:\n"
		for _, desc := range result.Errors() {
			errMsg += fmt.Sprintf("  - %s: %s\n", desc.Field(), desc.Description())
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// ValidateYAMLString validates a YAML string against the JSON schema
func ValidateYAMLString(yamlContent string) error {
	return ValidateYAMLBytes([]byte(yamlContent))
}

// ValidateYAMLBytes validates YAML bytes against the JSON schema
func ValidateYAMLBytes(yamlContent []byte) error {
	var config Config
	if err := yaml.Unmarshal(yamlContent, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return ValidateConfig(&config)
}
