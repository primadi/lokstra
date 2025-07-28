package print_service

import (
	"os"

	"gopkg.in/yaml.v3"
)

var escapeMapByType map[string]map[string][]byte
var printerMap map[string]PrinterConfig
var defaultPrinter string

func LoadPrinterConfig(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	printerMap = map[string]PrinterConfig{}
	for _, p := range cfg.Printers {
		printerMap[p.Name] = p
	}
	escapeMapByType = cfg.PrinterTypes
	defaultPrinter = cfg.Default
	return nil
}
