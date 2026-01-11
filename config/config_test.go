package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_CommaSeparatedValues(t *testing.T) {
	// Create a temporary YAML file with comma-separated values
	content := `
groups:
  - name: pallets
    values: a, b, c, lud, prm, slp
  - name: e - g
    values: e, f, g, gft, hwk, hvc
size: 5
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(cfg.Groups))
	}

	// Check first group
	if cfg.Groups[0].Name != "pallets" {
		t.Errorf("expected group name 'pallets', got %q", cfg.Groups[0].Name)
	}
	expectedValues := []string{"a", "b", "c", "lud", "prm", "slp"}
	if len(cfg.Groups[0].Values) != len(expectedValues) {
		t.Errorf("expected %d values, got %d", len(expectedValues), len(cfg.Groups[0].Values))
	}
	for i, v := range expectedValues {
		if cfg.Groups[0].Values[i] != v {
			t.Errorf("value %d: expected %q, got %q", i, v, cfg.Groups[0].Values[i])
		}
	}

	if cfg.Size != 5 {
		t.Errorf("expected size 5, got %d", cfg.Size)
	}
}

func TestLoadConfig_ArrayValues(t *testing.T) {
	// Create a temporary YAML file with array values
	content := `
groups:
  - name: test
    values: [x, y, z]
size: 10
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(cfg.Groups))
	}

	expectedValues := []string{"x", "y", "z"}
	if len(cfg.Groups[0].Values) != len(expectedValues) {
		t.Errorf("expected %d values, got %d", len(expectedValues), len(cfg.Groups[0].Values))
	}
	for i, v := range expectedValues {
		if cfg.Groups[0].Values[i] != v {
			t.Errorf("value %d: expected %q, got %q", i, v, cfg.Groups[0].Values[i])
		}
	}
}

func TestSaveConfig_OutputsCommaSeparated(t *testing.T) {
	cfg := &Config{
		Groups: []Group{
			{Name: "test", Values: StringList{"a", "b", "c"}},
		},
		Size: 5,
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.yaml")
	if err := SaveConfig(cfg, tmpFile); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Read back and verify it's parseable
	loaded, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed after SaveConfig: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(loaded.Groups))
	}
	if len(loaded.Groups[0].Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(loaded.Groups[0].Values))
	}
}
