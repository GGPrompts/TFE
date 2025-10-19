package main

// Module: context_menu.go
// Purpose: Right-click context menu functionality
// Responsibilities:
// - Menu item definitions
// - Menu action execution
// - Menu rendering

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// contextMenuItem represents an item in the context menu
type contextMenuItem struct {
	label  string
	action string
}

// writeCDTarget writes the target directory to a file so the shell can cd after TFE exits
func writeCDTarget(path string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	targetFile := filepath.Join(homeDir, ".tfe_cd_target")
	return os.WriteFile(targetFile, []byte(path), 0644)
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
		items = append(items, contextMenuItem{"📁 New Folder...", "newfolder"})
		items = append(items, contextMenuItem{"📄 New File...", "newfile"})
		items = append(items, contextMenuItem{"📋 Copy Path", "copypath"})

		// Add separator and TUI tools if available
		hasTools := false
		if editorAvailable("lazygit") {
			if !hasTools {
				items = append(items, contextMenuItem{"─────────", "separator"})
				hasTools = true
			}
			items = append(items, contextMenuItem{"🌿 Git (lazygit)", "lazygit"})
		}
		if editorAvailable("lazydocker") {
			if !hasTools {
				items = append(items, contextMenuItem{"─────────", "separator"})
				hasTools = true
			}
			items = append(items, contextMenuItem{"🐋 Docker (lazydocker)", "lazydocker"})
		}
		if editorAvailable("lnav") {
			if !hasTools {
				items = append(items, contextMenuItem{"─────────", "separator"})
				hasTools = true
			}
			items = append(items, contextMenuItem{"📜 Logs (lnav)", "lnav"})
		}
		if editorAvailable("htop") {
			if !hasTools {
				items = append(items, contextMenuItem{"─────────", "separator"})
				hasTools = true
			}
			items = append(items, contextMenuItem{"📊 Processes (htop)", "htop"})
		}

		// Add separator and favorites
		items = append(items, contextMenuItem{"─────────", "separator"})
		items = append(items, contextMenuItem{"🗑️  Delete", "delete"})
		if m.isFavorite(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"⭐ Unfavorite", "togglefav"})
		} else {
			items = append(items, contextMenuItem{"☆ Add Favorite", "togglefav"})
		}
	} else {
		// File menu items
		items = append(items, contextMenuItem{"👁  Preview", "preview"})

		// Add "Open in Browser" for images and HTML files
		if isBrowserFile(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"🌐 Open in Browser", "browser"})
		}

		items = append(items, contextMenuItem{"✏  Edit", "edit"})

		// Add "Run Script" for executable files
		if isExecutableFile(*m.contextMenuFile) {
			items = append(items, contextMenuItem{"▶️  Run Script", "runscript"})
		}

		items = append(items, contextMenuItem{"📋 Copy Path", "copypath"})
		items = append(items, contextMenuItem{"🗑️  Delete", "delete"})
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
		// Quick CD: write directory to file and exit TFE so shell can cd
		if m.contextMenuFile.isDir {
			if err := writeCDTarget(m.contextMenuFile.path); err != nil {
				m.setStatusMessage(fmt.Sprintf("Failed to save directory for quick CD: %s", err), true)
				return m, tea.ClearScreen
			}
			return m, tea.Quit
		}
		return m, tea.ClearScreen

	case "preview":
		// Preview file
		if !m.contextMenuFile.isDir {
			m.loadPreview(m.contextMenuFile.path)
			m.viewMode = viewFullPreview
			m.calculateLayout() // Update widths for full-screen
			m.populatePreviewCache() // Repopulate cache with correct width
			// Disable mouse to allow text selection
			return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
		}
		return m, tea.ClearScreen

	case "browser":
		// Open in browser (images/HTML)
		if !m.contextMenuFile.isDir {
			return m, openInBrowser(m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "edit":
		// Edit file
		if !m.contextMenuFile.isDir {
			editor := getAvailableEditor()
			if editor == "" {
				m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
				return m, tea.ClearScreen
			}
			if editorAvailable("micro") {
				editor = "micro"
			}
			return m, openEditor(editor, m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "runscript":
		// Run executable script
		if !m.contextMenuFile.isDir && isExecutableFile(*m.contextMenuFile) {
			// Use bash to run the script
			scriptPath := m.contextMenuFile.path
			command := fmt.Sprintf("bash %s", scriptPath)
			return m, runCommand(command, filepath.Dir(scriptPath))
		}
		return m, tea.ClearScreen

	case "copypath":
		// Copy path to clipboard
		if err := copyToClipboard(m.contextMenuFile.path); err != nil {
			m.setStatusMessage(fmt.Sprintf("Failed to copy to clipboard: %s", err), true)
		} else {
			m.setStatusMessage("Path copied to clipboard", false)
		}
		return m, tea.ClearScreen

	case "togglefav":
		// Toggle favorite
		m.toggleFavorite(m.contextMenuFile.path)
		return m, tea.ClearScreen

	case "separator":
		// Separator is not selectable - shouldn't happen but handle gracefully
		return m, tea.ClearScreen

	case "lazygit":
		// Launch lazygit in the selected directory
		if m.contextMenuFile.isDir {
			return m, openTUITool("lazygit", m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "lazydocker":
		// Launch lazydocker in the selected directory
		if m.contextMenuFile.isDir {
			return m, openTUITool("lazydocker", m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "lnav":
		// Launch lnav in the selected directory
		if m.contextMenuFile.isDir {
			return m, openTUITool("lnav", m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "htop":
		// Launch htop (doesn't need directory context but launch from directory anyway)
		if m.contextMenuFile.isDir {
			return m, openTUITool("htop", m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "newfolder":
		// Create new folder in the selected directory
		if m.contextMenuFile.isDir {
			// Navigate to the directory first
			m.currentPath = m.contextMenuFile.path
			m.cursor = 0
			m.loadFiles()

			// Show input dialog for folder name
			m.dialog = dialogModel{
				dialogType: dialogInput,
				title:      "Create Directory",
				message:    "Enter directory name:",
				input:      "",
			}
			m.showDialog = true
		}
		return m, tea.ClearScreen

	case "newfile":
		// Create new file in the selected directory
		if m.contextMenuFile.isDir {
			// Navigate to the directory first
			m.currentPath = m.contextMenuFile.path
			m.cursor = 0
			m.loadFiles()

			// Show input dialog for file name
			m.dialog = dialogModel{
				dialogType: dialogInput,
				title:      "Create File",
				message:    "Enter filename:",
				input:      "",
			}
			m.showDialog = true
		}
		return m, tea.ClearScreen

	case "delete":
		// Delete the selected file or folder
		fileType := "file"
		if m.contextMenuFile.isDir {
			fileType = "directory"
		}
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Delete " + fileType,
			message:    fmt.Sprintf("Delete '%s'?\nThis cannot be undone.", m.contextMenuFile.name),
		}
		m.showDialog = true
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
		var line string

		// Style separators differently
		if item.action == "separator" {
			// Separator: dim color, not selectable
			separatorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("240")).
				Width(contentWidth)
			line = separatorStyle.Render(fmt.Sprintf("  %s  ", item.label))
		} else if i == m.contextMenuCursor {
			// Highlighted selected item
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Width(contentWidth)
			line = selectedStyle.Render(fmt.Sprintf("  %s  ", item.label))
		} else {
			// Normal items also need a background to cover underlying text
			normalStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Width(contentWidth)
			line = normalStyle.Render(fmt.Sprintf("  %s  ", item.label))
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

// isExecutableFile checks if a file is executable (has execute permission or is a shell script)
func isExecutableFile(file fileItem) bool {
	// Check file extension for common script types
	ext := strings.ToLower(filepath.Ext(file.path))
	if ext == ".sh" || ext == ".bash" || ext == ".zsh" || ext == ".fish" {
		return true
	}

	// Check if file has execute permission
	// The mode contains permission bits - check if any execute bit is set
	// 0111 = user, group, or other has execute permission
	if file.mode&0111 != 0 {
		return true
	}

	return false
}
