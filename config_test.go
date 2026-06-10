package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadConfig_CorruptedFile verifies that a corrupt config.toml is preserved
// (renamed aside) and a warning is returned, rather than being silently
// overwritten with defaults by a later persistConfig.
func TestLoadConfig_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configDir := filepath.Join(tmpDir, ".config", "tfe")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	cfgPath := filepath.Join(configDir, "config.toml")
	// Invalid TOML: unterminated string / bad syntax that the user customized
	corruptContent := "this is = not valid toml [[["
	if err := os.WriteFile(cfgPath, []byte(corruptContent), 0644); err != nil {
		t.Fatalf("Failed to write corrupt config: %v", err)
	}

	cfg, warning := loadConfig()

	// Should fall back to defaults
	if cfg.DefaultViewMode != defaultConfig().DefaultViewMode {
		t.Errorf("Corrupt config should fall back to defaults, got view mode %q", cfg.DefaultViewMode)
	}
	if warning == "" {
		t.Error("Corrupt config should return a non-empty warning")
	}

	// Original path must no longer hold the corrupt file (renamed aside)
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Error("Corrupt config file should have been renamed aside")
	}

	// A backup preserving the original content should exist
	matches, err := filepath.Glob(cfgPath + ".corrupt-*")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("Expected 1 corrupt backup, got %d", len(matches))
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(data) != corruptContent {
		t.Errorf("Backup should preserve original content, got %q", string(data))
	}
}

// TestLoadConfig_ValidFile verifies a well-formed config loads without warning
// and round-trips through saveConfig (atomic write).
func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := defaultConfig()
	cfg.SortOrder = "size"
	cfg.ShowHidden = true
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	loaded, warning := loadConfig()
	if warning != "" {
		t.Errorf("Valid config should not return a warning, got %q", warning)
	}
	if loaded.SortOrder != "size" {
		t.Errorf("Expected SortOrder 'size', got %q", loaded.SortOrder)
	}
	if !loaded.ShowHidden {
		t.Error("Expected ShowHidden true")
	}

	// No corrupt backup should be created for a valid file
	matches, _ := filepath.Glob(filepath.Join(tmpDir, ".config", "tfe", "config.toml.corrupt-*"))
	if len(matches) != 0 {
		t.Errorf("Valid config should not produce a corrupt backup, found %d", len(matches))
	}
}

// TestQuarantineCorruptFile verifies the helper renames the file aside and
// returns a timestamped path with the expected prefix.
func TestQuarantineCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.json")
	content := "original-corrupt-content"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	backupPath, err := quarantineCorruptFile(path)
	if err != nil {
		t.Fatalf("quarantineCorruptFile failed: %v", err)
	}

	if !strings.HasPrefix(backupPath, path+".corrupt-") {
		t.Errorf("Backup path %q should start with %q", backupPath, path+".corrupt-")
	}

	// Original should be gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Original file should have been renamed away")
	}

	// Backup should hold the original content
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(data) != content {
		t.Errorf("Backup content mismatch: got %q", string(data))
	}

	// Renaming a non-existent file should return an error
	if _, err := quarantineCorruptFile(filepath.Join(tmpDir, "does-not-exist")); err == nil {
		t.Error("Expected error quarantining a non-existent file")
	}
}
