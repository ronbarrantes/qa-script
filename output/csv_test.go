package output

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"qa-script/rules"
)

func TestWriteCSV_Spillover(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		itemCount int
		wantCols  int
		wantRows  int // data rows (excluding header)
	}{
		{
			name:      "no spillover - items less than size",
			size:      5,
			itemCount: 3,
			wantCols:  1,
			wantRows:  3,
		},
		{
			name:      "no spillover - items equal to size",
			size:      5,
			itemCount: 5,
			wantCols:  1,
			wantRows:  5,
		},
		{
			name:      "spillover - one extra item",
			size:      5,
			itemCount: 6,
			wantCols:  2,
			wantRows:  5, // maxRows capped at size
		},
		{
			name:      "spillover - multiple columns",
			size:      3,
			itemCount: 10,
			wantCols:  4, // ceil(10/3) = 4
			wantRows:  3,
		},
		{
			name:      "size 0 - no limit",
			size:      0,
			itemCount: 10,
			wantCols:  1,
			wantRows:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test data
			grouped := rules.TitleGroupedLocations{
				"TestGroup": make([]string, tt.itemCount),
			}
			for i := 0; i < tt.itemCount; i++ {
				grouped["TestGroup"][i] = string(rune('A' + i))
			}

			data := NewOutputData([]string{"TestGroup"}, grouped, nil, 0, tt.size)

			// Write to temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.csv")
			if err := WriteCSV(filePath, data); err != nil {
				t.Fatalf("WriteCSV failed: %v", err)
			}

			// Read and verify
			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("Failed to read CSV: %v", err)
			}

			// Verify header row (should have wantCols columns)
			if len(records) == 0 {
				t.Fatal("CSV has no rows")
			}
			gotCols := len(records[0])
			if gotCols != tt.wantCols {
				t.Errorf("got %d columns, want %d", gotCols, tt.wantCols)
			}

			// Verify data rows
			gotRows := len(records) - 1 // minus header
			if gotRows != tt.wantRows {
				t.Errorf("got %d data rows, want %d", gotRows, tt.wantRows)
			}
		})
	}
}

func TestWriteCSV_SpilloverLayout(t *testing.T) {
	// Test that items are laid out in newspaper-style columns
	// Items should fill down each column first, then move to next column
	grouped := rules.TitleGroupedLocations{
		"Group": []string{"A", "B", "C", "D", "E", "F", "G"},
	}

	data := NewOutputData([]string{"Group"}, grouped, nil, 0, 3) // size=3, 7 items -> 3 columns

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	if err := WriteCSV(filePath, data); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Expected layout with size=3:
	// Header: Group, Group, Group
	// Row 1:  A,     D,     G
	// Row 2:  B,     E,     ""
	// Row 3:  C,     F,     ""
	expected := [][]string{
		{"Group", "Group", "Group"},
		{"A", "D", "G"},
		{"B", "E", ""},
		{"C", "F", ""},
	}

	if len(records) != len(expected) {
		t.Fatalf("got %d rows, want %d", len(records), len(expected))
	}

	for i, row := range records {
		if len(row) != len(expected[i]) {
			t.Errorf("row %d: got %d cols, want %d", i, len(row), len(expected[i]))
			continue
		}
		for j, cell := range row {
			if cell != expected[i][j] {
				t.Errorf("row %d col %d: got %q, want %q", i, j, cell, expected[i][j])
			}
		}
	}
}

func TestWriteCSV_EmptyGroupsOmitted(t *testing.T) {
	// Test that empty groups are not included in the output
	grouped := rules.TitleGroupedLocations{
		"Pallets":   []string{},                 // Empty - should be omitted
		"Shelves":   []string{"S1", "S2", "S3"}, // Has items - should be included
		"FloorLocs": []string{},                 // Empty - should be omitted
		"Bins":      []string{"B1", "B2"},       // Has items - should be included
	}

	// TitleOrder includes all groups, but empty ones should be skipped
	data := NewOutputData([]string{"Pallets", "Shelves", "FloorLocs", "Bins"}, grouped, nil, 1, 0)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	if err := WriteCSV(filePath, data); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Header should be: Shelves, "", Bins (gap=1 between groups)
	header := records[0]
	expectedHeaders := []string{"Shelves", "", "Bins"}
	if len(header) != len(expectedHeaders) {
		t.Fatalf("got %d columns, want %d: %v", len(header), len(expectedHeaders), header)
	}

	for i, want := range expectedHeaders {
		if header[i] != want {
			t.Errorf("header[%d] = %q, want %q", i, header[i], want)
		}
	}

	// Verify "Pallets" and "FloorLocs" are NOT in the output
	for _, h := range header {
		if h == "Pallets" || h == "FloorLocs" {
			t.Errorf("empty group %q should not be in output", h)
		}
	}
}

func TestWriteCSV_GapColumns(t *testing.T) {
	grouped := rules.TitleGroupedLocations{
		"GroupA": []string{"A1", "A2"},
		"GroupB": []string{"B1", "B2"},
	}

	// gap=2 means 2 empty columns between groups
	data := NewOutputData([]string{"GroupA", "GroupB"}, grouped, nil, 2, 0)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	if err := WriteCSV(filePath, data); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Expected: GroupA, "", "", GroupB (1 col + 2 gap + 1 col = 4 total)
	header := records[0]
	if len(header) != 4 {
		t.Errorf("got %d columns, want 4", len(header))
	}

	// Check header structure
	if header[0] != "GroupA" || header[1] != "" || header[2] != "" || header[3] != "GroupB" {
		t.Errorf("unexpected header: %v", header)
	}
}
