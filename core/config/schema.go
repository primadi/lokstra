package config

import "embed"

//go:embed lokstra.json
var configSchemaFs embed.FS

var configSchema string

func init() {
	data, err := configSchemaFs.ReadFile("lokstra.json")
	if err != nil {
		panic(err)
	}
	configSchema = string(data)
}
