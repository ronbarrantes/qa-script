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
	"gopkg.in/yaml.v3"

	"qa-script/config"
	"qa-script/output"
	"qa-script/processor"
	"qa-script/rules"
)

// Constants for GUI config
const (
	appConfigDirName  = "qa-loc-priorities"
	guiRulesFileName  = "rules.yaml"
)

// App struct
type App struct {
	ctx            context.Context
	locationsFile  string
	prioritiesFile string
}

// getConfigDir returns the platform-specific config directory for the GUI app
// Mac/Linux: ~/.config/qa-loc-priorities/
// Windows: %LOCALAPPDATA%\qa-loc-priorities\
func getConfigDir() (string, error) {
	var baseDir string

	switch goruntime.GOOS {
	case "windows":
		// Use %LOCALAPPDATA% on Windows
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			// Fallback to UserHomeDir + AppData\Local
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("could not determine home directory: %w", err)
			}
			baseDir = filepath.Join(home, "AppData", "Local")
		}
	default:
		// Mac and Linux use ~/.config
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".config")
	}

	return filepath.Join(baseDir, appConfigDirName), nil
}

// getGUIRulesPath returns the full path to the GUI rules file in the config directory
// Creates the config directory if it doesn't exist
func getGUIRulesPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, guiRulesFileName), nil
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

	// Get rules from config directory (GUI always uses config dir, never creates files next to CSV)
	rulesPath, err := getGUIRulesPath()
	if err != nil {
		return "", fmt.Errorf("failed to get rules path: %w", err)
	}

	// If rules file doesn't exist, create from defaults
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		if err := a.createDefaultRulesFile(rulesPath); err != nil {
			return "", fmt.Errorf("failed to create default rules: %w", err)
		}
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
	outputDir := filepath.Dir(a.locationsFile)
	outputPath := filepath.Join(outputDir, "location_priorities.xlsx")

	// Write XLSX output
	if err := output.WriteXLSX(outputPath, outputData); err != nil {
		return "", fmt.Errorf("failed to write XLSX: %w", err)
	}

	return outputPath, nil
}

// createDefaultRulesFile creates a default rules file at the specified path
func (a *App) createDefaultRulesFile(path string) error {
	// Parse embedded defaults
	var cfg rules.Config
	if err := yaml.Unmarshal(config.DefaultRulesYAML, &cfg); err != nil {
		return fmt.Errorf("failed to parse default rules: %w", err)
	}

	// Convert to YAML struct with flow-style arrays
	yamlCfg := rulesConfigYAML{
		MaxRows:   cfg.MaxRows,
		ColumnGap: cfg.ColumnGap,
		Groups:    make([]rulesGroupYAML, len(cfg.Groups)),
	}
	for i, g := range cfg.Groups {
		yamlCfg.Groups[i] = rulesGroupYAML{
			Title:  g.Title,
			Values: g.Values,
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&yamlCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
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
	// URL decode common characters that may appear in file paths
	// Using manual replacement for common cases to avoid pulling in net/url
	replacements := map[string]string{
		"%20": " ",
		"%23": "#",
		"%25": "%",
		"%26": "&",
		"%27": "'",
		"%28": "(",
		"%29": ")",
		"%2B": "+",
		"%2C": ",",
		"%3B": ";",
		"%3D": "=",
		"%40": "@",
		"%5B": "[",
		"%5D": "]",
	}
	for encoded, decoded := range replacements {
		path = strings.ReplaceAll(path, encoded, decoded)
		// Also handle lowercase variants
		path = strings.ReplaceAll(path, strings.ToLower(encoded), decoded)
	}
	return path
}

// RulesConfigJS is a JavaScript-friendly version of the rules config
type RulesConfigJS struct {
	Groups    []GroupJS `json:"groups"`
	MaxRows   int       `json:"maxRows"`
	ColumnGap int       `json:"columnGap"`
}

// GroupJS is a JavaScript-friendly version of a group
type GroupJS struct {
	Title  string `json:"title"`
	Values string `json:"values"` // Comma-separated string for easier UI editing
}

// GetRulesConfig reads the current rules configuration from the platform-specific config directory
// Falls back to embedded defaults if no config file exists
// Mac/Linux: ~/.config/qa-loc-priorities/rules.yaml
// Windows: %LOCALAPPDATA%\qa-loc-priorities\rules.yaml
func (a *App) GetRulesConfig() (*RulesConfigJS, error) {
	// Always use platform-specific config directory
	rulesPath, err := getGUIRulesPath()
	if err != nil {
		// Fall back to embedded defaults if we can't get config path
		return a.getEmbeddedDefaults()
	}

	// Check if config file exists
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		// No config file yet, use embedded defaults
		return a.getEmbeddedDefaults()
	}

	// Load the rules file
	cfg, err := rules.LoadConfig(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Convert to JS-friendly format
	return a.configToJS(cfg), nil
}

// getEmbeddedDefaults returns the embedded default rules as RulesConfigJS
func (a *App) getEmbeddedDefaults() (*RulesConfigJS, error) {
	var cfg rules.Config
	if err := yaml.Unmarshal(config.DefaultRulesYAML, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse default rules: %w", err)
	}
	return a.configToJS(&cfg), nil
}

// configToJS converts a rules.Config to RulesConfigJS
func (a *App) configToJS(cfg *rules.Config) *RulesConfigJS {
	result := &RulesConfigJS{
		MaxRows:   cfg.MaxRows,
		ColumnGap: cfg.ColumnGap,
		Groups:    make([]GroupJS, len(cfg.Groups)),
	}
	for i, g := range cfg.Groups {
		result.Groups[i] = GroupJS{
			Title:  g.Title,
			Values: strings.Join(g.Values, ", "),
		}
	}
	return result
}

// rulesConfigYAML is used for YAML marshaling with flow-style arrays [a, b, c]
type rulesConfigYAML struct {
	Groups    []rulesGroupYAML `yaml:"groups"`
	MaxRows   int              `yaml:"max_rows"`
	ColumnGap int              `yaml:"column_gap"`
}

type rulesGroupYAML struct {
	Title  string   `yaml:"title"`
	Values []string `yaml:"values,flow"` // flow style: [a, b, c] instead of - a\n- b\n- c
}

// SaveRulesConfig saves the rules configuration to the platform-specific config directory
// Mac/Linux: ~/.config/qa-loc-priorities/rules.yaml
// Windows: %LOCALAPPDATA%\qa-loc-priorities\rules.yaml
func (a *App) SaveRulesConfig(configJS *RulesConfigJS) error {
	// Always save to platform-specific config directory
	rulesPath, err := getGUIRulesPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Convert from JS format to YAML struct with flow-style arrays
	cfg := rulesConfigYAML{
		MaxRows:   configJS.MaxRows,
		ColumnGap: configJS.ColumnGap,
		Groups:    make([]rulesGroupYAML, len(configJS.Groups)),
	}

	for i, g := range configJS.Groups {
		// Parse comma-separated values
		values := strings.Split(g.Values, ",")
		cleanValues := make([]string, 0, len(values))
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				cleanValues = append(cleanValues, v)
			}
		}

		cfg.Groups[i] = rulesGroupYAML{
			Title:  g.Title,
			Values: cleanValues,
		}
	}

	// Marshal to YAML with flow-style arrays
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	// Write to file
	if err := os.WriteFile(rulesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	return nil
}

// GetDefaultRulesConfig returns the embedded default rules configuration
func (a *App) GetDefaultRulesConfig() (*RulesConfigJS, error) {
	return a.getEmbeddedDefaults()
}
