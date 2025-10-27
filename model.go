package main

import (
	"os"
	"strings"
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
		terminalType:      detectTerminalType(), // Detect terminal for emoji width compensation
		displayMode:       modeTree,             // Tree view works better on narrow terminals
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
		gitReposScanDepth:   3,   // Default scan depth: 3 levels (safer)
		gitReposList:        make([]fileItem, 0),
		expandedDirs:        make(map[string]bool),
		// Prompt inline editing
		promptEditMode:       false,
		focusedVariableIndex: 0,
		filledVariables:      make(map[string]string),
		// Command history will be initialized below after loading from disk
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
			"pyradio":       editorAvailable("pyradio"),
			"micro":         editorAvailable("micro"), // Used in context menu edit action
			"textual-paint": editorAvailable("textual-paint"), // Used for new image creation
		},
		cachedMenus: nil, // Will be built on first access
		// Performance caching
		promptDirsCache: make(map[string]bool), // Cache for prompts filter performance
	}

	// Load command history from disk (supports per-directory and global history)
	commandHistoryByDir, commandHistoryGlobal := loadCommandHistory()
	m.commandHistoryByDir = commandHistoryByDir
	m.commandHistoryGlobal = commandHistoryGlobal
	// Build combined history for current directory
	m.rebuildCombinedHistory()

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
// Uses accordion-style layout: focused pane gets 60%, unfocused gets 40%
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
			// Focused pane gets 60%, unfocused gets 40%
			if m.focusedPane == leftPane {
				m.leftWidth = (m.width * 60) / 100  // 60%
				m.rightWidth = (m.width * 40) / 100 // 40%
			} else {
				m.leftWidth = (m.width * 40) / 100  // 40%
				m.rightWidth = (m.width * 60) / 100 // 60%
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

// detectTerminalType determines which terminal emulator is being used
// This is used for emoji width compensation (variation selector handling)
func detectTerminalType() terminalType {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	wtSession := os.Getenv("WT_SESSION")
	wezterm := os.Getenv("WEZTERM_EXECUTABLE")

	// Manual override for testing/debugging
	if override := os.Getenv("TFE_TERMINAL_TYPE"); override != "" {
		switch override {
		case "windows-terminal", "wt":
			return terminalWindowsTerminal
		case "wezterm":
			return terminalWezTerm
		case "kitty":
			return terminalKitty
		case "iterm2":
			return terminalITerm2
		case "xterm":
			return terminalXterm
		case "termux":
			return terminalTermux
		}
	}

	// Check for Windows Terminal (most reliable indicator)
	if wtSession != "" {
		return terminalWindowsTerminal
	}

	// Check for WezTerm
	if wezterm != "" || termProgram == "WezTerm" {
		return terminalWezTerm
	}

	// Check for Kitty
	if strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
		return terminalKitty
	}

	// Check for iTerm2
	if termProgram == "iTerm.app" {
		return terminalITerm2
	}

	// Check for xterm
	if term == "xterm" || term == "xterm-256color" {
		return terminalXterm
	}

	// Check for Termux (Android)
	if strings.Contains(os.Getenv("PREFIX"), "com.termux") {
		return terminalTermux
	}

	return terminalUnknown
}
