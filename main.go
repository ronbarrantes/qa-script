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
}
