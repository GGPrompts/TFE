package main

// Module: config.go
// Purpose: Unified configuration system for TFE settings
// Responsibilities:
// - Defining default configuration (matches current hardcoded behavior)
// - Loading configuration from ~/.config/tfe/config.toml
// - Saving configuration (write default file if none exists)
// - Providing sensible defaults for all settings

import (
	"bytes"
	"fmt"
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
	SortOrder       string `toml:"sort_order"`         // "name", "size", "modified", or "type"
	StartupDualPane   bool   `toml:"startup_dual_pane"`   // Open with preview pane visible
	StartupFocus      string `toml:"startup_focus"`       // "files" or "preview" — which pane is focused on startup
	FocusedPaneRatio  int    `toml:"focused_pane_ratio"`  // Focused pane width % in accordion layout (50-90, default 60)

	// External tools
	Editor string `toml:"editor"` // Preferred editor command (empty = use $EDITOR)

	// Profiles (launchable terminal sessions from the Profiles menu)
	Profiles []Profile `toml:"profiles,omitempty"` // Custom profiles; nil/empty = use defaults

	// Theme colors (optional — falls back to theme.toml or defaults)
	Theme *Theme `toml:"theme,omitempty"` // Inline theme; nil means not present in config
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
		StartupDualPane:    true,
		StartupFocus:       "files",
		FocusedPaneRatio:   60,
		Editor:             "",
		Profiles: []Profile{
			{Name: "Shell Here", Command: "bash"},
			{Name: "Claude Here", Command: "claude"},
		},
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
// The second return value is a non-empty warning message when the config
// file existed but could not be parsed; in that case the corrupt file is
// renamed aside (config.toml.corrupt-<timestamp>) so a later persistConfig
// cannot overwrite user customizations with defaults. Loaders run before the
// model exists, so they return the warning for initialModel to surface rather
// than calling setStatusMessage directly.
func loadConfig() (Config, string) {
	cfg := defaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create config directory and write default config
			dir := filepath.Dir(path)
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				return cfg, ""
			}
			_ = saveConfig(cfg)
		}
		return cfg, ""
	}

	if _, err := toml.Decode(string(data), &cfg); err != nil {
		// Parse failure: preserve the corrupt file before returning defaults,
		// otherwise the next persistConfig would overwrite user customizations
		// (custom [theme] sections, [[profiles]]) with defaults.
		if backupPath, renameErr := quarantineCorruptFile(path); renameErr == nil {
			return defaultConfig(), fmt.Sprintf("Config file was corrupt; backed up to %s", filepath.Base(backupPath))
		}
		return defaultConfig(), "Config file was corrupt and could not be backed up; not overwriting"
	}

	// Check whether [theme] section was actually present in the file
	var raw map[string]interface{}
	if _, err := toml.Decode(string(data), &raw); err == nil {
		if _, hasTheme := raw["theme"]; !hasTheme {
			// No [theme] section in config.toml — leave cfg.Theme nil
			cfg.Theme = nil
		}
	}

	return cfg, ""
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

	// Render to an in-memory buffer first, then write atomically so a crash
	// mid-write can't leave config.toml truncated/corrupt.
	var buf bytes.Buffer
	buf.WriteString("# TFE Configuration\n# See CLAUDE.md for documentation\n\n")
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}

	return atomicWriteFile(path, buf.Bytes(), 0644)
}

// persistConfig syncs the current model state to m.config and saves to disk.
// Called after any runtime toggle that should be remembered across sessions.
func (m *model) persistConfig() {
	m.config.PanelLock = m.panelsLocked
	m.config.ShowHidden = m.showHidden
	m.config.DarkMode = !m.forceLightTheme
	m.config.SortOrder = m.sortBy
	if err := saveConfig(m.config); err != nil {
		m.statusMessage = fmt.Sprintf("Failed to save config: %v", err)
	}
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
