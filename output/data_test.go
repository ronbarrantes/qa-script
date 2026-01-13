package output

import (
	"qa-script/rules"
	"testing"
)

func TestOutputData_IsPriority(t *testing.T) {
	data := &OutputData{
		PrioritySet: map[string]struct{}{
			"SS4:GFT33.A":  {},
			"SS11:HW405.C": {},
		},
	}

	tests := []struct {
		location string
		expected bool
	}{
		{"SS4:GFT33.A", true},
		{"SS11:HW405.C", true},
		{"PS2:CL106", false},
		{"", false},
		{"SS4:GFT33.A ", false}, // with trailing space - exact match required
	}

	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			result := data.IsPriority(tt.location)
			if result != tt.expected {
				t.Errorf("IsPriority(%q) = %v, want %v", tt.location, result, tt.expected)
			}
		})
	}
}

func TestNewOutputData(t *testing.T) {
	titleOrder := []string{"pallets", "efg"}
	grouped := rules.TitleGroupedLocations{
		"pallets": []string{"loc1", "loc2"},
		"efg":     []string{"loc3"},
	}
	priorities := []string{"loc1", "loc3"}

	data := NewOutputData(titleOrder, grouped, priorities)

	// Check titleOrder is set
	if len(data.TitleOrder) != 2 {
		t.Errorf("expected 2 titles, got %d", len(data.TitleOrder))
	}

	// Check grouped is set
	if len(data.Grouped["pallets"]) != 2 {
		t.Errorf("expected 2 items in pallets, got %d", len(data.Grouped["pallets"]))
	}

	// Check priority set is built correctly
	if !data.IsPriority("loc1") {
		t.Error("expected loc1 to be priority")
	}

	if !data.IsPriority("loc3") {
		t.Error("expected loc3 to be priority")
	}

	if data.IsPriority("loc2") {
		t.Error("expected loc2 to NOT be priority")
	}
}

func TestNewOutputData_EmptyPriorities(t *testing.T) {
	data := NewOutputData([]string{}, nil, []string{})

	if len(data.PrioritySet) != 0 {
		t.Errorf("expected empty priority set, got %d items", len(data.PrioritySet))
	}

	if data.IsPriority("anything") {
		t.Error("expected nothing to be priority with empty set")
	}
}

func TestNewOutputData_NilGrouped(t *testing.T) {
	// Should handle nil grouped without panic
	data := NewOutputData([]string{"test"}, nil, []string{"loc1"})

	if data.Grouped != nil {
		t.Error("expected Grouped to be nil")
	}

	// Priority should still work
	if !data.IsPriority("loc1") {
		t.Error("expected loc1 to be priority")
	}
}
