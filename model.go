package main

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

func initialModel() model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#5fd7ff"})

	m := model{
		currentPath:       cwd,
		cursor:            0,
		height:            24,
		width:             80,
		showHidden:        false,
		displayMode:       modeTree, // Tree view works better on narrow terminals
		sortBy:            "name",
		sortAsc:           true,
		viewMode:          viewSinglePane,
		focusedPane:       leftPane,
		lastClickIndex:    -1,
		preview: previewModel{
			maxPreview: 10000, // Max 10k lines
		},
		spinner:           s,
		loading:           false,
		favorites:         loadFavorites(),
		showFavoritesOnly: false,
		expandedDirs:      make(map[string]bool),
		commandFocused:    false, // Start in file browser mode, not command mode
	}

	m.loadFiles()
	m.calculateLayout()
	return m
}

// calculateLayout calculates left and right pane widths for dual-pane mode
func (m *model) calculateLayout() {
	if m.viewMode == viewSinglePane || m.viewMode == viewFullPreview {
		m.leftWidth = m.width
		m.rightWidth = 0
	} else {
		// 40/60 split for dual-pane
		m.leftWidth = m.width * 40 / 100
		m.rightWidth = m.width - m.leftWidth - 1 // -1 for separator
		if m.leftWidth < 20 {
			m.leftWidth = 20
		}
		if m.rightWidth < 30 {
			m.rightWidth = 30
		}
	}
}
