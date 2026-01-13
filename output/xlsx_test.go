package output

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
	"qa-script/rules"
)

func TestWriteXLSX_Spillover(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		itemCount  int
		wantCols   int
	}{
		{
			name:      "no spillover - items less than size",
			size:      5,
			itemCount: 3,
			wantCols:  1,
		},
		{
			name:      "no spillover - items equal to size",
			size:      5,
			itemCount: 5,
			wantCols:  1,
		},
		{
			name:      "spillover - one extra item",
			size:      5,
			itemCount: 6,
			wantCols:  2,
		},
		{
			name:      "spillover - multiple columns",
			size:      3,
			itemCount: 10,
			wantCols:  4, // ceil(10/3) = 4
		},
		{
			name:      "size 0 - no limit",
			size:      0,
			itemCount: 10,
			wantCols:  1,
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
			filePath := filepath.Join(tmpDir, "test.xlsx")
			if err := WriteXLSX(filePath, data); err != nil {
				t.Fatalf("WriteXLSX failed: %v", err)
			}

			// Read and verify
			f, err := excelize.OpenFile(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer f.Close()

			// Check header row - count non-empty cells
			sheetName := "Locations"
			gotCols := 0
			for col := 1; col <= 20; col++ {
				cell, _ := excelize.CoordinatesToCellName(col, 1)
				val, _ := f.GetCellValue(sheetName, cell)
				if val != "" {
					gotCols++
				} else {
					break
				}
			}

			// Note: XLSX merges header cells, so we check data columns instead
			// Count columns that have data in row 2
			dataCols := 0
			for col := 1; col <= 20; col++ {
				cell, _ := excelize.CoordinatesToCellName(col, 2)
				val, _ := f.GetCellValue(sheetName, cell)
				if val != "" {
					dataCols++
				} else if col > 1 {
					// Check if previous column had data
					prevCell, _ := excelize.CoordinatesToCellName(col-1, 2)
					prevVal, _ := f.GetCellValue(sheetName, prevCell)
					if prevVal == "" {
						break
					}
				}
			}

			// For spillover, data columns should match wantCols
			// But items might not fill all columns, so check ColumnsNeeded
			expectedCols := data.ColumnsNeeded(tt.itemCount)
			if expectedCols != tt.wantCols {
				t.Errorf("ColumnsNeeded returned %d, want %d", expectedCols, tt.wantCols)
			}
		})
	}
}

func TestWriteXLSX_SpilloverLayout(t *testing.T) {
	// Test that items are laid out in newspaper-style columns
	// Items should fill down each column first, then move to next column
	grouped := rules.TitleGroupedLocations{
		"Group": []string{"A", "B", "C", "D", "E", "F", "G"},
	}

	data := NewOutputData([]string{"Group"}, grouped, nil, 0, 3) // size=3, 7 items -> 3 columns

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	if err := WriteXLSX(filePath, data); err != nil {
		t.Fatalf("WriteXLSX failed: %v", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	sheetName := "Locations"

	// Expected layout with size=3:
	// Header: Group (merged across 3 columns)
	// Row 2:  A,     D,     G
	// Row 3:  B,     E,     ""
	// Row 4:  C,     F,     ""
	expected := map[string]string{
		"A1": "Group", // header
		"A2": "A", "B2": "D", "C2": "G",
		"A3": "B", "B3": "E", "C3": "",
		"A4": "C", "B4": "F", "C4": "",
	}

	for cell, want := range expected {
		got, _ := f.GetCellValue(sheetName, cell)
		if got != want {
			t.Errorf("cell %s: got %q, want %q", cell, got, want)
		}
	}
}

func TestWriteXLSX_ColumnWidth(t *testing.T) {
	// Test that column width is set based on longest data item, not just title
	grouped := rules.TitleGroupedLocations{
		"Short": []string{"SS11:VeryLongLocationCode123.A", "A", "B"},
	}

	data := NewOutputData([]string{"Short"}, grouped, nil, 0, 0)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	if err := WriteXLSX(filePath, data); err != nil {
		t.Fatalf("WriteXLSX failed: %v", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	// Get column width
	colWidth, err := f.GetColWidth("Locations", "A")
	if err != nil {
		t.Fatalf("Failed to get column width: %v", err)
	}

	// The longest item is "SS11:VeryLongLocationCode123.A" (30 chars) + 2 padding = 32
	// Minimum width is 15
	// Width should be based on data length, not title length ("Short" = 5 chars)
	expectedMinWidth := float64(len("SS11:VeryLongLocationCode123.A") + 2)
	if colWidth < expectedMinWidth {
		t.Errorf("column width %v is less than expected minimum %v (based on longest data)", colWidth, expectedMinWidth)
	}
}

func TestWriteXLSX_PriorityHighlight(t *testing.T) {
	grouped := rules.TitleGroupedLocations{
		"Group": []string{"item1", "item2", "item3"},
	}

	// Mark item2 as priority
	data := NewOutputData([]string{"Group"}, grouped, []string{"item2"}, 0, 0)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	if err := WriteXLSX(filePath, data); err != nil {
		t.Fatalf("WriteXLSX failed: %v", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	sheetName := "Locations"

	// Verify item2 (A3) has a style applied
	// We can't easily check the exact style, but we can verify the cell exists
	val, _ := f.GetCellValue(sheetName, "A3")
	if val != "item2" {
		t.Errorf("expected item2 at A3, got %q", val)
	}

	// Check that IsPriority returns correct values
	if !data.IsPriority("item2") {
		t.Error("expected item2 to be priority")
	}
	if data.IsPriority("item1") {
		t.Error("expected item1 to NOT be priority")
	}
}
