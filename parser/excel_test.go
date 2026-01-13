package parser

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

// createTestExcel creates a temporary Excel file for testing
func createTestExcel(t *testing.T, sheetName string, data [][]string) string {
	t.Helper()

	f := excelize.NewFile()
	defer f.Close()

	// Rename default sheet or create new one
	if sheetName != "Sheet1" {
		f.SetSheetName("Sheet1", sheetName)
	}

	// Write data
	for rowIdx, row := range data {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xlsx")
	if err := f.SaveAs(tmpFile); err != nil {
		t.Fatalf("failed to save test Excel file: %v", err)
	}

	return tmpFile
}

func TestParseExcel(t *testing.T) {
	data := [][]string{
		{"Name", "Location", "Status"},
		{"Item1", "SS4:AB100", "Active"},
		{"Item2", "PS2:CL200", "Inactive"},
		{"Item3", "ST5:GF300", "Active"},
	}

	tmpFile := createTestExcel(t, "TestSheet", data)

	// Test with specific sheet name
	excel, err := ParseExcel(tmpFile, "TestSheet")
	if err != nil {
		t.Fatalf("ParseExcel failed: %v", err)
	}

	// Check sheet name
	if excel.Sheet != "TestSheet" {
		t.Errorf("Sheet = %q, want %q", excel.Sheet, "TestSheet")
	}

	// Check headers
	expectedHeaders := []string{"Name", "Location", "Status"}
	if !reflect.DeepEqual(excel.Headers, expectedHeaders) {
		t.Errorf("Headers = %v, want %v", excel.Headers, expectedHeaders)
	}

	// Check row count
	if len(excel.Rows) != 3 {
		t.Errorf("len(Rows) = %d, want 3", len(excel.Rows))
	}
}

func TestParseExcel_DefaultSheet(t *testing.T) {
	data := [][]string{
		{"Col1", "Col2"},
		{"A", "B"},
	}

	tmpFile := createTestExcel(t, "Sheet1", data)

	// Test with empty sheet name (should use first sheet)
	excel, err := ParseExcel(tmpFile, "")
	if err != nil {
		t.Fatalf("ParseExcel failed: %v", err)
	}

	if excel.Sheet != "Sheet1" {
		t.Errorf("Sheet = %q, want %q", excel.Sheet, "Sheet1")
	}
}

func TestParseExcel_FileNotFound(t *testing.T) {
	_, err := ParseExcel("/nonexistent/file.xlsx", "")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestExcelData_GetColumnIndex(t *testing.T) {
	excel := &ExcelData{
		Headers: []string{"Container Tag", "Current Location", "Status"},
		Rows:    [][]string{},
	}

	tests := []struct {
		column   string
		expected int
	}{
		{"Container Tag", 0},
		{"Current Location", 1},
		{"Status", 2},
		{"NotFound", -1},
		{"container tag", -1}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			result := excel.GetColumnIndex(tt.column)
			if result != tt.expected {
				t.Errorf("GetColumnIndex(%q) = %d, want %d", tt.column, result, tt.expected)
			}
		})
	}
}

func TestExcelData_GetColumnValues(t *testing.T) {
	excel := &ExcelData{
		Headers: []string{"Tag", "Location"},
		Rows: [][]string{
			{"QA_HOLD", "SS4:AB100"},
			{"ACTIVE", "PS2:CL200"},
			{"QA_HOLD", "ST5:GF300"},
		},
	}

	values, err := excel.GetColumnValues("Location")
	if err != nil {
		t.Fatalf("GetColumnValues failed: %v", err)
	}

	expected := []string{"SS4:AB100", "PS2:CL200", "ST5:GF300"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("GetColumnValues(Location) = %v, want %v", values, expected)
	}
}

func TestExcelData_GetColumnValues_NotFound(t *testing.T) {
	excel := &ExcelData{
		Headers: []string{"A", "B"},
		Rows:    [][]string{},
	}

	_, err := excel.GetColumnValues("NotAColumn")
	if err == nil {
		t.Error("expected error for nonexistent column, got nil")
	}
}

func TestExcelData_GetColumnValues_ShortRows(t *testing.T) {
	excel := &ExcelData{
		Headers: []string{"A", "B", "C"},
		Rows: [][]string{
			{"v1", "v2", "v3"},
			{"short"},
			{"x", "y", "z"},
		},
	}

	values, err := excel.GetColumnValues("C")
	if err != nil {
		t.Fatalf("GetColumnValues failed: %v", err)
	}

	expected := []string{"v3", "", "z"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("GetColumnValues(C) = %v, want %v", values, expected)
	}
}
