package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"qa-script/config"
	"qa-script/output"
	"qa-script/processor"
)

// App struct
type App struct {
	ctx            context.Context
	locationsFile  string
	prioritiesFile string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SetLocationsFile sets the locations CSV file path
func (a *App) SetLocationsFile(path string) error {
	// Clean up file:// prefix if present
	path = cleanFilePath(path)

	if !strings.HasSuffix(strings.ToLower(path), ".csv") {
		return fmt.Errorf("locations file must be a CSV file")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}
	a.locationsFile = path
	return nil
}

// SetPrioritiesFile sets the priorities XLSX file path
func (a *App) SetPrioritiesFile(path string) error {
	// Clean up file:// prefix if present
	path = cleanFilePath(path)

	if !strings.HasSuffix(strings.ToLower(path), ".xlsx") {
		return fmt.Errorf("priorities file must be an XLSX file")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}
	a.prioritiesFile = path
	return nil
}

// GetLocationsFile returns the current locations file path
func (a *App) GetLocationsFile() string {
	return a.locationsFile
}

// GetPrioritiesFile returns the current priorities file path
func (a *App) GetPrioritiesFile() string {
	return a.prioritiesFile
}

// Reset clears both file paths
func (a *App) Reset() {
	a.locationsFile = ""
	a.prioritiesFile = ""
}

// Process runs the location grouping and generates the output file
func (a *App) Process() (string, error) {
	if a.locationsFile == "" {
		return "", fmt.Errorf("no locations file selected")
	}
	if a.prioritiesFile == "" {
		return "", fmt.Errorf("no priorities file selected")
	}

	// Ensure rules.yaml exists in the same directory as locations file
	rulesDir := filepath.Dir(a.locationsFile)
	rulesPath, err := config.EnsureRulesFile(rulesDir)
	if err != nil {
		return "", fmt.Errorf("failed to ensure rules file: %w", err)
	}

	// Process the data
	result, err := processor.Process(a.locationsFile, a.prioritiesFile, rulesPath)
	if err != nil {
		return "", fmt.Errorf("processing failed: %w", err)
	}

	// Create output data
	outputData := output.NewOutputData(
		result.TitleOrder,
		result.TitleGroupedLocations,
		result.PriorityLocations,
		result.Gap,
		result.Size,
	)

	// Generate output filename in same directory as locations file
	outputPath := filepath.Join(rulesDir, "locations_output.xlsx")

	// Write XLSX output
	if err := output.WriteXLSX(outputPath, outputData); err != nil {
		return "", fmt.Errorf("failed to write XLSX: %w", err)
	}

	return outputPath, nil
}

// SelectLocationsFile opens a file dialog to select a CSV file
func (a *App) SelectLocationsFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Locations CSV File",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // User cancelled
	}
	if err := a.SetLocationsFile(path); err != nil {
		return "", err
	}
	return path, nil
}

// SelectPrioritiesFile opens a file dialog to select an XLSX file
func (a *App) SelectPrioritiesFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Priorities XLSX File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // User cancelled
	}
	if err := a.SetPrioritiesFile(path); err != nil {
		return "", err
	}
	return path, nil
}

// cleanFilePath removes file:// prefix and handles URL encoding
func cleanFilePath(path string) string {
	// Remove file:// prefix
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
		// On Windows, file URIs have form file:///C:/path/to/file
		// After trimming file://, we get /C:/path - need to remove the leading /
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
	}
	// URL decode common characters
	path = strings.ReplaceAll(path, "%20", " ")
	return path
}
