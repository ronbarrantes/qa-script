// Package rules defines parsing, sorting, and grouping rules for location strings.
//
// Location Format:
//   - Full format: PREFIX:CODE (e.g., "SS4:GF225.B")
//   - PREFIX (before colon): Ignored for sorting/grouping purposes
//   - CODE (after colon): Used for sorting and grouping
//
// Grouping Rules:
//   - 2-letter prefix (e.g., "GF" in "GF225.B"): Belongs to a specific group
//     determined by the first letter (e.g., "GF" -> "G" group)
//   - 3-letter prefix (e.g., "GFT" in "GFT245.B"): "Matching group" that can
//     be assigned to any of multiple groups defined in the YAML configuration
package rules

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// LocationCode represents the parsed components of a location string.
type LocationCode struct {
	Original string // The full original location string (e.g., "SS4:GF225.B")
	Prefix   string // Everything before the colon (e.g., "SS4")
	Code     string // Everything after the colon (e.g., "GF225.B")
	Letters  string // The letter prefix of the code (e.g., "GF" or "GFT")
	Number   string // The numeric portion of the code (e.g., "225")
	Suffix   string // Any suffix after the number (e.g., ".B")
}

// CodeType indicates whether a location code is a standard group or a matching group.
type CodeType int

const (
	// StandardGroup is a 2-letter code that belongs to a single specific group.
	StandardGroup CodeType = iota
	// MatchingGroup is a 3-letter code that can match multiple groups.
	MatchingGroup
	// UnknownGroup is a code that doesn't fit the standard patterns.
	UnknownGroup
)

// codePattern matches the location code format after the colon.
// Captures: (letters)(numbers)(suffix)
var codePattern = regexp.MustCompile(`^([A-Za-z]+)(\d+)(.*)$`)

// ParseLocation parses a full location string into its components.
func ParseLocation(location string) LocationCode {
	location = strings.TrimSpace(location)
	lc := LocationCode{Original: location}

	// Split on colon - everything after is the relevant code
	if idx := strings.Index(location, ":"); idx != -1 {
		lc.Prefix = location[:idx]
		lc.Code = location[idx+1:]
	} else {
		// No colon found, treat entire string as code
		lc.Code = location
	}

	// Parse the code portion
	matches := codePattern.FindStringSubmatch(lc.Code)
	if len(matches) >= 4 {
		lc.Letters = strings.ToUpper(matches[1])
		lc.Number = matches[2]
		lc.Suffix = matches[3]
	} else {
		// Couldn't parse, use the whole code as letters
		lc.Letters = strings.ToUpper(lc.Code)
	}

	return lc
}

// GetCodeType returns the type of the location code based on letter count.
func (lc LocationCode) GetCodeType() CodeType {
	letterCount := len(lc.Letters)
	switch {
	case letterCount == 2:
		return StandardGroup
	case letterCount >= 3:
		return MatchingGroup
	default:
		return UnknownGroup
	}
}

// GetPrimaryGroup returns the primary group identifier for a location.
// For 2-letter codes, this is the first letter (e.g., "GF" -> "G").
// For 3-letter codes, this is the full 3-letter code (e.g., "GFT" -> "GFT").
func (lc LocationCode) GetPrimaryGroup() string {
	if len(lc.Letters) == 0 {
		return ""
	}

	switch lc.GetCodeType() {
	case StandardGroup:
		// 2-letter code: group by first letter
		return string(lc.Letters[0])
	case MatchingGroup:
		// 3-letter code: use full code as group identifier
		return lc.Letters
	default:
		return lc.Letters
	}
}

// IsMatchingGroup returns true if this is a 3-letter matching group code.
func (lc LocationCode) IsMatchingGroup() bool {
	return lc.GetCodeType() == MatchingGroup
}

// SortKey returns the sortable portion of the location (everything after the colon).
func (lc LocationCode) SortKey() string {
	return lc.Code
}

// MatchesGroupValue checks if this location matches a given group value.
// Rules:
//   - Single letter value (e.g., "G"): Matches if first letter of code matches
//   - Multi-letter value (e.g., "GFT"): Matches if full Letters equals the value
func (lc LocationCode) MatchesGroupValue(groupValue string) bool {
	groupValue = strings.ToUpper(strings.TrimSpace(groupValue))
	if groupValue == "" || lc.Letters == "" {
		return false
	}

	// Single letter: prefix match (e.g., "G" matches "GF", "GA", etc.)
	if len(groupValue) == 1 {
		return strings.HasPrefix(lc.Letters, groupValue)
	}

	// Multi-letter: exact match
	return lc.Letters == groupValue
}

// SortLocations sorts a slice of location strings by their sort key (after the colon).
// The sort is case-insensitive and handles alphanumeric codes intelligently.
func SortLocations(locations []string) []string {
	sorted := make([]string, len(locations))
	copy(sorted, locations)

	sort.Slice(sorted, func(i, j int) bool {
		a := ParseLocation(sorted[i])
		b := ParseLocation(sorted[j])
		return compareLocationCodes(a, b) < 0
	})

	return sorted
}

// compareLocationCodes compares two LocationCode values for sorting.
// Returns negative if a < b, positive if a > b, zero if equal.
func compareLocationCodes(a, b LocationCode) int {
	// First compare by letters
	if cmp := strings.Compare(a.Letters, b.Letters); cmp != 0 {
		return cmp
	}

	// Then compare by number (numerically if possible)
	aNum := parseNumber(a.Number)
	bNum := parseNumber(b.Number)
	if aNum != bNum {
		if aNum < bNum {
			return -1
		}
		return 1
	}

	// Finally compare by suffix
	return strings.Compare(a.Suffix, b.Suffix)
}

// parseNumber extracts a numeric value from a string, returns 0 if not parseable.
func parseNumber(s string) int {
	var n int
	for _, r := range s {
		if unicode.IsDigit(r) {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

// ExtractLetterCode is a helper that returns just the letter portion of a location.
// This is useful for quick grouping checks.
func ExtractLetterCode(location string) (string, bool) {
	lc := ParseLocation(location)
	if lc.Letters == "" {
		return "", false
	}
	return lc.Letters, true
}

// ExtractSortKey returns the portion of location used for sorting (after colon).
func ExtractSortKey(location string) string {
	lc := ParseLocation(location)
	return lc.Code
}

// GroupLocations groups a slice of locations by their primary group.
// Returns a map where keys are group identifiers and values are location slices.
func GroupLocations(locations []string) map[string][]string {
	groups := make(map[string][]string)

	for _, loc := range locations {
		lc := ParseLocation(loc)
		group := lc.GetPrimaryGroup()
		if group == "" {
			group = "unknown"
		}
		groups[group] = append(groups[group], loc)
	}

	return groups
}

// SortGroupedLocations sorts each group's locations in place.
func SortGroupedLocations(groups map[string][]string) {
	for key := range groups {
		groups[key] = SortLocations(groups[key])
	}
}
