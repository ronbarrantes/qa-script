package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseCSV(t *testing.T) {
	// Create a temporary CSV file
	content := `Name,Location,Status
Item1,SS4:AB100,Active
Item2,PS2:CL200,Inactive
Item3,ST5:GF300,Active`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	csv, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	// Check headers
	expectedHeaders := []string{"Name", "Location", "Status"}
	if !reflect.DeepEqual(csv.Headers, expectedHeaders) {
		t.Errorf("Headers = %v, want %v", csv.Headers, expectedHeaders)
	}

	// Check row count
	if len(csv.Rows) != 3 {
		t.Errorf("len(Rows) = %d, want 3", len(csv.Rows))
	}

	// Check first row
	expectedRow := []string{"Item1", "SS4:AB100", "Active"}
	if !reflect.DeepEqual(csv.Rows[0], expectedRow) {
		t.Errorf("Rows[0] = %v, want %v", csv.Rows[0], expectedRow)
	}
}

func TestParseCSV_FileNotFound(t *testing.T) {
	_, err := ParseCSV("/nonexistent/file.csv")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestParseCSV_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.csv")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := ParseCSV(tmpFile)
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}
}

func TestCSVData_GetColumnIndex(t *testing.T) {
	csv := &CSVData{
		Headers: []string{"Name", "Location", "Status"},
		Rows:    [][]string{},
	}

	tests := []struct {
		column   string
		expected int
	}{
		{"Name", 0},
		{"Location", 1},
		{"Status", 2},
		{"NotFound", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			result := csv.GetColumnIndex(tt.column)
			if result != tt.expected {
				t.Errorf("GetColumnIndex(%q) = %d, want %d", tt.column, result, tt.expected)
			}
		})
	}
}

func TestCSVData_GetColumnValues(t *testing.T) {
	csv := &CSVData{
		Headers: []string{"Name", "Location", "Status"},
		Rows: [][]string{
			{"Item1", "SS4:AB100", "Active"},
			{"Item2", "PS2:CL200", "Inactive"},
			{"Item3", "ST5:GF300", "Active"},
		},
	}

	values, err := csv.GetColumnValues("Location")
	if err != nil {
		t.Fatalf("GetColumnValues failed: %v", err)
	}

	expected := []string{"SS4:AB100", "PS2:CL200", "ST5:GF300"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("GetColumnValues(Location) = %v, want %v", values, expected)
	}
}

func TestCSVData_GetColumnValues_NotFound(t *testing.T) {
	csv := &CSVData{
		Headers: []string{"Name", "Location"},
		Rows:    [][]string{},
	}

	_, err := csv.GetColumnValues("NotAColumn")
	if err == nil {
		t.Error("expected error for nonexistent column, got nil")
	}
}

func TestCSVData_GetColumnValues_ShortRows(t *testing.T) {
	// Test handling of rows shorter than the column index
	csv := &CSVData{
		Headers: []string{"A", "B", "C"},
		Rows: [][]string{
			{"val1", "val2", "val3"},
			{"short"}, // row shorter than expected
			{"x", "y", "z"},
		},
	}

	values, err := csv.GetColumnValues("C")
	if err != nil {
		t.Fatalf("GetColumnValues failed: %v", err)
	}

	// Should get "val3", "", "z" (empty for short row)
	expected := []string{"val3", "", "z"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("GetColumnValues(C) = %v, want %v", values, expected)
	}
}

func TestParseCSV_TrimsWhitespace(t *testing.T) {
	// Create a CSV file with extra whitespace in headers and values
	content := `  Name  , Location ,Status  
  Item1  ,  SS4:AB100  ,  Active  
Item2   ,PS2:CL200  , Inactive`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "whitespace.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	csv, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	// Check headers are trimmed
	expectedHeaders := []string{"Name", "Location", "Status"}
	if !reflect.DeepEqual(csv.Headers, expectedHeaders) {
		t.Errorf("Headers = %v, want %v", csv.Headers, expectedHeaders)
	}

	// Check first row values are trimmed
	expectedRow1 := []string{"Item1", "SS4:AB100", "Active"}
	if !reflect.DeepEqual(csv.Rows[0], expectedRow1) {
		t.Errorf("Rows[0] = %v, want %v", csv.Rows[0], expectedRow1)
	}

	// Check second row values are trimmed
	expectedRow2 := []string{"Item2", "PS2:CL200", "Inactive"}
	if !reflect.DeepEqual(csv.Rows[1], expectedRow2) {
		t.Errorf("Rows[1] = %v, want %v", csv.Rows[1], expectedRow2)
	}
}
