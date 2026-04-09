package main

// Module: theme.go
// Purpose: Configurable theme system for TFE colors
// Responsibilities:
// - Defining default theme (matches original hardcoded colors)
// - Loading theme from ~/.config/tfe/theme.toml
// - Converting theme colors to lipgloss AdaptiveColor values

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
)

// currentTheme holds the active theme used by all style definitions
var currentTheme Theme

// applyThemeMode forces lipgloss to use the configured terminal background mode
// and rebuilds shared styles so CLI/config theme choices override autodetection.
func applyThemeMode(darkMode bool) {
	lipgloss.SetHasDarkBackground(darkMode)
	initStyles()
}

func uiPanelBackground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#f4f6f8",
		Dark:  "#303030",
	}
}

func uiInputBackground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#e7edf3",
		Dark:  "#262626",
	}
}

func uiMutedText() lipgloss.AdaptiveColor {
	return currentTheme.Status.adaptiveColor()
}

func uiSubtleText() lipgloss.AdaptiveColor {
	return currentTheme.BorderUnfocused.adaptiveColor()
}

func uiBodyText() lipgloss.AdaptiveColor {
	return currentTheme.File.adaptiveColor()
}

func uiSuccessBackground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#1f7a1f",
		Dark:  "#1f7a1f",
	}
}

func uiSuccessForeground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#ffffff",
		Dark:  "#ffffff",
	}
}

func uiErrorBackground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#b42318",
		Dark:  "#c62828",
	}
}

func uiInfoBackground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#005faf",
		Dark:  "#0087d7",
	}
}

func uiInfoForeground() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: "#ffffff",
		Dark:  "#ffffff",
	}
}

// adaptiveColor converts a ThemeColor to a lipgloss AdaptiveColor
func (tc ThemeColor) adaptiveColor() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: tc.Light,
		Dark:  tc.Dark,
	}
}

// defaultTheme returns the built-in theme matching TFE's original hardcoded colors
func defaultTheme() Theme {
	return Theme{
		// Core UI colors
		Title: ThemeColor{
			Light: "#0087d7", // Dark blue for light backgrounds
			Dark:  "#5fd7ff", // Bright cyan for dark backgrounds
		},
		Path: ThemeColor{
			Light: "#666666", // Medium gray for light
			Dark:  "#999999", // Light gray for dark
		},
		Status: ThemeColor{
			Light: "#444444",
			Dark:  "#AAAAAA",
		},
		// Selection colors
		SelectionBg: ThemeColor{
			Light: "#0087d7", // Dark blue background for light
			Dark:  "#00d7ff", // Bright cyan background for dark
		},
		SelectionFg: ThemeColor{
			Light: "#FFFFFF", // White text on dark blue
			Dark:  "#000000", // Black text on bright cyan
		},
		NarrowSelect: ThemeColor{
			Light: "#00AF00", // Bright green for light backgrounds
			Dark:  "#00FF00", // Matrix green for dark backgrounds
		},
		// File type colors
		Folder: ThemeColor{
			Light: "#005faf", // Dark blue for light
			Dark:  "#5fd7ff", // Bright cyan for dark
		},
		File: ThemeColor{
			Light: "#000000", // Black for light
			Dark:  "#FFFFFF", // White for dark
		},
		// Special file colors
		ClaudeContext: ThemeColor{
			Light: "#D75F00", // Darker orange for light
			Dark:  "#FF8700", // Bright orange for dark
		},
		Agents: ThemeColor{
			Light: "#8B00FF", // Darker purple for light
			Dark:  "#BD93F9", // Bright purple for dark
		},
		PromptsFolder: ThemeColor{
			Light: "#D7005F", // Dark magenta for light
			Dark:  "#FF79C6", // Bright pink for dark
		},
		ObsidianVault: ThemeColor{
			Light: "#008B8B", // Dark cyan for light
			Dark:  "#50FAE9", // Bright teal for dark
		},
		// Border colors
		BorderFocused: ThemeColor{
			Light: "#0087d7", // Blue (matches title/accent)
			Dark:  "#00d7ff", // Bright cyan
		},
		BorderUnfocused: ThemeColor{
			Light: "#999999", // Medium gray
			Dark:  "#585858", // Dark gray
		},
		// Alternating row background
		AlternateRow: ThemeColor{
			Light: "#eeeeee",
			Dark:  "#333333",
		},
		// Diff colors (git diff preview in changes mode)
		DiffAdded: ThemeColor{
			Light: "#006400", // Dark green for light backgrounds
			Dark:  "#00d700", // Bright green for dark backgrounds
		},
		DiffRemoved: ThemeColor{
			Light: "#8B0000", // Dark red for light backgrounds
			Dark:  "#ff5f5f", // Bright red for dark backgrounds
		},
		DiffHunkHeader: ThemeColor{
			Light: "#0087d7", // Dark cyan for light backgrounds
			Dark:  "#00d7ff", // Bright cyan for dark backgrounds
		},
		DiffMeta: ThemeColor{
			Light: "#888888", // Medium gray for light backgrounds
			Dark:  "#6c6c6c", // Dim gray for dark backgrounds
		},
	}
}

// loadTheme reads a theme from a TOML file at the given path.
// Missing fields in the file are filled with default values.
func loadTheme(path string) (Theme, error) {
	theme := defaultTheme()

	data, err := os.ReadFile(path)
	if err != nil {
		return theme, err
	}

	if _, err := toml.Decode(string(data), &theme); err != nil {
		return defaultTheme(), err
	}

	return theme, nil
}

// initTheme sets the global currentTheme variable.
// Priority order:
//  1. If configTheme is non-nil ([theme] section present in config.toml), use it.
//  2. Else if ~/.config/tfe/theme.toml exists, load from there (backwards compat).
//  3. Else use built-in defaults.
func initTheme(configTheme *Theme) {
	// 1. Config.toml [theme] section takes priority
	if configTheme != nil {
		currentTheme = *configTheme
		return
	}

	// 2. Backwards compat: try legacy theme.toml
	home, err := os.UserHomeDir()
	if err != nil {
		currentTheme = defaultTheme()
		return
	}

	themePath := filepath.Join(home, ".config", "tfe", "theme.toml")
	theme, err := loadTheme(themePath)
	if err != nil {
		// File not found or parse error: use defaults
		currentTheme = defaultTheme()
		return
	}

	currentTheme = theme
}

// initStyles rebuilds all lipgloss style variables from the current theme.
// Must be called after initTheme() sets currentTheme and after applyThemeMode()
// selects the active light/dark branch for adaptive colors.
func initStyles() {
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Title.adaptiveColor())

	pathStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Path.adaptiveColor()).
		PaddingLeft(2)

	statusStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Status.adaptiveColor()).
		PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Foreground(currentTheme.SelectionFg.adaptiveColor())

	narrowSelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.NarrowSelect.adaptiveColor())

	folderStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Folder.adaptiveColor())

	fileStyle = lipgloss.NewStyle().
		Foreground(currentTheme.File.adaptiveColor())

	claudeContextStyle = lipgloss.NewStyle().
		Foreground(currentTheme.ClaudeContext.adaptiveColor())

	agentsStyle = lipgloss.NewStyle().
		Foreground(currentTheme.Agents.adaptiveColor())

	promptsFolderStyle = lipgloss.NewStyle().
		Foreground(currentTheme.PromptsFolder.adaptiveColor())

	obsidianVaultStyle = lipgloss.NewStyle().
		Foreground(currentTheme.ObsidianVault.adaptiveColor())

	// Diff preview styles
	diffAddedStyle = lipgloss.NewStyle().
		Foreground(currentTheme.DiffAdded.adaptiveColor())

	diffRemovedStyle = lipgloss.NewStyle().
		Foreground(currentTheme.DiffRemoved.adaptiveColor())

	diffHunkHeaderStyle = lipgloss.NewStyle().
		Foreground(currentTheme.DiffHunkHeader.adaptiveColor()).
		Bold(true)

	diffMetaStyle = lipgloss.NewStyle().
		Foreground(currentTheme.DiffMeta.adaptiveColor()).
		Italic(true)
}
