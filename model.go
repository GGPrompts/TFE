package main

import (
	"os"
	"time"

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
		viewMode:          viewSinglePane, // Will be set to dual-pane if narrow terminal
		focusedPane:       leftPane,
		lastClickIndex:    -1,
		preview: previewModel{
			maxPreview: 10000, // Max 10k lines
		},
		spinner:             s,
		loading:             false,
		favorites:           loadFavorites(),
		showFavoritesOnly:   false,
		expandedDirs:        make(map[string]bool),
		commandHistory:      loadCommandHistory(), // Load from disk on startup
		commandCursorPos:    0,
		historyPos:          0,
		commandFocused:      false, // Start in file browser mode, not command mode
		previewMouseEnabled: true,  // Mouse enabled by default
		// Menu system
		startupTime:      time.Now(),
		menuOpen:         false,
		activeMenu:       "",
		selectedMenuItem: -1,
		menuBarFocused:   false,
		highlightedMenu:  "",
		// Menu caching - check tool availability once at startup (performance optimization)
		toolsAvailable: map[string]bool{
			"lazygit":       editorAvailable("lazygit"),
			"lazydocker":    editorAvailable("lazydocker"),
			"lnav":          editorAvailable("lnav"),
			"htop":          editorAvailable("htop"),
			"bottom":        editorAvailable("bottom"),
			"micro":         editorAvailable("micro"), // Used in context menu edit action
			"textual-paint": editorAvailable("textual-paint"), // Used for new image creation
		},
		cachedMenus: nil, // Will be built on first access
	}

	m.loadFiles()

	// Auto-enable dual-pane mode on narrow terminals (phones/Termux)
	// Dual-pane works better on narrow screens - less horizontal scrolling needed
	if m.width < 100 {
		m.viewMode = viewDualPane
	}

	m.calculateLayout()
	return m
}

// calculateLayout calculates left and right pane widths for dual-pane mode
// Uses accordion-style layout: focused pane gets 2/3, unfocused gets 1/3
// Exception: Vertical split (Detail view or narrow terminals) uses full width for both panes
func (m *model) calculateLayout() {
	if m.viewMode == viewSinglePane || m.viewMode == viewFullPreview {
		m.leftWidth = m.width
		m.rightWidth = 0
	} else {
		// Check if using vertical split (Detail always uses vertical, List/Tree on narrow terminals)
		useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

		if useVerticalSplit {
			// Vertical split (stacked layout) - set full width for both panes
			// (actual rendering uses full width for top and bottom panes)
			m.leftWidth = m.width   // Full width for top pane (file list)
			m.rightWidth = m.width  // Full width for bottom pane (preview)
		} else {
			// List/Tree view on wide terminals: accordion-style horizontal split
			// Focused pane gets 2/3, unfocused gets 1/3
			if m.focusedPane == leftPane {
				m.leftWidth = (m.width * 2) / 3  // 66%
				m.rightWidth = m.width / 3       // 33%
			} else {
				m.leftWidth = m.width / 3        // 33%
				m.rightWidth = (m.width * 2) / 3 // 66%
			}

			// Ensure minimum widths for usability
			if m.leftWidth < 30 {
				m.leftWidth = 30
			}
			if m.rightWidth < 30 {
				m.rightWidth = 30
			}

			// Adjust for separator (1 char between horizontal panes)
			if m.focusedPane == leftPane {
				m.rightWidth = m.width - m.leftWidth - 1
			} else {
				m.leftWidth = m.width - m.rightWidth - 1
			}
		}
	}
}
