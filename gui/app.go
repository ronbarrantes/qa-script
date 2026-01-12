package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx           context.Context
	csvPath       string
	excelPath     string
	csvValidated  bool
	excelValidated bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ValidationResult holds the result of file validation
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	FileName    string   `json:"fileName"`
	FilePath    string   `json:"filePath"`
	Message     string   `json:"message"`
	Headers     []string `json:"headers"`
	RowCount    int      `json:"rowCount"`
}

// ValidateCSV validates a CSV file for the required "Location" column
func (a *App) ValidateCSV(filePath string) ValidationResult {
	result := ValidationResult{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" {
		result.Message = "Invalid file type. Please drop a CSV file (.csv)"
		return result
	}

	// Open and parse CSV
	file, err := os.Open(filePath)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to open file: %v", err)
		return result
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		result.Message = fmt.Sprintf("Failed to parse CSV: %v", err)
		return result
	}

	if len(records) == 0 {
		result.Message = "CSV file is empty"
		return result
	}

	// Get headers
	headers := records[0]
	result.Headers = headers
	result.RowCount = len(records) - 1

	// Check for required "Location" column
	hasLocation := false
	for _, h := range headers {
		if strings.TrimSpace(h) == "Location" {
			hasLocation = true
			break
		}
	}

	if !hasLocation {
		result.Message = "Missing required column: 'Location'. Found columns: " + strings.Join(headers, ", ")
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Valid CSV file with %d data rows", result.RowCount)
	a.csvPath = filePath
	a.csvValidated = true

	return result
}

// ValidateExcel validates an Excel file for required columns
func (a *App) ValidateExcel(filePath string) ValidationResult {
	result := ValidationResult{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".xlsx" && ext != ".xls" {
		result.Message = "Invalid file type. Please drop an Excel file (.xlsx or .xls)"
		return result
	}

	// Open and parse Excel
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to open file: %v", err)
		return result
	}
	defer f.Close()

	// Get first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to read sheet: %v", err)
		return result
	}

	if len(rows) == 0 {
		result.Message = "Excel file is empty"
		return result
	}

	// Get headers
	headers := rows[0]
	result.Headers = headers
	result.RowCount = len(rows) - 1

	// Check for required columns
	hasContainerTag := false
	hasCurrentLocation := false
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h == "Container Tag" {
			hasContainerTag = true
		}
		if h == "Current Location" {
			hasCurrentLocation = true
		}
	}

	missingCols := []string{}
	if !hasContainerTag {
		missingCols = append(missingCols, "'Container Tag'")
	}
	if !hasCurrentLocation {
		missingCols = append(missingCols, "'Current Location'")
	}

	if len(missingCols) > 0 {
		result.Message = fmt.Sprintf("Missing required columns: %s. Found columns: %s", 
			strings.Join(missingCols, ", "), strings.Join(headers, ", "))
		return result
	}

	result.Valid = true
	result.Message = fmt.Sprintf("Valid Excel file with %d data rows", result.RowCount)
	a.excelPath = filePath
	a.excelValidated = true

	return result
}

// FileSelection holds the selected file paths
type FileSelection struct {
	CSVPath   string `json:"csvPath"`
	ExcelPath string `json:"excelPath"`
}

// GetSelectedFiles returns the currently selected and validated files
func (a *App) GetSelectedFiles() FileSelection {
	return FileSelection{
		CSVPath:   a.csvPath,
		ExcelPath: a.excelPath,
	}
}

// ClearCSV clears the CSV selection
func (a *App) ClearCSV() {
	a.csvPath = ""
	a.csvValidated = false
}

// ClearExcel clears the Excel selection
func (a *App) ClearExcel() {
	a.excelPath = ""
	a.excelValidated = false
}

// ProcessFiles initiates the file processing and returns the result
func (a *App) ProcessFiles() map[string]interface{} {
	result := make(map[string]interface{})
	
	if !a.csvValidated {
		result["success"] = false
		result["error"] = "CSV file not validated. Please drop a valid CSV file."
		return result
	}

	result["success"] = true
	result["csvPath"] = a.csvPath
	result["excelPath"] = a.excelPath
	result["message"] = "Files ready for processing"
	
	return result
}

// Cancel cancels the operation and quits the app
func (a *App) Cancel() {
	runtime.Quit(a.ctx)
}

// OpenFileDialog opens a file dialog for selecting files
func (a *App) OpenFileDialog(fileType string) string {
	var filters []runtime.FileFilter
	var title string
	
	if fileType == "csv" {
		title = "Select CSV File"
		filters = []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		}
	} else {
		title = "Select Excel File"
		filters = []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx, *.xls)", Pattern: "*.xlsx;*.xls"},
		}
	}
	
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
		Filters: filters,
	})
	
	if err != nil {
		return ""
	}
	
	return filePath
}
