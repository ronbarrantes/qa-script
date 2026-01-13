package location

import (
	"reflect"
	"testing"
)

func TestGetSortKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with colon", "PS2:CL106", "CL106"},
		{"with colon and suffix", "SS4:GF225.B", "GF225.B"},
		{"no colon", "ABC123", "ABC123"},
		{"empty string", "", ""},
		{"colon at start", ":ABC", "ABC"},
		{"colon at end", "ABC:", ""},
		{"multiple colons", "A:B:C", "B:C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSortKey(tt.input)
			if result != tt.expected {
				t.Errorf("GetSortKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUniqueAndSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "removes duplicates",
			input:    []string{"PS2:AB100", "PS2:AB100", "SS4:AB100"},
			expected: []string{"PS2:AB100", "SS4:AB100"},
		},
		{
			name:     "sorts by part after colon",
			input:    []string{"SS4:GF225", "PS2:AB100", "ST5:CL106"},
			expected: []string{"PS2:AB100", "ST5:CL106", "SS4:GF225"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"PS2:AB100"},
			expected: []string{"PS2:AB100"},
		},
		{
			name:     "all duplicates",
			input:    []string{"PS2:AB100", "PS2:AB100", "PS2:AB100"},
			expected: []string{"PS2:AB100"},
		},
		{
			name:     "mixed with and without colons",
			input:    []string{"ABC", "PS2:AB100", "DEF"},
			expected: []string{"PS2:AB100", "ABC", "DEF"}, // AB100 < ABC < DEF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueAndSort(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UniqueAndSort(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUniqueAndSort_PreservesOriginal(t *testing.T) {
	// Verify the original slice is not modified
	input := []string{"SS4:GF225", "PS2:AB100"}
	original := make([]string, len(input))
	copy(original, input)

	UniqueAndSort(input)

	if !reflect.DeepEqual(input, original) {
		t.Errorf("UniqueAndSort modified the original slice: got %v, want %v", input, original)
	}
}
