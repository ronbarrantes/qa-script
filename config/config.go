package config

import (
	"fmt"
	"os"
	"strings"

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
	Name   string     `yaml:"name"`
	Values StringList `yaml:"values"`
}

// StringList is a custom type that can unmarshal from either a YAML array or
// a comma-separated string (e.g., "a, b, c" or [a, b, c]).
type StringList []string

// UnmarshalYAML implements custom unmarshaling to support both formats:
// - YAML array: [a, b, c] or multi-line list
// - Comma-separated string: "a, b, c"
func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshal as a sequence (array) first
	if value.Kind == yaml.SequenceNode {
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*s = list
		return nil
	}

	// Otherwise, treat as a comma-separated string
	if value.Kind == yaml.ScalarNode {
		var str string
		if err := value.Decode(&str); err != nil {
			return err
		}
		// Split by comma and trim whitespace
		parts := strings.Split(str, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		*s = result
		return nil
	}

	return fmt.Errorf("values must be a list or comma-separated string")
}

// MarshalYAML outputs the StringList as a comma-separated string for cleaner YAML.
func (s StringList) MarshalYAML() (interface{}, error) {
	return strings.Join(s, ", "), nil
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
	// Add header comment and example structure
	content := `# QA Script configuration
#
# groups:
#   - name: <group name shown in Excel header>
#     values: <comma-separated location codes, e.g., a, b, c, lud, prm>
#
# size: maximum number of rows before spilling into a new column per group
#

groups: []

size: 20

output_sheet: Groups
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
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
