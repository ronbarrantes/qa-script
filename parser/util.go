package parser

import "strings"

// trimStringSlice trims leading and trailing whitespace from all strings in a slice
func trimStringSlice(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = strings.TrimSpace(s)
	}
	return result
}
