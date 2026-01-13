package output

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// WriteXLSX writes the grouped locations to an Excel file
// Each column is a title group with the title as the header
// Priority locations are highlighted in yellow
func WriteXLSX(filePath string, data *OutputData) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Locations"
	f.SetSheetName("Sheet1", sheetName)

	// Create styles
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDDDDD"}, Pattern: 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create header style: %w", err)
	}

	priorityStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFF00"}, Pattern: 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create priority style: %w", err)
	}

	// Build headers (titles + unassigned if it has items)
	headers := make([]string, 0, len(data.TitleOrder)+1)
	headers = append(headers, data.TitleOrder...)
	if len(data.Grouped["unassigned"]) > 0 {
		headers = append(headers, "unassigned")
	}

	// Find the maximum number of rows needed
	maxRows := 0
	for _, title := range headers {
		if locs := data.Grouped[title]; len(locs) > maxRows {
			maxRows = len(locs)
		}
	}

	// Write headers (row 1)
	for col, title := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cell, title)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Write data rows (starting at row 2)
	for row := 0; row < maxRows; row++ {
		for col, title := range headers {
			locs := data.Grouped[title]
			if row < len(locs) {
				loc := locs[row]
				cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
				f.SetCellValue(sheetName, cell, loc)

				// Highlight priority locations in yellow
				if data.IsPriority(loc) {
					f.SetCellStyle(sheetName, cell, cell, priorityStyle)
				}
			}
		}
	}

	// Auto-fit column widths (approximate)
	for col, title := range headers {
		colName, _ := excelize.ColumnNumberToName(col + 1)
		// Set width based on header length or a minimum
		width := float64(len(title) + 5)
		if width < 15 {
			width = 15
		}
		f.SetColWidth(sheetName, colName, colName, width)
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
