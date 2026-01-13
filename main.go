package main

import (
	"fmt"
	"log"

	"qa-script/location"
	"qa-script/parser"
)

func main() {
	// Parse the CSV file
	csv, err := parser.ParseCSV("mock_data/TEST_Locations.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Get the Location column values
	locations, err := csv.GetColumnValues("Location")
	if err != nil {
		log.Fatal(err)
	}

	// Get unique sorted locations
	uniqueLocations := location.UniqueAndSort(locations)

	// Log the results
	fmt.Println("=== Unique Locations (sorted) ===")
	for i, loc := range uniqueLocations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique locations (from %d rows)\n", len(uniqueLocations), len(csv.Rows))

	// Parse the Excel file
	excel, err := parser.ParseExcel("mock_data/TEST_Priorities.xlsx", "")
	if err != nil {
		log.Fatal(err)
	}

	// Get column indices
	tagIdx := excel.GetColumnIndex("Container Tag")
	if tagIdx == -1 {
		log.Fatal("Container Tag column not found")
	}
	locIdx := excel.GetColumnIndex("Current Location")
	if locIdx == -1 {
		log.Fatal("Current Location column not found")
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

	// Convert to sorted slice
	priorityLocations := location.UniqueAndSort(mapKeys(prioritySet))

	// Log the priority locations
	fmt.Println("\n=== Priority Locations (QA_HOLD_PICKING) ===")
	for i, loc := range priorityLocations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique priority locations\n", len(priorityLocations))
}

// mapKeys extracts keys from a set map
func mapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
