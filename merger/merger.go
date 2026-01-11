package merger

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"qa-script/config"
	"qa-script/parser"
)

// MergeAndWrite combines CSV and Excel data according to the config and writes the output
func MergeAndWrite(csvData *parser.CSVData, excelData *parser.ExcelData, outputPath string, cfg *config.Config) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set the sheet name
	sheetName := cfg.Output.SheetName
	if sheetName == "" {
		sheetName = "Combined"
	}

	// Rename the default sheet
	f.SetSheetName("Sheet1", sheetName)

	// Convert source data to maps for easier lookup
	csvMaps := csvData.ToMap()
	excelMaps := excelData.ToMap()

	// Write headers based on column mapping
	for colIdx, mapping := range cfg.Output.ColumnMapping {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, mapping.OutputName)
	}

	// Determine the number of rows to write
	// Use the maximum of CSV and Excel rows
	maxRows := len(csvMaps)
	if len(excelMaps) > maxRows {
		maxRows = len(excelMaps)
	}

	// Write data rows
	for rowIdx := 0; rowIdx < maxRows; rowIdx++ {
		for colIdx, mapping := range cfg.Output.ColumnMapping {
			var value string

			switch mapping.Source {
			case "csv":
				if rowIdx < len(csvMaps) {
					if val, ok := csvMaps[rowIdx][mapping.SourceColumn]; ok && val != "" {
						value = val
					} else {
						value = mapping.Default
					}
				} else {
					value = mapping.Default
				}
			case "excel":
				if rowIdx < len(excelMaps) {
					if val, ok := excelMaps[rowIdx][mapping.SourceColumn]; ok && val != "" {
						value = val
					} else {
						value = mapping.Default
					}
				} else {
					value = mapping.Default
				}
			default:
				value = mapping.Default
			}

			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2) // +2 because row 1 is headers
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save the file
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save output file: %w", err)
	}

	return nil
}

// MergeByKey combines data from CSV and Excel based on a common key column
func MergeByKey(csvData *parser.CSVData, excelData *parser.ExcelData, outputPath string, cfg *config.Config, keyColumn string) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set the sheet name
	sheetName := cfg.Output.SheetName
	if sheetName == "" {
		sheetName = "Combined"
	}

	f.SetSheetName("Sheet1", sheetName)

	// Build index from Excel data by key
	excelIndex := make(map[string]map[string]string)
	for _, row := range excelData.ToMap() {
		if key, ok := row[keyColumn]; ok && key != "" {
			excelIndex[key] = row
		}
	}

	// Write headers
	for colIdx, mapping := range cfg.Output.ColumnMapping {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, mapping.OutputName)
	}

	// Write data rows, joining on the key
	csvMaps := csvData.ToMap()
	for rowIdx, csvRow := range csvMaps {
		keyValue := csvRow[keyColumn]
		excelRow := excelIndex[keyValue] // May be nil if no match

		for colIdx, mapping := range cfg.Output.ColumnMapping {
			var value string

			switch mapping.Source {
			case "csv":
				if val, ok := csvRow[mapping.SourceColumn]; ok && val != "" {
					value = val
				} else {
					value = mapping.Default
				}
			case "excel":
				if excelRow != nil {
					if val, ok := excelRow[mapping.SourceColumn]; ok && val != "" {
						value = val
					} else {
						value = mapping.Default
					}
				} else {
					value = mapping.Default
				}
			default:
				value = mapping.Default
			}

			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save output file: %w", err)
	}

	return nil
}
