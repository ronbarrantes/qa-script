package output

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// WriteXLSX writes the grouped locations to an Excel file
// Each column is a title group with the title as the header
// Priority locations are highlighted in yellow
// Gap columns are inserted between groups based on data.ColumnGap
// MaxRows controls max rows per column before spillover
// Headers are merged across spillover columns
func WriteXLSX(filePath string, data *OutputData) (err error) {
	f := excelize.NewFile()
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close excel file: %w", closeErr)
		}
	}()

	sheetName := "Locations"
	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return fmt.Errorf("failed to set sheet name: %w", err)
	}

	// Create styles
	// Note: excelize expects hex colors WITHOUT the # prefix
	// Header: bold black text, centered horizontally
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return fmt.Errorf("failed to create header style: %w", err)
	}

	// Priority: bright yellow background (FFFF00 = pure yellow)
	priorityStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create priority style: %w", err)
	}

	// Build group titles (only including groups that have items)
	groupTitles := make([]string, 0, len(data.TitleOrder)+1)
	for _, title := range data.TitleOrder {
		if len(data.Grouped[title]) > 0 {
			groupTitles = append(groupTitles, title)
		}
	}
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

		// With spillover, max rows is capped at MaxRows (or actual count if less)
		rowsForGroup := len(locs)
		if data.MaxRows > 0 && rowsForGroup > data.MaxRows {
			rowsForGroup = data.MaxRows
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
		startCell, err := excelize.CoordinatesToCellName(startCol, 1)
		if err != nil {
			return fmt.Errorf("failed to get start cell name: %w", err)
		}
		endCell, err := excelize.CoordinatesToCellName(startCol+cols-1, 1)
		if err != nil {
			return fmt.Errorf("failed to get end cell name: %w", err)
		}

		if err := f.SetCellValue(sheetName, startCell, title); err != nil {
			return fmt.Errorf("failed to set header value: %w", err)
		}
		if err := f.SetCellStyle(sheetName, startCell, endCell, headerStyle); err != nil {
			return fmt.Errorf("failed to set header style: %w", err)
		}

		if cols > 1 {
			if err := f.MergeCell(sheetName, startCell, endCell); err != nil {
				return fmt.Errorf("failed to merge header cells: %w", err)
			}
		}

		// Calculate max width needed for this group's data
		maxDataWidth := len(title)
		for _, loc := range locs {
			if len(loc) > maxDataWidth {
				maxDataWidth = len(loc)
			}
		}

		// Set column widths for all columns in this group
		for c := 0; c < cols; c++ {
			colName, err := excelize.ColumnNumberToName(startCol + c)
			if err != nil {
				return fmt.Errorf("failed to get column name: %w", err)
			}
			// Add padding for comfortable reading
			width := float64(maxDataWidth + 2)
			if width < 15 {
				width = 15
			}
			if err := f.SetColWidth(sheetName, colName, colName, width); err != nil {
				return fmt.Errorf("failed to set column width: %w", err)
			}
		}

		// Write data with spillover
		for idx, loc := range locs {
			var locCol, locRow int
			if data.MaxRows > 0 {
				locCol = startCol + (idx / data.MaxRows)
				locRow = 2 + (idx % data.MaxRows)
			} else {
				locCol = startCol
				locRow = 2 + idx
			}

			cell, err := excelize.CoordinatesToCellName(locCol, locRow)
			if err != nil {
				return fmt.Errorf("failed to get cell name: %w", err)
			}
			if err := f.SetCellValue(sheetName, cell, loc); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}

			// Highlight priority locations in yellow
			if data.IsPriority(loc) {
				if err := f.SetCellStyle(sheetName, cell, cell, priorityStyle); err != nil {
					return fmt.Errorf("failed to set priority style: %w", err)
				}
			}
		}

		// Move to next group's starting column
		col = startCol + cols

		// Add gap columns after each group except the last
		if i < len(groupTitles)-1 {
			col += data.ColumnGap
		}
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
