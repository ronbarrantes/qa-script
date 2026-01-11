package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
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
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// First row is headers
	data := &CSVData{
		Headers: records[0],
		Rows:    records[1:],
	}

	return data, nil
}

// GetColumnIndex returns the index of a column by name
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

// GetUniqueColumnValues returns de-duplicated column values (preserving first-seen order).
func (c *CSVData) GetUniqueColumnValues(columnName string) ([]string, error) {
	values, err := c.GetColumnValues(columnName)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out, nil
}

// ToMap converts rows to a slice of maps with header keys
func (c *CSVData) ToMap() []map[string]string {
	result := make([]map[string]string, len(c.Rows))
	for i, row := range c.Rows {
		rowMap := make(map[string]string)
		for j, header := range c.Headers {
			if j < len(row) {
				rowMap[header] = row[j]
			}
		}
		result[i] = rowMap
	}
	return result
}
