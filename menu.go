package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		"edit": {
			Label: "Edit",
			Items: []MenuItem{
				{Label: "🗑️  Delete", Action: "delete", Shortcut: "F8"},
			},
		},
		"view": {
			Label: "View",
			Items: []MenuItem{
				{Label: "📄 List", Action: "display-list", Shortcut: "1", IsCheckable: true, IsChecked: m.displayMode == modeList},
				{Label: "📋 Details", Action: "display-detail", Shortcut: "2", IsCheckable: true, IsChecked: m.displayMode == modeDetail},
				{Label: "🌳 Tree", Action: "display-tree", Shortcut: "3", IsCheckable: true, IsChecked: m.displayMode == modeTree},
				{IsSeparator: true},
				{Label: "⬌ Preview Pane", Action: "toggle-dual-pane", Shortcut: "Tab/Space", IsCheckable: true, IsChecked: m.viewMode == viewDualPane},
				{Label: "👁️  Show Hidden Files", Action: "toggle-hidden", Shortcut: "H or .", IsCheckable: true, IsChecked: m.showHidden},
				{IsSeparator: true},
				{Label: "📝 Prompts Library", Action: "toggle-prompts", Shortcut: "F11", IsCheckable: true, IsChecked: m.showPromptsOnly},
				{Label: "⭐ Favorites", Action: "toggle-favorites", Shortcut: "F6", IsCheckable: true, IsChecked: m.showFavoritesOnly},
				{Label: "🗑️  Trash", Action: "toggle-trash", Shortcut: "F12", IsCheckable: true, IsChecked: m.showTrashOnly},
				{IsSeparator: true},
				{Label: "🔄 Refresh", Action: "refresh", Shortcut: "F5"},
			},
		},
		"tools": {
			Label: "Tools",
			Items: []MenuItem{
				{Label: ">_ Command Prompt", Action: "toggle-command", Shortcut: ":", IsCheckable: true, IsChecked: m.commandFocused},
				{Label: "🔍 Search in Folder", Action: "toggle-search", Shortcut: "/"},
				{Label: "🎯 Fuzzy Search", Action: "fuzzy-search", Shortcut: "Ctrl+P"},
			},
		},
		"help": {
			Label: "Help",
			Items: []MenuItem{
				{Label: "⌨️  Keyboard Shortcuts", Action: "show-hotkeys", Shortcut: "F1"},
				{Label: "ℹ️  About TFE", Action: "show-about"},
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

	// Add Games Launcher
	if hasTools {
		// Only add separator if we added TUI tools above
		toolsMenu.Items = append(toolsMenu.Items, MenuItem{IsSeparator: true})
	}
	toolsMenu.Items = append(toolsMenu.Items, MenuItem{Label: "🎮 Games Launcher", Action: "launch-games"})

	menus["tools"] = toolsMenu

	// The performance win comes from using m.toolsAvailable instead of editorAvailable()
	// which eliminates 5 filesystem lookups per render (was causing dropdown lag)
	return menus
}

// getMenuOrder returns the order of menus in the menu bar
func getMenuOrder() []string {
	return []string{"file", "edit", "view", "tools", "help"}
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

	// Menu bar styles
	menuActiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	menuHighlightedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("240")).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
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
		// F for File, E for Edit, V for View, T for Tools, H for Help
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

	// Menu item styles
	menuItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")) // Add background to prevent transparency issues

	menuItemSelectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39")).
		Bold(true)

	menuItemDisabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")) // Add background to prevent transparency issues

	// Build dropdown panel
	var lines []string
	maxWidth := 0

	// First pass: calculate max width
	for _, item := range menu.Items {
		if item.IsSeparator {
			continue
		}
		width := lipgloss.Width(item.Label) // Use lipgloss.Width for accurate emoji/unicode width
		if item.IsCheckable {
			width += lipgloss.Width("✓ ") // Use actual width of checkmark + space
		}
		if item.Shortcut != "" {
			width += lipgloss.Width(item.Shortcut) + 3 // Use lipgloss.Width for shortcut too
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
			labelWidth -= lipgloss.Width(shortcut) + 1 // Use lipgloss.Width for accurate width
		}

		line := " " + padRight(label, labelWidth)
		if shortcut != "" {
			line += " " + shortcut
		}
		line += " "

		lines = append(lines, itemStyle.Render(line))
	}

	// Create dropdown panel
	dropdown := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(maxWidth).
		Render(strings.Join(lines, "\n"))

	return dropdown
}

// getMenuXPosition calculates the X position for a menu
func (m model) getMenuXPosition(menuKey string) int {
	menus := m.getMenus()
	menuOrder := getMenuOrder()

	menuActiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
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
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
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

	return x >= menuX && x < menuX+maxWidth && y >= 1 && y < 1+height
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

	// Edit menu
	case "toggle-favorites":
		m.showFavoritesOnly = !m.showFavoritesOnly
		m.cursor = 0
		if m.showFavoritesOnly {
			m.loadFiles()
		}

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

	case "display-detail":
		m.displayMode = modeDetail
		m.detailScrollX = 0 // Reset scroll when switching to detail view

	case "display-tree":
		m.displayMode = modeTree

	case "toggle-dual-pane":
		if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
		} else {
			m.viewMode = viewDualPane
		}
		m.calculateLayout()
		m.populatePreviewCache()

	case "toggle-hidden":
		m.showHidden = !m.showHidden
		m.loadFiles()

	case "refresh":
		m.loadFiles()
		m.setStatusMessage("Refreshed", false)

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

	case "toggle-prompts":
		m.showPromptsOnly = !m.showPromptsOnly
		m.cursor = 0
		m.loadFiles()

	case "launch-games":
		// Launch TUIClassics game launcher
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.setStatusMessage("Error: Could not find home directory", true)
		} else {
			classicsPath := filepath.Join(homeDir, "TUIClassics", "bin", "classics")

			// Check if classics launcher exists
			if _, err := os.Stat(classicsPath); err == nil {
				// Close menu and launch
				m.menuOpen = false
				m.activeMenu = ""
				m.selectedMenuItem = -1
				return m, openTUITool(classicsPath, filepath.Dir(classicsPath))
			}

			// If classics doesn't exist, check for individual games
			binDir := filepath.Join(homeDir, "TUIClassics", "bin")
			if entries, err := os.ReadDir(binDir); err == nil && len(entries) > 0 {
				// Find first executable game
				for _, entry := range entries {
					if !entry.IsDir() {
						gamePath := filepath.Join(binDir, entry.Name())
						if info, err := os.Stat(gamePath); err == nil && info.Mode()&0111 != 0 {
							// Close menu and launch
							m.menuOpen = false
							m.activeMenu = ""
							m.selectedMenuItem = -1
							return m, openTUITool(gamePath, filepath.Dir(gamePath))
						}
					}
				}
			}

			m.setStatusMessage("TUIClassics not found. Install from: github.com/GGPrompts/TUIClassics", true)
		}

	case "toggle-trash":
		m.showTrashOnly = !m.showTrashOnly
		m.cursor = 0
		m.loadFiles()

	// Help menu
	case "show-hotkeys":
		// F1 functionality: Show hotkeys reference with context-aware navigation
		hotkeysPath := filepath.Join(filepath.Dir(m.currentPath), "HOTKEYS.md")
		// Try to find HOTKEYS.md in the TFE directory
		// First check if it exists in current directory
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
		m.setStatusMessage("Action: "+action+" (not implemented)", false)
	}

	// Close menu after action (unless already closed for tools that launch)
	m.menuOpen = false
	m.activeMenu = ""
	m.selectedMenuItem = -1

	return m, nil
}

// padRight pads a string with spaces to reach the desired width
func padRight(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}
