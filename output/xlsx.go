package output

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// WriteXLSX writes the grouped locations to an Excel file
// Each column is a title group with the title as the header
// Priority locations are highlighted in yellow
// Gap columns are inserted between groups based on data.Gap
// Size controls max rows per column before spillover
// Headers are merged across spillover columns
func WriteXLSX(filePath string, data *OutputData) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Locations"
	f.SetSheetName("Sheet1", sheetName)

	// Create styles
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#333333"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
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

	// Build group titles (titles + unassigned if it has items)
	groupTitles := make([]string, 0, len(data.TitleOrder)+1)
	groupTitles = append(groupTitles, data.TitleOrder...)
	if len(data.Grouped["unassigned"]) > 0 {
		groupTitles = append(groupTitles, "unassigned")
	}

	// Calculate columns needed for each group and max rows
	groupColumns := make([]int, len(groupTitles))
	maxRows := 0
	for i, title := range groupTitles {
		locs := data.Grouped[title]
		cols := data.ColumnsNeeded(len(locs))
		groupColumns[i] = cols

		// With spillover, max rows is capped at Size (or actual count if less)
		rowsForGroup := len(locs)
		if data.Size > 0 && rowsForGroup > data.Size {
			rowsForGroup = data.Size
		}
		if rowsForGroup > maxRows {
			maxRows = rowsForGroup
		}
	}

	// Track current column position (1-based for Excel)
	col := 1

	// Write each group
	for i, title := range groupTitles {
		locs := data.Grouped[title]
		cols := groupColumns[i]
		startCol := col

		// Write header - merge if multiple columns
		startCell, _ := excelize.CoordinatesToCellName(startCol, 1)
		endCell, _ := excelize.CoordinatesToCellName(startCol+cols-1, 1)

		f.SetCellValue(sheetName, startCell, title)
		f.SetCellStyle(sheetName, startCell, endCell, headerStyle)

		if cols > 1 {
			f.MergeCell(sheetName, startCell, endCell)
		}

		// Set column widths for all columns in this group
		for c := 0; c < cols; c++ {
			colName, _ := excelize.ColumnNumberToName(startCol + c)
			width := float64(len(title) + 5)
			if width < 15 {
				width = 15
			}
			f.SetColWidth(sheetName, colName, colName, width)
		}

		// Write data with spillover
		for idx, loc := range locs {
			var locCol, locRow int
			if data.Size > 0 {
				locCol = startCol + (idx / data.Size)
				locRow = 2 + (idx % data.Size)
			} else {
				locCol = startCol
				locRow = 2 + idx
			}

			cell, _ := excelize.CoordinatesToCellName(locCol, locRow)
			f.SetCellValue(sheetName, cell, loc)

			// Highlight priority locations in yellow
			if data.IsPriority(loc) {
				f.SetCellStyle(sheetName, cell, cell, priorityStyle)
			}
		}

		// Move to next group's starting column
		col = startCol + cols

		// Add gap columns after each group except the last
		if i < len(groupTitles)-1 {
			col += data.Gap
		}
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
