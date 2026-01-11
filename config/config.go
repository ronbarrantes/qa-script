package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML template configuration
type Config struct {
	// Version of the configuration schema
	Version string `yaml:"version"`

	// CSV configuration
	CSV CSVConfig `yaml:"csv"`

	// Excel configuration
	Excel ExcelConfig `yaml:"excel"`

	// Output configuration
	Output OutputConfig `yaml:"output"`

	// ExcelSheet is the specific sheet to read from the Excel input file
	ExcelSheet string `yaml:"excel_sheet"`
}

// CSVConfig holds CSV-specific settings
type CSVConfig struct {
	// Columns to include from CSV (empty means all)
	Columns []string `yaml:"columns"`

	// SkipRows is the number of rows to skip at the beginning
	SkipRows int `yaml:"skip_rows"`

	// Delimiter is the CSV delimiter (default: comma)
	Delimiter string `yaml:"delimiter"`
}

// ExcelConfig holds Excel-specific settings
type ExcelConfig struct {
	// Columns to include from Excel (empty means all)
	Columns []string `yaml:"columns"`

	// Sheet is the sheet name to read from
	Sheet string `yaml:"sheet"`

	// SkipRows is the number of rows to skip at the beginning
	SkipRows int `yaml:"skip_rows"`
}

// OutputConfig holds output file settings
type OutputConfig struct {
	// SheetName is the name of the output sheet
	SheetName string `yaml:"sheet_name"`

	// ColumnMapping maps source columns to output columns
	// Key: output column name, Value: source specification (e.g., "csv:ColumnA" or "excel:ColumnB")
	ColumnMapping []ColumnMap `yaml:"column_mapping"`
}

// ColumnMap represents a mapping from source to output column
type ColumnMap struct {
	// OutputName is the name of the column in the output file
	OutputName string `yaml:"output_name"`

	// Source specifies where the data comes from: "csv" or "excel"
	Source string `yaml:"source"`

	// SourceColumn is the column name in the source file
	SourceColumn string `yaml:"source_column"`

	// Default value if the source column is empty
	Default string `yaml:"default,omitempty"`
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

// GenerateTemplate creates a default YAML template file
func GenerateTemplate(filePath string) error {
	defaultConfig := Config{
		Version: "1.0",
		CSV: CSVConfig{
			Columns:   []string{}, // Empty means include all
			SkipRows:  0,
			Delimiter: ",",
		},
		Excel: ExcelConfig{
			Columns:  []string{}, // Empty means include all
			Sheet:    "",         // Empty means first sheet
			SkipRows: 0,
		},
		ExcelSheet: "", // Empty means first sheet
		Output: OutputConfig{
			SheetName: "Combined",
			ColumnMapping: []ColumnMap{
				{
					OutputName:   "ID",
					Source:       "csv",
					SourceColumn: "id",
					Default:      "",
				},
				{
					OutputName:   "Name",
					Source:       "excel",
					SourceColumn: "name",
					Default:      "N/A",
				},
				{
					OutputName:   "Value",
					Source:       "csv",
					SourceColumn: "value",
					Default:      "0",
				},
			},
		},
	}

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# QA Script Configuration Template
# This file defines how CSV and Excel files are combined into the output.
#
# Modify the column_mapping section to specify:
#   - output_name: Column name in the output Excel file
#   - source: Where the data comes from ("csv" or "excel")
#   - source_column: The column name in the source file
#   - default: Optional default value if source is empty
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
