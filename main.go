package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

// getSortKey extracts the part after the colon for sorting
func getSortKey(location string) string {
	if idx := strings.Index(location, ":"); idx != -1 {
		return location[idx+1:]
	}
	return location
}

func main() {
	// Open the CSV file
	file, err := os.Open("mock_data/TEST_Locations.csv")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV:", err)
	}

	// Find the Location column index
	headers := records[0]
	locationIdx := -1
	for i, header := range headers {
		if header == "Location" {
			locationIdx = i
			break
		}
	}

	if locationIdx == -1 {
		log.Fatal("Location column not found")
	}

	// Use a set (map) to store unique locations
	locationSet := make(map[string]struct{})
	for _, row := range records[1:] {
		location := row[locationIdx]
		locationSet[location] = struct{}{}
	}

	// Convert set to sorted slice (sort by part after colon)
	locations := make([]string, 0, len(locationSet))
	for loc := range locationSet {
		locations = append(locations, loc)
	}
	sort.Slice(locations, func(i, j int) bool {
		return getSortKey(locations[i]) < getSortKey(locations[j])
	})

	// Log the unique sorted locations
	fmt.Println("=== Unique Locations (sorted) ===")
	for i, loc := range locations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique locations (from %d rows)\n", len(locations), len(records)-1)
}
