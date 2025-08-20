package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

func main() {
	// Load schema
	schemaLoader := gojsonschema.NewReferenceLoader("file:///schema/lokstra.json")

	// Find YAML config files
	pattern := "cmd/projects/*/config/*.yaml"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error finding config files: %v\n", err)
		return
	}

	for _, configPath := range matches {
		fmt.Printf("Validating: %s\n", configPath)

		// Read YAML file
		yamlData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("  Error reading file: %v\n", err)
			continue
		}

		// Convert YAML to JSON for validation
		var yamlContent interface{}
		err = yaml.Unmarshal(yamlData, &yamlContent)
		if err != nil {
			fmt.Printf("  Error parsing YAML: %v\n", err)
			continue
		}

		jsonData, err := json.Marshal(yamlContent)
		if err != nil {
			fmt.Printf("  Error converting to JSON: %v\n", err)
			continue
		}

		// Validate against schema
		documentLoader := gojsonschema.NewBytesLoader(jsonData)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			fmt.Printf("  Validation error: %v\n", err)
			continue
		}

		if result.Valid() {
			fmt.Printf("  ✅ Valid configuration\n")
		} else {
			fmt.Printf("  ❌ Invalid configuration:\n")
			for _, desc := range result.Errors() {
				fmt.Printf("    - %s\n", desc)
			}
		}
		fmt.Println()
	}
}
