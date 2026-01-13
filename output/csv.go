package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"qa-script/rules"
)

// WriteCSV writes the grouped locations to a CSV file
// Each column is a title group with the title as the header
func WriteCSV(filePath string, titleOrder []string, grouped rules.TitleGroupedLocations) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Find the maximum number of rows needed
	maxRows := 0
	for _, title := range titleOrder {
		if locs := grouped[title]; len(locs) > maxRows {
			maxRows = len(locs)
		}
	}
	// Check unassigned too
	if locs := grouped["unassigned"]; len(locs) > maxRows {
		maxRows = len(locs)
	}

	// Build headers (titles + unassigned if it has items)
	headers := make([]string, 0, len(titleOrder)+1)
	headers = append(headers, titleOrder...)
	if len(grouped["unassigned"]) > 0 {
		headers = append(headers, "unassigned")
	}

	// Write header row
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write data rows
	for row := 0; row < maxRows; row++ {
		record := make([]string, len(headers))
		for col, title := range headers {
			locs := grouped[title]
			if row < len(locs) {
				record[col] = locs[row]
			} else {
				record[col] = ""
			}
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row %d: %w", row, err)
		}
	}

	return nil
}
