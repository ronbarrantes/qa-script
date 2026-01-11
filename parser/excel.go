package parser

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ExcelData represents parsed Excel data with headers and rows
type ExcelData struct {
	Headers []string
	Rows    [][]string
	Sheet   string
}

// ParseExcel reads an Excel file and returns the parsed data
func ParseExcel(filePath string, sheetName string) (*ExcelData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// If no sheet name specified, use the first sheet
	if sheetName == "" {
		sheetName = f.GetSheetName(0)
	}

	// Get all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %q: %w", sheetName, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %q is empty", sheetName)
	}

	// First row is headers
	data := &ExcelData{
		Headers: rows[0],
		Rows:    rows[1:],
		Sheet:   sheetName,
	}

	return data, nil
}

// GetColumnIndex returns the index of a column by name
func (e *ExcelData) GetColumnIndex(columnName string) int {
	for i, header := range e.Headers {
		if header == columnName {
			return i
		}
	}
	return -1
}

// GetColumnValues returns all values for a given column name
func (e *ExcelData) GetColumnValues(columnName string) ([]string, error) {
	idx := e.GetColumnIndex(columnName)
	if idx == -1 {
		return nil, fmt.Errorf("column %q not found", columnName)
	}

	values := make([]string, len(e.Rows))
	for i, row := range e.Rows {
		if idx < len(row) {
			values[i] = row[idx]
		}
	}
	return values, nil
}

// ToMap converts rows to a slice of maps with header keys
func (e *ExcelData) ToMap() []map[string]string {
	result := make([]map[string]string, len(e.Rows))
	for i, row := range e.Rows {
		rowMap := make(map[string]string)
		for j, header := range e.Headers {
			if j < len(row) {
				rowMap[header] = row[j]
			}
		}
		result[i] = rowMap
	}
	return result
}

// GetSheetNames returns all sheet names in the Excel file
func GetSheetNames(filePath string) ([]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	return f.GetSheetList(), nil
}
