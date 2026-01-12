package parser

import (
	"encoding/csv"
	"fmt"
	"os"
)

// CSVData represents parsed CSV data with headers and rows
type CSVData struct {
	Headers []string
	Rows    [][]string
}

// ParseCSV reads a CSV file and returns the parsed data
func ParseCSV(filePath string) (*CSVData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	return &CSVData{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}

// GetColumnIndex returns the index of a column by name, or -1 if not found
func (c *CSVData) GetColumnIndex(columnName string) int {
	for i, header := range c.Headers {
		if header == columnName {
			return i
		}
	}
	return -1
}

// GetColumnValues returns all values for a given column name
func (c *CSVData) GetColumnValues(columnName string) ([]string, error) {
	idx := c.GetColumnIndex(columnName)
	if idx == -1 {
		return nil, fmt.Errorf("column %q not found", columnName)
	}

	values := make([]string, len(c.Rows))
	for i, row := range c.Rows {
		if idx < len(row) {
			values[i] = row[idx]
		}
	}
	return values, nil
}
