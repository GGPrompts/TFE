package main

import (
	"strings"

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
// - Helper functions for input processing

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
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

// cleanBracketedPaste removes bracketed paste escape sequences from input
// Bracketed paste sequences are: ESC[200~ (start) and ESC[201~ (end)
func cleanBracketedPaste(s string) string {
	// Remove common bracketed paste escape sequences (with and without ESC)
	s = strings.ReplaceAll(s, "\x1b[200~", "")
	s = strings.ReplaceAll(s, "\x1b[201~", "")
	s = strings.ReplaceAll(s, "\x1b[200", "")
	s = strings.ReplaceAll(s, "\x1b[201", "")
	s = strings.ReplaceAll(s, "[200~", "")
	s = strings.ReplaceAll(s, "[201~", "")
	s = strings.ReplaceAll(s, "[200", "")
	s = strings.ReplaceAll(s, "[201", "")
	// Remove any standalone bracketed paste markers that might slip through
	if s == "[" || s == "]" || s == "~" {
		return ""
	}
	return s
}

// isBracketedPasteMarker checks if the input is a bracketed paste sequence marker
func isBracketedPasteMarker(s string) bool {
	markers := []string{
		"\x1b[200~", "\x1b[201~",
		"[200~", "[201~",
		"\x1b[200", "\x1b[201",
		"[200", "[201",
	}
	for _, marker := range markers {
		if strings.Contains(s, marker) {
			return true
		}
	}
	// Check for partial markers that might come through
	if len(s) <= 5 && (strings.HasPrefix(s, "[20") || strings.HasPrefix(s, "\x1b[20")) {
		return true
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
		m.calculateGridLayout()  // Recalculate grid columns on resize
		m.calculateLayout()      // Recalculate pane layout on resize
		m.populatePreviewCache() // Repopulate cache with new width

	case spinner.TickMsg:
		// Update spinner animation
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

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
		// Force a refresh and re-enable mouse support (shell commands may disable it)
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)
	}

	return m, nil
}
