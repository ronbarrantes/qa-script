package rules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractLetterPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"3 letters", "SS4:GFT22.B", "GFT"},
		{"2 letters", "SS4:GF225.C", "GF"},
		{"2 letters simple", "PS2:CL106", "CL"},
		{"3 letters with suffix", "PS1:LUD215", "LUD"},
		{"no colon", "ABC123", ""},
		{"empty string", "", ""},
		{"only letters after colon", "X:ABC", "ABC"},
		{"starts with number after colon", "X:123ABC", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLetterPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("extractLetterPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsMultiLetterKey(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"a", false},
		{"ab", false},
		{"abc", true},
		{"abcd", true},
		{"gft", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isMultiLetterKey(tt.input)
			if result != tt.expected {
				t.Errorf("isMultiLetterKey(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfig_GetAllKeys(t *testing.T) {
	config := &Config{
		Groups: []Group{
			{Title: "pallets", Values: []string{"a", "B", "LUD"}},
			{Title: "efg", Values: []string{"E", "gft"}},
		},
	}

	expected := []string{"a", "b", "lud", "e", "gft"}
	result := config.GetAllKeys()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetAllKeys() = %v, want %v", result, expected)
	}
}

func TestConfig_GetTitlesInOrder(t *testing.T) {
	config := &Config{
		Groups: []Group{
			{Title: "pallets", Values: []string{"a", "b"}},
			{Title: "efg", Values: []string{"e", "f"}},
			{Title: "hjkl", Values: []string{"h", "j"}},
		},
	}

	expected := []string{"pallets", "efg", "hjkl"}
	result := config.GetTitlesInOrder()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetTitlesInOrder() = %v, want %v", result, expected)
	}
}

func TestGroupLocations(t *testing.T) {
	config := &Config{
		Groups: []Group{
			{Title: "pallets", Values: []string{"a", "c", "lud"}},
			{Title: "efg", Values: []string{"e", "g", "gft"}},
		},
	}

	locations := []string{
		"ST5:AQ211",   // should match 'a'
		"PS2:CL106",   // should match 'c'
		"SS4:GFT33.A", // should match 'gft' (exact 3-letter)
		"SS4:GF225.B", // should match 'g' (2-letter prefix)
		"PS1:LUD215",  // should match 'lud' (exact 3-letter)
		"SS4:XY999",   // should be unassigned
	}

	result := GroupLocations(locations, config)

	// Check specific assignments
	tests := []struct {
		key      string
		expected []string
	}{
		{"a", []string{"ST5:AQ211"}},
		{"c", []string{"PS2:CL106"}},
		{"gft", []string{"SS4:GFT33.A"}},
		{"g", []string{"SS4:GF225.B"}},
		{"lud", []string{"PS1:LUD215"}},
		{"unassigned", []string{"SS4:XY999"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if !reflect.DeepEqual(result[tt.key], tt.expected) {
				t.Errorf("GroupLocations()[%q] = %v, want %v", tt.key, result[tt.key], tt.expected)
			}
		})
	}
}

func TestGroupLocations_ThreeLetterExactMatch(t *testing.T) {
	// Test that 3-letter prefixes match exactly, not by first letter
	config := &Config{
		Groups: []Group{
			{Title: "test", Values: []string{"g", "gft"}},
		},
	}

	locations := []string{
		"SS4:GFT33.A", // should match 'gft', NOT 'g'
		"SS4:GF225.B", // should match 'g'
	}

	result := GroupLocations(locations, config)

	if len(result["gft"]) != 1 || result["gft"][0] != "SS4:GFT33.A" {
		t.Errorf("Expected GFT33.A to match 'gft', got %v", result["gft"])
	}

	if len(result["g"]) != 1 || result["g"][0] != "SS4:GF225.B" {
		t.Errorf("Expected GF225.B to match 'g', got %v", result["g"])
	}
}

func TestGroupByTitle(t *testing.T) {
	config := &Config{
		Groups: []Group{
			{Title: "pallets", Values: []string{"a", "c"}},
			{Title: "efg", Values: []string{"e", "gft"}},
		},
	}

	grouped := GroupedLocations{
		"a":          []string{"ST5:AQ211"},
		"c":          []string{"PS2:CL106"},
		"gft":        []string{"SS4:GFT33.A"},
		"e":          []string{"SS4:EF205.G"},
		"unassigned": []string{"SS4:XY999"},
	}

	result := GroupByTitle(grouped, config)

	// pallets should have items from 'a' and 'c'
	expectedPallets := []string{"ST5:AQ211", "PS2:CL106"}
	if !reflect.DeepEqual(result["pallets"], expectedPallets) {
		t.Errorf("GroupByTitle()[pallets] = %v, want %v", result["pallets"], expectedPallets)
	}

	// efg should have items from 'e' and 'gft'
	expectedEfg := []string{"SS4:EF205.G", "SS4:GFT33.A"}
	if !reflect.DeepEqual(result["efg"], expectedEfg) {
		t.Errorf("GroupByTitle()[efg] = %v, want %v", result["efg"], expectedEfg)
	}

	// unassigned should be preserved
	if !reflect.DeepEqual(result["unassigned"], []string{"SS4:XY999"}) {
		t.Errorf("GroupByTitle()[unassigned] = %v, want [SS4:XY999]", result["unassigned"])
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary YAML file
	content := `groups:
  - title: pallets
    values: [a, b, c]
  - title: efg
    values: [e, f, g]
max_rows: 20
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	config, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(config.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(config.Groups))
	}

	if config.Groups[0].Title != "pallets" {
		t.Errorf("expected first group title 'pallets', got %q", config.Groups[0].Title)
	}

	if config.MaxRows != 20 {
		t.Errorf("expected max_rows 20, got %d", config.MaxRows)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestGetNonEmptyGroups(t *testing.T) {
	grouped := GroupedLocations{
		"a":          []string{"loc1"},
		"b":          []string{},
		"c":          []string{"loc2", "loc3"},
		"unassigned": []string{},
	}

	result := grouped.GetNonEmptyGroups()

	if len(result) != 2 {
		t.Errorf("expected 2 non-empty groups, got %d", len(result))
	}

	if _, exists := result["a"]; !exists {
		t.Error("expected 'a' to be in result")
	}

	if _, exists := result["c"]; !exists {
		t.Error("expected 'c' to be in result")
	}

	if _, exists := result["b"]; exists {
		t.Error("expected 'b' to NOT be in result")
	}
}
