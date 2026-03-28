package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// menu.go - Dropdown Menu System
// Purpose: Menu bar rendering and interaction logic
// When to extend: Add new menus or menu items here

// getMenus returns all available menus with current state
// Uses cached tool availability to avoid repeated filesystem lookups (performance optimization)
func (m model) getMenus() map[string]Menu {
	// Build menus with current state
	// Performance: Uses m.toolsAvailable (cached at startup) instead of editorAvailable()
	// This eliminates 5 filesystem lookups per render (was causing lag)
	// Build File menu dynamically (add New Image if textual-paint is available)
	fileMenuItems := []MenuItem{
		{Label: "📁 New Folder...", Action: "new-folder", Shortcut: "F7"},
		{Label: "📄 New File...", Action: "new-file"},
	}

	// Add "New Image" if textual-paint is available
	if m.toolsAvailable["textual-paint"] {
		fileMenuItems = append(fileMenuItems, MenuItem{Label: "🎨 New Image", Action: "new-image"})
	}

	// Add "New Prompt" for creating prompt templates
	fileMenuItems = append(fileMenuItems, MenuItem{Label: "📝 New Prompt...", Action: "new-prompt"})

	fileMenuItems = append(fileMenuItems,
		MenuItem{Label: "📂 Open", Action: "open", Shortcut: "Enter"},
		MenuItem{IsSeparator: true},
		MenuItem{Label: "📋 Copy Path", Action: "copy-path", Shortcut: "F5"},
		MenuItem{IsSeparator: true},
		MenuItem{Label: "🚪 Exit", Action: "quit", Shortcut: "F10"},
	)

	menus := map[string]Menu{
		"file": {
			Label: "File",
			Items: fileMenuItems,
		},
		"profiles": {
			Label: "Profiles",
			Items: m.buildProfileMenuItems(),
		},
		"view": {
			Label: "View",
			Items: []MenuItem{
				{Label: "📄 List", Action: "display-list", Shortcut: "1", IsCheckable: true, IsChecked: m.displayMode == modeList},
				{Label: "📋 Details", Action: "display-detail", Shortcut: "2", IsCheckable: true, IsChecked: m.displayMode == modeDetail},
				{Label: "🌳 Tree", Action: "display-tree", Shortcut: "3", IsCheckable: true, IsChecked: m.displayMode == modeTree},
				{Label: "  └─ Collapse All", Action: "collapse-all-tree", Shortcut: "Ctrl+W"},
				{IsSeparator: true},
				{Label: "⬌ Preview Pane", Action: "toggle-dual-pane", Shortcut: "Tab/Space", IsCheckable: true, IsChecked: m.viewMode == viewDualPane},
				{Label: "🔒 Lock Panel Widths", Action: "toggle-panel-lock", Shortcut: "Ctrl+L", IsCheckable: true, IsChecked: m.panelsLocked},
				{Label: "👁  Show Hidden Files", Action: "toggle-hidden", Shortcut: "H or .", IsCheckable: true, IsChecked: m.showHidden},
				{IsSeparator: true},
				{Label: "📝 Prompts Library", Action: "toggle-prompts", Shortcut: "F11", IsCheckable: true, IsChecked: m.showPromptsOnly},
				{Label: "⭐ Favorites", Action: "toggle-favorites", Shortcut: "F6", IsCheckable: true, IsChecked: m.showFavoritesOnly},
				{Label: "🔀 Git Repositories", Action: "toggle-git-repos", IsCheckable: true, IsChecked: m.showGitReposOnly},
				{Label: "⚡ Git Changes", Action: "toggle-changes", Shortcut: "Ctrl+G", IsCheckable: true, IsChecked: m.showChangesOnly},
				{Label: "🗑  Trash", Action: "toggle-trash", Shortcut: "F12", IsCheckable: true, IsChecked: m.showTrashOnly},
			},
		},
		"go": {
			Label: "Go",
			Items: []MenuItem{
				{Label: "🏠 Home (~)", Action: "go-home", Shortcut: "~"},
				{Label: "⭐ Favorites", Action: "go-favorites", Shortcut: "F6"},
				{Label: "📝 Prompts", Action: "go-prompts", Shortcut: "F11"},
				{Label: "🔀 Git Repos", Action: "go-git-repos"},
				{Label: "🗑  Trash", Action: "go-trash", Shortcut: "F12"},
				{IsSeparator: true},
				{Label: "📂 Quick CD", Action: "go-quickcd", Shortcut: "Ctrl+D"},
				{Label: "🎯 Fuzzy Search", Action: "go-fuzzy", Shortcut: "Ctrl+P"},
			},
		},
		"git": {
			Label: "Git",
			Items: []MenuItem{
				{Label: "⚡ Changes Mode", Action: "git-changes-mode", Shortcut: "Ctrl+G", IsCheckable: true, IsChecked: m.showChangesOnly},
				{Label: "📋 Toggle Diff", Action: "git-toggle-diff", Shortcut: "d", IsCheckable: true, IsChecked: m.showDiffPreview},
				{IsSeparator: true},
				{Label: "⬇  Pull", Action: "git-pull"},
				{Label: "⬆  Push", Action: "git-push"},
				{Label: "🔄 Sync", Action: "git-sync"},
				{Label: "📡 Fetch", Action: "git-fetch"},
				{IsSeparator: true},
				{Label: "📋 Yank Diff", Action: "git-yank-diff", Shortcut: "y"},
				{Label: "📋 Yank All Diffs", Action: "git-yank-all-diffs", Shortcut: "Y"},
			},
		},
		"tools": {
			Label: "Tools",
			Items: []MenuItem{
				{Label: ">_ Command Prompt", Action: "toggle-command", Shortcut: ":", IsCheckable: true, IsChecked: m.commandFocused},
				{Label: "🔍 Search in Folder", Action: "toggle-search", Shortcut: "/"},
				{Label: "🎯 Fuzzy Search", Action: "fuzzy-search", Shortcut: "Ctrl+P"},
				{IsSeparator: true},
				{Label: "🔄 Pull & Rebuild TFE", Action: "pull-rebuild", Shortcut: ""},
			},
		},
		"settings": {
			Label: "Settings",
			Items: []MenuItem{
				{Label: "🌙 Dark Mode", Action: "settings-dark-mode", IsCheckable: true, IsChecked: !m.forceLightTheme},
				{Label: "🔒 Panel Lock", Action: "settings-panel-lock", Shortcut: "Ctrl+L", IsCheckable: true, IsChecked: m.panelsLocked},
				{Label: "👁  File Watcher", Action: "settings-file-watcher", IsCheckable: true, IsChecked: m.watcherActive},
				{Label: "📁 Show Hidden Files", Action: "settings-show-hidden", IsCheckable: true, IsChecked: m.showHidden},
				{IsSeparator: true},
				{Label: "⚙  Open Settings Panel...", Action: "settings-open-panel", Shortcut: "Ctrl+,"},
			},
		},
		"help": {
			Label: "Help",
			Items: []MenuItem{
				{Label: "⌨  Keyboard Shortcuts", Action: "show-hotkeys", Shortcut: "F1"},
				{Label: "ℹ  About TFE", Action: "show-about"},
				{IsSeparator: true},
				{Label: "🔗 GitHub Repository", Action: "open-github"},
			},
		},
	}

	// Add TUI tools to Tools menu if available (using cached availability)
	toolsMenu := menus["tools"]
	hasTools := false

	// Use cached tool availability instead of filesystem lookups (performance optimization)
	if m.toolsAvailable["lazygit"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "🌿 Git (lazygit)", Action: "lazygit"})
	}
	if m.toolsAvailable["lazydocker"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "🐋 Docker (lazydocker)", Action: "lazydocker"})
	}
	if m.toolsAvailable["lnav"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "📜 Logs (lnav)", Action: "lnav"})
	}
	if m.toolsAvailable["htop"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "📊 Processes (htop)", Action: "htop"})
	}
	if m.toolsAvailable["bottom"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "📊 Monitor (bottom)", Action: "bottom"})
	}
	if m.toolsAvailable["pyradio"] {
		if !hasTools {
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
			hasTools = true
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "📻 Radio (pyradio)", Action: "pyradio"})
	}

	// Add Games Launcher (only if TUIClassics is installed)
	if m.tuiClassicsPath != "" {
		if hasTools {
			// Only add separator if we added TUI tools above
			toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
		}
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "🎮 Games Launcher", Action: "launch-games"})
	}

	menus["tools"] = toolsMenu

	// The performance win comes from using m.toolsAvailable instead of editorAvailable()
	// which eliminates filesystem lookups per render (was causing dropdown lag)
	return menus
}

// buildProfileMenuItems builds menu items from configured profiles
func (m model) buildProfileMenuItems() []MenuItem {
	profiles := m.config.Profiles
	if len(profiles) == 0 {
		// Fallback to defaults if config has no profiles
		profiles = []Profile{
			{Name: "Shell Here", Command: "bash"},
			{Name: "Claude Here", Command: "claude"},
		}
	}

	var items []MenuItem
	for i, p := range profiles {
		label := p.Name
		hint := ""
		if p.Dir != "" {
			hint = p.Dir
		}
		items = append(items, MenuItem{
			Label:    label,
			Action:   fmt.Sprintf("profile-%d", i),
			Shortcut: hint,
		})
	}

	// Separator + Edit Profiles...
	items = append(items, MenuItem{IsSeparator: true})
	items = append(items, MenuItem{Label: "Edit Profiles...", Action: "edit-profiles"})

	return items
}

// writePostCommand writes the command to execute after TFE exits
func writePostCommand(command string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	targetFile := filepath.Join(homeDir, ".tfe_post_command")
	return os.WriteFile(targetFile, []byte(command), 0600)
}

// getMenuOrder returns the order of menus in the menu bar
func getMenuOrder() []string {
	return []string{"file", "profiles", "view", "go", "git", "tools", "settings", "help"}
}

// getPreviousMenu returns the menu key to the left of the current menu (with wrapping)
func getPreviousMenu(current string) string {
	order := getMenuOrder()
	for i, key := range order {
		if key == current {
			if i == 0 {
				return order[len(order)-1] // Wrap to last menu
			}
			return order[i-1]
		}
	}
	return order[0] // Fallback to first menu
}

// getNextMenu returns the menu key to the right of the current menu (with wrapping)
func getNextMenu(current string) string {
	order := getMenuOrder()
	for i, key := range order {
		if key == current {
			if i == len(order)-1 {
				return order[0] // Wrap to first menu
			}
			return order[i+1]
		}
	}
	return order[0] // Fallback to first menu
}

// getFirstSelectableMenuItem returns the index of the first non-separator item in the menu
// Returns 0 if no valid items found (fallback)
func (m model) getFirstSelectableMenuItem(menuKey string) int {
	menus := m.getMenus()
	menu, exists := menus[menuKey]
	if !exists {
		return 0
	}

	// Find first non-separator item
	for i, item := range menu.Items {
		if !item.IsSeparator {
			return i
		}
	}

	// Fallback to 0 if all items are separators (shouldn't happen)
	return 0
}

// renderMenuBar renders the menu bar (replaces GitHub link after 5s)
func (m model) renderMenuBar() string {
	menus := m.getMenus()
	menuOrder := getMenuOrder()

	var renderedMenus []string

	// Menu bar styles - use theme colors for light/dark theme support
	menuActiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.SelectionFg.adaptiveColor()).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	menuHighlightedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
		Background(lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#404040"}).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.Title.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	// Styles for first menu (no left padding for left alignment)
	menuActiveStyleFirst := menuActiveStyle.Copy().Padding(0, 1, 0, 0)
	menuHighlightedStyleFirst := menuHighlightedStyle.Copy().Padding(0, 1, 0, 0)
	menuInactiveStyleFirst := menuInactiveStyle.Copy().Padding(0, 1, 0, 0)

	for i, menuKey := range menuOrder {
		menu := menus[menuKey]
		isFirst := i == 0

		// Style based on state: active (open) > highlighted (focused) > inactive
		var style lipgloss.Style
		if m.activeMenu == menuKey && m.menuOpen {
			if isFirst {
				style = menuActiveStyleFirst
			} else {
				style = menuActiveStyle
			}
		} else if m.highlightedMenu == menuKey && m.menuBarFocused {
			if isFirst {
				style = menuHighlightedStyleFirst
			} else {
				style = menuHighlightedStyle
			}
		} else {
			if isFirst {
				style = menuInactiveStyleFirst
			} else {
				style = menuInactiveStyle
			}
		}

		// Render menu label with underlined first letter (hotkey indicator)
		// F for File, A for AI, V for View, G for Go/Git, T for Tools, H for Help
		label := menu.Label
		if len(label) > 0 {
			// Create underline style based on current style (without padding)
			baseStyle := style.Copy().Padding(0, 0)
			underlineStyle := baseStyle.Copy().Underline(true)
			firstLetter := underlineStyle.Render(string(label[0]))
			restOfLabel := baseStyle.Render(label[1:])
			// Apply padding to the combined result
			renderedMenu := style.Render(firstLetter + restOfLabel)
			renderedMenus = append(renderedMenus, renderedMenu)
		} else {
			renderedMenu := style.Render(menu.Label)
			renderedMenus = append(renderedMenus, renderedMenu)
		}
	}

	// Join with single space
	menuBarContent := strings.Join(renderedMenus, " ")
	padding := m.width - lipgloss.Width(menuBarContent)
	if padding < 0 {
		padding = 0
	}

	return menuBarContent + strings.Repeat(" ", padding)
}

// renderActiveDropdown renders the currently active dropdown menu
func (m model) renderActiveDropdown() string {
	if !m.menuOpen || m.activeMenu == "" {
		return ""
	}

	menus := m.getMenus()
	menu, exists := menus[m.activeMenu]
	if !exists {
		return ""
	}

	// Menu item styles - use theme colors for light/dark theme support
	menuItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}).
		Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#303030"})

	menuItemSelectedStyle := lipgloss.NewStyle().
		Foreground(currentTheme.SelectionFg.adaptiveColor()).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Bold(true)

	menuItemDisabledStyle := lipgloss.NewStyle().
		Foreground(currentTheme.BorderUnfocused.adaptiveColor()).
		Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#303030"})

	// Build dropdown panel
	var lines []string
	maxWidth := 0

	// First pass: calculate max width using terminal-aware width
	for _, item := range menu.Items {
		if item.IsSeparator {
			continue
		}
		width := m.visualWidthCompensated(item.Label) // Use terminal-aware width for emoji
		if item.IsCheckable {
			width += m.visualWidthCompensated("✓ ") // Use terminal-aware width of checkmark + space
		}
		if item.Shortcut != "" {
			width += m.visualWidthCompensated(item.Shortcut) + 3 // Use terminal-aware width for shortcut
		}
		if width > maxWidth {
			maxWidth = width
		}
	}

	// Add padding
	maxWidth += 4 // 2 chars padding on each side
	if maxWidth < 20 {
		maxWidth = 20
	}

	// Second pass: render items
	for i, item := range menu.Items {
		if item.IsSeparator {
			lines = append(lines, strings.Repeat("─", maxWidth-2))
			continue
		}

		// Determine style
		var itemStyle lipgloss.Style
		if item.Disabled {
			itemStyle = menuItemDisabledStyle
		} else if i == m.selectedMenuItem {
			itemStyle = menuItemSelectedStyle
		} else {
			itemStyle = menuItemStyle
		}

		// Build item line
		label := item.Label
		if item.IsCheckable {
			if item.IsChecked {
				label = "✓ " + label
			} else {
				label = "  " + label
			}
		}
		shortcut := item.Shortcut

		// Pad label
		labelWidth := maxWidth - 4
		if shortcut != "" {
			labelWidth -= m.visualWidthCompensated(shortcut) + 1 // Use terminal-aware width
		}

		line := " " + m.padRight(label, labelWidth)
		if shortcut != "" {
			line += " " + shortcut
		}
		line += " "

		lines = append(lines, itemStyle.Render(line))
	}

	// Create dropdown panel
	dropdown := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentTheme.BorderUnfocused.adaptiveColor()).
		Width(maxWidth).
		Render(strings.Join(lines, "\n"))

	return dropdown
}

// getMenuXPosition calculates the X position for a menu
func (m model) getMenuXPosition(menuKey string) int {
	menus := m.getMenus()
	menuOrder := getMenuOrder()

	menuActiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.SelectionFg.adaptiveColor()).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.Title.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	// Styles for first menu (no left padding)
	menuActiveStyleFirst := menuActiveStyle.Copy().Padding(0, 1, 0, 0)
	menuInactiveStyleFirst := menuInactiveStyle.Copy().Padding(0, 1, 0, 0)

	xPos := 0
	for i, key := range menuOrder {
		if key == menuKey {
			return xPos
		}
		menu := menus[key]
		isFirst := i == 0
		// Use actual rendered width (matching renderMenuBar logic with underlined first letter)
		var style lipgloss.Style
		if m.activeMenu == key && m.menuOpen {
			if isFirst {
				style = menuActiveStyleFirst
			} else {
				style = menuActiveStyle
			}
		} else {
			if isFirst {
				style = menuInactiveStyleFirst
			} else {
				style = menuInactiveStyle
			}
		}
		// Calculate width with underlined first letter (same as renderMenuBar)
		label := menu.Label
		var renderedMenu string
		if len(label) > 0 {
			baseStyle := style.Copy().Padding(0, 0)
			underlineStyle := baseStyle.Copy().Underline(true)
			firstLetter := underlineStyle.Render(string(label[0]))
			restOfLabel := baseStyle.Render(label[1:])
			renderedMenu = style.Render(firstLetter + restOfLabel)
		} else {
			renderedMenu = style.Render(menu.Label)
		}
		xPos += lipgloss.Width(renderedMenu) + 1 // +1 for space separator
	}
	return xPos
}

// isInMenuBar checks if position is in the menu bar (line 0)
func (m model) isInMenuBar(x, y int) bool {
	// Menu bar is on line 0 (first line)
	return y == 0
}

// getMenuAtPosition returns which menu is at the given X position
func (m model) getMenuAtPosition(x int) string {
	menus := m.getMenus()
	menuOrder := getMenuOrder()

	menuActiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.SelectionFg.adaptiveColor()).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(currentTheme.Title.adaptiveColor()).
		Bold(true).
		Padding(0, 1)

	// Styles for first menu (no left padding)
	menuActiveStyleFirst := menuActiveStyle.Copy().Padding(0, 1, 0, 0)
	menuInactiveStyleFirst := menuInactiveStyle.Copy().Padding(0, 1, 0, 0)

	xPos := 0
	for i, menuKey := range menuOrder {
		menu := menus[menuKey]
		isFirst := i == 0

		// Calculate actual rendered width (matching renderMenuBar logic with underlined first letter)
		var style lipgloss.Style
		if m.activeMenu == menuKey && m.menuOpen {
			if isFirst {
				style = menuActiveStyleFirst
			} else {
				style = menuActiveStyle
			}
		} else {
			if isFirst {
				style = menuInactiveStyleFirst
			} else {
				style = menuInactiveStyle
			}
		}
		// Calculate width with underlined first letter (same as renderMenuBar)
		label := menu.Label
		var renderedMenu string
		if len(label) > 0 {
			baseStyle := style.Copy().Padding(0, 0)
			underlineStyle := baseStyle.Copy().Underline(true)
			firstLetter := underlineStyle.Render(string(label[0]))
			restOfLabel := baseStyle.Render(label[1:])
			renderedMenu = style.Render(firstLetter + restOfLabel)
		} else {
			renderedMenu = style.Render(menu.Label)
		}
		menuWidth := lipgloss.Width(renderedMenu)

		if x >= xPos && x < xPos+menuWidth {
			return menuKey
		}

		xPos += menuWidth + 1 // +1 for space separator
	}

	return ""
}

// isInDropdown checks if position is within the active dropdown
func (m model) isInDropdown(x, y int) bool {
	if !m.menuOpen || m.activeMenu == "" {
		return false
	}

	// Dropdown starts at y=1 (after menu bar)
	if y < 1 {
		return false
	}

	menus := m.getMenus()
	menu, exists := menus[m.activeMenu]
	if !exists {
		return false
	}

	// Calculate dropdown bounds
	menuX := m.getMenuXPosition(m.activeMenu)

	// Count items for height
	height := 0
	for range menu.Items {
		height++
	}
	height += 2 // borders

	// Estimate width (will be at least 20)
	maxWidth := 20
	for _, item := range menu.Items {
		if item.IsSeparator {
			continue
		}
		width := len(item.Label)
		if item.IsCheckable {
			width += 2
		}
		if item.Shortcut != "" {
			width += len(item.Shortcut) + 3
		}
		width += 4 // padding
		if width > maxWidth {
			maxWidth = width
		}
	}

	// Clamp X so dropdown stays within terminal width (matches overlay logic)
	totalWidth := maxWidth + 2 // +2 for border
	if menuX+totalWidth > m.width {
		menuX = m.width - totalWidth
		if menuX < 0 {
			menuX = 0
		}
	}

	return x >= menuX && x < menuX+totalWidth && y >= 1 && y < 1+height
}

// getMenuItemAtPosition returns the menu item index at the given Y position in dropdown
func (m model) getMenuItemAtPosition(y int) int {
	if !m.menuOpen || m.activeMenu == "" {
		return -1
	}

	// Dropdown content starts at y=2 (after border)
	itemY := y - 2
	if itemY < 0 {
		return -1
	}

	menus := m.getMenus()
	menu, exists := menus[m.activeMenu]
	if !exists {
		return -1
	}

	if itemY >= len(menu.Items) {
		return -1
	}

	return itemY
}

// executeMenuAction executes a menu item action
func (m model) executeMenuAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	// File menu
	case "new-folder":
		// Create new folder in current directory
		m.dialog = dialogModel{
			dialogType: dialogInput,
			title:      "Create Directory",
			message:    "Enter directory name:",
			input:      "",
		}
		m.showDialog = true

	case "new-file":
		// Create new file in current directory
		m.dialog = dialogModel{
			dialogType: dialogInput,
			title:      "Create File",
			message:    "Enter filename:",
			input:      "",
		}
		m.showDialog = true

	case "new-image":
		// Launch textual-paint with blank canvas (it will handle save dialog)
		// Close menu before launching
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openImageEditorNew(m.currentPath)

	case "new-prompt":
		// Create new prompt file with template
		promptTemplate := `---
name: My Prompt Template
description: Brief description of what this prompt does
inputs:
  variable1:
    type: string
    description: Description of first variable
  variable2:
    type: string
    description: Description of second variable
---

# System Prompt

You are a helpful assistant.

# User Request

{{variable1}}

Additional context: {{variable2}}

# Instructions

1. First instruction
2. Second instruction
3. Third instruction
`
		// Generate filename with timestamp to avoid conflicts
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("new-prompt-%s.prompty", timestamp)
		filepath := filepath.Join(m.currentPath, filename)

		// Write template to file
		if err := os.WriteFile(filepath, []byte(promptTemplate), 0644); err != nil {
			m.setStatusMessage(fmt.Sprintf("Error creating prompt: %s", err), true)
		} else {
			// File created successfully - open in editor
			m.setStatusMessage(fmt.Sprintf("Created %s", filename), false)
			m.loadFiles() // Refresh file list

			// Open in editor
			editor := getAvailableEditor()
			if editor == "" {
				m.setStatusMessage("Prompt created but no editor available (tried micro, nano, vim, vi)", true)
			} else {
				// Use cached micro check (performance optimization)
				if m.toolsAvailable["micro"] {
					editor = "micro"
				}
				// Close menu before launching editor
				m.menuOpen = false
				m.activeMenu = ""
				m.selectedMenuItem = -1
				return m, openEditor(editor, filepath)
			}
		}

	case "open":
		// Open selected file/folder (same as Enter key)
		file := m.getCurrentFile()
		if file != nil {
			if file.isDir {
				m.currentPath = file.path
				m.cursor = 0
				m.loadFiles()
			} else {
				// Preview file
				m.loadPreview(file.path)
				m.viewMode = viewFullPreview
				m.calculateLayout()
				m.populatePreviewCache()
			}
		}

	case "copy-path":
		// Copy current directory path to clipboard
		if err := copyToClipboard(m.currentPath); err != nil {
			m.setStatusMessage(fmt.Sprintf("Failed to copy: %s", err), true)
		} else {
			m.setStatusMessage("Path copied to clipboard", false)
		}

	case "quit":
		return m, tea.Quit

	// View menu
	case "toggle-favorites":
		// Auto-exit trash mode when toggling favorites
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showFavoritesOnly = !m.showFavoritesOnly
		m.cursor = 0
		m.loadFiles()

	case "delete":
		// Delete selected file/folder
		file := m.getCurrentFile()
		if file != nil && file.name != ".." {
			m.dialog = dialogModel{
				dialogType: dialogConfirm,
				title:      "Move to Trash",
				message:    fmt.Sprintf("Move '%s' to trash?", file.name),
			}
			m.showDialog = true
		}

	// View menu
	case "display-list":
		m.displayMode = modeList
		m.expandedDirs = make(map[string]bool) // Reset tree expansion when leaving tree view
		m.calculateLayout() // Recalculate widths for new display mode

	case "display-detail":
		m.displayMode = modeDetail
		m.detailScrollX = 0 // Reset scroll when switching to detail view
		m.expandedDirs = make(map[string]bool) // Reset tree expansion when leaving tree view
		m.calculateLayout() // Recalculate widths for detail view columns

	case "display-tree":
		m.displayMode = modeTree
		m.calculateLayout() // Recalculate widths for new display mode

	case "collapse-all-tree":
		// Collapse all expanded folders in tree view
		if m.displayMode == modeTree {
			m.expandedDirs = make(map[string]bool)
			m.setStatusMessage("All folders collapsed", false)
		} else {
			m.setStatusMessage("Collapse all only works in tree view (press 3)", false)
		}

	case "toggle-dual-pane":
		if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
		} else {
			m.viewMode = viewDualPane
		}
		m.calculateLayout()
		m.populatePreviewCache()

	case "toggle-panel-lock":
		if m.viewMode == viewDualPane {
			m.panelsLocked = !m.panelsLocked
			if !m.panelsLocked {
				m.calculateLayout()
				m.populatePreviewCache()
			}
			m.persistConfig()
		} else {
			m.setStatusMessage("Panel lock only works in dual-pane mode", false)
		}

	case "toggle-hidden":
		m.showHidden = !m.showHidden
		m.loadFiles()
		m.persistConfig()

	case "pull-rebuild":
		// Pull latest TFE code, rebuild, and exit (so user can restart with new version)

		// Find TFE repository using smart discovery
		tfeRepoPath := findTFERepository()

		if tfeRepoPath == "" {
			m.setStatusMessage("❌ TFE git repository not found. Checked common locations and current binary path.", true)
			return m, nil
		}

		// Show confirmation dialog
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Pull & Rebuild TFE",
			message:    fmt.Sprintf("This will:\n• Run 'git pull' in %s\n• Rebuild and install TFE\n• Exit TFE (you'll need to restart)\n\nContinue?", tfeRepoPath),
		}
		m.showDialog = true

	// Tools menu
	case "toggle-command":
		m.commandFocused = !m.commandFocused
		if !m.commandFocused {
			m.commandInput = ""
		}

	case "toggle-search":
		// Toggle directory filter search
		m.searchMode = !m.searchMode
		if !m.searchMode {
			m.searchQuery = ""
			m.filteredIndices = nil
			m.cursor = 0
		}

	case "fuzzy-search":
		// Close menu and set fuzzy search active
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		m.fuzzySearchActive = true
		// Clear screen before launching fuzzy search to ensure clean terminal state
		return m, tea.Sequence(
			tea.ClearScreen,
			m.launchFuzzySearch(),
		)

	case "lazygit":
		// Launch lazygit in current directory
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("lazygit", m.currentPath)

	case "lazydocker":
		// Launch lazydocker in current directory
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("lazydocker", m.currentPath)

	case "lnav":
		// Launch lnav in current directory
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("lnav", m.currentPath)

	case "htop":
		// Launch htop
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("htop", m.currentPath)

	case "bottom":
		// Launch bottom system monitor
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("bottom", m.currentPath)

	case "pyradio":
		// Launch pyradio internet radio player
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, openTUITool("pyradio", m.currentPath)

	// Profiles menu
	case "edit-profiles":
		// Open config file in editor
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		cfgPath, err := configPath()
		if err != nil {
			m.setStatusMessage(fmt.Sprintf("Failed to find config path: %s", err), true)
			return m, nil
		}
		editor := getAvailableEditor()
		if editor == "" {
			m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
			return m, nil
		}
		return m, openEditor(editor, cfgPath)

	case "toggle-prompts":
		// Auto-exit trash mode when toggling prompts filter
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showPromptsOnly = !m.showPromptsOnly
		m.cursor = 0
		m.loadFiles()

	case "launch-games":
		// Launch TUIClassics game launcher
		if m.tuiClassicsPath != "" {
			// Close menu and launch
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			return m, openTUITool(m.tuiClassicsPath, filepath.Dir(m.tuiClassicsPath))
		}

		// Not found - show helpful message
		m.setStatusMessage("TUIClassics not found. Install: git clone https://github.com/GGPrompts/TUIClassics ~/TUIClassics && cd ~/TUIClassics && make build", true)

	case "toggle-git-repos":
		// Auto-exit trash mode when toggling git repos filter
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}

		m.showGitReposOnly = !m.showGitReposOnly

		// If turning ON, scan for repos recursively from current directory
		if m.showGitReposOnly {
			// Auto-switch to Detail view when enabling git repos filter
			m.displayMode = modeDetail
			m.detailScrollX = 0 // Reset scroll
			m.calculateLayout() // Recalculate widths for detail view

			m.setStatusMessage("🔍 Scanning for git repositories (depth 3, max 50)...", false)
			m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
			m.gitReposLastScan = time.Now()
			m.gitReposScanRoot = m.currentPath
			m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
		}

		m.cursor = 0
		m.loadFiles()

	case "toggle-changes":
		// Auto-exit trash mode when toggling git changes filter
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}

		m.showChangesOnly = !m.showChangesOnly

		if m.showChangesOnly {
			changed, err := m.getChangedFiles()
			if err != nil {
				m.setStatusMessage(err.Error(), true)
				m.showChangesOnly = false
			} else {
				m.changedFiles = changed
				// Load agent sessions and build file-to-agent map
				m.agentSessions = getAgentSessions()
				m.agentFileMap = buildAgentFileMap(changed, m.agentSessions)
				m.displayMode = modeDetail
				m.detailScrollX = 0
				m.showDiffPreview = true
				m.calculateLayout()
				m.setStatusMessage(fmt.Sprintf("Git changes: %d files (d: toggle diff)", len(changed)), false)
			}
		} else {
			m.showDiffPreview = false
			m.agentSessions = nil
			m.agentFileMap = nil
		}

		m.cursor = 0
		m.loadFiles()

	case "toggle-trash":
		// Navigate to trash view (or exit if already in trash)
		if m.showTrashOnly {
			// Already in trash - exit and restore previous path
			m.showTrashOnly = false
			if m.trashRestorePath != "" {
				m.currentPath = m.trashRestorePath
				m.trashRestorePath = ""
			}
			m.cursor = 0
			m.loadFiles()
		} else {
			// Enter trash view - save current path
			m.trashRestorePath = m.currentPath
			m.showTrashOnly = true
			m.showFavoritesOnly = false
			m.showPromptsOnly = false
			m.showChangesOnly = false
			m.showDiffPreview = false
			m.cursor = 0
			m.loadFiles()
		}

	// Go menu
	case "go-home":
		// Navigate to home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			if m.showTrashOnly {
				m.showTrashOnly = false
				m.trashRestorePath = ""
			}
			m.currentPath = homeDir
			m.cursor = 0
			m.showFavoritesOnly = false
			m.showPromptsOnly = false
			m.showGitReposOnly = false
			m.showChangesOnly = false
			m.showDiffPreview = false
			m.loadFiles()
		} else {
			m.setStatusMessage("Error: Could not find home directory", true)
		}

	case "go-favorites":
		// Toggle favorites view (same as F6)
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showFavoritesOnly = !m.showFavoritesOnly
		m.cursor = 0
		m.loadFiles()

	case "go-prompts":
		// Toggle prompts view (same as F11)
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showPromptsOnly = !m.showPromptsOnly
		m.cursor = 0
		m.loadFiles()

		// Auto-expand ~/.prompts when filter is turned on
		if m.showPromptsOnly {
			if homeDir, err := os.UserHomeDir(); err == nil {
				globalPromptsDir := filepath.Join(homeDir, ".prompts")
				if info, err := os.Stat(globalPromptsDir); err == nil && info.IsDir() {
					m.expandedDirs[globalPromptsDir] = true
				} else {
					m.setStatusMessage("💡 Tip: Create ~/.prompts/ folder for global prompts (see helper below)", false)
				}
			}
		}

	case "go-git-repos":
		// Toggle git repos view (same as toggle-git-repos)
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}

		m.showGitReposOnly = !m.showGitReposOnly

		if m.showGitReposOnly {
			m.displayMode = modeDetail
			m.detailScrollX = 0
			m.calculateLayout()

			m.setStatusMessage("🔍 Scanning for git repositories (depth 3, max 50)...", false)
			m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
			m.gitReposLastScan = time.Now()
			m.gitReposScanRoot = m.currentPath
			m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
		}

		m.cursor = 0
		m.loadFiles()

	case "go-trash":
		// Toggle trash view (same as F12)
		if m.showTrashOnly {
			m.showTrashOnly = false
			if m.trashRestorePath != "" {
				m.currentPath = m.trashRestorePath
				m.trashRestorePath = ""
			}
			m.cursor = 0
			m.loadFiles()
		} else {
			m.trashRestorePath = m.currentPath
			m.showTrashOnly = true
			m.showFavoritesOnly = false
			m.showPromptsOnly = false
			m.showChangesOnly = false
			m.showDiffPreview = false
			m.cursor = 0
			m.loadFiles()
		}

	case "go-quickcd":
		// Quick CD: write current directory as CD target and quit
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		if isTermux() && !hasParentShell() {
			return m, termuxNewSession("exec bash -l", m.currentPath)
		}
		if err := writeCDTarget(m.currentPath); err != nil {
			m.setStatusMessage(fmt.Sprintf("Failed to save directory for quick CD: %s", err), true)
			return m, tea.ClearScreen
		}
		return m, tea.Quit

	case "go-fuzzy":
		// Fuzzy search (same as Ctrl+P)
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		m.fuzzySearchActive = true
		return m, tea.Sequence(
			tea.ClearScreen,
			m.launchFuzzySearch(),
		)

	// Git menu
	case "git-changes-mode":
		// Toggle git changes mode (same as Ctrl+G)
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}

		m.showChangesOnly = !m.showChangesOnly

		if m.showChangesOnly {
			changed, err := m.getChangedFiles()
			if err != nil {
				m.setStatusMessage(err.Error(), true)
				m.showChangesOnly = false
			} else {
				m.changedFiles = changed
				m.agentSessions = getAgentSessions()
				m.agentFileMap = buildAgentFileMap(changed, m.agentSessions)
				m.displayMode = modeDetail
				m.detailScrollX = 0
				m.showDiffPreview = true
				m.calculateLayout()
				m.setStatusMessage(fmt.Sprintf("Git changes: %d files (d: toggle diff)", len(changed)), false)
			}
		} else {
			m.showDiffPreview = false
			m.agentSessions = nil
			m.agentFileMap = nil
		}

		m.cursor = 0
		m.loadFiles()

	case "git-toggle-diff":
		// Toggle diff preview in changes mode
		if m.showChangesOnly {
			m.showDiffPreview = !m.showDiffPreview
			if m.showDiffPreview {
				m.setStatusMessage("Diff preview enabled", false)
			} else {
				m.setStatusMessage("File preview enabled", false)
			}
		} else {
			m.setStatusMessage("Toggle diff only works in Changes Mode (Ctrl+G)", false)
		}

	case "git-pull":
		// Git pull in current directory's git root
		gitRoot := m.resolveGitRoot()
		if gitRoot != "" {
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			return m, gitPull(gitRoot)
		}
		m.setStatusMessage("Not in a git repository", true)

	case "git-push":
		// Git push in current directory's git root
		gitRoot := m.resolveGitRoot()
		if gitRoot != "" {
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			return m, gitPush(gitRoot)
		}
		m.setStatusMessage("Not in a git repository", true)

	case "git-sync":
		// Git sync (pull + push) in current directory's git root
		gitRoot := m.resolveGitRoot()
		if gitRoot != "" {
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			return m, gitSync(gitRoot)
		}
		m.setStatusMessage("Not in a git repository", true)

	case "git-fetch":
		// Git fetch in current directory's git root
		gitRoot := m.resolveGitRoot()
		if gitRoot != "" {
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			return m, gitFetch(gitRoot)
		}
		m.setStatusMessage("Not in a git repository", true)

	case "git-yank-diff":
		// Yank current file's diff to clipboard (same as 'y' in changes mode)
		if m.showChangesOnly {
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				statusCode := extractGitStatusCode(currentFile.name)
				diff, err := m.getFileDiff(currentFile.path, statusCode)
				if err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to get diff: %s", err), true)
				} else {
					// Format as markdown with file path header and diff code fence
					gitRoot := m.resolveGitRoot()
					relPath := currentFile.path
					if gitRoot != "" {
						if rp, err := filepath.Rel(gitRoot, currentFile.path); err == nil {
							relPath = rp
						}
					}
					markdown := fmt.Sprintf("## %s\n\n```diff\n%s```\n", relPath, diff)
					if err := copyToClipboard(markdown); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy diff: %s", err), true)
					} else {
						m.setStatusMessage(fmt.Sprintf("Copied diff for %s to clipboard", filepath.Base(relPath)), false)
					}
				}
			}
		} else {
			m.setStatusMessage("Yank diff only works in Changes Mode (Ctrl+G)", false)
		}

	case "git-yank-all-diffs":
		// Yank all diffs to clipboard (same as 'Y' in changes mode)
		if m.showChangesOnly && len(m.changedFiles) > 0 {
			gitRoot := m.resolveGitRoot()
			var allDiffs strings.Builder
			copied := 0
			for _, f := range m.changedFiles {
				if f.isDir {
					continue
				}
				statusCode := extractGitStatusCode(f.name)
				diff, err := m.getFileDiff(f.path, statusCode)
				if err != nil || diff == "" {
					continue
				}
				relPath := f.path
				if gitRoot != "" {
					if rel, err := filepath.Rel(gitRoot, f.path); err == nil {
						relPath = rel
					}
				}
				if copied > 0 {
					allDiffs.WriteString("\n---\n\n")
				}
				allDiffs.WriteString(fmt.Sprintf("## %s\n\n```diff\n%s```\n", relPath, diff))
				copied++
			}
			if copied > 0 {
				if err := copyToClipboard(allDiffs.String()); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to copy diffs: %s", err), true)
				} else {
					m.setStatusMessage(fmt.Sprintf("Copied diffs for %d files to clipboard", copied), false)
				}
			} else {
				m.setStatusMessage("No diffs available to copy", false)
			}
		} else {
			m.setStatusMessage("Yank all diffs only works in Changes Mode (Ctrl+G)", false)
		}

	// Settings menu
	case "settings-dark-mode":
		m.setConfigBool("dark_mode", m.forceLightTheme) // toggle: if light, set dark=true; if dark, set dark=false
		m.persistConfig()
		// Invalidate glamour cache so markdown re-renders with correct theme
		m.glamourRenderer = nil
		m.glamourRendererWidth = 0
		m.preview.cacheValid = false

	case "settings-panel-lock":
		m.setConfigBool("panel_lock", !m.panelsLocked)
		m.persistConfig()
		if !m.panelsLocked {
			m.calculateLayout()
			m.populatePreviewCache()
		}

	case "settings-file-watcher":
		m.setConfigBool("file_watcher_enabled", !m.watcherActive)
		m.persistConfig()

	case "settings-show-hidden":
		m.setConfigBool("show_hidden", !m.showHidden)
		m.persistConfig()

	case "settings-open-panel":
		m.dialog = dialogModel{
			dialogType: dialogSettings,
			title:      "Settings",
		}
		m.showDialog = true
		m.settingsCategory = 0
		m.settingsCursor = 0

	// Help menu
	case "show-hotkeys":
		// F1 functionality: Show hotkeys reference with context-aware navigation
		// First check if it exists in current directory
		hotkeysPath := filepath.Join(m.currentPath, "HOTKEYS.md")
		if _, err := os.Stat(hotkeysPath); os.IsNotExist(err) {
			// Try executable directory
			if exePath, err := os.Executable(); err == nil {
				hotkeysPath = filepath.Join(filepath.Dir(exePath), "HOTKEYS.md")
			}
		}
		// Load and show the hotkeys file if it exists
		if _, err := os.Stat(hotkeysPath); err == nil {
			m.loadPreview(hotkeysPath)

			// Context-aware help: Jump to relevant section based on current mode
			sectionName := m.getHelpSectionName()
			if sectionLine := findSectionLine(m.preview.content, sectionName); sectionLine >= 0 {
				m.preview.scrollPos = sectionLine
			}

			m.viewMode = viewFullPreview
			m.searchMode = false // Disable search mode in preview
			m.calculateLayout() // Update widths for full-screen
			m.populatePreviewCache() // Repopulate cache with correct width
			// Close menu before showing help
			m.menuOpen = false
			m.activeMenu = ""
			m.selectedMenuItem = -1
			// Return early with ClearScreen command
			return m, tea.ClearScreen
		} else {
			m.setStatusMessage("HOTKEYS.md not found", true)
		}

	case "show-about":
		m.setStatusMessage("TFE v"+Version+" - Terminal File Explorer | github.com/GGPrompts/TFE", false)

	case "open-github":
		m.setStatusMessage("GitHub: https://github.com/GGPrompts/TFE", false)

	default:
		// Handle profile-N actions dynamically
		if strings.HasPrefix(action, "profile-") {
			idxStr := strings.TrimPrefix(action, "profile-")
			idx := 0
			for _, ch := range idxStr {
				idx = idx*10 + int(ch-'0')
			}
			profiles := m.config.Profiles
			if idx >= 0 && idx < len(profiles) {
				profile := profiles[idx]
				m.menuOpen = false
				m.activeMenu = ""
				m.selectedMenuItem = -1

				// Determine target directory
				targetDir := m.currentPath
				if profile.Dir != "" {
					targetDir = profile.Dir
				}

				// Write CD target so the wrapper changes to the right directory
				if err := writeCDTarget(targetDir); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to write CD target: %s", err), true)
					return m, nil
				}

				// Write the post-exit command for the wrapper to execute
				if err := writePostCommand(profile.Command); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to write post command: %s", err), true)
					return m, nil
				}

				return m, tea.Quit
			}
		}
		m.setStatusMessage("Action: "+action+" (not implemented)", false)
	}

	// Close menu after action (unless already closed for tools that launch)
	m.menuOpen = false
	m.activeMenu = ""
	m.selectedMenuItem = -1

	return m, nil
}

// padRight pads a string with spaces to reach the desired width using terminal-aware width
func (m model) padRight(s string, width int) string {
	currentWidth := m.visualWidthCompensated(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}
