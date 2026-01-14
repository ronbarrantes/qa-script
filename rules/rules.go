package rules

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Config represents the rules.yaml structure
type Config struct {
	Groups []Group `yaml:"groups"`
	Size   int     `yaml:"size"`
	Gap    int     `yaml:"gap"` // Number of empty columns between groups (0 = no gap)
}

// Group represents a single group in the config
type Group struct {
	Title  string   `yaml:"title"`
	Values []string `yaml:"values"`
}

// GroupedLocations maps rule keys to their assigned locations
type GroupedLocations map[string][]string

// LoadConfig reads and parses the rules.yaml file
func LoadConfig(path string) (*Config, error) {
	fmt.Printf("Loading rules from: %s\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
	}

	fmt.Printf("Read %d bytes from rules file\n", len(data))

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Log what was loaded for debugging
	fmt.Printf("Loaded rules: %d groups, size=%d, gap=%d\n", len(config.Groups), config.Size, config.Gap)
	for _, g := range config.Groups {
		fmt.Printf("  - Group '%s': %v\n", g.Title, g.Values)
	}

	return &config, nil
}

// GetAllKeys returns all values from all groups as lowercase keys
func (c *Config) GetAllKeys() []string {
	var keys []string
	for _, group := range c.Groups {
		for _, value := range group.Values {
			keys = append(keys, strings.ToLower(value))
		}
	}
	return keys
}

// GetTitlesInOrder returns all group titles in the order they appear in the config
func (c *Config) GetTitlesInOrder() []string {
	titles := make([]string, len(c.Groups))
	for i, group := range c.Groups {
		titles[i] = group.Title
	}
	return titles
}

// extractLetterPrefix extracts the letter-only prefix after the colon
// e.g., "SS4:GFT22.B" -> "GFT", "SS4:GF225.C" -> "GF"
func extractLetterPrefix(location string) string {
	// Find the part after the colon
	idx := strings.Index(location, ":")
	if idx == -1 {
		return ""
	}
	afterColon := location[idx+1:]

	// Extract leading letters only (no digits)
	var letters strings.Builder
	for _, r := range afterColon {
		if unicode.IsLetter(r) {
			letters.WriteRune(r)
		} else {
			break
		}
	}
	return letters.String()
}

// isMultiLetterKey checks if a key is 3+ letters (not a single letter key)
func isMultiLetterKey(key string) bool {
	return len(key) >= 3
}

// GroupLocations assigns locations to groups based on the rules
func GroupLocations(locations []string, config *Config) GroupedLocations {
	// Build a set of valid keys from config
	validKeys := make(map[string]bool)
	for _, key := range config.GetAllKeys() {
		validKeys[strings.ToLower(key)] = true
	}

	// Initialize result map with all keys + unassigned
	result := make(GroupedLocations)
	for key := range validKeys {
		result[key] = []string{}
	}
	result["unassigned"] = []string{}

	// Pattern to check if prefix has only letters (no digits)
	lettersOnly := regexp.MustCompile(`^[A-Za-z]+$`)

	// Assign each location to a group
	for _, loc := range locations {
		prefix := extractLetterPrefix(loc)
		if prefix == "" {
			result["unassigned"] = append(result["unassigned"], loc)
			continue
		}

		prefixLower := strings.ToLower(prefix)
		assigned := false

		// Rule 1: If prefix is 3+ letters, try exact match
		if len(prefix) >= 3 && lettersOnly.MatchString(prefix) {
			if validKeys[prefixLower] {
				result[prefixLower] = append(result[prefixLower], loc)
				assigned = true
			}
		}

		// Rule 2: If prefix is 2 letters (or 3+ didn't match), use first letter
		if !assigned && len(prefix) >= 2 {
			firstLetter := strings.ToLower(string(prefix[0]))
			if validKeys[firstLetter] {
				result[firstLetter] = append(result[firstLetter], loc)
				assigned = true
			}
		}

		// Rule 3: Anything else goes to unassigned
		if !assigned {
			result["unassigned"] = append(result["unassigned"], loc)
		}
	}

	return result
}

// GetNonEmptyGroups returns only groups that have locations assigned
func (g GroupedLocations) GetNonEmptyGroups() GroupedLocations {
	result := make(GroupedLocations)
	for key, locs := range g {
		if len(locs) > 0 {
			result[key] = locs
		}
	}
	return result
}

// TitleGroupedLocations maps group titles to their assigned locations
type TitleGroupedLocations map[string][]string

// GroupByTitle takes grouped locations and re-groups them by the config group titles
// e.g., "pallets" -> all locations from [a, b, c, lud, prm, slp]
func GroupByTitle(grouped GroupedLocations, config *Config) TitleGroupedLocations {
	result := make(TitleGroupedLocations)

	// Initialize with all titles + unassigned
	for _, group := range config.Groups {
		result[group.Title] = []string{}
	}
	result["unassigned"] = []string{}

	// For each group, collect locations from its values
	for _, group := range config.Groups {
		for _, value := range group.Values {
			valueLower := strings.ToLower(value)
			if locs, exists := grouped[valueLower]; exists {
				result[group.Title] = append(result[group.Title], locs...)
			}
		}
	}

	// Copy unassigned locations
	if locs, exists := grouped["unassigned"]; exists {
		result["unassigned"] = append(result["unassigned"], locs...)
	}

	return result
}
