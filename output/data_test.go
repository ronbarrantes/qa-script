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

	data := NewOutputData(titleOrder, grouped, priorities, 1, 20)

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
	data := NewOutputData([]string{}, nil, []string{}, 0, 0)

	if len(data.PrioritySet) != 0 {
		t.Errorf("expected empty priority set, got %d items", len(data.PrioritySet))
	}

	if data.IsPriority("anything") {
		t.Error("expected nothing to be priority with empty set")
	}
}

func TestNewOutputData_NilGrouped(t *testing.T) {
	// Should handle nil grouped without panic
	data := NewOutputData([]string{"test"}, nil, []string{"loc1"}, 0, 0)

	if data.Grouped != nil {
		t.Error("expected Grouped to be nil")
	}

	// Priority should still work
	if !data.IsPriority("loc1") {
		t.Error("expected loc1 to be priority")
	}
}

func TestNewOutputData_Gap(t *testing.T) {
	data := NewOutputData([]string{"test"}, nil, []string{}, 2, 0)

	if data.Gap != 2 {
		t.Errorf("expected Gap to be 2, got %d", data.Gap)
	}
}

func TestNewOutputData_MaxRows(t *testing.T) {
	data := NewOutputData([]string{"test"}, nil, []string{}, 0, 20)

	if data.MaxRows != 20 {
		t.Errorf("expected MaxRows to be 20, got %d", data.MaxRows)
	}
}

func TestOutputData_ColumnsNeeded(t *testing.T) {
	tests := []struct {
		name      string
		maxRows   int
		itemCount int
		expected  int
	}{
		{"no spillover needed", 20, 10, 1},
		{"exactly at max_rows", 3, 3, 1},
		{"needs 2 columns", 3, 4, 2},
		{"needs 4 columns", 3, 11, 4},
		{"max_rows 0 means no limit", 0, 100, 1},
		{"empty group", 3, 0, 1},
		{"max_rows 1 item per column", 1, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &OutputData{MaxRows: tt.maxRows}
			result := data.ColumnsNeeded(tt.itemCount)
			if result != tt.expected {
				t.Errorf("ColumnsNeeded(%d) with max_rows %d = %d, want %d",
					tt.itemCount, tt.maxRows, result, tt.expected)
			}
		})
	}
}
