package main

// Module: context_menu.go
// Purpose: Right-click context menu functionality
// Responsibilities:
// - Menu item definitions
// - Menu action execution
// - Menu rendering

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// contextMenuItem represents an item in the context menu
type contextMenuItem struct {
	label  string
	action string
}

// getContextMenuItems returns the list of menu items for the current file
func (m model) getContextMenuItems() []contextMenuItem {
	if m.contextMenuFile == nil {
		return []contextMenuItem{}
	}

	items := []contextMenuItem{}

	if m.contextMenuFile.isDir {
		// Directory menu items
		items = append(items, contextMenuItem{"📂 Open", "open"})
		items = append(items, contextMenuItem{"📂 Quick CD", "quickcd"})
		items = append(items, contextMenuItem{"📋 Copy Path", "copypath"})
		if m.isFavorite(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"⭐ Unfavorite", "togglefav"})
		} else {
			items = append(items, contextMenuItem{"☆ Add Favorite", "togglefav"})
		}
	} else {
		// File menu items
		items = append(items, contextMenuItem{"👁  Preview", "preview"})
		items = append(items, contextMenuItem{"✏  Edit", "edit"})
		items = append(items, contextMenuItem{"📋 Copy Path", "copypath"})
		if m.isFavorite(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"⭐ Unfavorite", "togglefav"})
		} else {
			items = append(items, contextMenuItem{"☆ Add Favorite", "togglefav"})
		}
	}

	return items
}

// executeContextMenuAction executes the action for the currently selected menu item
func (m model) executeContextMenuAction() (tea.Model, tea.Cmd) {
	if m.contextMenuFile == nil {
		m.contextMenuOpen = false
		return m, nil
	}

	items := m.getContextMenuItems()
	if m.contextMenuCursor >= len(items) {
		m.contextMenuOpen = false
		return m, nil
	}

	action := items[m.contextMenuCursor].action

	// Close menu
	m.contextMenuOpen = false

	// Execute action
	switch action {
	case "open":
		// Navigate into directory
		if m.contextMenuFile.isDir {
			m.currentPath = m.contextMenuFile.path
			m.cursor = 0
			m.loadFiles()
		}
		return m, tea.ClearScreen

	case "quickcd":
		// Change to directory
		if m.contextMenuFile.isDir {
			m.currentPath = m.contextMenuFile.path
			m.cursor = 0
			m.loadFiles()
		}
		return m, tea.ClearScreen

	case "preview":
		// Preview file
		if !m.contextMenuFile.isDir {
			m.loadPreview(m.contextMenuFile.path)
			m.viewMode = viewFullPreview
			// Disable mouse to allow text selection
			return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
		}
		return m, tea.ClearScreen

	case "edit":
		// Edit file
		if !m.contextMenuFile.isDir {
			editor := getAvailableEditor()
			if editor == "" {
				return m, tea.ClearScreen
			}
			if editorAvailable("micro") {
				editor = "micro"
			}
			return m, openEditor(editor, m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "copypath":
		// Copy path to clipboard
		_ = copyToClipboard(m.contextMenuFile.path)
		return m, tea.ClearScreen

	case "togglefav":
		// Toggle favorite
		m.toggleFavorite(m.contextMenuFile.path)
		return m, tea.ClearScreen
	}

	return m, tea.ClearScreen
}

// renderContextMenu renders the context menu at the stored position
func (m model) renderContextMenu() string {
	if !m.contextMenuOpen || m.contextMenuFile == nil {
		return ""
	}

	items := m.getContextMenuItems()
	if len(items) == 0 {
		return ""
	}

	// Calculate menu dimensions - make it wider to ensure full coverage
	maxWidth := 0
	for _, item := range items {
		// Use visual width (accounting for emoji)
		width := visualWidth(item.label)
		if width > maxWidth {
			maxWidth = width
		}
	}
	// Content width (without borders/padding)
	contentWidth := maxWidth + 4 // Add internal spacing

	// Build menu content with consistent width
	var menuLines []string
	for i, item := range items {
		// Add padding to each line
		line := fmt.Sprintf("  %s  ", item.label)

		// Highlight selected item
		if i == m.contextMenuCursor {
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Width(contentWidth)
			line = selectedStyle.Render(line)
		} else {
			// Normal items also need a background to cover underlying text
			normalStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Width(contentWidth)
			line = normalStyle.Render(line)
		}
		menuLines = append(menuLines, line)
	}

	// Create menu box - don't set width here, let it fit the content
	menuContent := strings.Join(menuLines, "\n")

	// Apply border and background
	menuStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		BorderBackground(lipgloss.Color("236")).  // Background for border area
		Background(lipgloss.Color("236"))         // Background for content area

	return menuStyle.Render(menuContent)
}
