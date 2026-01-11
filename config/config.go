package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration for grouping locations into an output Excel.
type Config struct {
	// Groups define how CSV locations are grouped in the output Excel.
	Groups []Group `yaml:"groups"`

	// Size is the maximum number of rows in a column for a group before spilling into the next column.
	Size int `yaml:"size"`

	// ExcelSheet optionally selects the sheet to read from the input Excel file.
	// Empty means the first sheet.
	ExcelSheet string `yaml:"excel_sheet,omitempty"`

	// OutputSheet optionally sets the output sheet name.
	// Empty means "Groups".
	OutputSheet string `yaml:"output_sheet,omitempty"`
}

// Group represents one output group (e.g. "a-f") and which location codes belong to it.
type Group struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values"`
}

// LoadConfig reads and parses a YAML configuration file
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GenerateTemplate creates a default YAML template file.
func GenerateTemplate(filePath string) error {
	defaultConfig := Config{Size: 20, OutputSheet: "Groups", Groups: []Group{}}
	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# QA Script configuration
#
# groups:
#   - name: <group name shown in Excel header>
#     values: [<location-code>, <location-code>, ...]
#
# size: maximum number of rows before spilling into a new column per group
#
`

	if err := os.WriteFile(filePath, []byte(header+string(data)), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveConfig writes the configuration to a YAML file
func SaveConfig(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
