package main

import (
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Module: update.go
// Purpose: Main update dispatcher and initialization
// Responsibilities:
// - Bubbletea initialization (Init)
// - Main message dispatcher (Update)
// - Window resize handling
// - Editor/command finished message handling
// - Spinner tick handling
// - Helper function: isSpecialKey() for detecting non-printable keys

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(), // Start landing page animation
	)
}

// tickMsg for landing page animation
type tickMsg struct{}

// tickCmd creates a command that sends tick messages for animation
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// stripANSI removes ANSI escape codes from a string
// This prevents pasted styled text from corrupting the command line
func stripANSI(s string) string {
	// Match all common ANSI escape sequences:
	// - CSI sequences: ESC [ ... (letter or @)
	// - OSC sequences: ESC ] ... (BEL or ST)
	// - Other escape sequences: ESC followed by various characters
	ansiRegex := regexp.MustCompile(`\x1b(\[[0-9;?]*[a-zA-Z@]|\][^\x07\x1b]*(\x07|\x1b\\)|[>=<>()#])`)
	cleaned := ansiRegex.ReplaceAllString(s, "")

	// Also strip terminal response sequences that may appear without ESC prefix
	// These can leak in when terminal responds to queries (e.g., color capability checks)
	// Patterns: ";rgb:xxxx/xxxx", "1;rgb:xxxx/xxxx/xxxx", "0;rgb:...", numeric response codes
	// Match anywhere in the string, not just exact matches
	responseRegex := regexp.MustCompile(`;?rgb:[0-9a-fA-F/]+|\d+;rgb:[0-9a-fA-F/]+|\d+;\d+(?:;\d+)*`)
	cleaned = responseRegex.ReplaceAllString(cleaned, "")

	return cleaned
}

// isSpecialKey checks if a key string represents a special (non-printable) key
// that should not be added to command input
func isSpecialKey(key string) bool {
	specialKeys := []string{
		"up", "down", "left", "right",
		"home", "end", "pageup", "pagedown", "pgup", "pgdn",
		"delete", "insert",
		"backspace", "enter", "return", "tab", "esc", "escape",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
		"ctrl+c", "ctrl+h", "ctrl+d", "ctrl+z",
		"alt+", "ctrl+", // Prefixes for modifier combinations
		"shift+",
	}

	// Check exact matches
	for _, special := range specialKeys {
		if key == special {
			return true
		}
		// Check if key starts with modifier prefix (like "ctrl+a", "alt+x")
		if len(special) > 0 && special[len(special)-1] == '+' && len(key) > len(special) {
			if key[:len(special)] == special {
				return true
			}
		}
	}

	return false
}

// Update is the main message dispatcher
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update tree items cache if in tree mode (before processing events)
	if m.displayMode == modeTree {
		m.updateTreeItems()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dispatch to keyboard event handler
		return m.handleKeyEvent(msg)

	case tea.MouseMsg:
		// Dispatch to mouse event handler
		return m.handleMouseEvent(msg)

	case tea.WindowSizeMsg:
		// Handle window resize
		m.height = msg.Height
		m.width = msg.Width

		m.calculateLayout()      // Recalculate pane layout on resize
		m.populatePreviewCache() // Repopulate cache with new width

		// Reset horizontal scroll on window resize
		m.detailScrollX = 0

	case tickMsg:
		// Background refresh for git repos (every 60 seconds)
		if m.showGitReposOnly && !m.gitReposLastScan.IsZero() {
			elapsed := time.Since(m.gitReposLastScan)
			if elapsed >= 60*time.Second {
				// Re-scan from the same root directory
				m.gitReposList = m.scanGitReposRecursive(m.gitReposScanRoot, m.gitReposScanDepth, 50)
				m.gitReposLastScan = time.Now()
			}
		}

		return m, tickCmd() // Continue animation

	case spinner.TickMsg:
		// Update spinner animation
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statusTimeoutMsg:
		// Status message timeout - force full screen redraw
		// Clear screen to ensure proper redraw of footer
		return m, tea.ClearScreen

	case editorFinishedMsg:
		// Editor has closed, we're back in TFE
		// Refresh file list in case file was modified
		m.loadFiles()
		// Force a refresh and re-enable mouse support (external editors disable it)
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case commandFinishedMsg:
		// Command has finished, we're back in TFE
		// Refresh file list in case command modified files
		m.loadFiles()
		// Force a refresh and restore terminal state (alt screen + mouse support)
		return m, tea.Batch(
			tea.EnterAltScreen,       // Re-enter alternate screen (required!)
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case fuzzySearchResultMsg:
		// Fuzzy search completed
		m.fuzzySearchActive = false
		if msg.err == nil && msg.selected != "" {
			m.navigateToFuzzyResult(msg.selected)
		}
		// Force a refresh and re-enable mouse support
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case markdownRenderedMsg:
		// Markdown rendering completed in background
		// Just return to trigger a re-render with the cached content
		return m, nil

	case browserOpenedMsg:
		// Browser opened (or failed to open)
		if msg.success {
			m.setStatusMessage("✓ Opened in browser", false)
		} else {
			m.setStatusMessage("Failed to open in browser: "+msg.err.Error(), true)
		}
		return m, statusTimeoutCmd()

	case fileExplorerOpenedMsg:
		// File explorer opened (or failed to open)
		if msg.success {
			m.setStatusMessage("✓ Opened in file explorer", false)
		} else {
			m.setStatusMessage("Failed to open file explorer: "+msg.err.Error(), true)
		}
		return m, statusTimeoutCmd()
	}

	return m, nil
}
