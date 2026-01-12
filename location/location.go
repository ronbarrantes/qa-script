package location

import (
	"sort"
	"strings"
)

// GetSortKey extracts the part after the colon for sorting
// e.g., "PS2:CL106" -> "CL106"
func GetSortKey(loc string) string {
	if idx := strings.Index(loc, ":"); idx != -1 {
		return loc[idx+1:]
	}
	return loc
}

// UniqueAndSort takes a slice of locations, removes duplicates, and sorts them
// by the part after the colon
func UniqueAndSort(locations []string) []string {
	// Use a set to remove duplicates
	locationSet := make(map[string]struct{})
	for _, loc := range locations {
		locationSet[loc] = struct{}{}
	}

	// Convert to slice
	unique := make([]string, 0, len(locationSet))
	for loc := range locationSet {
		unique = append(unique, loc)
	}

	// Sort by the part after the colon
	sort.Slice(unique, func(i, j int) bool {
		return GetSortKey(unique[i]) < GetSortKey(unique[j])
	})

	return unique
}
