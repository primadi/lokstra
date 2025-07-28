package print_service

type GlobalConfig struct {
	PrinterTypes map[string]map[string][]byte `yaml:"printer_types"`
	Printers     []PrinterConfig              `yaml:"printers"`
	Default      string                       `yaml:"default"`
}

type PrinterConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Device string `yaml:"device"`
}
