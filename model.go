package main

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

func initialModel() model {
	// Load unified configuration from ~/.config/tfe/config.toml
	cfg := loadConfig()

	// Initialize theme from config [theme] section, falling back to theme.toml
	initTheme(cfg.Theme)
	initStyles()

	// CLI flags override config values
	// Check if --light or --dark was explicitly passed
	for _, arg := range os.Args[1:] {
		if arg == "--light" {
			cfg.DarkMode = false
		} else if arg == "--dark" {
			cfg.DarkMode = true
		}
	}

	// Environment variable overrides config for auto-changes
	if os.Getenv("TFE_AUTO_CHANGES") == "1" {
		cfg.AutoChanges = true
	}

	// Use startPath from CLI if provided, otherwise use current working directory
	initialPath := startPath
	if initialPath == "" {
		var err error
		initialPath, err = os.Getwd()
		if err != nil {
			initialPath = "."
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(currentTheme.Title.adaptiveColor())

	m := model{
		currentPath:     initialPath,
		cursor:          0,
		height:          24,
		width:           80,
		showHidden:      cfg.ShowHidden,
		inTmux:          isInsideTmux(),       // Detect if running inside tmux
		terminalType:    detectTerminalType(), // Detect terminal for emoji width compensation
		forceLightTheme: !cfg.DarkMode,        // Config dark_mode maps to inverse of forceLightTheme
		displayMode:       parseViewMode(cfg.DefaultViewMode),
		sortBy:            cfg.SortOrder,
		sortAsc:           true,
		viewMode:          viewSinglePane, // Will be set to dual-pane if terminal width >= 100
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
			"tmux":          editorAvailable("tmux"),          // Used for tmux quad split
			// AI tools (claude, codex, gemini, opencode)
			"claude":        editorAvailable("claude"),
			"codex":         editorAvailable("codex"),
			"gemini":        editorAvailable("gemini"),
			"opencode":      editorAvailable("opencode"),
		},
		tuiClassicsPath: getTUIClassicsPath(), // Cache TUIClassics path (checks multiple locations)
		cachedMenus:     nil,                   // Will be built on first access
		// Performance caching
		promptDirsCache: make(map[string]bool), // Cache for prompts filter performance
		// Agent auto-watch: enable via config or TFE_AUTO_CHANGES=1
		agentAutoWatch:         cfg.AutoChanges,
		lastKnownAgentSessions: make(map[string]string),
		// Unified configuration
		config: cfg,
	}

	// Load command history from disk (supports per-directory and global history)
	commandHistoryByDir, commandHistoryGlobal := loadCommandHistory()
	m.commandHistoryByDir = commandHistoryByDir
	m.commandHistoryGlobal = commandHistoryGlobal
	// Build combined history for current directory
	m.rebuildCombinedHistory()

	// Apply config settings that need post-init setup
	m.panelsLocked = cfg.PanelLock

	m.loadFiles()

	// Initialize file watcher (fsnotify) for live directory refresh
	if cfg.FileWatcherEnabled {
		m.initWatcher()
	}

	// If a file was specified on CLI, find and select it
	if selectFile != "" {
		for i, f := range m.files {
			if f.name == selectFile {
				m.cursor = i
				// Load preview for the selected file if --preview flag is set
				if autoPreview && !f.isDir {
					m.loadPreview(f.path)
				}
				break
			}
		}
	}

	// Auto-enable dual-pane mode on WIDE terminals only
	// Narrow terminals (Termux, phones, small windows) stay in single-pane mode
	// Rationale:
	//   - Wide terminals: Side-by-side dual-pane is great (file list + preview)
	//   - Narrow terminals: Even with vertical-split dual-pane, limited vertical space
	//     means cramped file list AND cramped preview (especially with keyboard up)
	//   - Single-pane: Full screen for file list, press Enter for full-screen preview
	//     Much better use of limited screen space on mobile devices
	if m.width >= 100 {
		m.viewMode = viewDualPane
	}
	// else: keep viewSinglePane (default from line 32)

	// If standalone preview file was specified (tfe --preview /path/to/file),
	// enter preview-only mode: minimal UI showing just the file content
	if previewFile != "" {
		m.previewOnly = true
		m.viewMode = viewFullPreview
		m.loadPreview(previewFile)
		m.calculateLayout()
		m.populatePreviewCache()
		return m
	}

	// If --preview flag was set, enable dual-pane to show preview
	// and focus the preview pane (gives it 60% width in accordion layout)
	if autoPreview {
		m.viewMode = viewDualPane
		m.focusedPane = rightPane
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
		m.panelsLocked = false // Reset lock when leaving dual-pane
	} else {
		// Check if using vertical split (Detail always uses vertical, List/Tree on narrow terminals)
		useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

		if useVerticalSplit {
			// Vertical split (stacked layout) - set full width for both panes
			// (actual rendering uses full width for top and bottom panes)
			m.leftWidth = m.width   // Full width for top pane (file list)
			m.rightWidth = m.width  // Full width for bottom pane (preview)
		} else if m.panelsLocked && m.leftWidth > 0 && m.rightWidth > 0 {
			// Panels locked: keep current widths, only adjust for terminal resize
			total := m.leftWidth + m.rightWidth + 1 // +1 for separator
			if total != m.width {
				// Terminal was resized - scale proportionally
				ratio := float64(m.leftWidth) / float64(m.leftWidth+m.rightWidth)
				m.leftWidth = int(float64(m.width-1) * ratio)
				if m.leftWidth < 30 {
					m.leftWidth = 30
				}
				m.rightWidth = m.width - m.leftWidth - 1
				if m.rightWidth < 30 {
					m.rightWidth = 30
					m.leftWidth = m.width - m.rightWidth - 1
				}
			}
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

// verticalSplitHeights returns (topHeight, bottomHeight) for vertical split mode.
// Respects panelsLocked: when locked, preserves the current ratio instead of
// applying accordion focus-based sizing.
func (m *model) verticalSplitHeights(maxVisible int) (int, int) {
	if m.panelsLocked && m.lockedTopRatio > 0 {
		// Locked: use saved ratio
		topHeight := int(float64(maxVisible) * m.lockedTopRatio)
		if topHeight < 3 {
			topHeight = 3
		}
		bottomHeight := maxVisible - topHeight
		if bottomHeight < 3 {
			bottomHeight = 3
			topHeight = maxVisible - bottomHeight
		}
		return topHeight, bottomHeight
	}

	// Accordion: focused pane gets 2/3
	var topHeight, bottomHeight int
	if m.focusedPane == leftPane {
		topHeight = (maxVisible * 2) / 3
		bottomHeight = maxVisible - topHeight
	} else {
		bottomHeight = (maxVisible * 2) / 3
		topHeight = maxVisible - bottomHeight
	}
	return topHeight, bottomHeight
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

	// Check for Termux (Android) FIRST - it sets TERM=xterm-256color
	// So we need to detect it before the xterm check
	// Multiple detection methods for better compatibility across Termux versions
	if os.Getenv("TERMUX_VERSION") != "" ||
		os.Getenv("TERMUX_APP_PID") != "" ||
		strings.Contains(os.Getenv("PREFIX"), "com.termux") {
		return terminalTermux
	}

	// Check for xterm BEFORE Windows Terminal
	// When xterm PTY runs inside Windows Terminal, TERM=xterm-256color but WT_SESSION is also set
	// Prioritize TERM to detect the actual rendering terminal, not the wrapper
	if term == "xterm" || term == "xterm-256color" {
		return terminalXterm
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

	// Check for Windows Terminal LAST (as a fallback)
	// This catches native Windows Terminal when no other terminal is detected
	if wtSession != "" {
		return terminalWindowsTerminal
	}

	return terminalUnknown
}
