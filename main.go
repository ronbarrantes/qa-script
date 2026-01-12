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

	// Log the Excel data
	fmt.Println("\n=== Priorities (Excel) ===")
	fmt.Printf("Sheet: %s\n", excel.Sheet)
	fmt.Printf("Headers: %v\n", excel.Headers)
	fmt.Println("---")
	for i, row := range excel.Rows {
		fmt.Printf("%d: %v\n", i+1, row)
	}
	fmt.Printf("\nTotal: %d rows\n", len(excel.Rows))
}
