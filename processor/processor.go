package processor

import (
	"fmt"

	"qa-script/location"
	"qa-script/parser"
	"qa-script/rules"
)

// Result holds the processed data from both files
type Result struct {
	Locations         []string                 // Unique sorted locations from CSV
	PriorityLocations []string                 // Unique sorted locations with QA_HOLD_PICKING
	GroupedLocations  rules.GroupedLocations   // Locations grouped by rules
	TotalCSVRows      int
	TotalExcelRows    int
}

// LoadLocations reads the CSV file and returns unique sorted locations
func LoadLocations(csvPath string) ([]string, int, error) {
	csv, err := parser.ParseCSV(csvPath)
	if err != nil {
		return nil, 0, err
	}

	locations, err := csv.GetColumnValues("Location")
	if err != nil {
		return nil, 0, err
	}

	uniqueLocations := location.UniqueAndSort(locations)
	return uniqueLocations, len(csv.Rows), nil
}

// LoadPriorities reads the Excel file and returns unique sorted locations
// filtered by Container Tag == QA_HOLD_PICKING
func LoadPriorities(excelPath string) ([]string, int, error) {
	excel, err := parser.ParseExcel(excelPath, "")
	if err != nil {
		return nil, 0, err
	}

	tagIdx := excel.GetColumnIndex("Container Tag")
	if tagIdx == -1 {
		return nil, 0, fmt.Errorf("column %q not found", "Container Tag")
	}

	locIdx := excel.GetColumnIndex("Current Location")
	if locIdx == -1 {
		return nil, 0, fmt.Errorf("column %q not found", "Current Location")
	}

	// Filter by QA_HOLD_PICKING and extract unique Current Locations
	prioritySet := make(map[string]struct{})
	for _, row := range excel.Rows {
		if tagIdx < len(row) && row[tagIdx] == "QA_HOLD_PICKING" {
			if locIdx < len(row) {
				prioritySet[row[locIdx]] = struct{}{}
			}
		}
	}

	// Convert to slice and sort
	priorities := make([]string, 0, len(prioritySet))
	for loc := range prioritySet {
		priorities = append(priorities, loc)
	}

	return location.UniqueAndSort(priorities), len(excel.Rows), nil
}

// Process loads both files and returns the combined result
func Process(csvPath, excelPath, rulesPath string) (*Result, error) {
	locations, csvRows, err := LoadLocations(csvPath)
	if err != nil {
		return nil, fmt.Errorf("loading locations: %w", err)
	}

	priorities, excelRows, err := LoadPriorities(excelPath)
	if err != nil {
		return nil, fmt.Errorf("loading priorities: %w", err)
	}

	// Load rules and group locations
	config, err := rules.LoadConfig(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("loading rules: %w", err)
	}

	grouped := rules.GroupLocations(locations, config)

	return &Result{
		Locations:         locations,
		PriorityLocations: priorities,
		GroupedLocations:  grouped,
		TotalCSVRows:      csvRows,
		TotalExcelRows:    excelRows,
	}, nil
}
