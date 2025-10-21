package main

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// menu.go - Dropdown Menu System
// Purpose: Menu bar rendering and interaction logic
// When to extend: Add new menus or menu items here

// getMenus returns all available menus with current state
func (m model) getMenus() map[string]Menu {
	return map[string]Menu{
		"navigate": {
			Label: "Navigate",
			Items: []MenuItem{
				{Label: "🏠 Home", Action: "home", Shortcut: "🏠 button"},
				{Label: "⭐ Favorites", Action: "toggle-favorites", Shortcut: "F6", IsCheckable: true, IsChecked: m.showFavoritesOnly},
				{Label: "🗑️ Trash", Action: "toggle-trash", Shortcut: "F12", IsCheckable: true, IsChecked: m.showTrashOnly},
			},
		},
		"view": {
			Label: "View",
			Items: []MenuItem{
				{Label: "Display Mode: List", Action: "display-list", Shortcut: "1 or F9", IsCheckable: true, IsChecked: m.displayMode == modeList},
				{Label: "Display Mode: Detail", Action: "display-detail", Shortcut: "2 or F9", IsCheckable: true, IsChecked: m.displayMode == modeDetail},
				{Label: "Display Mode: Tree", Action: "display-tree", Shortcut: "3 or F9", IsCheckable: true, IsChecked: m.displayMode == modeTree},
				{IsSeparator: true},
				{Label: "⬌ Dual Pane", Action: "toggle-dual-pane", Shortcut: "Tab/Space", IsCheckable: true, IsChecked: m.viewMode == viewDualPane},
				{Label: "📝 Prompts Filter", Action: "toggle-prompts", Shortcut: "F11", IsCheckable: true, IsChecked: m.showPromptsOnly},
				{IsSeparator: true},
				{Label: "Show Hidden Files", Action: "toggle-hidden", Shortcut: "H or .", IsCheckable: true, IsChecked: m.showHidden},
			},
		},
		"tools": {
			Label: "Tools",
			Items: []MenuItem{
				{Label: ">_ Command Mode", Action: "toggle-command", Shortcut: ":", IsCheckable: true, IsChecked: m.commandFocused},
				{Label: "🔍 Search", Action: "toggle-search", Shortcut: "/ or Ctrl+F"},
				{Label: "🎯 Fuzzy Search", Action: "fuzzy-search", Shortcut: "Ctrl+P"},
				{IsSeparator: true},
				{Label: "🎮 Games Launcher", Action: "launch-games", Shortcut: "🎮 button"},
			},
		},
		"help": {
			Label: "Help",
			Items: []MenuItem{
				{Label: "⌨️ Keyboard Shortcuts", Action: "show-hotkeys", Shortcut: "F1"},
				{Label: "ℹ️ About TFE", Action: "show-about"},
				{IsSeparator: true},
				{Label: "🔗 GitHub Repository", Action: "open-github"},
			},
		},
	}
}

// getMenuOrder returns the order of menus in the menu bar
func getMenuOrder() []string {
	return []string{"navigate", "view", "tools", "help"}
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

	menuInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	for _, menuKey := range menuOrder {
		menu := menus[menuKey]

		// Style based on active state
		var style lipgloss.Style
		if m.activeMenu == menuKey && m.menuOpen {
			style = menuActiveStyle
		} else {
			style = menuInactiveStyle
		}

		// Render menu label
		renderedMenu := style.Render(menu.Label)
		renderedMenus = append(renderedMenus, renderedMenu)
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
		Foreground(lipgloss.Color("252"))

	menuItemSelectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39")).
		Bold(true)

	menuItemDisabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Build dropdown panel
	var lines []string
	maxWidth := 0

	// First pass: calculate max width
	for _, item := range menu.Items {
		if item.IsSeparator {
			continue
		}
		width := len(item.Label)
		if item.IsCheckable {
			width += 2 // "✓ " or "  "
		}
		if item.Shortcut != "" {
			width += len(item.Shortcut) + 3 // spacing
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
			labelWidth -= len(shortcut) + 1
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

	xPos := 0
	for _, key := range menuOrder {
		if key == menuKey {
			return xPos
		}
		menu := menus[key]
		// Use actual rendered width
		var style lipgloss.Style
		if m.activeMenu == key && m.menuOpen {
			style = menuActiveStyle
		} else {
			style = menuInactiveStyle
		}
		renderedMenu := style.Render(menu.Label)
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

	xPos := 0
	for _, menuKey := range menuOrder {
		menu := menus[menuKey]

		// Calculate actual rendered width
		var style lipgloss.Style
		if m.activeMenu == menuKey && m.menuOpen {
			style = menuActiveStyle
		} else {
			style = menuInactiveStyle
		}
		renderedMenu := style.Render(menu.Label)
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
	// Navigate menu
	case "home":
		homeDir, err := os.UserHomeDir()
		if err == nil {
			m.currentPath = homeDir
			m.cursor = 0
			m.loadFiles()
		}

	case "toggle-favorites":
		m.showFavoritesOnly = !m.showFavoritesOnly
		m.cursor = 0
		if m.showFavoritesOnly {
			m.loadFiles()
		}

	case "toggle-trash":
		m.showTrashOnly = !m.showTrashOnly
		m.cursor = 0
		m.loadFiles()

	// View menu
	case "display-list":
		m.displayMode = modeList

	case "display-detail":
		m.displayMode = modeDetail

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

	case "toggle-prompts":
		m.showPromptsOnly = !m.showPromptsOnly
		m.cursor = 0
		m.loadFiles()

	case "toggle-hidden":
		m.showHidden = !m.showHidden
		m.loadFiles()

	// Tools menu
	case "toggle-command":
		m.commandFocused = !m.commandFocused
		if !m.commandFocused {
			m.commandInput = ""
		}

	case "toggle-search":
		// Context-aware search
		if m.viewMode == viewFullPreview || (m.viewMode == viewDualPane && m.focusedPane == rightPane) {
			// Toggle in-file search
			m.preview.searchActive = !m.preview.searchActive
			if !m.preview.searchActive {
				m.preview.searchQuery = ""
				m.preview.searchMatches = nil
				m.preview.currentMatch = -1
			}
		} else {
			// Toggle directory filter search
			m.searchMode = !m.searchMode
			if !m.searchMode {
				m.searchQuery = ""
				m.filteredIndices = nil
				m.cursor = 0
			}
		}

	case "fuzzy-search":
		m.menuOpen = false
		m.activeMenu = ""
		m.selectedMenuItem = -1
		return m, m.launchFuzzySearch()

	case "launch-games":
		m.setStatusMessage("Games launcher - Click 🎮 button or use Ctrl+G", false)

	// Help menu
	case "show-hotkeys":
		m.setStatusMessage("Press F1 to view keyboard shortcuts (feature to be implemented)", false)

	case "show-about":
		m.setStatusMessage("TFE v"+Version+" - Terminal File Explorer | github.com/GGPrompts/TFE", false)

	case "open-github":
		m.setStatusMessage("GitHub: https://github.com/GGPrompts/TFE", false)

	default:
		m.setStatusMessage("Action: "+action+" (not implemented)", false)
	}

	// Close menu after action
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
