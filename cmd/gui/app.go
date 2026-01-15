package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
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

	// Ensure qa_loc_rules.yaml exists in the same directory as locations file
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
		result.ColumnGap,
		result.MaxRows,
	)

	// Generate output filename in same directory as locations file
	outputPath := filepath.Join(rulesDir, "location_priorities.xlsx")

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

// ShortenPath returns a display-friendly shortened path
// Replaces home directory with ~ for cleaner display
func (a *App) ShortenPath(path string) string {
	if path == "" {
		return ""
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, just return the filename
		return filepath.Base(path)
	}

	// Normalize paths for comparison
	cleanPath := filepath.Clean(path)
	cleanHome := filepath.Clean(homeDir)

	// Check if path starts with home directory
	if strings.HasPrefix(cleanPath, cleanHome) {
		// Replace home directory with ~
		relativePart := strings.TrimPrefix(cleanPath, cleanHome)
		return "~" + relativePart
	}

	// If not under home directory, show just the parent folder + filename
	// This handles Windows paths like C:\Users\... without showing the full path
	dir := filepath.Dir(cleanPath)
	parent := filepath.Base(dir)
	filename := filepath.Base(cleanPath)

	if parent != "" && parent != "." {
		return filepath.Join(parent, filename)
	}

	return filename
}

// OpenFile opens a file with the system's default application
func (a *App) OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("no file path provided")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default: // Linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
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
