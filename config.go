package main

// Module: config.go
// Purpose: Unified configuration system for TFE settings
// Responsibilities:
// - Defining default configuration (matches current hardcoded behavior)
// - Loading configuration from ~/.config/tfe/config.toml
// - Saving configuration (write default file if none exists)
// - Providing sensible defaults for all settings

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all persistent TFE settings loaded from config.toml
type Config struct {
	// Appearance
	DarkMode bool `toml:"dark_mode"` // true = dark theme (default), false = light theme

	// Behavior
	AutoChanges        bool `toml:"auto_changes"`         // Auto-open changes mode when agent finishes (TFE_AUTO_CHANGES)
	FileWatcherEnabled bool `toml:"file_watcher_enabled"` // Enable fsnotify file watcher for live refresh

	// View defaults
	DefaultViewMode string `toml:"default_view_mode"` // "tree", "list", or "detail"
	PanelLock       bool   `toml:"panel_lock"`         // Lock panel widths (disable accordion)
	ShowHidden      bool   `toml:"show_hidden"`        // Show hidden files by default
	SortOrder       string `toml:"sort_order"`         // "name", "size", or "modified"

	// External tools
	Editor string `toml:"editor"` // Preferred editor command (empty = use $EDITOR)
}

// defaultConfig returns the built-in configuration matching TFE's current hardcoded behavior
func defaultConfig() Config {
	return Config{
		DarkMode:           true,
		AutoChanges:        false,
		FileWatcherEnabled: true,
		DefaultViewMode:    "tree",
		PanelLock:          false,
		ShowHidden:         false,
		SortOrder:          "name",
		Editor:             "",
	}
}

// configPath returns the path to ~/.config/tfe/config.toml
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tfe", "config.toml"), nil
}

// loadConfig reads configuration from ~/.config/tfe/config.toml.
// If the file doesn't exist, it creates the config directory and writes
// a default config file. Missing fields in an existing file are filled
// with default values.
func loadConfig() Config {
	cfg := defaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create config directory and write default config
			dir := filepath.Dir(path)
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				return cfg
			}
			_ = saveConfig(cfg)
		}
		return cfg
	}

	if _, err := toml.Decode(string(data), &cfg); err != nil {
		// Parse error: return defaults
		return defaultConfig()
	}

	return cfg
}

// saveConfig writes the configuration to ~/.config/tfe/config.toml
func saveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	if _, err := f.WriteString("# TFE Configuration\n# See CLAUDE.md for documentation\n\n"); err != nil {
		return err
	}

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// persistConfig syncs the current model state to m.config and saves to disk.
// Called after any runtime toggle that should be remembered across sessions.
func (m *model) persistConfig() {
	m.config.PanelLock = m.panelsLocked
	m.config.ShowHidden = m.showHidden
	m.config.DarkMode = !m.forceLightTheme
	m.config.SortOrder = m.sortBy
	_ = saveConfig(m.config)
}

// parseViewMode converts a config string to a displayMode value
func parseViewMode(s string) displayMode {
	switch s {
	case "list":
		return modeList
	case "detail":
		return modeDetail
	case "tree":
		return modeTree
	default:
		return modeTree
	}
}
