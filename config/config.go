package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed default_rules.yaml
var DefaultRulesYAML []byte

const DefaultRulesFileName = "qa_loc_rules.yaml"

// EnsureRulesFile checks if qa_loc_rules.yaml exists, creates it from embedded default if not.
// Returns the path to the rules file and a boolean indicating if a new file was created.
func EnsureRulesFile(dir string) (string, error) {
	rulesPath := filepath.Join(dir, DefaultRulesFileName)

	// Check if file exists
	_, err := os.Stat(rulesPath)
	if err == nil {
		// File exists
		return rulesPath, nil
	}

	if !os.IsNotExist(err) {
		// Some other error (e.g., permission denied)
		return "", fmt.Errorf("failed to check rules file: %w", err)
	}

	// File doesn't exist, create from embedded default
	if err := os.WriteFile(rulesPath, DefaultRulesYAML, 0644); err != nil {
		return "", fmt.Errorf("failed to create rules file: %w", err)
	}

	return rulesPath, nil
}

// GetRulesPath returns the path to qa_loc_rules.yaml in the given directory
// Does not create the file, just returns the path
func GetRulesPath(dir string) string {
	return filepath.Join(dir, DefaultRulesFileName)
}

// GetDefaultRules returns the embedded default rules as a string
func GetDefaultRules() string {
	return string(DefaultRulesYAML)
}
