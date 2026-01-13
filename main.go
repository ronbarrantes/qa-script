package main

import (
	"fmt"
	"log"
	"sort"

	"qa-script/processor"
)

func main() {
	result, err := processor.Process(
		"mock_data/TEST_Locations.csv",
		"mock_data/TEST_Priorities.xlsx",
		"rules.yaml",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Log locations
	fmt.Println("=== Unique Locations (sorted) ===")
	for i, loc := range result.Locations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique locations (from %d rows)\n", len(result.Locations), result.TotalCSVRows)

	// Log priority locations
	fmt.Println("\n=== Priority Locations (QA_HOLD_PICKING) ===")
	for i, loc := range result.PriorityLocations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique priority locations\n", len(result.PriorityLocations))

	// Log grouped locations by value keys
	fmt.Println("\n=== Grouped Locations (by value) ===")
	var keys []string
	for key := range result.GroupedLocations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		locs := result.GroupedLocations[key]
		if len(locs) > 0 {
			fmt.Printf("\n[%s] (%d locations)\n", key, len(locs))
			for _, loc := range locs {
				fmt.Printf("  - %s\n", loc)
			}
		}
	}

	// Log grouped locations by title
	fmt.Println("\n=== Grouped Locations (by title) ===")
	var titles []string
	for title := range result.TitleGroupedLocations {
		titles = append(titles, title)
	}
	sort.Strings(titles)

	for _, title := range titles {
		locs := result.TitleGroupedLocations[title]
		if len(locs) > 0 {
			fmt.Printf("\n[%s] (%d locations)\n", title, len(locs))
			for _, loc := range locs {
				fmt.Printf("  - %s\n", loc)
			}
		}
	}
}
