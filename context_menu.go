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

	// Special menu for trash view
	if m.showTrashOnly {
		items = append(items, contextMenuItem{"♻️  Restore", "restore"})
		items = append(items, contextMenuItem{"🗑️  Delete Permanently", "permanent_delete"})
		items = append(items, contextMenuItem{"─────────", "separator"})
		items = append(items, contextMenuItem{"🧹 Empty Trash", "empty_trash"})
		return items
	}

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
		if editorAvailable("bottom") {
			if !hasTools {
				items = append(items, contextMenuItem{"─────────", "separator"})
				hasTools = true
			}
			items = append(items, contextMenuItem{"📊 Monitor (bottom)", "bottom"})
		}

		// Add separator and favorites
		items = append(items, contextMenuItem{"─────────", "separator"})
		items = append(items, contextMenuItem{"📋 Copy to...", "copy"})
		items = append(items, contextMenuItem{"✏️  Rename...", "rename"})
		items = append(items, contextMenuItem{"🗑️  Delete", "delete"})
		if m.isFavorite(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"⭐ Unfavorite", "togglefav"})
		} else {
			items = append(items, contextMenuItem{"☆ Add Favorite", "togglefav"})
		}
	} else {
		// File menu items
		items = append(items, contextMenuItem{"👁  Preview", "preview"})

		// Add image-specific options
		if isImageFile(m.contextMenuFile.path) {
			items = append(items, contextMenuItem{"🖼️  View Image", "viewimage"})
			items = append(items, contextMenuItem{"🎨 Edit Image", "editimage"})
			items = append(items, contextMenuItem{"🌐 Open in Browser", "browser"})
		} else if isHTMLFile(m.contextMenuFile.path) {
			// Add "Open in Browser" for HTML files only
			items = append(items, contextMenuItem{"🌐 Open in Browser", "browser"})
		}

		items = append(items, contextMenuItem{"✏  Edit", "edit"})

		// Add "Run Script" for executable files
		if isExecutableFile(*m.contextMenuFile) {
			items = append(items, contextMenuItem{"▶️  Run Script", "runscript"})
		}

		items = append(items, contextMenuItem{"📋 Copy Path", "copypath"})
		items = append(items, contextMenuItem{"📋 Copy to...", "copy"})
		items = append(items, contextMenuItem{"✏️  Rename...", "rename"})
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

	case "viewimage":
		// View image in TUI viewer (viu, timg, chafa)
		if !m.contextMenuFile.isDir && isImageFile(m.contextMenuFile.path) {
			return m, openImageViewer(m.contextMenuFile.path)
		}
		return m, tea.ClearScreen

	case "editimage":
		// Edit image in TUI editor (textual-paint)
		if !m.contextMenuFile.isDir && isImageFile(m.contextMenuFile.path) {
			return m, openImageEditor(m.contextMenuFile.path)
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
			// Use bash to run the script safely (no command injection)
			scriptPath := m.contextMenuFile.path
			return m, runScript(scriptPath)
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

	case "bottom":
		// Launch bottom system monitor
		if m.contextMenuFile.isDir {
			return m, openTUITool("bottom", m.contextMenuFile.path)
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

	case "restore":
		// Restore item from trash
		if err := restoreFromTrash(m.contextMenuFile.path); err != nil {
			m.setStatusMessage(fmt.Sprintf("Failed to restore: %s", err), true)
		} else {
			m.setStatusMessage("Item restored successfully", false)
			m.loadFiles() // Refresh trash view
		}
		return m, tea.ClearScreen

	case "permanent_delete":
		// Permanently delete item from trash
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Permanently Delete",
			message:    fmt.Sprintf("Permanently delete '%s'?\nThis CANNOT be undone!", m.contextMenuFile.name),
		}
		m.showDialog = true
		return m, tea.ClearScreen

	case "empty_trash":
		// Empty entire trash
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Empty Trash",
			message:    "Permanently delete ALL items in trash?\nThis CANNOT be undone!",
		}
		m.showDialog = true
		return m, tea.ClearScreen

	case "copy":
		// Copy file or folder to destination
		m.dialog = dialogModel{
			dialogType: dialogInput,
			title:      "Copy File",
			message:    fmt.Sprintf("Copy '%s' to:", m.contextMenuFile.name),
			input:      "", // User types destination path
		}
		m.showDialog = true
		return m, tea.ClearScreen

	case "rename":
		// Rename the selected file or folder
		m.dialog = dialogModel{
			dialogType: dialogInput,
			title:      "Rename",
			message:    "New name:",
			input:      m.contextMenuFile.name, // Pre-fill current name
		}
		m.showDialog = true
		return m, tea.ClearScreen

	case "delete":
		// Delete the selected file or folder (move to trash)
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Move to Trash",
			message:    fmt.Sprintf("Move '%s' to trash?", m.contextMenuFile.name),
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

	// Calculate menu dimensions - find the longest item
	maxWidth := 0
	for _, item := range items {
		// Count runes, not bytes (better emoji support)
		width := len([]rune(item.label))
		if width > maxWidth {
			maxWidth = width
		}
	}

	// Fixed menu width (wider for better appearance)
	menuWidth := maxWidth + 4 // 2 spaces padding on each side

	// Build menu content with consistent width
	var menuLines []string
	for i, item := range items {
		// Pad all labels to the same width with spaces (ensures even borders)
		labelWidth := len([]rune(item.label))
		padding := maxWidth - labelWidth
		paddedLabel := item.label + strings.Repeat(" ", padding)

		var line string

		// Style separators differently
		if item.action == "separator" {
			// Separator: dim color, not selectable
			separatorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("240")).
				Width(menuWidth) // Force exact width
			line = separatorStyle.Render(fmt.Sprintf("  %s  ", paddedLabel))
		} else if i == m.contextMenuCursor {
			// Highlighted selected item
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Width(menuWidth) // Force exact width
			line = selectedStyle.Render(fmt.Sprintf("  %s  ", paddedLabel))
		} else {
			// Normal items also need a background to cover underlying text
			normalStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Width(menuWidth) // Force exact width
			line = normalStyle.Render(fmt.Sprintf("  %s  ", paddedLabel))
		}
		menuLines = append(menuLines, line)
	}

	// Create menu box with consistent-width content
	menuContent := strings.Join(menuLines, "\n")

	// Apply border (content width is already fixed)
	menuStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		BorderBackground(lipgloss.Color("236")).
		Background(lipgloss.Color("236"))

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
