package output

import (
	"encoding/csv"
	"fmt"
	"os"
)

// WriteCSV writes the grouped locations to a CSV file
// Each column is a title group with the title as the header
// Gap columns are inserted between groups based on data.Gap
// Size controls max rows per column before spillover
func WriteCSV(filePath string, data *OutputData) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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

	// Build headers with spillover columns and gap columns
	headers := make([]string, 0)
	for i, title := range groupTitles {
		// Add columns for this group (repeat title for spillover columns)
		for c := 0; c < groupColumns[i]; c++ {
			headers = append(headers, title)
		}
		// Add gap columns after each group except the last
		if i < len(groupTitles)-1 {
			for g := 0; g < data.Gap; g++ {
				headers = append(headers, "")
			}
		}
	}

	// Write header row
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write data rows
	for row := 0; row < maxRows; row++ {
		record := make([]string, 0)
		for i, title := range groupTitles {
			locs := data.Grouped[title]
			cols := groupColumns[i]

			// Fill each spillover column for this group
			for c := 0; c < cols; c++ {
				idx := c*data.Size + row
				if data.Size <= 0 {
					idx = row // no spillover, just use row index
				}
				if idx < len(locs) {
					record = append(record, locs[idx])
				} else {
					record = append(record, "")
				}
			}

			// Add gap columns after each group except the last
			if i < len(groupTitles)-1 {
				for g := 0; g < data.Gap; g++ {
					record = append(record, "")
				}
			}
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", row, err)
		}
	}

	return nil
}
