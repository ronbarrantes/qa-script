package rules

import (
	"testing"
)

func TestParseLocation(t *testing.T) {
	tests := []struct {
		input    string
		wantCode string
		wantLets string
		wantNum  string
		wantSuf  string
	}{
		{"SS4:GF225.B", "GF225.B", "GF", "225", ".B"},
		{"SS4:GFT245.B", "GFT245.B", "GFT", "245", ".B"},
		{"ABC:HVC12", "HVC12", "HVC", "12", ""},
		{"X:AB100", "AB100", "AB", "100", ""},
		{"NoColon", "NoColon", "NOCOLON", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lc := ParseLocation(tt.input)
			if lc.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", lc.Code, tt.wantCode)
			}
			if lc.Letters != tt.wantLets {
				t.Errorf("Letters = %q, want %q", lc.Letters, tt.wantLets)
			}
			if lc.Number != tt.wantNum {
				t.Errorf("Number = %q, want %q", lc.Number, tt.wantNum)
			}
			if lc.Suffix != tt.wantSuf {
				t.Errorf("Suffix = %q, want %q", lc.Suffix, tt.wantSuf)
			}
		})
	}
}

func TestGetCodeType(t *testing.T) {
	tests := []struct {
		input    string
		wantType CodeType
	}{
		{"SS4:GF225.B", StandardGroup},   // 2 letters = standard
		{"SS4:GFT245.B", MatchingGroup},  // 3 letters = matching
		{"X:AB100", StandardGroup},       // 2 letters = standard
		{"X:HVC12", MatchingGroup},       // 3 letters = matching
		{"X:A1", UnknownGroup},           // 1 letter = unknown
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lc := ParseLocation(tt.input)
			if got := lc.GetCodeType(); got != tt.wantType {
				t.Errorf("GetCodeType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestGetPrimaryGroup(t *testing.T) {
	tests := []struct {
		input     string
		wantGroup string
	}{
		{"SS4:GF225.B", "G"},    // 2 letters: first letter is group
		{"SS4:GFT245.B", "GFT"}, // 3 letters: full code is group
		{"X:AB100", "A"},       // 2 letters: first letter is group
		{"X:HVC12", "HVC"},     // 3 letters: full code is group
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lc := ParseLocation(tt.input)
			if got := lc.GetPrimaryGroup(); got != tt.wantGroup {
				t.Errorf("GetPrimaryGroup() = %q, want %q", got, tt.wantGroup)
			}
		})
	}
}

func TestMatchesGroupValue(t *testing.T) {
	tests := []struct {
		location   string
		groupValue string
		want       bool
	}{
		// 2-letter codes with single-letter group values
		{"SS4:GF225.B", "G", true},   // GF starts with G
		{"SS4:GF225.B", "H", false},  // GF doesn't start with H
		{"SS4:AB100.A", "A", true},   // AB starts with A

		// 3-letter codes with exact match
		{"SS4:GFT245.B", "GFT", true},  // Exact match
		{"SS4:GFT245.B", "G", true},    // Also matches prefix
		{"SS4:GFT245.B", "HVC", false}, // Doesn't match

		// Exact code matching
		{"SS4:GF225.B", "GF", true},   // Exact match for 2-letter
		{"SS4:GF225.B", "AB", false},  // Not a match
	}

	for _, tt := range tests {
		t.Run(tt.location+"_"+tt.groupValue, func(t *testing.T) {
			lc := ParseLocation(tt.location)
			if got := lc.MatchesGroupValue(tt.groupValue); got != tt.want {
				t.Errorf("MatchesGroupValue(%q) = %v, want %v", tt.groupValue, got, tt.want)
			}
		})
	}
}

func TestSortLocations(t *testing.T) {
	input := []string{
		"SS4:GF300.B",
		"SS4:GF25.A",
		"SS4:GF100.C",
		"SS4:AB50",
		"SS4:GF100.A",
	}

	sorted := SortLocations(input)

	// Should be sorted by letters first, then numerically, then by suffix
	expected := []string{
		"SS4:AB50",
		"SS4:GF25.A",
		"SS4:GF100.A",
		"SS4:GF100.C",
		"SS4:GF300.B",
	}

	for i, want := range expected {
		if sorted[i] != want {
			t.Errorf("sorted[%d] = %q, want %q", i, sorted[i], want)
		}
	}
}

func TestIsMatchingGroup(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"SS4:GF225.B", false},  // 2 letters = not matching group
		{"SS4:GFT245.B", true},  // 3 letters = matching group
		{"X:HVC12", true},       // 3 letters = matching group
		{"X:AB100", false},      // 2 letters = not matching group
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lc := ParseLocation(tt.input)
			if got := lc.IsMatchingGroup(); got != tt.want {
				t.Errorf("IsMatchingGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupLocations(t *testing.T) {
	input := []string{
		"SS4:GF225.B",
		"SS4:GFT245.B",
		"SS4:AB100",
		"SS4:GF300.A",
		"SS4:HVC12",
	}

	groups := GroupLocations(input)

	// Check that groups are created correctly
	// 2-letter codes group by first letter
	// 3-letter codes use full code
	if len(groups["G"]) != 2 { // GF225 and GF300
		t.Errorf("Expected 2 items in G group, got %d", len(groups["G"]))
	}
	if len(groups["GFT"]) != 1 {
		t.Errorf("Expected 1 item in GFT group, got %d", len(groups["GFT"]))
	}
	if len(groups["A"]) != 1 { // AB100
		t.Errorf("Expected 1 item in A group, got %d", len(groups["A"]))
	}
	if len(groups["HVC"]) != 1 {
		t.Errorf("Expected 1 item in HVC group, got %d", len(groups["HVC"]))
	}
}

func TestExtractSortKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SS4:GF225.B", "GF225.B"},
		{"ABC:XYZ123", "XYZ123"},
		{"NoColon", "NoColon"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ExtractSortKey(tt.input); got != tt.want {
				t.Errorf("ExtractSortKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
