package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed default_rules.yaml
var DefaultRulesYAML []byte

const DefaultRulesFileName = "rules.yaml"

// GetDefaultRulesDir returns the default directory for storing rules.yaml
// This is the user's Documents folder, which is a standard location for user configuration
func GetDefaultRulesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Use Documents folder on all platforms
	var documentsDir string
	switch runtime.GOOS {
	case "windows":
		// On Windows, Documents is typically under the user profile
		documentsDir = filepath.Join(homeDir, "Documents")
	default:
		// On macOS and Linux, use ~/Documents
		documentsDir = filepath.Join(homeDir, "Documents")
	}

	// Ensure the Documents directory exists
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create documents directory: %w", err)
	}

	return documentsDir, nil
}

// GetDefaultRulesPath returns the full path to the default rules.yaml location
func GetDefaultRulesPath() (string, error) {
	dir, err := GetDefaultRulesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DefaultRulesFileName), nil
}

// EnsureRulesFile checks if rules.yaml exists, creates it from embedded default if not
// Returns the path to the rules file
func EnsureRulesFile(dir string) (string, error) {
	rulesPath := filepath.Join(dir, DefaultRulesFileName)

	// Check if file exists
	info, err := os.Stat(rulesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - create from embedded default
			if err := os.WriteFile(rulesPath, DefaultRulesYAML, 0644); err != nil {
				return "", fmt.Errorf("failed to create rules file: %w", err)
			}
			fmt.Printf("Created default rules file: %s\n", rulesPath)
		} else {
			// Some other error (permission, etc.) - return the error
			return "", fmt.Errorf("failed to check rules file: %w", err)
		}
	} else {
		// File exists - log its details for debugging
		fmt.Printf("Using existing rules file: %s (size: %d bytes, modified: %s)\n",
			rulesPath, info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
	}

	return rulesPath, nil
}

// EnsureDefaultRulesFile ensures rules.yaml exists in the default Documents location
// Returns the path to the rules file
func EnsureDefaultRulesFile() (string, error) {
	dir, err := GetDefaultRulesDir()
	if err != nil {
		return "", err
	}
	return EnsureRulesFile(dir)
}

// GetRulesPath returns the path to rules.yaml in the given directory
// Does not create the file, just returns the path
func GetRulesPath(dir string) string {
	return filepath.Join(dir, DefaultRulesFileName)
}

// GetDefaultRules returns the embedded default rules as a string
func GetDefaultRules() string {
	return string(DefaultRulesYAML)
}
