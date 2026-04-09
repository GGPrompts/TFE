package main

// Module: settings.go
// Purpose: Settings panel dialog for interactive configuration editing
// Responsibilities:
// - Rendering the settings overlay (categories, toggle/select/input widgets)
// - Handling keyboard navigation within the settings panel
// - Persisting changes immediately to config.toml via saveConfig()
// - Syncing config changes back to the model's runtime fields

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// settingsItem represents a single editable setting
type settingsItem struct {
	label   string       // Display label
	key     string       // Config key identifier
	kind    settingsKind // Widget type
	options []string     // For select widgets: list of valid values
}

// settingsKind represents the type of settings widget
type settingsKind int

const (
	settingsToggle settingsKind = iota // Boolean on/off
	settingsSelect                     // Pick from list of options
	settingsString                     // Free-text input
)

// settingsCategories defines the category tabs
var settingsCategoryNames = []string{"General", "Appearance", "File Watcher"}

// settingsByCategory returns the settings items for a given category index
func settingsByCategory(cat int) []settingsItem {
	switch cat {
	case 0: // General
		return []settingsItem{
			{label: "Show Hidden Files", key: "show_hidden", kind: settingsToggle},
			{label: "Panel Lock", key: "panel_lock", kind: settingsToggle},
			{label: "Sort Order", key: "sort_order", kind: settingsSelect, options: []string{"name", "size", "modified"}},
			{label: "Default View Mode", key: "default_view_mode", kind: settingsSelect, options: []string{"tree", "list", "detail"}},
			{label: "Editor", key: "editor", kind: settingsString},
		}
	case 1: // Appearance
		return []settingsItem{
			{label: "Dark Mode", key: "dark_mode", kind: settingsToggle},
		}
	case 2: // File Watcher
		return []settingsItem{
			{label: "File Watcher Enabled", key: "file_watcher_enabled", kind: settingsToggle},
			{label: "Auto Changes (Agent)", key: "auto_changes", kind: settingsToggle},
		}
	default:
		return nil
	}
}

// getConfigBool returns a boolean config value by key
func (m model) getConfigBool(key string) bool {
	switch key {
	case "dark_mode":
		return m.config.DarkMode
	case "file_watcher_enabled":
		return m.config.FileWatcherEnabled
	case "show_hidden":
		return m.config.ShowHidden
	case "panel_lock":
		return m.config.PanelLock
	case "auto_changes":
		return m.config.AutoChanges
	default:
		return false
	}
}

// setConfigBool sets a boolean config value by key and syncs to model
func (m *model) setConfigBool(key string, val bool) {
	switch key {
	case "dark_mode":
		m.config.DarkMode = val
		m.forceLightTheme = !val
		applyThemeMode(val)
	case "file_watcher_enabled":
		m.config.FileWatcherEnabled = val
		if val && !m.watcherActive {
			m.initWatcher()
		} else if !val && m.watcherActive {
			m.stopWatcher()
		}
	case "show_hidden":
		m.config.ShowHidden = val
		m.showHidden = val
		m.loadFiles()
	case "panel_lock":
		m.config.PanelLock = val
		m.panelsLocked = val
	case "auto_changes":
		m.config.AutoChanges = val
		m.agentAutoWatch = val
	}
}

// getConfigString returns a string config value by key
func (m model) getConfigString(key string) string {
	switch key {
	case "sort_order":
		return m.config.SortOrder
	case "default_view_mode":
		return m.config.DefaultViewMode
	case "editor":
		return m.config.Editor
	default:
		return ""
	}
}

// setConfigString sets a string config value by key and syncs to model
func (m *model) setConfigString(key, val string) {
	switch key {
	case "sort_order":
		m.config.SortOrder = val
		m.sortBy = val
		m.loadFiles()
	case "default_view_mode":
		m.config.DefaultViewMode = val
		// Note: don't change current view mode — this sets the default for next launch
	case "editor":
		m.config.Editor = val
	}
}

// renderSettingsPanel renders the settings overlay dialog
func (m model) renderSettingsPanel() string {
	// Calculate panel dimensions (large overlay)
	panelWidth := m.width - 10
	if panelWidth > 72 {
		panelWidth = 72
	}
	if panelWidth < 40 {
		panelWidth = 40
	}
	innerWidth := panelWidth - 6 // Account for border + padding

	// -- Styles --
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentTheme.BorderFocused.adaptiveColor()).
		Background(uiPanelBackground()).
		Padding(1, 2).
		Width(panelWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Title.adaptiveColor()).
		Align(lipgloss.Center).
		Width(innerWidth)

	tabActiveStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.SelectionFg.adaptiveColor()).
		Background(currentTheme.SelectionBg.adaptiveColor()).
		Padding(0, 1)

	tabInactiveStyle := lipgloss.NewStyle().
		Foreground(uiMutedText()).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(uiBodyText())

	selectedLabelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(currentTheme.Title.adaptiveColor())

	valueOnStyle := lipgloss.NewStyle().
		Foreground(currentTheme.DiffAdded.adaptiveColor()).
		Bold(true)

	valueOffStyle := lipgloss.NewStyle().
		Foreground(uiSubtleText())

	valueSelectStyle := lipgloss.NewStyle().
		Foreground(currentTheme.Title.adaptiveColor())

	editingStyle := lipgloss.NewStyle().
		Foreground(uiBodyText()).
		Background(uiInputBackground()).
		Padding(0, 1)

	hintStyle := lipgloss.NewStyle().
		Foreground(uiMutedText()).
		Align(lipgloss.Center).
		Width(innerWidth)

	separatorStyle := lipgloss.NewStyle().
		Foreground(uiSubtleText())

	// -- Build content --
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Settings"))
	content.WriteString("\n\n")

	// Category tabs
	var tabs []string
	for i, name := range settingsCategoryNames {
		if i == m.settingsCategory {
			tabs = append(tabs, tabActiveStyle.Render(name))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(name))
		}
	}
	tabLine := strings.Join(tabs, " ")
	content.WriteString(tabLine)
	content.WriteString("\n")
	content.WriteString(separatorStyle.Render(strings.Repeat("\u2500", innerWidth)))
	content.WriteString("\n\n")

	// Settings items for current category
	items := settingsByCategory(m.settingsCategory)
	for i, item := range items {
		cursor := "  "
		lStyle := labelStyle
		if i == m.settingsCursor {
			cursor = "> "
			lStyle = selectedLabelStyle
		}

		label := lStyle.Render(item.label)

		var valueStr string
		switch item.kind {
		case settingsToggle:
			val := m.getConfigBool(item.key)
			if val {
				valueStr = valueOnStyle.Render("[ON]")
			} else {
				valueStr = valueOffStyle.Render("[OFF]")
			}

		case settingsSelect:
			current := m.getConfigString(item.key)
			valueStr = valueSelectStyle.Render(current)

		case settingsString:
			if i == m.settingsCursor && m.settingsEditing {
				valueStr = editingStyle.Render(m.settingsInput + "\u2588")
			} else {
				val := m.getConfigString(item.key)
				if val == "" {
					val = "(default)"
				}
				valueStr = valueSelectStyle.Render(val)
			}
		}

		// Calculate spacing to right-align the value
		labelWidth := visualWidth(cursor) + visualWidth(item.label)
		valWidth := visualWidth(valueStr)
		padding := innerWidth - labelWidth - valWidth
		if padding < 1 {
			padding = 1
		}

		line := cursor + label + strings.Repeat(" ", padding) + valueStr
		content.WriteString(line)
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Hint line
	hints := "j/k: navigate | Tab: category | Esc: close"
	if len(items) > 0 {
		item := items[m.settingsCursor]
		switch item.kind {
		case settingsToggle:
			hints = "Space/Enter: toggle | Tab: category | Esc: close"
		case settingsSelect:
			hints = "Space/Enter: cycle | Tab: category | Esc: close"
		case settingsString:
			if m.settingsEditing {
				hints = "Enter: confirm | Esc: cancel editing"
			} else {
				hints = "Enter: edit | Tab: category | Esc: close"
			}
		}
	}
	content.WriteString(hintStyle.Render(hints))

	return borderStyle.Render(content.String())
}

// getSettingsPanelGeometry returns the dialog position and dimensions for mouse hit-testing.
// Returns (dialogX, dialogY, panelWidth, panelHeight).
func (m model) getSettingsPanelGeometry() (int, int, int, int) {
	panelWidth := m.width - 10
	if panelWidth > 72 {
		panelWidth = 72
	}
	if panelWidth < 40 {
		panelWidth = 40
	}

	// Use the same position as getDialogPosition to stay in sync
	dialogX, dialogY := m.getDialogPosition()

	// Height: match getDialogPosition's calculation
	items := settingsByCategory(m.settingsCategory)
	panelHeight := len(items) + 10

	return dialogX, dialogY, panelWidth, panelHeight
}

// handleSettingsMouseEvent processes mouse input when the settings panel is open.
// Returns (model, cmd, handled). If handled is true, the caller should not process further.
func (m model) handleSettingsMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	dialogX, dialogY, panelWidth, panelHeight := m.getSettingsPanelGeometry()

	// Check if click is inside the dialog bounds
	inDialog := msg.X >= dialogX && msg.X < dialogX+panelWidth &&
		msg.Y >= dialogY && msg.Y < dialogY+panelHeight

	// Relative position inside the dialog content area (after border + padding)
	relY := msg.Y - dialogY - 2 // -1 border, -1 padding
	relX := msg.X - dialogX - 3 // -1 border, -2 padding

	items := settingsByCategory(m.settingsCategory)

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// Scroll up through settings items
		if len(items) > 0 {
			m.settingsCursor--
			if m.settingsCursor < 0 {
				m.settingsCursor = len(items) - 1
			}
		}
		return m, nil, true

	case tea.MouseButtonWheelDown:
		// Scroll down through settings items
		if len(items) > 0 {
			m.settingsCursor++
			if m.settingsCursor >= len(items) {
				m.settingsCursor = 0
			}
		}
		return m, nil, true

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionRelease {
			return m, nil, true
		}

		// Click outside dialog = close
		if !inDialog {
			m.showDialog = false
			m.dialog = dialogModel{}
			m.settingsEditing = false
			return m, tea.ClearScreen, true
		}

		// Row 2 (relY == 2) is the category tab bar
		if relY == 2 && relX >= 0 {
			// Hit-test each tab: tabs are rendered with padding(0,1) so each is len(name)+2, separated by space
			tabX := 0
			for i, name := range settingsCategoryNames {
				tabWidth := len(name) + 2 // +2 for padding(0,1)
				if relX >= tabX && relX < tabX+tabWidth {
					m.settingsCategory = i
					m.settingsCursor = 0
					m.settingsEditing = false
					return m, nil, true
				}
				tabX += tabWidth + 1 // +1 for space separator
			}
			return m, nil, true
		}

		// Rows 5+ are settings items (relY == 4 is the blank line after separator)
		itemStartRow := 5
		itemIndex := relY - itemStartRow
		if itemIndex >= 0 && itemIndex < len(items) {
			m.settingsCursor = itemIndex

			// Click acts like Enter/Space — toggle, cycle, or enter edit mode
			item := items[itemIndex]
			switch item.kind {
			case settingsToggle:
				current := m.getConfigBool(item.key)
				m.setConfigBool(item.key, !current)
				if err := saveConfig(m.config); err != nil {
					m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
				}
			case settingsSelect:
				current := m.getConfigString(item.key)
				for idx, opt := range item.options {
					if opt == current {
						next := (idx + 1) % len(item.options)
						m.setConfigString(item.key, item.options[next])
						if err := saveConfig(m.config); err != nil {
							m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
						}
						break
					}
				}
			case settingsString:
				m.settingsEditing = true
				m.settingsInput = m.getConfigString(item.key)
			}
			return m, nil, true
		}

		// Click somewhere else inside the dialog — consume but do nothing
		return m, nil, true

	case tea.MouseButtonRight:
		// Consume right clicks so they don't open context menu behind the panel
		return m, nil, true
	}

	// Consume all other mouse events while settings is open
	return m, nil, true
}

// handleSettingsKeyEvent processes keyboard input when the settings panel is open
func (m model) handleSettingsKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := settingsByCategory(m.settingsCategory)

	// If editing a string field, handle text input
	if m.settingsEditing {
		switch msg.String() {
		case "esc":
			// Cancel editing, revert to original value
			m.settingsEditing = false
			m.settingsInput = ""
			return m, nil
		case "enter":
			// Confirm edit
			if m.settingsCursor < len(items) {
				item := items[m.settingsCursor]
				m.setConfigString(item.key, m.settingsInput)
				if err := saveConfig(m.config); err != nil {
					m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
				}
			}
			m.settingsEditing = false
			m.settingsInput = ""
			return m, nil
		case "backspace":
			if len(m.settingsInput) > 0 {
				m.settingsInput = m.settingsInput[:len(m.settingsInput)-1]
			}
			return m, nil
		default:
			// Append typed characters
			if len(msg.Runes) > 0 {
				m.settingsInput += string(msg.Runes)
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		// Close settings panel
		m.showDialog = false
		m.dialog = dialogModel{}
		m.settingsEditing = false
		return m, tea.ClearScreen

	case "tab", "shift+tab":
		// Switch category
		numCats := len(settingsCategoryNames)
		if msg.String() == "tab" {
			m.settingsCategory = (m.settingsCategory + 1) % numCats
		} else {
			m.settingsCategory = (m.settingsCategory - 1 + numCats) % numCats
		}
		m.settingsCursor = 0
		return m, nil

	case "j", "down":
		if len(items) > 0 {
			m.settingsCursor = (m.settingsCursor + 1) % len(items)
		}
		return m, nil

	case "k", "up":
		if len(items) > 0 {
			m.settingsCursor = (m.settingsCursor - 1 + len(items)) % len(items)
		}
		return m, nil

	case "enter", " ":
		if m.settingsCursor >= len(items) {
			return m, nil
		}
		item := items[m.settingsCursor]

		switch item.kind {
		case settingsToggle:
			current := m.getConfigBool(item.key)
			m.setConfigBool(item.key, !current)
			if err := saveConfig(m.config); err != nil {
				m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
			}

		case settingsSelect:
			current := m.getConfigString(item.key)
			// Cycle to next option
			for idx, opt := range item.options {
				if opt == current {
					next := (idx + 1) % len(item.options)
					m.setConfigString(item.key, item.options[next])
					if err := saveConfig(m.config); err != nil {
						m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
					}
					break
				}
			}

		case settingsString:
			// Enter edit mode
			m.settingsEditing = true
			m.settingsInput = m.getConfigString(item.key)
		}
		return m, nil

	case "l", "right":
		// For select widgets, cycle forward
		if m.settingsCursor < len(items) {
			item := items[m.settingsCursor]
			if item.kind == settingsSelect {
				current := m.getConfigString(item.key)
				for idx, opt := range item.options {
					if opt == current {
						next := (idx + 1) % len(item.options)
						m.setConfigString(item.key, item.options[next])
						if err := saveConfig(m.config); err != nil {
							m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
						}
						break
					}
				}
			}
		}
		return m, nil

	case "h", "left":
		// For select widgets, cycle backward
		if m.settingsCursor < len(items) {
			item := items[m.settingsCursor]
			if item.kind == settingsSelect {
				current := m.getConfigString(item.key)
				for idx, opt := range item.options {
					if opt == current {
						prev := (idx - 1 + len(item.options)) % len(item.options)
						m.setConfigString(item.key, item.options[prev])
						if err := saveConfig(m.config); err != nil {
							m.statusMessage = fmt.Sprintf("Failed to save: %v", err)
						}
						break
					}
				}
			}
		}
		return m, nil
	}

	return m, nil
}
